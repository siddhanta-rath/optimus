package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/mattn/go-isatty"
	"github.com/odpf/optimus/config"
	"github.com/odpf/optimus/models"
	"github.com/odpf/optimus/store/local"
	"github.com/odpf/salt/log"
	"github.com/odpf/salt/term"
	"github.com/spf13/afero"
	cli "github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var (
	prologueContents = `optimus %s

Optimus is an easy-to-use, reliable, and performant workflow orchestrator for 
data transformation, data modeling, pipelines, and data quality management.

For passing authentication details, use one of the following environment
variables:
1. OPTIMUS_AUTH_BASIC_TOKEN
2. OPTIMUS_AUTH_BEARER_TOKEN
`
	disableColoredOut = false

	// colored print
	coloredNotice  = fmt.Sprintf
	coloredError   = fmt.Sprintf
	coloredSuccess = fmt.Sprintf

	ErrServerNotReachable = func(host string) error {
		return fmt.Errorf(`Unable to reach optimus server at %s, this can happen due to following reasons:
1. Check if you are connected to internet
2. Is the host correctly configured in optimus config
3. Is OPTIMUS_HOST env incorrectly set
4. Is Optimus server currently unreachable`, host)
	}
)

const (
	GRPCMaxClientSendSize      = 45 << 20 // 45MB
	GRPCMaxClientRecvSize      = 45 << 20 // 45MB
	GRPCMaxRetry          uint = 3

	OptimusDialTimeout = time.Second * 2
)

func programPrologue(ver string) string {
	return fmt.Sprintf(prologueContents, ver)
}

// JobSpecRepository represents a storage interface for Job specifications locally
type JobSpecRepository interface {
	SaveAt(models.JobSpec, string) error
	Save(models.JobSpec) error
	GetByName(string) (models.JobSpec, error)
	GetAll() ([]models.JobSpec, error)
}

// New constructs the 'root' command.It houses all other sub commands
// default output of logging should go to stdout
// interactive output like progress bars should go to stderr
// unless the stdout/err is a tty, colors/progressbar should be disabled
func New(plainLog log.Logger, jsonLog log.Logger, conf config.Provider, pluginRepo models.PluginRepository, dsRepo models.DatastoreRepo) *cli.Command {
	disableColoredOut = !isTerminal(os.Stdout)
	var cmd = &cli.Command{
		Use:  "optimus",
		Long: programPrologue(config.Version),
		PersistentPreRun: func(cmd *cli.Command, args []string) {
			// initialise color if not requested to be disabled
			cs := term.NewColorScheme()
			if !disableColoredOut {
				coloredNotice = func(s string, a ...interface{}) string {
					return cs.Yellowf(s, a...)
				}
				coloredError = func(s string, a ...interface{}) string {
					return cs.Redf(s, a...)
				}
				coloredSuccess = func(s string, a ...interface{}) string {
					return cs.Greenf(s, a...)
				}
			}
		},
		SilenceUsage: true,
	}
	cmd.PersistentFlags().BoolVar(&disableColoredOut, "no-color", disableColoredOut, "disable colored output")

	//init local specs
	var jobSpecRepo JobSpecRepository
	jobSpecFs := afero.NewBasePathFs(afero.NewOsFs(), conf.GetJob().Path)
	if conf.GetJob().Path != "" {
		jobSpecRepo = local.NewJobSpecRepository(
			jobSpecFs,
			local.NewJobSpecAdapter(pluginRepo),
		)
	}
	datastoreSpecsFs := map[string]afero.Fs{}
	for _, dsConfig := range conf.GetDatastore() {
		datastoreSpecsFs[dsConfig.Type] = afero.NewBasePathFs(afero.NewOsFs(), dsConfig.Path)
	}

	cmd.AddCommand(versionCommand(plainLog, conf.GetHost(), pluginRepo))
	cmd.AddCommand(configCommand(plainLog, dsRepo))
	cmd.AddCommand(jobCommand(plainLog, jobSpecFs, pluginRepo, conf))
	cmd.AddCommand(deployCommand(plainLog, conf, jobSpecRepo, pluginRepo, dsRepo, datastoreSpecsFs))
	cmd.AddCommand(resourceCommand(plainLog, datastoreSpecsFs, dsRepo))
	cmd.AddCommand(serveCommand(jsonLog, conf))
	cmd.AddCommand(replayCommand(plainLog, conf))
	cmd.AddCommand(backupCommand(plainLog, dsRepo, conf))
	cmd.AddCommand(adminCommand(plainLog, conf))

	addExtensionCommand(cmd, plainLog)
	return cmd
}

func createConnection(ctx context.Context, host string) (*grpc.ClientConn, error) {
	retryOpts := []grpc_retry.CallOption{
		grpc_retry.WithBackoff(grpc_retry.BackoffExponential(100 * time.Millisecond)),
		grpc_retry.WithMax(GRPCMaxRetry),
	}
	var opts []grpc.DialOption
	opts = append(opts,
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallSendMsgSize(GRPCMaxClientSendSize),
			grpc.MaxCallRecvMsgSize(GRPCMaxClientRecvSize),
		),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			grpc_retry.UnaryClientInterceptor(retryOpts...),
			otelgrpc.UnaryClientInterceptor(),
			grpc_prometheus.UnaryClientInterceptor,
		)),
		grpc.WithStreamInterceptor(grpc_middleware.ChainStreamClient(
			otelgrpc.StreamClientInterceptor(),
			grpc_prometheus.StreamClientInterceptor,
		)),
	)

	// pass rpc credentials
	if token := os.Getenv("OPTIMUS_AUTH_BASIC_TOKEN"); token != "" {
		base64Token := base64.StdEncoding.EncodeToString([]byte(token))
		opts = append(opts, grpc.WithPerRPCCredentials(&BasicAuthentication{
			Token: base64Token,
		}))
	} else if token := os.Getenv("OPTIMUS_AUTH_BEARER_TOKEN"); token != "" {
		opts = append(opts, grpc.WithPerRPCCredentials(&BearerAuthentication{
			Token: token,
		}))
	}
	return grpc.DialContext(ctx, host, opts...)
}

type BearerAuthentication struct {
	Token string
}

func (a *BearerAuthentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", a.Token),
	}, nil
}

func (a *BearerAuthentication) RequireTransportSecurity() bool {
	return false
}

type BasicAuthentication struct {
	Token string
}

func (a *BasicAuthentication) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{
		"Authorization": fmt.Sprintf("Basic %s", a.Token),
	}, nil
}

func (a *BasicAuthentication) RequireTransportSecurity() bool {
	return false
}

func isTerminal(f *os.File) bool {
	return isatty.IsTerminal(f.Fd()) || isatty.IsCygwinTerminal(f.Fd())
}

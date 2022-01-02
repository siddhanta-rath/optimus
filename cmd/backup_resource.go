package cmd

import (
	"context"
	"fmt"

	"github.com/AlecAivazis/survey/v2"
	pb "github.com/odpf/optimus/api/proto/odpf/optimus"
	"github.com/odpf/optimus/config"
	"github.com/odpf/optimus/models"
	"github.com/odpf/salt/log"
	"github.com/pkg/errors"
	cli "github.com/spf13/cobra"
	"google.golang.org/grpc"
)

func backupResourceSubCommand(l log.Logger, datastoreRepo models.DatastoreRepo, conf config.Provider) *cli.Command {
	var (
		backupCmd = &cli.Command{
			Use:   "resource",
			Short: "Backup a resource",
		}
		project          = conf.GetProject().Name
		namespace        = conf.GetNamespace().Name
		dryRun           = false
		ignoreDownstream = false
		allDownstream    = false
		skipConfirm      = false
	)

	backupCmd.Flags().BoolVarP(&dryRun, "dry-run", "d", dryRun, "Only do a trial run with no permanent changes")
	backupCmd.Flags().StringVarP(&project, "project", "p", project, "Project name of optimus managed repository")
	backupCmd.Flags().StringVarP(&namespace, "namespace", "n", namespace, "Namespace of the resource within project")
	backupCmd.Flags().BoolVar(&skipConfirm, "confirm", skipConfirm, "Skip asking for confirmation")
	backupCmd.Flags().BoolVarP(&allDownstream, "all-downstream", "", allDownstream, "Run backup for all downstreams across namespaces")
	backupCmd.Flags().BoolVar(&ignoreDownstream, "ignore-downstream", ignoreDownstream, "Do not take backups for dependent downstream resources")

	backupCmd.RunE = func(cmd *cli.Command, args []string) error {
		availableStorer := []string{}
		for _, s := range datastoreRepo.GetAll() {
			availableStorer = append(availableStorer, s.Name())
		}
		var storerName string
		if err := survey.AskOne(&survey.Select{
			Message: "Select supported datastore?",
			Options: availableStorer,
		}, &storerName); err != nil {
			return err
		}

		var qs = []*survey.Question{
			{
				Name: "name",
				Prompt: &survey.Input{
					Message: "What is the resource name?",
					Help:    "Input urn of the resource",
				},
				Validate: survey.ComposeValidators(validateNoSlash, survey.MinLength(3),
					survey.MaxLength(1024)),
			},
			{
				Name: "description",
				Prompt: &survey.Input{
					Message: "Why is this backup needed?",
					Help:    "Describe intention to help identify the backup",
				},
				Validate: survey.ComposeValidators(survey.MinLength(3)),
			},
		}
		inputs := map[string]interface{}{}
		if err := survey.Ask(qs, &inputs); err != nil {
			return err
		}
		resourceName := inputs["name"].(string)
		description := inputs["description"].(string)

		var allowedDownstreamNamespaces []string
		if !ignoreDownstream {
			if allDownstream {
				allowedDownstreamNamespaces = []string{"*"}
			} else {
				allowedDownstreamNamespaces = []string{namespace}
			}
		}

		backupDryRunRequest := &pb.BackupDryRunRequest{
			ProjectName:                 project,
			Namespace:                   namespace,
			ResourceName:                resourceName,
			DatastoreName:               storerName,
			Description:                 description,
			AllowedDownstreamNamespaces: allowedDownstreamNamespaces,
		}
		if err := runBackupDryRunRequest(l, conf, backupDryRunRequest, !ignoreDownstream); err != nil {
			l.Info(coloredNotice("Unable to run backup dry run"))
			return err
		}
		if dryRun {
			//if only dry run, exit now
			return nil
		}

		if !skipConfirm {
			proceedWithBackup := "Yes"
			if err := survey.AskOne(&survey.Select{
				Message: "Proceed the backup?",
				Options: []string{"Yes", "No"},
				Default: "Yes",
			}, &proceedWithBackup); err != nil {
				return err
			}
			if proceedWithBackup == "No" {
				l.Info("aborting...")
				return nil
			}
		}

		backupRequest := &pb.BackupRequest{
			ProjectName:                 project,
			Namespace:                   namespace,
			ResourceName:                resourceName,
			DatastoreName:               storerName,
			Description:                 description,
			AllowedDownstreamNamespaces: allowedDownstreamNamespaces,
		}

		for _, ds := range conf.GetDatastore() {
			if ds.Type == storerName {
				backupRequest.Config = ds.Backup
			}
		}
		return runBackupRequest(l, conf, backupRequest)
	}
	return backupCmd
}

func runBackupDryRunRequest(l log.Logger, conf config.Provider, backupRequest *pb.BackupDryRunRequest, backupDownstream bool) (err error) {
	dialTimeoutCtx, dialCancel := context.WithTimeout(context.Background(), OptimusDialTimeout)
	defer dialCancel()

	var conn *grpc.ClientConn
	if conn, err = createConnection(dialTimeoutCtx, conf.GetHost()); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			l.Error(ErrServerNotReachable(conf.GetHost()).Error())
		}
		return err
	}
	defer conn.Close()

	requestTimeoutCtx, requestCancel := context.WithTimeout(context.Background(), backupTimeout)
	defer requestCancel()

	runtime := pb.NewRuntimeServiceClient(conn)

	spinner := NewProgressBar()
	spinner.Start("please wait...")
	backupDryRunResponse, err := runtime.BackupDryRun(requestTimeoutCtx, backupRequest)
	spinner.Stop()
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			l.Error(coloredError("Backup dry run took too long, timing out"))
		}
		return errors.Wrapf(err, "request failed to backup job %s", backupRequest.ResourceName)
	}

	printBackupDryRunResponse(l, backupRequest, backupDryRunResponse, backupDownstream)
	return nil
}

func runBackupRequest(l log.Logger, conf config.Provider, backupRequest *pb.BackupRequest) (err error) {
	dialTimeoutCtx, dialCancel := context.WithTimeout(context.Background(), OptimusDialTimeout)
	defer dialCancel()

	var conn *grpc.ClientConn
	if conn, err = createConnection(dialTimeoutCtx, conf.GetHost()); err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			l.Error(ErrServerNotReachable(conf.GetHost()).Error())
		}
		return err
	}
	defer conn.Close()

	requestTimeout, requestCancel := context.WithTimeout(context.Background(), backupTimeout)
	defer requestCancel()

	runtime := pb.NewRuntimeServiceClient(conn)

	spinner := NewProgressBar()
	spinner.Start("please wait...")
	backupResponse, err := runtime.Backup(requestTimeout, backupRequest)
	spinner.Stop()

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			l.Info("Backup took too long, timing out")
		}
		return errors.Wrapf(err, "request failed to backup job %s", backupRequest.ResourceName)
	}

	printBackupResponse(l, backupResponse)
	return nil
}

func printBackupResponse(l log.Logger, backupResponse *pb.BackupResponse) {
	l.Info(coloredSuccess("Resource backup completed successfully:"))
	for counter, result := range backupResponse.Urn {
		l.Info(fmt.Sprintf("%d. %s", counter+1, result))
	}
}

func printBackupDryRunResponse(l log.Logger, backupRequest *pb.BackupDryRunRequest, backupResponse *pb.BackupDryRunResponse,
	backupDownstream bool) {
	if !backupDownstream {
		l.Info(coloredNotice(fmt.Sprintf("Backup list for %s. Downstreams will be ignored.", backupRequest.ResourceName)))
	} else {
		l.Info(coloredNotice(fmt.Sprintf("Backup list for %s. Supported downstreams will be included.", backupRequest.ResourceName)))
	}
	for counter, resource := range backupResponse.ResourceName {
		l.Info(fmt.Sprintf("%d. %s", counter+1, resource))
	}

	if len(backupResponse.IgnoredResources) > 0 {
		l.Info("\nThese resources will be ignored:")
	}
	for counter, ignoredResource := range backupResponse.IgnoredResources {
		l.Info(fmt.Sprintf("%d. %s", counter+1, ignoredResource))
	}
}

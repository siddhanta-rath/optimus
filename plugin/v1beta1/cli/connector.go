package cli

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/go-plugin"
	"github.com/odpf/optimus/plugin/v1beta1/base"

	"github.com/odpf/optimus/models"

	hplugin "github.com/hashicorp/go-plugin"
	pbp "github.com/odpf/optimus/api/proto/odpf/optimus/plugins/v1beta1"
	"google.golang.org/grpc"
)

var _ hplugin.GRPCPlugin = &Connector{}

type Connector struct {
	hplugin.NetRPCUnsupportedPlugin
	hplugin.GRPCPlugin

	impl   models.CommandLineMod
	logger hclog.Logger
}

func (p *Connector) GRPCServer(broker *hplugin.GRPCBroker, s *grpc.Server) error {
	pbp.RegisterCLIModServiceServer(s, &GRPCServer{
		Impl: p.impl,
	})
	return nil
}

func (p *Connector) GRPCClient(ctx context.Context, broker *hplugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{
		client: pbp.NewCLIModServiceClient(c),
		baseClient: &base.GRPCClient{
			Client: pbp.NewBaseServiceClient(c),
			Logger: p.logger,
		},
	}, nil
}

func NewPlugin(impl interface{}, logger hclog.Logger) *Connector {
	return &Connector{
		impl:   impl.(models.CommandLineMod),
		logger: logger,
	}
}

func NewPluginClient(logger hclog.Logger) *Connector {
	return &Connector{
		logger: logger,
	}
}

func Serve(t interface{}, logger hclog.Logger) {
	hplugin.Serve(&hplugin.ServeConfig{
		HandshakeConfig: base.Handshake,
		Plugins: map[string]plugin.Plugin{
			models.PluginTypeBase:      base.NewPlugin(t, logger),
			models.ModTypeCLI.String(): NewPlugin(t, logger),
		},
		GRPCServer: plugin.DefaultGRPCServer,
		Logger:     logger,
	})
}

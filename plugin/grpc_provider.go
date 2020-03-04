package plugin

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"

	proto "github.com/hashicorp/terraform-plugin-sdk/v2/internal/tfplugin5"
)

// GRPCProviderPlugin implements plugin.GRPCPlugin for the go-plugin package.
type gRPCProviderPlugin struct {
	plugin.Plugin
	GRPCProvider func() proto.ProviderServer
}

// this exists only to satisfy the go-plugin.GRPCPlugin interface
// that interface should likely be split
func (p *gRPCProviderPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return nil, nil
}

func (p *gRPCProviderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterProviderServer(s, p.GRPCProvider())
	return nil
}

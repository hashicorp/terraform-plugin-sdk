package plugin

import (
	"context"

	plugin "github.com/hashicorp/go-plugin"
	proto "github.com/hashicorp/terraform-plugin-sdk/internal/tfplugin5"
	"google.golang.org/grpc"
)

// GRPCProviderPlugin implements plugin.GRPCPlugin for the go-plugin package.
type GRPCProviderPlugin struct {
	plugin.Plugin
	GRPCProvider func() proto.ProviderServer
}

// this exists only to satisfy the go-plugin.GRPCPlugin interface
// that interface should likely be split
func (p *GRPCProviderPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return nil, nil
}

func (p *GRPCProviderPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterProviderServer(s, p.GRPCProvider())
	return nil
}

package plugin

import (
	"github.com/hashicorp/go-plugin"
	grpcplugin "github.com/hashicorp/terraform-plugin-sdk/internal/helper/plugin"
	proto "github.com/hashicorp/terraform-plugin-sdk/internal/tfplugin5"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

const (
	// The constants below are the names of the plugins that can be dispensed
	// from the plugin server.
	ProviderPluginName = "provider"
)

// Handshake is the HandshakeConfig used to configure clients and servers.
var Handshake = plugin.HandshakeConfig{
	// The magic cookie values should NEVER be changed.
	MagicCookieKey:   "TF_PLUGIN_MAGIC_COOKIE",
	MagicCookieValue: "d602bf8f470bc67ca7faa0386276bbdd4330efaf76d1a219cb4d6991ca9872b2",
}

type ProviderFunc func() terraform.ResourceProvider
type GRPCProviderFunc func() proto.ProviderServer

// ServeOpts are the configurations to serve a plugin.
type ServeOpts struct {
	ProviderFunc ProviderFunc

	// Wrapped versions of the above plugins will automatically shimmed and
	// added to the GRPC functions when possible.
	GRPCProviderFunc GRPCProviderFunc
}

// Serve serves a plugin. This function never returns and should be the final
// function called in the main function of the plugin.
func Serve(opts *ServeOpts) {
	// since the plugins may not yet be aware of the new protocol, we
	// automatically wrap the plugins in the grpc shims.
	if opts.GRPCProviderFunc == nil && opts.ProviderFunc != nil {
		opts.GRPCProviderFunc = func() proto.ProviderServer {
			return grpcplugin.NewGRPCProviderServerShim(opts.ProviderFunc())
		}
	}
	VersionedPlugins[5][ProviderPluginName] = &GRPCProviderPlugin{
		GRPCProvider: opts.GRPCProviderFunc,
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig:  Handshake,
		VersionedPlugins: VersionedPlugins,
		GRPCServer:       plugin.DefaultGRPCServer,
	})
}

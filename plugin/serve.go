package plugin

import (
	"github.com/hashicorp/go-plugin"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
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
	// providers implemented with the legacy API are
	// automatically wrapped/shimmed to the grpc type.
	if opts.GRPCProviderFunc == nil && opts.ProviderFunc != nil {
		provider := grpcplugin.NewGRPCProviderServerShim(opts.ProviderFunc())
		if provider == nil {
			panic("Plugin could not be converted to grpcplugin.GRPCProviderServer")
		}
		opts.GRPCProviderFunc = func() proto.ProviderServer {
			return provider
		}
	}

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: Handshake,
		VersionedPlugins: map[int]plugin.PluginSet{
			4: legacyPluginMap(opts),
			5: {
				ProviderPluginName: &GRPCProviderPlugin{
					GRPCProvider: opts.GRPCProviderFunc,
				},
			},
		},
		GRPCServer: plugin.DefaultGRPCServer,
	})
}

// pluginMap returns the legacy map[string]plugin.Plugin to use for configuring
// a plugin server or client.
func legacyPluginMap(opts *ServeOpts) map[string]plugin.Plugin {
	var p plugin.Plugin
	if schema.IsProto5() {
		p = &plugin.NetRPCUnsupportedPlugin{}

	} else {
		p = &ResourceProviderPlugin{
			ResourceProvider: opts.ProviderFunc,
		}
	}
	return map[string]plugin.Plugin{
		ProviderPluginName: p,
	}
}

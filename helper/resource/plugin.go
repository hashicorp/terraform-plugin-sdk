package resource

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/diagutils"
	grpcplugin "github.com/hashicorp/terraform-plugin-sdk/v2/internal/helper/plugin"
	proto "github.com/hashicorp/terraform-plugin-sdk/v2/internal/tfplugin5"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tftest "github.com/hashicorp/terraform-plugin-test/v2"
	testing "github.com/mitchellh/go-testing-interface"
)

func runProviderCommand(t testing.T, f func() error, wd *tftest.WorkingDir, factories map[string]func() (*schema.Provider, error)) error {
	t.Helper()

	// Run the provider in the same process as the test runner using the
	// reattach behavior in Terraform. This ensures we get test coverage
	// and enables the use of delve as a debugger.

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// this is needed so Terraform doesn't default to expecting protocol 4;
	// we're skipping the handshake because Terraform didn't launch the
	// plugin.
	os.Setenv("PLUGIN_PROTOCOL_VERSIONS", "5")

	var namespaces []string
	host := "registry.terraform.io"
	if v := os.Getenv("TF_ACC_PROVIDER_NAMESPACE"); v != "" {
		namespaces = append(namespaces, v)
	} else {
		// unfortunately, we need to populate both of them
		// Terraform 0.12.26 and higher uses the legacy mode ("-")
		// Terraform 0.13.0 and higher uses the default mode ("hashicorp")
		// because of the change in how providers are addressed in 0.13
		namespaces = append(namespaces, "-", "hashicorp")
	}
	if v := os.Getenv("TF_ACC_PROVIDER_HOST"); v != "" {
		host = v
	}

	// Spin up gRPC servers for every provider factory, start a
	// WaitGroup to listen for all of the close channels.
	wg := sync.WaitGroup{}
	wg.Add(len(factories))
	reattachInfo := map[string]plugin.ReattachConfig{}
	for providerName, factory := range factories {
		// providerName may be returned as terraform-provider-foo, and we need
		// just foo. So let's fix that.
		providerName = strings.TrimPrefix(providerName, "terraform-provider-")

		provider, err := factory()
		if err != nil {
			return fmt.Errorf("unable to create provider %q from factory: %w", providerName, err)
		}

		// PT: should this actually be called here? does it not get called by TF itself already?
		diags := provider.Configure(ctx, terraform.NewResourceConfigRaw(nil))
		if diags.HasError() {
			return fmt.Errorf("unable to configure provider %q: %w", providerName, diagutils.ErrorDiags(diags))
		}

		opts := &plugin.ServeOpts{
			GRPCProviderFunc: func() proto.ProviderServer {
				return grpcplugin.NewGRPCProviderServer(provider)
			},
			Logger: hclog.New(&hclog.LoggerOptions{
				Name:   "plugintest",
				Level:  hclog.Trace,
				Output: ioutil.Discard,
			}),
		}

		config, closeCh, err := plugin.DebugServe(ctx, opts)
		if err != nil {
			return fmt.Errorf("unable to server provider %q: %w", providerName, err)
		}

		go func(c <-chan struct{}) {
			<-c
			wg.Done()
		}(closeCh)

		// Copy reattach info for any additional namespaces needed for testing
		for _, ns := range namespaces {
			reattachInfo[strings.TrimSuffix(host, "/")+"/"+
				strings.TrimSuffix(ns, "/")+"/"+
				providerName] = config
		}
	}

	// plugin.DebugServe hijacks our log output location, so let's reset it
	logging.SetOutput()

	reattachStr, err := json.Marshal(reattachInfo)
	if err != nil {
		return err
	}
	wd.Setenv("TF_REATTACH_PROVIDERS", string(reattachStr))

	// ok, let's call whatever Terraform command the test was trying to
	// call, now that we know it'll attach back to that server we just
	// started.
	err = f()
	if err != nil {
		log.Printf("[WARN] Got error running Terraform: %s", err)
	}

	// cancel the server so it'll return. Otherwise, this closeCh won't get
	// closed, and we'll hang here.
	cancel()

	// wait for the server to actually shut down; it may take a moment for
	// it to clean up, or whatever.
	// TODO: add a timeout here?
	wg.Wait()

	// once we've run the Terraform command, let's remove the reattach
	// information from the WorkingDir's environment. The WorkingDir will
	// persist until the next call, but the server in the reattach info
	// doesn't exist anymore at this point, so the reattach info is no
	// longer valid. In theory it should be overwritten in the next call,
	// but just to avoid any confusing bug reports, let's just unset the
	// environment variable altogether.
	wd.Unsetenv("TF_REATTACH_PROVIDERS")

	// return any error returned from the orchestration code running
	// Terraform commands
	return err
}

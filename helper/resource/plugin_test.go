package resource

import (
	"context"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/plugintest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func testRunProviderCommand(t *testing.T, ctx context.Context, opts *plugin.ServeOpts, wd *plugintest.WorkingDir, f func() error) {
	t.Helper()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	t.Log("starting provider")

	config, _, err := plugin.DebugServe(ctx, opts)

	if err != nil {
		t.Fatalf("unable to serve provider: %s", err)
	}

	tfexecConfig := tfexec.ReattachConfig{
		Protocol:        config.Protocol,
		ProtocolVersion: config.ProtocolVersion,
		Pid:             config.Pid,
		Test:            config.Test,
		Addr: tfexec.ReattachConfigAddr{
			Network: config.Addr.Network,
			String:  config.Addr.String,
		},
	}

	reattachInfo := map[string]tfexec.ReattachConfig{
		opts.ProviderAddr: tfexecConfig,
	}

	wd.SetReattachInfo(ctx, reattachInfo)

	t.Log("running command")

	if err := f(); err != nil {
		t.Errorf("error running Terraform: %s", err)
	}

	wd.UnsetReattachInfo()
}

func TestPanicAtTheDisco(t *testing.T) {
	t.Setenv("CHECKPOINT_DISABLE", "1")

	currentDir, err := os.Getwd()

	if err != nil {
		t.Fatalf("unable to get working directory: %s", err)
	}

	ctx := context.Background()

	helper := plugintest.AutoInitProviderHelper(ctx, currentDir)
	wd := helper.RequireNewWorkingDir(ctx, t)
	t.Logf("working directory: %s", wd.GetHelper().WorkingDirectory())

	if err := wd.SetConfig(ctx, `resource "test_thing" "test" {}`); err != nil {
		t.Fatalf("error setting configuration: %s", err)
	}

	provider := &schema.Provider{
		ResourcesMap: map[string]*schema.Resource{
			"test_thing": {
				CreateContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
					d.SetId("id")

					return nil
				},
				DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
					return nil
				},
				ReadContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
					return nil
				},
				Schema: map[string]*schema.Schema{
					"id": {
						Computed: true,
						Type:     schema.TypeString,
					},
				},
			},
		},
	}
	providerAddress := "registry.terraform.io/hashicorp/test"

	grpcProviderServer := schema.NewGRPCProviderServer(provider)
	// Prevent goroutine leak
	//defer grpcProviderServer.StopProvider(ctx, nil) //nolint:errcheck // does not return errors

	opts := &plugin.ServeOpts{
		GRPCProviderFunc: func() tfprotov5.ProviderServer {
			return grpcProviderServer
		},
		Logger: hclog.New(&hclog.LoggerOptions{
			Name:   "plugintest",
			Level:  hclog.Trace,
			Output: io.Discard,
		}),
		NoLogOutputOverride: true,
		UseTFLogSink:        t,
		ProviderAddr:        providerAddress,
	}

	t.Log("Calling init")
	testRunProviderCommand(t, ctx, opts, wd, func() error {
		return wd.Init(ctx)
	})

	t.Log("Calling plan")
	testRunProviderCommand(t, ctx, opts, wd, func() error {
		return wd.CreatePlan(ctx)
	})

	t.Log("Calling apply")
	testRunProviderCommand(t, ctx, opts, wd, func() error {
		return wd.Apply(ctx)
	})

	t.Log("Calling show")
	testRunProviderCommand(t, ctx, opts, wd, func() error {
		_, err := wd.State(ctx)
		return err
	})

	t.Log("Calling refresh")
	testRunProviderCommand(t, ctx, opts, wd, func() error {
		return wd.Refresh(ctx)
	})
}

func TestProtoV5ProviderFactoriesMerge(t *testing.T) {
	t.Parallel()

	testProviderFactory1 := func() (tfprotov5.ProviderServer, error) {
		return nil, nil
	}
	testProviderFactory2 := func() (tfprotov5.ProviderServer, error) {
		return nil, nil
	}

	// Function pointers do not play well with go-cmp, so convert these
	// into their stringified address for comparison.
	transformer := cmp.Transformer(
		"protov5ProviderFactory",
		func(pf protov5ProviderFactory) string {
			return fmt.Sprintf("%v", pf)
		},
	)

	testCases := map[string]struct {
		pf       protov5ProviderFactories
		others   []protov5ProviderFactories
		expected protov5ProviderFactories
	}{
		"no-overlap": {
			pf: protov5ProviderFactories{
				"test1": testProviderFactory1,
			},
			others: []protov5ProviderFactories{
				{
					"test2": testProviderFactory1,
				},
				{
					"test3": testProviderFactory1,
				},
			},
			expected: protov5ProviderFactories{
				"test1": testProviderFactory1,
				"test2": testProviderFactory1,
				"test3": testProviderFactory1,
			},
		},
		"overlap": {
			pf: protov5ProviderFactories{
				"test": testProviderFactory1,
			},
			others: []protov5ProviderFactories{
				{
					"test": testProviderFactory1,
				},
				{
					"test": testProviderFactory2,
				},
			},
			expected: protov5ProviderFactories{
				"test": testProviderFactory2,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.pf.merge(testCase.others...)

			if diff := cmp.Diff(got, testCase.expected, transformer); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestProtoV6ProviderFactoriesMerge(t *testing.T) {
	t.Parallel()

	testProviderFactory1 := func() (tfprotov6.ProviderServer, error) {
		return nil, nil
	}
	testProviderFactory2 := func() (tfprotov6.ProviderServer, error) {
		return nil, nil
	}

	// Function pointers do not play well with go-cmp, so convert these
	// into their stringified address for comparison.
	transformer := cmp.Transformer(
		"protov6ProviderFactory",
		func(pf protov6ProviderFactory) string {
			return fmt.Sprintf("%v", pf)
		},
	)

	testCases := map[string]struct {
		pf       protov6ProviderFactories
		others   []protov6ProviderFactories
		expected protov6ProviderFactories
	}{
		"no-overlap": {
			pf: protov6ProviderFactories{
				"test1": testProviderFactory1,
			},
			others: []protov6ProviderFactories{
				{
					"test2": testProviderFactory1,
				},
				{
					"test3": testProviderFactory1,
				},
			},
			expected: protov6ProviderFactories{
				"test1": testProviderFactory1,
				"test2": testProviderFactory1,
				"test3": testProviderFactory1,
			},
		},
		"overlap": {
			pf: protov6ProviderFactories{
				"test": testProviderFactory1,
			},
			others: []protov6ProviderFactories{
				{
					"test": testProviderFactory1,
				},
				{
					"test": testProviderFactory2,
				},
			},
			expected: protov6ProviderFactories{
				"test": testProviderFactory2,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.pf.merge(testCase.others...)

			if diff := cmp.Diff(got, testCase.expected, transformer); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestSdkProviderFactoriesMerge(t *testing.T) {
	t.Parallel()

	testProviderFactory1 := func() (*schema.Provider, error) {
		return nil, nil
	}
	testProviderFactory2 := func() (*schema.Provider, error) {
		return nil, nil
	}

	// Function pointers do not play well with go-cmp, so convert these
	// into their stringified address for comparison.
	transformer := cmp.Transformer(
		"sdkProviderFactory",
		func(pf sdkProviderFactory) string {
			return fmt.Sprintf("%v", pf)
		},
	)

	testCases := map[string]struct {
		pf       sdkProviderFactories
		others   []sdkProviderFactories
		expected sdkProviderFactories
	}{
		"no-overlap": {
			pf: sdkProviderFactories{
				"test1": testProviderFactory1,
			},
			others: []sdkProviderFactories{
				{
					"test2": testProviderFactory1,
				},
				{
					"test3": testProviderFactory1,
				},
			},
			expected: sdkProviderFactories{
				"test1": testProviderFactory1,
				"test2": testProviderFactory1,
				"test3": testProviderFactory1,
			},
		},
		"overlap": {
			pf: sdkProviderFactories{
				"test": testProviderFactory1,
			},
			others: []sdkProviderFactories{
				{
					"test": testProviderFactory1,
				},
				{
					"test": testProviderFactory2,
				},
			},
			expected: sdkProviderFactories{
				"test": testProviderFactory2,
			},
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := testCase.pf.merge(testCase.others...)

			if diff := cmp.Diff(got, testCase.expected, transformer); diff != "" {
				t.Errorf("unexpected difference: %s", diff)
			}
		})
	}
}

func TestRunProviderCommand(t *testing.T) {
	currentDir, err := os.Getwd()

	if err != nil {
		t.Fatalf("unable to get working directory: %s", err)
	}

	ctx := context.Background()
	funcCalled := false
	helper := plugintest.AutoInitProviderHelper(ctx, currentDir)

	err = runProviderCommand(
		ctx,
		t,
		func() error {
			funcCalled = true
			return nil
		},
		helper.RequireNewWorkingDir(ctx, t),
		&providerFactories{
			legacy: map[string]func() (*schema.Provider, error){
				"examplecloud": func() (*schema.Provider, error) { //nolint:unparam // required signature
					return &schema.Provider{
						ResourcesMap: map[string]*schema.Resource{
							"examplecloud_thing": {
								CreateContext: func(ctx context.Context, d *schema.ResourceData, i interface{}) diag.Diagnostics {
									d.SetId("id")

									return nil
								},
								DeleteContext: func(_ context.Context, _ *schema.ResourceData, _ interface{}) diag.Diagnostics {
									return nil
								},
								ReadContext: func(_ context.Context, d *schema.ResourceData, _ interface{}) diag.Diagnostics {
									return nil
								},
								Schema: map[string]*schema.Schema{
									"id": {
										Computed: true,
										Type:     schema.TypeString,
									},
								},
							},
						},
					}, nil
				},
			},
		},
	)

	if err != nil {
		t.Fatal(err)
	}

	if !funcCalled {
		t.Error("expected func to be called")
	}
}

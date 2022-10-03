package resource

import (
	"context"

	"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/logging"
	"github.com/hashicorp/terraform-plugin-sdk/v2/internal/plugintest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testStepNewRefreshState(ctx context.Context, t testing.T, wd *plugintest.WorkingDir, step TestStep, cfg string, providers *providerFactories) error {
	t.Helper()

	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true

	var err error
	err = runProviderCommand(ctx, t, func() error {
		_, err = getState(ctx, t, wd)
		if err != nil {
			return err
		}
		return nil
	}, wd, providers)
	if err != nil {
		t.Fatalf("Error getting state: %s", err)
	}

	if step.Config == "" {
		logging.HelperResourceTrace(ctx, "Using prior TestStep Config for refresh")

		step.Config = cfg
		if step.Config == "" {
			t.Fatal("Cannot refresh state with no specified config")
		}
	}

	err = wd.SetConfig(ctx, step.Config)
	if err != nil {
		t.Fatalf("Error setting test config: %s", err)
	}

	logging.HelperResourceDebug(ctx, "Running Terraform CLI init and refresh")

	err = runProviderCommand(ctx, t, func() error {
		return wd.Init(ctx)
	}, wd, providers)
	if err != nil {
		t.Fatalf("Error running init: %s", err)
	}

	err = runProviderCommand(ctx, t, func() error {
		return wd.Refresh(ctx)
	}, wd, providers)
	if err != nil {
		return err
	}

	var refreshState *terraform.State
	err = runProviderCommand(ctx, t, func() error {
		refreshState, err = getState(ctx, t, wd)
		if err != nil {
			return err
		}
		return nil
	}, wd, providers)
	if err != nil {
		t.Fatalf("Error getting state: %s", err)
	}

	// Go through the refreshed state and verify
	if step.Check != nil {
		logging.HelperResourceDebug(ctx, "Calling TestStep Check for RefreshState")

		if err := step.Check(refreshState); err != nil {
			t.Fatal(err)
		}

		logging.HelperResourceDebug(ctx, "Called TestStep Check for RefreshState")
	}

	return nil
}

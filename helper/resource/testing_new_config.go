package resource

import (
	"encoding/json"
	"log"

	"github.com/davecgh/go-spew/spew"
	tfjson "github.com/hashicorp/terraform-json"
	tftest "github.com/hashicorp/terraform-plugin-test/v2"
	testing "github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func testStepNewConfig(t testing.T, c TestCase, wd *tftest.WorkingDir, step TestStep, stepNo int) error {
	t.Helper()

	log.Printf("[DEBUG] starting test step %d", stepNo)
	defer log.Printf("[DEBUG] finished test step %d", stepNo)

	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true

	var idRefreshCheck *terraform.ResourceState
	idRefresh := c.IDRefreshName != ""

	if !step.Destroy {
		log.Printf("[DEBUG] getting state for non-destroy step #%d", stepNo)
		var state *terraform.State
		err := runProviderCommand(t, func() error {
			state = getState(t, wd)
			return nil
		}, wd, c.ProviderFactories)
		if err != nil {
			log.Printf("[DEBUG] error getting state for non-destroy step #%d", stepNo)
			return err
		}
		log.Printf("[DEBUG] running taint for non-destroy step #%d", stepNo)
		if err := testStepTaint(state, step); err != nil {
			t.Fatalf("Error when tainting resources: %s", err)
		}
		log.Printf("[DEBUG] ran taint for non-destroy step #%d", stepNo)
	}

	log.Printf("[DEBUG] setting config for step #%d", stepNo)
	wd.RequireSetConfig(t, step.Config)
	log.Printf("[DEBUG] set config for step #%d", stepNo)

	if !step.PlanOnly {
		log.Printf("[DEBUG] applying step #%d", stepNo)
		err := runProviderCommand(t, func() error {
			return wd.Apply()
		}, wd, c.ProviderFactories)
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] applied step #%d", stepNo)

		log.Printf("[DEBUG] getting state for step #%d", stepNo)
		var state *terraform.State
		err = runProviderCommand(t, func() error {
			state = getState(t, wd)
			return nil
		}, wd, c.ProviderFactories)
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] got state for step #%d", stepNo)
		if step.Check != nil {
			state.IsBinaryDrivenTest = true
			if err := step.Check(state); err != nil {
				t.Fatal(err)
			}
		}
	}

	// Test for perpetual diffs by performing a plan, a refresh, and another plan

	// do a plan
	log.Printf("[DEBUG] running plan to test for perpetual diffs in step #%d", stepNo)
	err := runProviderCommand(t, func() error {
		wd.RequireCreatePlan(t)
		return nil
	}, wd, c.ProviderFactories)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] ran plan to test for perpetual diffs in step #%d", stepNo)

	var plan *tfjson.Plan
	err = runProviderCommand(t, func() error {
		plan = wd.RequireSavedPlan(t)
		return nil
	}, wd, c.ProviderFactories)
	if err != nil {
		return err
	}

	if !planIsEmpty(plan) {
		if step.ExpectNonEmptyPlan {
			t.Log("[INFO] Got non-empty plan, as expected")
		} else {
			planJSON, err := json.Marshal(plan)
			if err != nil {
				t.Fatalf("After applying test step #%d, the plan was not empty, and we couldn't marshal it to JSON (err: %v). %s", stepNo, err, spewConf.Sdump(plan))
			} else {
				t.Fatalf("After applying test step #%d, the plan was not empty. %s", stepNo, string(planJSON))
			}
		}
	}

	// do a refresh
	if !c.PreventPostDestroyRefresh {
		log.Printf("[DEBUG] running refresh to test for perpetual diffs in step #%d", stepNo)
		err := runProviderCommand(t, func() error {
			wd.RequireRefresh(t)
			return nil
		}, wd, c.ProviderFactories)
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] ran refresh to test for perpetual diffs in step #%d", stepNo)
	}

	log.Printf("[DEBUG] running second plan to test for perpetual diffs in step #%d", stepNo)
	// do another plan
	err = runProviderCommand(t, func() error {
		wd.RequireCreatePlan(t)
		return nil
	}, wd, c.ProviderFactories)
	if err != nil {
		return err
	}
	log.Printf("[DEBUG] ran second plan to test for perpetual diffs in step #%d", stepNo)

	err = runProviderCommand(t, func() error {
		plan = wd.RequireSavedPlan(t)
		return nil
	}, wd, c.ProviderFactories)
	if err != nil {
		return err
	}

	// check if plan is empty
	if !planIsEmpty(plan) {
		if step.ExpectNonEmptyPlan {
			t.Log("[INFO] Got non-empty plan, as expected")
		} else {
			planJSON, err := json.Marshal(plan)
			if err != nil {
				t.Fatalf("After applying test step #%d and performing a `terraform refresh`, the plan was not empty, and we couldn't marshal it to JSON (err: %v). %s", stepNo, err, spewConf.Sdump(plan))
			} else {
				t.Fatalf("After apply test step #%d and performing a `terraform refresh`, the plan was not empty. %s", stepNo, string(planJSON))
			}
		}
	}

	// ID-ONLY REFRESH
	// If we've never checked an id-only refresh and our state isn't
	// empty, find the first resource and test it.
	var state *terraform.State
	err = runProviderCommand(t, func() error {
		state = getState(t, wd)
		return nil
	}, wd, c.ProviderFactories)
	if err != nil {
		return err
	}
	if idRefresh && idRefreshCheck == nil && !state.Empty() {
		// Find the first non-nil resource in the state
		for _, m := range state.Modules {
			if len(m.Resources) > 0 {
				if v, ok := m.Resources[c.IDRefreshName]; ok {
					idRefreshCheck = v
				}

				break
			}
		}

		// If we have an instance to check for refreshes, do it
		// immediately. We do it in the middle of another test
		// because it shouldn't affect the overall state (refresh
		// is read-only semantically) and we want to fail early if
		// this fails. If refresh isn't read-only, then this will have
		// caught a different bug.
		if idRefreshCheck != nil {
			if err := testIDRefresh(c, t, wd, step, idRefreshCheck); err != nil {
				t.Fatalf(
					"[ERROR] Test: ID-only test failed: %s", err)
			}
		}
	}

	return nil
}

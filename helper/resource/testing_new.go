package resource

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"
	"testing"

	tftest "github.com/apparentlymart/terraform-plugin-test"
	"github.com/davecgh/go-spew/spew"
	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-sdk/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/internal/addrs"
	"github.com/hashicorp/terraform-plugin-sdk/internal/tfdiags"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func shimAttributeValues(flatmap map[string]string, currentKey string, value interface{}) {
	switch v := value.(type) {
	case nil:
		// omit the entry altogether
	case bool:
		flatmap[currentKey] = strconv.FormatBool(v)
	case float64:
		flatmap[currentKey] = strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		flatmap[currentKey] = v
	case map[string]interface{}:
		if currentKey != "" {
			currentKey += "."
		}
		for key, val := range v {
			shimAttributeValues(flatmap, fmt.Sprintf("%s%s", currentKey, key), val)
		}
		flatmap[currentKey+"%"] = strconv.Itoa(len(v))
	case []interface{}:
		if currentKey != "" {
			currentKey += "."
		}
		for i, val := range v {
			shimAttributeValues(flatmap, fmt.Sprintf("%s%d", currentKey, i), val)
		}
		flatmap[currentKey+"#"] = strconv.Itoa(len(v))
	default:
		panic("Unknown json type")
	}
}

func shimStateModule(state *terraform.State, newModule *tfjson.StateModule) error {
	var path addrs.ModuleInstance
	var diags tfdiags.Diagnostics
	if newModule.Address == "" {
		path = addrs.RootModuleInstance
	} else {
		path, diags = addrs.ParseModuleInstanceStr(newModule.Address)
		if diags.HasErrors() {
			return diags.Err()
		}
	}

	mod := state.AddModule(path)
	for _, res := range newModule.Resources {
		resState := &terraform.ResourceState{
			Provider: res.ProviderName,
			Type:     res.Type,
		}

		flatmap := make(map[string]string)
		shimAttributeValues(flatmap, "", res.AttributeValues)

		if _, exists := flatmap["id"]; !exists {
			return errors.New("attributes had no id")
		}

		resState.Primary = &terraform.InstanceState{
			Tainted:    res.Tainted,
			ID:         flatmap["id"],
			Attributes: flatmap,
			Meta: map[string]interface{}{
				"schema_version": res.SchemaVersion,
			},
		}

		resState.Dependencies = res.DependsOn

		idx := ""
		switch v := res.Index.(type) {
		case int:
			idx = fmt.Sprintf(".%d", v)
		case string:
			idx = "." + v
		}

		mod.Resources[res.Address+idx] = resState
	}

	for _, child := range newModule.ChildModules {
		if err := shimStateModule(state, child); err != nil {
			return err
		}
	}
	return nil
}

func shimTFJson(jsonState *tfjson.State) (*terraform.State, error) {
	state := terraform.NewState()
	state.TFVersion = jsonState.TerraformVersion
	if jsonState.Values == nil {
		// the state is empty
		return state, nil
	}

	if err := shimStateModule(state, jsonState.Values.RootModule); err != nil {
		return nil, err
	}

	// shimming of lists and maps might be incorrect
	for key, output := range jsonState.Values.Outputs {
		outputType := ""
		switch output.Value.(type) {
		case string:
			outputType = "string"
		case []interface{}:
			outputType = "list"
		case map[string]interface{}:
			outputType = "map"
		default:
			return nil, errors.New("output was not expected type")
		}

		state.RootModule().Outputs[key] = &terraform.OutputState{
			Type:      outputType,
			Value:     output.Value,
			Sensitive: output.Sensitive,
		}
	}

	return state, nil
}

func getState(t *testing.T, wd *tftest.WorkingDir) *terraform.State {
	jsonState := wd.RequireState(t)
	state, err := shimTFJson(jsonState)
	if err != nil {
		t.Fatal(err)
	}
	return state
}

func RunLegacyTest(t *testing.T, c TestCase, providers map[string]terraform.ResourceProvider) {
	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true
	wd := acctest.TestHelper.RequireNewWorkingDir(t)

	defer func() {
		// destroy step
		// TODO probably better to combine this with TestStep.Destroy implementation as in old framework

		wd.RequireDestroy(t)

		if c.CheckDestroy != nil {
			statePostDestroy := getState(t, wd)

			if err := c.CheckDestroy(statePostDestroy); err != nil {
				t.Fatal(err)
			}
		}
		wd.Close()
	}()

	providerCfg := testProviderConfig(c)

	wd.RequireSetConfig(t, providerCfg)
	wd.RequireInit(t)

	// use this to track last step succesfully applied
	// acts as default for import tests
	var appliedCfg string

	for i, step := range c.Steps {

		if step.PreConfig != nil {
			step.PreConfig()
		}

		// TODOs
		if step.Destroy {
			t.Fatal("TODO: TestStep.Destroy")
		}
		if step.PreventDiskCleanup {
			t.Fatal("TODO: TestStep.PreventDiskCleanup")
		}

		if step.SkipFunc != nil {
			skip, err := step.SkipFunc()
			if err != nil {
				t.Fatal(err)
			}
			if skip {
				log.Printf("[WARN] Skipping step %d", i)
				continue
			}
		}

		if step.ImportState {
			step.providers = providers
			err := testStepNewImportState(t, c, wd, step, appliedCfg)
			if err != nil {
				t.Fatal(err)
			}
			continue
		}

		if step.Config != "" {
			err := testStepNewConfig(t, c, wd, step)
			if step.ExpectError != nil {
				if err == nil {
					t.Fatal("Expected an error but got none")
				}
				if !step.ExpectError.MatchString(err.Error()) {
					t.Fatalf("Expected an error with pattern, no match on: %s", err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
			}
			appliedCfg = step.Config
			continue
		}

		t.Fatal("Unsupported test mode")
	}
	// If we never checked an id-only refresh, it is a failure.
	// TODO KEM: why is this here? does this ever happen?
	// if idRefresh {
	// 	if len(c.Steps) > 0 && idRefreshCheck == nil {
	// 		t.Error("ID-only refresh check never ran.")
	// 	}
	// }
}

func planIsEmpty(plan *tfjson.Plan) bool {
	for _, rc := range plan.ResourceChanges {
		for _, a := range rc.Change.Actions {
			if a != tfjson.ActionNoop {
				return false
			}
		}
	}
	return true
}

func testStepNewImportState(t *testing.T, c TestCase, wd *tftest.WorkingDir, step TestStep, cfg string) error {
	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true

	if step.ResourceName == "" {
		t.Fatal("ResourceName is required for an import state test")
	}

	// get state from check sequence
	state := getState(t, wd)

	// Determine the ID to import
	var importId string
	switch {
	case step.ImportStateIdFunc != nil:
		var err error
		importId, err = step.ImportStateIdFunc(state)
		if err != nil {
			t.Fatal(err)
		}
	case step.ImportStateId != "":
		importId = step.ImportStateId
	default:
		resource, err := testResource(step, state)
		if err != nil {
			t.Fatal(err)
		}
		importId = resource.Primary.ID
	}
	importId = step.ImportStateIdPrefix + importId

	// create working directory for import tests
	if step.Config == "" {
		// I can't understand how the previous framework
		// managed to set this to just an empty provider block cfg
		// it must have somehow piggy backed the last non import step config??

		/*
			if step.ImportState {
				if step.Config == "" {
					step.Config = testProviderConfig(c)
				}

				// Can optionally set step.Config in addition to
				// step.ImportState, to provide config for the import.
				state, err = testStepImportState(opts, state, step)
			}
		*/

		// this is what I think should be done
		step.Config = cfg
		if step.Config == "" {
			t.Fatal("Cannot import state with no specified config")
		}
	}
	importWd := acctest.TestHelper.RequireNewWorkingDir(t)
	defer importWd.Close()
	importWd.RequireSetConfig(t, step.Config)
	importWd.RequireInit(t)
	importWd.RequireImport(t, step.ResourceName, importId)
	importState := getState(t, wd)

	// Go through the imported state and verify
	if step.ImportStateCheck != nil {
		var states []*terraform.InstanceState
		for _, r := range importState.RootModule().Resources {
			if r.Primary != nil {
				is := r.Primary.DeepCopy()
				is.Ephemeral.Type = r.Type // otherwise the check function cannot see the type
				states = append(states, is)
			}
		}
		if err := step.ImportStateCheck(states); err != nil {
			t.Fatal(err)
		}
	}

	// TODO: this was copy pasted from the old framework
	// perhaps it can be cleaned up
	// Verify that all the states match
	if step.ImportStateVerify {
		// attach providers for ImportStateVerify
		// step.providers = providers
		t.Logf("step.providers: %+v", step.providers)

		new := importState.RootModule().Resources
		old := state.RootModule().Resources

		for _, r := range new {
			// Find the existing resource
			var oldR *terraform.ResourceState
			for _, r2 := range old {
				if r2.Primary != nil && r2.Primary.ID == r.Primary.ID && r2.Type == r.Type {
					oldR = r2
					break
				}
			}
			if oldR == nil {
				t.Fatalf(
					"Failed state verification, resource with ID %s not found",
					r.Primary.ID)
			}

			// We'll try our best to find the schema for this resource type
			// so we can ignore Removed fields during validation. If we fail
			// to find the schema then we won't ignore them and so the test
			// will need to rely on explicit ImportStateVerifyIgnore, though
			// this shouldn't happen in any reasonable case.
			// KEM CHANGE FROM OLD FRAMEWORK: Fail test if this happens.
			var rsrcSchema *schema.Resource
			providerAddr, diags := addrs.ParseAbsProviderConfigStr("provider." + r.Provider + "." + r.Type)
			if diags.HasErrors() {
				t.Fatalf("Failed to find schema for resource with ID %s", r.Primary)
			}

			providerType := providerAddr.ProviderConfig.Type
			if provider, ok := step.providers[providerType]; ok {
				if provider, ok := provider.(*schema.Provider); ok {
					rsrcSchema = provider.ResourcesMap[r.Type]
				}
			}

			// don't add empty flatmapped containers, so we can more easily
			// compare the attributes
			skipEmpty := func(k, v string) bool {
				if strings.HasSuffix(k, ".#") || strings.HasSuffix(k, ".%") {
					if v == "0" {
						return true
					}
				}
				return false
			}

			// Compare their attributes
			actual := make(map[string]string)
			for k, v := range r.Primary.Attributes {
				if skipEmpty(k, v) {
					continue
				}
				actual[k] = v
			}

			expected := make(map[string]string)
			for k, v := range oldR.Primary.Attributes {
				if skipEmpty(k, v) {
					continue
				}
				expected[k] = v
			}

			// Remove fields we're ignoring
			for _, v := range step.ImportStateVerifyIgnore {
				for k := range actual {
					if strings.HasPrefix(k, v) {
						delete(actual, k)
					}
				}
				for k := range expected {
					if strings.HasPrefix(k, v) {
						delete(expected, k)
					}
				}
			}

			// Also remove any attributes that are marked as "Removed" in the
			// schema, if we have a schema to check that against.
			if rsrcSchema != nil {
				for k := range actual {
					for _, schema := range rsrcSchema.SchemasForFlatmapPath(k) {
						if schema.Removed != "" {
							delete(actual, k)
							break
						}
					}
				}
				for k := range expected {
					for _, schema := range rsrcSchema.SchemasForFlatmapPath(k) {
						if schema.Removed != "" {
							delete(expected, k)
							break
						}
					}
				}
			}

			if !reflect.DeepEqual(actual, expected) {
				// Determine only the different attributes
				for k, v := range expected {
					if av, ok := actual[k]; ok && v == av {
						delete(expected, k)
						delete(actual, k)
					}
				}

				t.Fatalf(
					"ImportStateVerify attributes not equivalent. Difference is shown below. Top is actual, bottom is expected."+
						"\n\n%s\n\n%s",
					spewConf.Sdump(actual), spewConf.Sdump(expected))
			}
		}
	}

	return nil
}

func testStepNewConfig(t *testing.T, c TestCase, wd *tftest.WorkingDir, step TestStep) error {
	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true

	var idRefreshCheck *terraform.ResourceState
	idRefresh := c.IDRefreshName != ""

	if !step.Destroy {
		state := getState(t, wd)
		if err := testStepTaint(state, step); err != nil {
			t.Fatalf("Error when tainting resources: %s", err)
		}
	}

	wd.RequireSetConfig(t, step.Config)

	if !step.PlanOnly {
		err := wd.Apply()
		if err != nil {
			return err
		}

		state := getState(t, wd)
		if step.Check != nil {
			if err := step.Check(state); err != nil {
				t.Fatal(err)
			}
		}
	}

	// ID-ONLY REFRESH
	// If we've never checked an id-only refresh and our state isn't
	// empty, find the first resource and test it.
	state := getState(t, wd)
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
			t.Logf(
				"[WARN] Test: Running ID-only refresh check on %s",
				idRefreshCheck.Primary.ID)
			if err := testIDRefresh(c, t, wd, step, idRefreshCheck); err != nil {
				t.Fatalf(
					"[ERROR] Test: ID-only test failed: %s", err)
			}
		}
	}

	// DESTROY / CLEANUP

	// do a plan
	wd.RequireCreatePlan(t)
	plan := wd.RequireSavedPlan(t)

	if !planIsEmpty(plan) {
		if step.ExpectNonEmptyPlan {
			t.Log("[INFO] Got non-empty plan, as expected")
		} else {

			t.Fatalf("After applying this step, the plan was not empty. %s", spewConf.Sdump(plan)) // TODO error message
		}
	}

	// do a refresh
	if !c.PreventPostDestroyRefresh {
		wd.RequireRefresh(t)
	}

	// TODO deal with the data resources instantiated during refresh

	// do another plan
	wd.RequireCreatePlan(t)
	plan = wd.RequireSavedPlan(t)

	// check if plan is empty
	if !planIsEmpty(plan) {
		if step.ExpectNonEmptyPlan {
			t.Log("[INFO] Got non-empty plan, as expected")
		} else {

			t.Fatalf("After applying this step, the plan was not empty. %s", spewConf.Sdump(plan)) // TODO error message
		}
	}

	return nil
}

func testIDRefresh(c TestCase, t *testing.T, wd *tftest.WorkingDir, step TestStep, r *terraform.ResourceState) error {
	spewConf := spew.NewDefaultConfig()
	spewConf.SortKeys = true

	// Build the state. The state is just the resource with an ID. There
	// are no attributes. We only set what is needed to perform a refresh.
	state := terraform.NewState()
	state.RootModule().Resources = make(map[string]*terraform.ResourceState)
	state.RootModule().Resources[c.IDRefreshName] = &terraform.ResourceState{}

	// Temporarily set the config to a minimal provider config for the refresh
	// test. After the refresh we can reset it.
	cfg := testProviderConfig(c)
	wd.RequireSetConfig(t, cfg)
	defer wd.RequireSetConfig(t, step.Config)

	// Refresh!
	wd.RequireRefresh(t)
	state = getState(t, wd)

	// Verify attribute equivalence.
	actualR := state.RootModule().Resources[c.IDRefreshName]
	if actualR == nil {
		return fmt.Errorf("Resource gone!")
	}
	if actualR.Primary == nil {
		return fmt.Errorf("Resource has no primary instance")
	}
	actual := actualR.Primary.Attributes
	expected := r.Primary.Attributes
	// Remove fields we're ignoring
	for _, v := range c.IDRefreshIgnore {
		for k, _ := range actual {
			if strings.HasPrefix(k, v) {
				delete(actual, k)
			}
		}
		for k, _ := range expected {
			if strings.HasPrefix(k, v) {
				delete(expected, k)
			}
		}
	}

	if !reflect.DeepEqual(actual, expected) {
		// Determine only the different attributes
		for k, v := range expected {
			if av, ok := actual[k]; ok && v == av {
				delete(expected, k)
				delete(actual, k)
			}
		}

		spewConf := spew.NewDefaultConfig()
		spewConf.SortKeys = true
		return fmt.Errorf(
			"Attributes not equivalent. Difference is shown below. Top is actual, bottom is expected."+
				"\n\n%s\n\n%s",
			spewConf.Sdump(actual), spewConf.Sdump(expected))
	}

	return nil
}

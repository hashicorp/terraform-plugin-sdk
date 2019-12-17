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

package resource

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"testing"

	tfjson "github.com/hashicorp/terraform-json"
	"github.com/hashicorp/terraform-plugin-sdk/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/internal/addrs"
	"github.com/hashicorp/terraform-plugin-sdk/internal/tfdiags"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func shimAttributeValues(flatmap map[string]string, currentKey string, value interface{}) {
	switch v := value.(type) {
	case nil:
		// this might not be how terraform 0.11 behaved?
		// I wonder if we should omit the entry altogether?
		flatmap[currentKey] = ""
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

func RunLegacyTest(t *testing.T, c TestCase) {
	wd := acctest.TestHelper.RequireNewWorkingDir(t)
	defer wd.Close()

	providerCfg := testProviderConfig(c)

	wd.RequireSetConfig(t, providerCfg)
	wd.RequireInit(t)

	defer func() {
		// destroy if state exists
		if wd.RequireState(t).Values != nil {
			wd.RequireDestroy(t)
		}
	}()

	for i, step := range c.Steps {
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

		// use this to track last step succesfully applied
		// acts as default for import tests
		var appliedCfg string

		if step.ImportState {
			if step.ResourceName == "" {
				t.Fatal("ResourceName is required for an import state test")
			}

			// get state from check sequence
			jsonState := wd.RequireState(t)
			state, err := shimTFJson(jsonState)
			if err != nil {
				t.Fatal(err)
			}

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

			if step.PreConfig != nil {
				step.PreConfig()
			}
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
				step.Config = appliedCfg
				if step.Config == "" {
					t.Fatal("Cannot import state with no specified config")
				}
			}
			importWd := acctest.TestHelper.RequireNewWorkingDir(t)
			importWd.RequireSetConfig(t, step.Config)
			importWd.RequireInit(t)
			importWd.RequireImport(t, step.ResourceName, importId)
			importStateJson := importWd.RequireState(t)
			importState, err := shimTFJson(importStateJson)
			if err != nil {
				t.Fatal(err)
			}

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

			if step.ImportStateVerify {
				t.Fatal("TODO: ImportStateVerify")
			}

			importWd.Close()
			continue
		}

		if step.Config != "" {
			if step.PreConfig != nil {
				step.PreConfig()
			}
			wd.RequireSetConfig(t, step.Config)
			err := wd.Apply()

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
				jsonState := wd.RequireState(t)
				state, err := shimTFJson(jsonState)
				if err != nil {
					t.Fatal(err)
				}
				if err := step.Check(state); err != nil {
					t.Fatal(err)
				}
				appliedCfg = step.Config
			}
			continue
		}

		t.Fatal("Unsupported test mode")
	}
}

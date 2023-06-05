// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package resource

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	testinginterface "github.com/mitchellh/go-testing-interface"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	if err := os.Setenv(EnvTfAcc, "1"); err != nil {
		panic(err)
	}
}

// testExpectTFatal provides a wrapper for logic which should call
// (*testing.T).Fatal() or (*testing.T).Fatalf().
//
// Since we do not want the wrapping test to fail when an expected test error
// occurs, it is required that the testLogic passed in uses
// github.com/mitchellh/go-testing-interface.RuntimeT instead of the real
// *testing.T.
//
// If Fatal() or Fatalf() is not called in the logic, the real (*testing.T).Fatal() will
// be called to fail the test.
func testExpectTFatal(t *testing.T, testLogic func()) {
	t.Helper()

	var recoverIface interface{}

	func() {
		defer func() {
			recoverIface = recover()
		}()

		testLogic()
	}()

	if recoverIface == nil {
		t.Fatalf("expected t.Fatal(), got none")
	}

	recoverStr, ok := recoverIface.(string)

	if !ok {
		t.Fatalf("expected string from recover(), got: %v (%T)", recoverIface, recoverIface)
	}

	// this string is hardcoded in github.com/mitchellh/go-testing-interface
	if !strings.HasPrefix(recoverStr, "testing.T failed, see logs for output") {
		t.Fatalf("expected t.Fatal(), got: %s", recoverStr)
	}
}

func TestParallelTest(t *testing.T) {
	mt := new(mockT)

	ParallelTest(mt, TestCase{
		IsUnitTest: true,
		ProviderFactories: map[string]func() (*schema.Provider, error){
			"test": func() (*schema.Provider, error) { //nolint:unparam // required signature
				return nil, nil
			},
		},
		Steps: []TestStep{
			{
				Config: "# not empty",
			},
		},
	})

	if !mt.ParallelCalled {
		t.Fatal("Parallel() not called")
	}
}

func TestComposeAggregateTestCheckFunc(t *testing.T) {
	check1 := func(s *terraform.State) error {
		return errors.New("Error 1")
	}

	check2 := func(s *terraform.State) error {
		return errors.New("Error 2")
	}

	f := ComposeAggregateTestCheckFunc(check1, check2)
	err := f(nil)
	if err == nil {
		t.Fatalf("Expected errors")
	}

	multi := err.(*multierror.Error)
	if !strings.Contains(multi.Errors[0].Error(), "Error 1") {
		t.Fatalf("Expected Error 1, Got %s", multi.Errors[0])
	}
	if !strings.Contains(multi.Errors[1].Error(), "Error 2") {
		t.Fatalf("Expected Error 2, Got %s", multi.Errors[1])
	}
}

func TestComposeTestCheckFunc(t *testing.T) {
	cases := []struct {
		F      []TestCheckFunc
		Result string
	}{
		{
			F: []TestCheckFunc{
				func(*terraform.State) error { return nil },
			},
			Result: "",
		},

		{
			F: []TestCheckFunc{
				func(*terraform.State) error {
					return fmt.Errorf("error")
				},
				func(*terraform.State) error { return nil },
			},
			Result: "Check 1/2 error: error",
		},

		{
			F: []TestCheckFunc{
				func(*terraform.State) error { return nil },
				func(*terraform.State) error {
					return fmt.Errorf("error")
				},
			},
			Result: "Check 2/2 error: error",
		},

		{
			F: []TestCheckFunc{
				func(*terraform.State) error { return nil },
				func(*terraform.State) error { return nil },
			},
			Result: "",
		},
	}

	for i, tc := range cases {
		f := ComposeTestCheckFunc(tc.F...)
		err := f(nil)
		if err == nil {
			err = fmt.Errorf("")
		}
		if tc.Result != err.Error() {
			t.Fatalf("Case %d bad: %s", i, err)
		}
	}
}

// mockT implements TestT for testing
type mockT struct {
	testinginterface.RuntimeT

	ParallelCalled bool
}

func (t *mockT) Parallel() {
	t.ParallelCalled = true
}

func (t *mockT) Name() string {
	return "MockedName"
}

func TestTest_Main(t *testing.T) {
	flag.Parse()
	if *flagSweep == "" {
		// Tests for the TestMain method used for Sweepers will panic without the -sweep
		// flag specified. Mock the value for now
		*flagSweep = "us-east-1"
	}

	cases := []struct {
		Name            string
		Sweepers        map[string]*Sweeper
		ExpectedRunList []string
		SweepRun        string
	}{
		{
			Name: "basic passing",
			Sweepers: map[string]*Sweeper{
				"aws_dummy": {
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_dummy"},
		},
	}

	for _, tc := range cases {
		// reset sweepers
		sweeperFuncs = map[string]*Sweeper{}

		t.Run(tc.Name, func(t *testing.T) {
			for n, s := range tc.Sweepers {
				AddTestSweepers(n, s)
			}
			*flagSweepRun = tc.SweepRun

			// Verify it does not exit/panic
			TestMain(&testing.M{})
		})
	}
}

func TestFilterSweepers(t *testing.T) {
	cases := []struct {
		Name             string
		Sweepers         map[string]*Sweeper
		Filter           string
		ExpectedSweepers []string
	}{
		{
			Name: "normal",
			Sweepers: map[string]*Sweeper{
				"aws_dummy": {
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
			},
			ExpectedSweepers: []string{"aws_dummy"},
		},
		{
			Name: "with dep",
			Sweepers: map[string]*Sweeper{
				"aws_dummy": {
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
				"aws_top": {
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": {
					Name: "aws_sub",
					F:    mockSweeperFunc,
				},
			},
			ExpectedSweepers: []string{"aws_dummy", "aws_sub", "aws_top"},
		},
		{
			Name: "with filter",
			Sweepers: map[string]*Sweeper{
				"aws_dummy": {
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
				"aws_top": {
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": {
					Name: "aws_sub",
					F:    mockSweeperFunc,
				},
			},
			ExpectedSweepers: []string{"aws_dummy"},
			Filter:           "aws_dummy",
		},
		{
			Name: "with two filters",
			Sweepers: map[string]*Sweeper{
				"aws_dummy": {
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
				"aws_top": {
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": {
					Name: "aws_sub",
					F:    mockSweeperFunc,
				},
			},
			ExpectedSweepers: []string{"aws_dummy", "aws_sub"},
			Filter:           "aws_dummy,aws_sub",
		},
		{
			Name: "with dep and filter",
			Sweepers: map[string]*Sweeper{
				"aws_dummy": {
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
				"aws_top": {
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": {
					Name: "aws_sub",
					F:    mockSweeperFunc,
				},
			},
			ExpectedSweepers: []string{"aws_sub", "aws_top"},
			Filter:           "aws_top",
		},
		{
			Name: "with non-matching filter",
			Sweepers: map[string]*Sweeper{
				"aws_dummy": {
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
				"aws_top": {
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": {
					Name: "aws_sub",
					F:    mockSweeperFunc,
				},
			},
			Filter: "none",
		},
		{
			Name: "with nested depenencies and top level filter",
			Sweepers: map[string]*Sweeper{
				"not_matching": {
					Name: "not_matching",
					F:    mockSweeperFunc,
				},
				"matching_level1": {
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": {
					Name:         "matching_level2",
					Dependencies: []string{"matching_level3"},
					F:            mockSweeperFunc,
				},
				"matching_level3": {
					Name: "matching_level3",
					F:    mockSweeperFunc,
				},
			},
			ExpectedSweepers: []string{"matching_level1", "matching_level2", "matching_level3"},
			Filter:           "matching_level1",
		},
		{
			Name: "with nested depenencies and middle level filter",
			Sweepers: map[string]*Sweeper{
				"not_matching": {
					Name: "not_matching",
					F:    mockSweeperFunc,
				},
				"matching_level1": {
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": {
					Name:         "matching_level2",
					Dependencies: []string{"matching_level3"},
					F:            mockSweeperFunc,
				},
				"matching_level3": {
					Name: "matching_level3",
					F:    mockSweeperFunc,
				},
			},
			ExpectedSweepers: []string{"matching_level2", "matching_level3"},
			Filter:           "matching_level2",
		},
		{
			Name: "with nested depenencies and bottom level filter",
			Sweepers: map[string]*Sweeper{
				"not_matching": {
					Name: "not_matching",
					F:    mockSweeperFunc,
				},
				"matching_level1": {
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": {
					Name:         "matching_level2",
					Dependencies: []string{"matching_level3"},
					F:            mockSweeperFunc,
				},
				"matching_level3": {
					Name: "matching_level3",
					F:    mockSweeperFunc,
				},
			},
			ExpectedSweepers: []string{"matching_level3"},
			Filter:           "matching_level3",
		},
	}

	for _, tc := range cases {
		// reset sweepers
		sweeperFuncs = map[string]*Sweeper{}

		t.Run(tc.Name, func(t *testing.T) {
			actualSweepers := filterSweepers(tc.Filter, tc.Sweepers)

			var keys []string
			for k := range actualSweepers {
				keys = append(keys, k)
			}

			sort.Strings(keys)
			sort.Strings(tc.ExpectedSweepers)
			if !reflect.DeepEqual(keys, tc.ExpectedSweepers) {
				t.Fatalf("Expected keys mismatch, expected:\n%#v\ngot:\n%#v\n", tc.ExpectedSweepers, keys)
			}
		})
	}
}

func TestFilterSweeperWithDependencies(t *testing.T) {
	cases := []struct {
		Name             string
		Sweepers         map[string]*Sweeper
		StartingSweeper  string
		ExpectedSweepers []string
	}{
		{
			Name: "no dependencies",
			Sweepers: map[string]*Sweeper{
				"matching_level1": {
					Name: "matching_level1",
					F:    mockSweeperFunc,
				},
				"non_matching": {
					Name: "non_matching",
					F:    mockSweeperFunc,
				},
			},
			StartingSweeper:  "matching_level1",
			ExpectedSweepers: []string{"matching_level1"},
		},
		{
			Name: "one level one dependency",
			Sweepers: map[string]*Sweeper{
				"matching_level1": {
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": {
					Name: "matching_level2",
					F:    mockSweeperFunc,
				},
			},
			StartingSweeper:  "matching_level1",
			ExpectedSweepers: []string{"matching_level1", "matching_level2"},
		},
		{
			Name: "one level multiple dependencies",
			Sweepers: map[string]*Sweeper{
				"matching_level1": {
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2a", "matching_level2b"},
					F:            mockSweeperFunc,
				},
				"matching_level2a": {
					Name: "matching_level2a",
					F:    mockSweeperFunc,
				},
				"matching_level2b": {
					Name: "matching_level2b",
					F:    mockSweeperFunc,
				},
			},
			StartingSweeper:  "matching_level1",
			ExpectedSweepers: []string{"matching_level1", "matching_level2a", "matching_level2b"},
		},
		{
			Name: "multiple level one dependency",
			Sweepers: map[string]*Sweeper{
				"matching_level1": {
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": {
					Name:         "matching_level2",
					Dependencies: []string{"matching_level3"},
					F:            mockSweeperFunc,
				},
				"matching_level3": {
					Name: "matching_level3",
					F:    mockSweeperFunc,
				},
			},
			StartingSweeper:  "matching_level1",
			ExpectedSweepers: []string{"matching_level1", "matching_level2", "matching_level3"},
		},
		{
			Name: "multiple level multiple dependencies",
			Sweepers: map[string]*Sweeper{
				"matching_level1": {
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2a", "matching_level2b"},
					F:            mockSweeperFunc,
				},
				"matching_level2a": {
					Name:         "matching_level2a",
					Dependencies: []string{"matching_level3a", "matching_level3b"},
					F:            mockSweeperFunc,
				},
				"matching_level2b": {
					Name:         "matching_level2b",
					Dependencies: []string{"matching_level3c", "matching_level3d"},
					F:            mockSweeperFunc,
				},
				"matching_level3a": {
					Name: "matching_level3a",
					F:    mockSweeperFunc,
				},
				"matching_level3b": {
					Name: "matching_level3b",
					F:    mockSweeperFunc,
				},
				"matching_level3c": {
					Name: "matching_level3c",
					F:    mockSweeperFunc,
				},
				"matching_level3d": {
					Name: "matching_level3d",
					F:    mockSweeperFunc,
				},
			},
			StartingSweeper:  "matching_level1",
			ExpectedSweepers: []string{"matching_level1", "matching_level2a", "matching_level2b", "matching_level3a", "matching_level3b", "matching_level3c", "matching_level3d"},
		},
		{
			Name: "no parents one level",
			Sweepers: map[string]*Sweeper{
				"matching_level1": {
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": {
					Name:         "matching_level2",
					Dependencies: []string{"matching_level3"},
					F:            mockSweeperFunc,
				},
				"matching_level3": {
					Name: "matching_level3",
					F:    mockSweeperFunc,
				},
			},
			StartingSweeper:  "matching_level2",
			ExpectedSweepers: []string{"matching_level2", "matching_level3"},
		},
		{
			Name: "no parents multiple level",
			Sweepers: map[string]*Sweeper{
				"matching_level1": {
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": {
					Name:         "matching_level2",
					Dependencies: []string{"matching_level3"},
					F:            mockSweeperFunc,
				},
				"matching_level3": {
					Name: "matching_level3",
					F:    mockSweeperFunc,
				},
			},
			StartingSweeper:  "matching_level3",
			ExpectedSweepers: []string{"matching_level3"},
		},
		{
			Name: "one level missing dependency",
			Sweepers: map[string]*Sweeper{
				"matching_level1": {
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2a", "matching_level2c"},
					F:            mockSweeperFunc,
				},
				"matching_level2a": {
					Name: "matching_level2a",
					F:    mockSweeperFunc,
				},
				"matching_level2b": {
					Name: "matching_level2b",
					F:    mockSweeperFunc,
				},
			},
			StartingSweeper:  "matching_level1",
			ExpectedSweepers: []string{"matching_level1", "matching_level2a"},
		},
		{
			Name: "multiple level missing dependencies",
			Sweepers: map[string]*Sweeper{
				"matching_level1": {
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2a", "matching_level2b", "matching_level2c"},
					F:            mockSweeperFunc,
				},
				"matching_level2a": {
					Name:         "matching_level2a",
					Dependencies: []string{"matching_level3a", "matching_level3e"},
					F:            mockSweeperFunc,
				},
				"matching_level2b": {
					Name:         "matching_level2b",
					Dependencies: []string{"matching_level3c", "matching_level3f"},
					F:            mockSweeperFunc,
				},
				"matching_level3a": {
					Name: "matching_level3a",
					F:    mockSweeperFunc,
				},
				"matching_level3b": {
					Name: "matching_level3b",
					F:    mockSweeperFunc,
				},
				"matching_level3c": {
					Name: "matching_level3c",
					F:    mockSweeperFunc,
				},
				"matching_level3d": {
					Name: "matching_level3d",
					F:    mockSweeperFunc,
				},
			},
			StartingSweeper:  "matching_level1",
			ExpectedSweepers: []string{"matching_level1", "matching_level2a", "matching_level2b", "matching_level3a", "matching_level3c"},
		},
	}

	for _, tc := range cases {
		// reset sweepers
		sweeperFuncs = map[string]*Sweeper{}

		t.Run(tc.Name, func(t *testing.T) {
			actualSweepers := filterSweeperWithDependencies(tc.StartingSweeper, tc.Sweepers)

			var keys []string
			for k := range actualSweepers {
				keys = append(keys, k)
			}

			sort.Strings(keys)
			sort.Strings(tc.ExpectedSweepers)
			if !reflect.DeepEqual(keys, tc.ExpectedSweepers) {
				t.Fatalf("Expected keys mismatch, expected:\n%#v\ngot:\n%#v\n", tc.ExpectedSweepers, keys)
			}
		})
	}
}

func TestRunSweepers(t *testing.T) {
	cases := []struct {
		Name            string
		Sweepers        map[string]*Sweeper
		ExpectedRunList []string
		SweepRun        string
		AllowFailures   bool
		ExpectError     bool
	}{
		{
			Name: "single",
			Sweepers: map[string]*Sweeper{
				"aws_dummy": {
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_dummy"},
		},
		{
			Name: "multiple",
			Sweepers: map[string]*Sweeper{
				"aws_one": {
					Name: "aws_one",
					F:    mockSweeperFunc,
				},
				"aws_two": {
					Name: "aws_two",
					F:    mockSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_one", "aws_two"},
		},
		{
			Name: "multiple with dep",
			Sweepers: map[string]*Sweeper{
				"aws_dummy": {
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
				"aws_top": {
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": {
					Name: "aws_sub",
					F:    mockSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_dummy", "aws_sub", "aws_top"},
		},
		{
			Name: "failing dep",
			Sweepers: map[string]*Sweeper{
				"aws_top": {
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": {
					Name: "aws_sub",
					F:    mockFailingSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_sub"},
			ExpectError:     true,
		},
		{
			Name: "failing dep allow failures",
			Sweepers: map[string]*Sweeper{
				"aws_top": {
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": {
					Name: "aws_sub",
					F:    mockFailingSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_sub", "aws_top"},
			AllowFailures:   true,
			ExpectError:     true,
		},
		{
			Name: "failing top",
			Sweepers: map[string]*Sweeper{
				"aws_top": {
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockFailingSweeperFunc,
				},
				"aws_sub": {
					Name: "aws_sub",
					F:    mockSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_sub", "aws_top"},
			ExpectError:     true,
		},
		{
			Name: "failing top allow failures",
			Sweepers: map[string]*Sweeper{
				"aws_top": {
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockFailingSweeperFunc,
				},
				"aws_sub": {
					Name: "aws_sub",
					F:    mockSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_sub", "aws_top"},
			AllowFailures:   true,
			ExpectError:     true,
		},
		{
			Name: "failing top and dep",
			Sweepers: map[string]*Sweeper{
				"aws_top": {
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockFailingSweeperFunc,
				},
				"aws_sub": {
					Name: "aws_sub",
					F:    mockFailingSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_sub"},
			ExpectError:     true,
		},
		{
			Name: "failing top and dep allow failues",
			Sweepers: map[string]*Sweeper{
				"aws_top": {
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockFailingSweeperFunc,
				},
				"aws_sub": {
					Name: "aws_sub",
					F:    mockFailingSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_sub", "aws_top"},
			AllowFailures:   true,
			ExpectError:     true,
		},
	}

	for _, tc := range cases {
		// reset sweepers
		sweeperFuncs = map[string]*Sweeper{}

		t.Run(tc.Name, func(t *testing.T) {
			sweeperRunList, err := runSweepers([]string{"test"}, tc.Sweepers, tc.AllowFailures)
			fmt.Printf("sweeperRunList: %#v\n", sweeperRunList)

			if err == nil && tc.ExpectError {
				t.Fatalf("expected error, did not receive error")
			}

			if err != nil && !tc.ExpectError {
				t.Fatalf("did not expect error, received error: %s", err)
			}

			// get list of tests ran from sweeperRunList keys
			var keys []string
			for k := range sweeperRunList["test"] {
				keys = append(keys, k)
			}

			sort.Strings(keys)
			sort.Strings(tc.ExpectedRunList)
			if !reflect.DeepEqual(keys, tc.ExpectedRunList) {
				t.Fatalf("Expected keys mismatch, expected:\n%#v\ngot:\n%#v\n", tc.ExpectedRunList, keys)
			}
		})
	}
}

func mockFailingSweeperFunc(s string) error {
	return errors.New("failing sweeper")
}

func mockSweeperFunc(s string) error {
	return nil
}

func TestTestCheckResourceAttr(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		state         *terraform.State
		key           string
		value         string
		expectedError error
	}{
		"attribute not found": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{},
								},
							},
						},
					},
				},
			},
			key:           "nonexistent",
			value:         "test-value",
			expectedError: fmt.Errorf("Attribute 'nonexistent' not found"),
		},
		"bool attribute match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_bool_attribute": "true",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_bool_attribute",
			value: "true",
		},
		"bool attribute mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_bool_attribute": "true",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_bool_attribute",
			value:         "false",
			expectedError: fmt.Errorf("Attribute 'test_bool_attribute' expected \"false\", got \"true\""),
		},
		"float attribute match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_float_attribute": "1.2",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_float_attribute",
			value: "1.2",
		},
		"float attribute mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_float_attribute": "1.2",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_float_attribute",
			value:         "3.4",
			expectedError: fmt.Errorf("Attribute 'test_float_attribute' expected \"3.4\", got \"1.2\""),
		},
		"integer attribute match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_integer_attribute": "123",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_integer_attribute",
			value: "123",
		},
		"integer attribute mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_integer_attribute": "123",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_integer_attribute",
			value:         "456",
			expectedError: fmt.Errorf("Attribute 'test_integer_attribute' expected \"456\", got \"123\""),
		},
		"list attribute directly": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_list_attribute",
			value:         "test-value",
			expectedError: fmt.Errorf("list or set attribute 'test_list_attribute' must be checked by element count key (test_list_attribute.#) or element value keys (e.g. test_list_attribute.0)"),
		},
		"list attribute element count match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_list_attribute.#",
			value: "1",
		},
		"list attribute element count mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_list_attribute.#",
			value:         "2",
			expectedError: fmt.Errorf("Attribute 'test_list_attribute.#' expected \"2\", got \"1\""),
		},
		"list attribute element count match 0 when empty": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "0",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_list_attribute.#",
			value: "0",
		},
		// Special case with .# and value 0
		"list attribute element count match 0 when missing": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{},
								},
							},
						},
					},
				},
			},
			key:   "test_list_attribute.#",
			value: "0",
		},
		"list attribute element value match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_list_attribute.0",
			value: "test-value",
		},
		"list attribute element value mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_list_attribute.0",
			value:         "not-test-value",
			expectedError: fmt.Errorf("Attribute 'test_list_attribute.0' expected \"not-test-value\", got \"test-value\""),
		},
		"map attribute directly": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_map_attribute",
			value:         "test-value",
			expectedError: fmt.Errorf("map attribute 'test_map_attribute' must be checked by element count key (test_map_attribute.%%) or element value keys (e.g. test_map_attribute.examplekey)"),
		},
		"map attribute element count match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_map_attribute.%",
			value: "1",
		},
		"map attribute element count mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_map_attribute.%",
			value:         "2",
			expectedError: fmt.Errorf("Attribute 'test_map_attribute.%%' expected \"2\", got \"1\""),
		},
		"map attribute element count match 0 when empty": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%": "0",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_map_attribute.%",
			value: "0",
		},
		// Special case with .% and value 0
		"map attribute element count match 0 when missing": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{},
								},
							},
						},
					},
				},
			},
			key:   "test_map_attribute.%",
			value: "0",
		},
		"map attribute element value match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_map_attribute.testkey1",
			value: "test-value-1",
		},
		"map attribute element value mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_map_attribute.testkey1",
			value:         "test-value-2",
			expectedError: fmt.Errorf("Attribute 'test_map_attribute.testkey1' expected \"test-value-2\", got \"test-value-1\""),
		},
		"set attribute indexing error": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_set_attribute.#":                         "1",
										"test_set_attribute.101.test_string_attribute": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_set_attribute.101.nonexistent",
			value:         "test-value",
			expectedError: fmt.Errorf("likely indexes into TypeSet"),
		},
		"string attribute match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_string_attribute": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_string_attribute",
			value: "test-value",
		},
		"string attribute mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_string_attribute": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_string_attribute",
			value:         "not-test-value",
			expectedError: fmt.Errorf("Attribute 'test_string_attribute' expected \"not-test-value\", got \"test-value\""),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := TestCheckResourceAttr("test_resource", testCase.key, testCase.value)(testCase.state)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("expected error: %s", testCase.expectedError)
			}
		})
	}
}

func TestTestCheckResourceAttrWith(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		state         *terraform.State
		key           string
		value         string
		expectedError error
	}{
		"attribute not found": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{},
								},
							},
						},
					},
				},
			},
			key:           "nonexistent",
			value:         "test-value",
			expectedError: fmt.Errorf("Attribute 'nonexistent' expected to be set"),
		},
		"bool attribute match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_bool_attribute": "true",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_bool_attribute",
			value: "true",
		},
		"bool attribute mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_bool_attribute": "true",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_bool_attribute",
			value:         "false",
			expectedError: fmt.Errorf("attribute 'test_bool_attribute' expected 'false', got 'true'"),
		},
		"list attribute directly": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_list_attribute",
			value:         "test-value",
			expectedError: fmt.Errorf("list or set attribute 'test_list_attribute' must be checked by element count key (test_list_attribute.#) or element value keys (e.g. test_list_attribute.0)"),
		},
		"list attribute element count match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_list_attribute.#",
			value: "1",
		},
		"list attribute element count mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_list_attribute.#",
			value:         "2",
			expectedError: fmt.Errorf("attribute 'test_list_attribute.#' expected '2', got '1'"),
		},
		"list attribute element count match 0 when empty": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "0",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_list_attribute.#",
			value: "0",
		},
		"list attribute element value match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_list_attribute.0",
			value: "test-value",
		},
		"list attribute element value mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_list_attribute.0",
			value:         "not-test-value",
			expectedError: fmt.Errorf("attribute 'test_list_attribute.0' expected 'not-test-value', got 'test-value'"),
		},
		"map attribute directly": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_map_attribute",
			value:         "test-value",
			expectedError: fmt.Errorf("map attribute 'test_map_attribute' must be checked by element count key (test_map_attribute.%%) or element value keys (e.g. test_map_attribute.examplekey)"),
		},
		"map attribute element count match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_map_attribute.%",
			value: "1",
		},
		"map attribute element count mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_map_attribute.%",
			value:         "2",
			expectedError: fmt.Errorf("attribute 'test_map_attribute.%%' expected '2', got '1'"),
		},
		"map attribute element count match 0 when empty": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%": "0",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_map_attribute.%",
			value: "0",
		},
		"map attribute element value match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_map_attribute.testkey1",
			value: "test-value-1",
		},
		"map attribute element value mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_map_attribute.testkey1",
			value:         "test-value-2",
			expectedError: fmt.Errorf("attribute 'test_map_attribute.testkey1' expected 'test-value-2', got 'test-value-1'"),
		},
		"set attribute indexing error": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_set_attribute.#":                         "1",
										"test_set_attribute.101.test_string_attribute": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_set_attribute.101.nonexistent",
			value:         "test-value",
			expectedError: fmt.Errorf("likely indexes into TypeSet"),
		},
		"string attribute match": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_string_attribute": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:   "test_string_attribute",
			value: "test-value",
		},
		"string attribute mismatch": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_string_attribute": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_string_attribute",
			value:         "not-test-value",
			expectedError: fmt.Errorf("attribute 'test_string_attribute' expected 'not-test-value', got 'test-value'"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := TestCheckResourceAttrWith("test_resource", testCase.key, func(v string) error {
				if testCase.value != v {
					return fmt.Errorf("attribute '%s' expected '%s', got '%s'", testCase.key, testCase.value, v)
				}
				return nil
			})(testCase.state)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("expected error: %s", testCase.expectedError)
			}
		})
	}
}

func TestTestCheckNoResourceAttr(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		state         *terraform.State
		key           string
		expectedError error
	}{
		"attribute not found": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{},
								},
							},
						},
					},
				},
			},
			key: "nonexistent",
		},
		"attribute found": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_bool_attribute": "true",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_bool_attribute",
			expectedError: fmt.Errorf("Attribute 'test_bool_attribute' found when not expected"),
		},
		"list attribute directly": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_list_attribute",
			expectedError: fmt.Errorf("list or set attribute 'test_list_attribute' must be checked by element count key (test_list_attribute.#) or element value keys (e.g. test_list_attribute.0)"),
		},
		// Special case with .# and value 0
		"list attribute element count match 0 when empty": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "0",
									},
								},
							},
						},
					},
				},
			},
			key: "test_list_attribute.#",
		},
		"list attribute element count mismatch 0 when non-empty": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_list_attribute.#",
			expectedError: fmt.Errorf("Attribute 'test_list_attribute.#' found when not expected"),
		},
		"map attribute directly": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_map_attribute",
			expectedError: fmt.Errorf("map attribute 'test_map_attribute' must be checked by element count key (test_map_attribute.%%) or element value keys (e.g. test_map_attribute.examplekey)"),
		},
		// Special case with .% and value 0
		"map attribute element count match 0 when empty": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%": "0",
									},
								},
							},
						},
					},
				},
			},
			key: "test_map_attribute.%",
		},
		"map attribute element count mismatch 0 when non-empty": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_map_attribute.%",
			expectedError: fmt.Errorf("Attribute 'test_map_attribute.%%' found when not expected"),
		},
		"set attribute indexing error": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_set_attribute.#":                         "1",
										"test_set_attribute.101.test_string_attribute": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_set_attribute.101.test_string_attribute",
			expectedError: fmt.Errorf("likely indexes into TypeSet"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := TestCheckNoResourceAttr("test_resource", testCase.key)(testCase.state)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("expected error: %s", testCase.expectedError)
			}
		})
	}
}

func TestTestCheckResourceAttrPair(t *testing.T) {
	tests := map[string]struct {
		nameFirst  string
		keyFirst   string
		nameSecond string
		keySecond  string
		state      *terraform.State
		wantErr    string
	}{
		"self": {
			nameFirst:  "test.a",
			keyFirst:   "a",
			nameSecond: "test.a",
			keySecond:  "a",
			state: &terraform.State{
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test.a": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"a": "boop",
									},
								},
							},
							"test.b": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"b": "boop",
									},
								},
							},
						},
					},
				},
			},
			wantErr: `comparing self: resource test.a attribute a`,
		},
		"exist match": {
			nameFirst:  "test.a",
			keyFirst:   "a",
			nameSecond: "test.b",
			keySecond:  "b",
			state: &terraform.State{
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test.a": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"a": "boop",
									},
								},
							},
							"test.b": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"b": "boop",
									},
								},
							},
						},
					},
				},
			},
			wantErr: ``,
		},
		"nonexist match": {
			nameFirst:  "test.a",
			keyFirst:   "a",
			nameSecond: "test.b",
			keySecond:  "b",
			state: &terraform.State{
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test.a": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{},
								},
							},
							"test.b": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{},
								},
							},
						},
					},
				},
			},
			wantErr: ``,
		},
		"exist nonmatch": {
			nameFirst:  "test.a",
			keyFirst:   "a",
			nameSecond: "test.b",
			keySecond:  "b",
			state: &terraform.State{
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test.a": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"a": "beep",
									},
								},
							},
							"test.b": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"b": "boop",
									},
								},
							},
						},
					},
				},
			},
			wantErr: `test.a: Attribute 'a' expected "boop", got "beep"`,
		},
		"inconsistent exist a": {
			nameFirst:  "test.a",
			keyFirst:   "a",
			nameSecond: "test.b",
			keySecond:  "b",
			state: &terraform.State{
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test.a": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"a": "beep",
									},
								},
							},
							"test.b": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{},
								},
							},
						},
					},
				},
			},
			wantErr: `test.a: Attribute "a" is "beep", but "b" is not set in test.b`,
		},
		"inconsistent exist b": {
			nameFirst:  "test.a",
			keyFirst:   "a",
			nameSecond: "test.b",
			keySecond:  "b",
			state: &terraform.State{
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test.a": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{},
								},
							},
							"test.b": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"b": "boop",
									},
								},
							},
						},
					},
				},
			},
			wantErr: `test.a: Attribute "a" not set, but "b" is set in test.b as "boop"`,
		},
		"unset and 0 equal list": {
			nameFirst:  "test.a",
			keyFirst:   "a.#",
			nameSecond: "test.b",
			keySecond:  "a.#",
			state: &terraform.State{
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test.a": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"a.#": "0",
									},
								},
							},
							"test.b": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{},
								},
							},
						},
					},
				},
			},
			wantErr: ``,
		},
		"unset and 0 equal map": {
			nameFirst:  "test.a",
			keyFirst:   "a.%",
			nameSecond: "test.b",
			keySecond:  "a.%",
			state: &terraform.State{
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test.a": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"a.%": "0",
									},
								},
							},
							"test.b": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{},
								},
							},
						},
					},
				},
			},
			wantErr: ``,
		},
		"count equal": {
			nameFirst:  "test.a",
			keyFirst:   "a.%",
			nameSecond: "test.b",
			keySecond:  "a.%",
			state: &terraform.State{
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test.a": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"a.%": "1",
									},
								},
							},
							"test.b": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"a.%": "1",
									}},
							},
						},
					},
				},
			},
			wantErr: ``,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fn := TestCheckResourceAttrPair(test.nameFirst, test.keyFirst, test.nameSecond, test.keySecond)
			err := fn(test.state)

			if test.wantErr != "" {
				if err == nil {
					t.Fatalf("succeeded; want error\nwant: %s", test.wantErr)
				}
				if got, want := err.Error(), test.wantErr; got != want {
					t.Fatalf("wrong error\ngot:  %s\nwant: %s", got, want)
				}
				return
			}

			if err != nil {
				t.Fatalf("failed; want success\ngot: %s", err.Error())
			}
		})
	}
}

func TestTestCheckResourceAttrSet(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		state         *terraform.State
		key           string
		expectedError error
	}{
		"attribute not found": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{},
								},
							},
						},
					},
				},
			},
			key:           "nonexistent",
			expectedError: fmt.Errorf("Attribute 'nonexistent' expected to be set"),
		},
		"attribute found": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_bool_attribute": "true",
									},
								},
							},
						},
					},
				},
			},
			key: "test_bool_attribute",
		},
		"list attribute directly": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_list_attribute.#": "1",
										"test_list_attribute.0": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_list_attribute",
			expectedError: fmt.Errorf("list or set attribute 'test_list_attribute' must be checked by element count key (test_list_attribute.#) or element value keys (e.g. test_list_attribute.0)"),
		},
		"map attribute directly": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_map_attribute.%":        "1",
										"test_map_attribute.testkey1": "test-value-1",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_map_attribute",
			expectedError: fmt.Errorf("map attribute 'test_map_attribute' must be checked by element count key (test_map_attribute.%%) or element value keys (e.g. test_map_attribute.examplekey)"),
		},
		"set attribute indexing error": {
			state: &terraform.State{
				IsBinaryDrivenTest: true, // Always true now
				Modules: []*terraform.ModuleState{
					{
						Path: []string{"root"},
						Resources: map[string]*terraform.ResourceState{
							"test_resource": {
								Primary: &terraform.InstanceState{
									Attributes: map[string]string{
										"test_set_attribute.#":                         "1",
										"test_set_attribute.101.test_string_attribute": "test-value",
									},
								},
							},
						},
					},
				},
			},
			key:           "test_set_attribute.101.nonexistent",
			expectedError: fmt.Errorf("likely indexes into TypeSet"),
		},
	}

	for name, testCase := range testCases {
		name, testCase := name, testCase

		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := TestCheckResourceAttrSet("test_resource", testCase.key)(testCase.state)

			if err != nil {
				if testCase.expectedError == nil {
					t.Fatalf("expected no error, got: %s", err)
				}

				if !strings.Contains(err.Error(), testCase.expectedError.Error()) {
					t.Fatalf("expected error %q, got: %s", testCase.expectedError, err)
				}
			}

			if err == nil && testCase.expectedError != nil {
				t.Fatalf("expected error: %s", testCase.expectedError)
			}
		})
	}
}

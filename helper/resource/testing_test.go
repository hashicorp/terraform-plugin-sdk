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
	if err := os.Setenv(TestEnvVar, "1"); err != nil {
		panic(err)
	}
}

func TestParallelTest(t *testing.T) {
	mt := new(mockT)

	ParallelTest(mt, TestCase{IsUnitTest: true})

	if !mt.ParallelCalled {
		t.Fatal("Parallel() not called")
	}
}

func TestTest_factoryError(t *testing.T) {
	resourceFactoryError := fmt.Errorf("resource factory error")

	factory := func() (*schema.Provider, error) {
		return nil, resourceFactoryError
	}
	mt := new(mockT)
	recovered := false

	func() {
		defer func() {
			r := recover()
			// this string is hardcoded in github.com/mitchellh/go-testing-interface
			if s, ok := r.(string); !ok || !strings.HasPrefix(s, "testing.T failed, see logs for output") {
				panic(r)
			}
			recovered = true
		}()
		Test(mt, TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"test": factory,
			},
			IsUnitTest: true,
		})
	}()

	if !recovered {
		t.Fatalf("test should've failed fatally")
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

func TestCheckResourceAttr_empty(t *testing.T) {
	s := terraform.NewState()
	s.AddModuleState(&terraform.ModuleState{
		Path: []string{"root"},
		Resources: map[string]*terraform.ResourceState{
			"test_resource": {
				Primary: &terraform.InstanceState{
					Attributes: map[string]string{
						"empty_list.#": "0",
						"empty_map.%":  "0",
					},
				},
			},
		},
	})

	for _, key := range []string{
		"empty_list.#",
		"empty_map.%",
		"missing_list.#",
		"missing_map.%",
	} {
		t.Run(key, func(t *testing.T) {
			check := TestCheckResourceAttr("test_resource", key, "0")
			if err := check(s); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCheckNoResourceAttr_empty(t *testing.T) {
	s := terraform.NewState()
	s.AddModuleState(&terraform.ModuleState{
		Path: []string{"root"},
		Resources: map[string]*terraform.ResourceState{
			"test_resource": {
				Primary: &terraform.InstanceState{
					Attributes: map[string]string{
						"empty_list.#": "0",
						"empty_map.%":  "0",
					},
				},
			},
		},
	})

	for _, key := range []string{
		"empty_list.#",
		"empty_map.%",
		"missing_list.#",
		"missing_map.%",
	} {
		t.Run(key, func(t *testing.T) {
			check := TestCheckNoResourceAttr("test_resource", key)
			if err := check(s); err != nil {
				t.Fatal(err)
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

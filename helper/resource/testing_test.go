package resource

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/internal/plugin/discovery"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func init() {
	// TODO: Remove when we remove the guard on id checks
	if err := os.Setenv("TF_ACC_IDONLY", "1"); err != nil {
		panic(err)
	}

	if err := os.Setenv(TestEnvVar, "1"); err != nil {
		panic(err)
	}
}

func TestParallelTest(t *testing.T) {
	mt := new(mockT)
	ParallelTest(mt, TestCase{})

	if !mt.ParallelCalled {
		t.Fatal("Parallel() not called")
	}
}

func TestTest_factoryError(t *testing.T) {
	resourceFactoryError := fmt.Errorf("resource factory error")

	factory := func() (terraform.ResourceProvider, error) {
		return nil, resourceFactoryError
	}

	mt := new(mockT)
	Test(mt, TestCase{
		ProviderFactories: map[string]terraform.ResourceProviderFactory{
			"test": factory,
		},
		Steps: []TestStep{
			TestStep{
				ExpectError: regexp.MustCompile("resource factory error"),
			},
		},
	})

	if !mt.failed() {
		t.Fatal("test should've failed")
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
	ErrorCalled    bool
	ErrorArgs      []interface{}
	FatalCalled    bool
	FatalArgs      []interface{}
	ParallelCalled bool
	SkipCalled     bool
	SkipArgs       []interface{}

	f bool
}

func (t *mockT) Error(args ...interface{}) {
	t.ErrorCalled = true
	t.ErrorArgs = args
	t.f = true
}

func (t *mockT) Fatal(args ...interface{}) {
	t.FatalCalled = true
	t.FatalArgs = args
	t.f = true
}

func (t *mockT) Parallel() {
	t.ParallelCalled = true
}

func (t *mockT) Skip(args ...interface{}) {
	t.SkipCalled = true
	t.SkipArgs = args
	t.f = true
}

func (t *mockT) Name() string {
	return "MockedName"
}

func (t *mockT) failed() bool {
	return t.f
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
				"aws_dummy": &Sweeper{
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
				"aws_dummy": &Sweeper{
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
			},
			ExpectedSweepers: []string{"aws_dummy"},
		},
		{
			Name: "with dep",
			Sweepers: map[string]*Sweeper{
				"aws_dummy": &Sweeper{
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
				"aws_top": &Sweeper{
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": &Sweeper{
					Name: "aws_sub",
					F:    mockSweeperFunc,
				},
			},
			ExpectedSweepers: []string{"aws_dummy", "aws_sub", "aws_top"},
		},
		{
			Name: "with filter",
			Sweepers: map[string]*Sweeper{
				"aws_dummy": &Sweeper{
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
				"aws_top": &Sweeper{
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": &Sweeper{
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
				"aws_dummy": &Sweeper{
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
				"aws_top": &Sweeper{
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": &Sweeper{
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
				"aws_dummy": &Sweeper{
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
				"aws_top": &Sweeper{
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": &Sweeper{
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
				"aws_dummy": &Sweeper{
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
				"aws_top": &Sweeper{
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": &Sweeper{
					Name: "aws_sub",
					F:    mockSweeperFunc,
				},
			},
			Filter: "none",
		},
		{
			Name: "with nested depenencies and top level filter",
			Sweepers: map[string]*Sweeper{
				"not_matching": &Sweeper{
					Name: "not_matching",
					F:    mockSweeperFunc,
				},
				"matching_level1": &Sweeper{
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": &Sweeper{
					Name:         "matching_level2",
					Dependencies: []string{"matching_level3"},
					F:            mockSweeperFunc,
				},
				"matching_level3": &Sweeper{
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
				"not_matching": &Sweeper{
					Name: "not_matching",
					F:    mockSweeperFunc,
				},
				"matching_level1": &Sweeper{
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": &Sweeper{
					Name:         "matching_level2",
					Dependencies: []string{"matching_level3"},
					F:            mockSweeperFunc,
				},
				"matching_level3": &Sweeper{
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
				"not_matching": &Sweeper{
					Name: "not_matching",
					F:    mockSweeperFunc,
				},
				"matching_level1": &Sweeper{
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": &Sweeper{
					Name:         "matching_level2",
					Dependencies: []string{"matching_level3"},
					F:            mockSweeperFunc,
				},
				"matching_level3": &Sweeper{
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
			for k, _ := range actualSweepers {
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
				"matching_level1": &Sweeper{
					Name: "matching_level1",
					F:    mockSweeperFunc,
				},
				"non_matching": &Sweeper{
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
				"matching_level1": &Sweeper{
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": &Sweeper{
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
				"matching_level1": &Sweeper{
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2a", "matching_level2b"},
					F:            mockSweeperFunc,
				},
				"matching_level2a": &Sweeper{
					Name: "matching_level2a",
					F:    mockSweeperFunc,
				},
				"matching_level2b": &Sweeper{
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
				"matching_level1": &Sweeper{
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": &Sweeper{
					Name:         "matching_level2",
					Dependencies: []string{"matching_level3"},
					F:            mockSweeperFunc,
				},
				"matching_level3": &Sweeper{
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
				"matching_level1": &Sweeper{
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2a", "matching_level2b"},
					F:            mockSweeperFunc,
				},
				"matching_level2a": &Sweeper{
					Name:         "matching_level2a",
					Dependencies: []string{"matching_level3a", "matching_level3b"},
					F:            mockSweeperFunc,
				},
				"matching_level2b": &Sweeper{
					Name:         "matching_level2b",
					Dependencies: []string{"matching_level3c", "matching_level3d"},
					F:            mockSweeperFunc,
				},
				"matching_level3a": &Sweeper{
					Name: "matching_level3a",
					F:    mockSweeperFunc,
				},
				"matching_level3b": &Sweeper{
					Name: "matching_level3b",
					F:    mockSweeperFunc,
				},
				"matching_level3c": &Sweeper{
					Name: "matching_level3c",
					F:    mockSweeperFunc,
				},
				"matching_level3d": &Sweeper{
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
				"matching_level1": &Sweeper{
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": &Sweeper{
					Name:         "matching_level2",
					Dependencies: []string{"matching_level3"},
					F:            mockSweeperFunc,
				},
				"matching_level3": &Sweeper{
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
				"matching_level1": &Sweeper{
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2"},
					F:            mockSweeperFunc,
				},
				"matching_level2": &Sweeper{
					Name:         "matching_level2",
					Dependencies: []string{"matching_level3"},
					F:            mockSweeperFunc,
				},
				"matching_level3": &Sweeper{
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
				"matching_level1": &Sweeper{
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2a", "matching_level2c"},
					F:            mockSweeperFunc,
				},
				"matching_level2a": &Sweeper{
					Name: "matching_level2a",
					F:    mockSweeperFunc,
				},
				"matching_level2b": &Sweeper{
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
				"matching_level1": &Sweeper{
					Name:         "matching_level1",
					Dependencies: []string{"matching_level2a", "matching_level2b", "matching_level2c"},
					F:            mockSweeperFunc,
				},
				"matching_level2a": &Sweeper{
					Name:         "matching_level2a",
					Dependencies: []string{"matching_level3a", "matching_level3e"},
					F:            mockSweeperFunc,
				},
				"matching_level2b": &Sweeper{
					Name:         "matching_level2b",
					Dependencies: []string{"matching_level3c", "matching_level3f"},
					F:            mockSweeperFunc,
				},
				"matching_level3a": &Sweeper{
					Name: "matching_level3a",
					F:    mockSweeperFunc,
				},
				"matching_level3b": &Sweeper{
					Name: "matching_level3b",
					F:    mockSweeperFunc,
				},
				"matching_level3c": &Sweeper{
					Name: "matching_level3c",
					F:    mockSweeperFunc,
				},
				"matching_level3d": &Sweeper{
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
			for k, _ := range actualSweepers {
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
				"aws_dummy": &Sweeper{
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_dummy"},
		},
		{
			Name: "multiple",
			Sweepers: map[string]*Sweeper{
				"aws_one": &Sweeper{
					Name: "aws_one",
					F:    mockSweeperFunc,
				},
				"aws_two": &Sweeper{
					Name: "aws_two",
					F:    mockSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_one", "aws_two"},
		},
		{
			Name: "multiple with dep",
			Sweepers: map[string]*Sweeper{
				"aws_dummy": &Sweeper{
					Name: "aws_dummy",
					F:    mockSweeperFunc,
				},
				"aws_top": &Sweeper{
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": &Sweeper{
					Name: "aws_sub",
					F:    mockSweeperFunc,
				},
			},
			ExpectedRunList: []string{"aws_dummy", "aws_sub", "aws_top"},
		},
		{
			Name: "failing dep",
			Sweepers: map[string]*Sweeper{
				"aws_top": &Sweeper{
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": &Sweeper{
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
				"aws_top": &Sweeper{
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockSweeperFunc,
				},
				"aws_sub": &Sweeper{
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
				"aws_top": &Sweeper{
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockFailingSweeperFunc,
				},
				"aws_sub": &Sweeper{
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
				"aws_top": &Sweeper{
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockFailingSweeperFunc,
				},
				"aws_sub": &Sweeper{
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
				"aws_top": &Sweeper{
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockFailingSweeperFunc,
				},
				"aws_sub": &Sweeper{
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
				"aws_top": &Sweeper{
					Name:         "aws_top",
					Dependencies: []string{"aws_sub"},
					F:            mockFailingSweeperFunc,
				},
				"aws_sub": &Sweeper{
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
			for k, _ := range sweeperRunList["test"] {
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

func TestTestProviderResolver(t *testing.T) {
	stubProvider := func(name string) terraform.ResourceProvider {
		return &schema.Provider{
			Schema: map[string]*schema.Schema{
				name: &schema.Schema{
					Type:     schema.TypeString,
					Required: true,
				},
			},
		}
	}

	c := TestCase{
		ProviderFactories: map[string]terraform.ResourceProviderFactory{
			"foo": terraform.ResourceProviderFactoryFixed(stubProvider("foo")),
			"bar": terraform.ResourceProviderFactoryFixed(stubProvider("bar")),
		},
		Providers: map[string]terraform.ResourceProvider{
			"baz": stubProvider("baz"),
			"bop": stubProvider("bop"),
		},
	}

	resolver, err := testProviderResolver(c)
	if err != nil {
		t.Fatal(err)
	}

	reqd := discovery.PluginRequirements{
		"foo": &discovery.PluginConstraints{},
		"bar": &discovery.PluginConstraints{},
		"baz": &discovery.PluginConstraints{},
		"bop": &discovery.PluginConstraints{},
	}

	factories, errs := resolver.ResolveProviders(reqd)
	if len(errs) != 0 {
		for _, err := range errs {
			t.Error(err)
		}
		t.Fatal("unexpected errors")
	}

	for name := range reqd {
		t.Run(name, func(t *testing.T) {
			pf, ok := factories[name]
			if !ok {
				t.Fatalf("no factory for %q", name)
			}
			p, err := pf()
			if err != nil {
				t.Fatal(err)
			}
			resp := p.GetSchema()
			_, ok = resp.Provider.Block.Attributes[name]
			if !ok {
				var has string
				for k := range resp.Provider.Block.Attributes {
					has = k
					break
				}
				if has != "" {
					t.Errorf("provider %q does not have the expected schema attribute %q (but has %q)", name, name, has)
				} else {
					t.Errorf("provider %q does not have the expected schema attribute %q", name, name)
				}
			}
		})
	}
}

func TestCheckResourceAttr_empty(t *testing.T) {
	s := terraform.NewState()
	s.AddModuleState(&terraform.ModuleState{
		Path: []string{"root"},
		Resources: map[string]*terraform.ResourceState{
			"test_resource": &terraform.ResourceState{
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
			"test_resource": &terraform.ResourceState{
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
		state   *terraform.State
		wantErr string
	}{
		"exist match": {
			&terraform.State{
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
			``,
		},
		"nonexist match": {
			&terraform.State{
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
			``,
		},
		"exist nonmatch": {
			&terraform.State{
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
			`test.a: Attribute 'a' expected "boop", got "beep"`,
		},
		"inconsistent exist a": {
			&terraform.State{
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
			`test.a: Attribute "a" is "beep", but "b" is not set in test.b`,
		},
		"inconsistent exist b": {
			&terraform.State{
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
			`test.a: Attribute "a" not set, but "b" is set in test.b as "boop"`,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fn := TestCheckResourceAttrPair("test.a", "a", "test.b", "b")
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

func TestTestCheckResourceAttrPairCount(t *testing.T) {
	tests := map[string]struct {
		state   *terraform.State
		attr    string
		wantErr string
	}{
		"unset and 0 equal list": {
			&terraform.State{
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
			"a.#",
			``,
		},
		"unset and 0 equal map": {
			&terraform.State{
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
			"a.%",
			``,
		},
		"count equal": {
			&terraform.State{
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
			"a.%",
			``,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			fn := TestCheckResourceAttrPair("test.a", test.attr, "test.b", test.attr)
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

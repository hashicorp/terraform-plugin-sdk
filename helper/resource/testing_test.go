package resource

import (
	"errors"
	"flag"
	"fmt"
	"reflect"
	"sort"
	"testing"
)

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
			ExpectedSweepers: []string{"aws_top"},
			Filter:           "aws_top",
		},
		{
			Name: "filter and none",
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

package resource

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

// flagSweep is a flag available when running tests on the command line. It
// contains a comma seperated list of regions to for the sweeper functions to
// run in.  This flag bypasses the normal Test path and instead runs functions designed to
// clean up any leaked resources a testing environment could have created. It is
// a best effort attempt, and relies on Provider authors to implement "Sweeper"
// methods for resources.

// Adding Sweeper methods with AddTestSweepers will
// construct a list of sweeper funcs to be called here. We iterate through
// regions provided by the sweep flag, and for each region we iterate through the
// tests, and exit on any errors. At time of writing, sweepers are ran
// sequentially, however they can list dependencies to be ran first. We track
// the sweepers that have been ran, so as to not run a sweeper twice for a given
// region.
//
// WARNING:
// Sweepers are designed to be destructive. You should not use the -sweep flag
// in any environment that is not strictly a test environment. Resources will be
// destroyed.

var flagSweep = flag.String("sweep", "", "List of Regions to run available Sweepers")
var flagSweepAllowFailures = flag.Bool("sweep-allow-failures", false, "Enable to allow Sweeper Tests to continue after failures")
var flagSweepRun = flag.String("sweep-run", "", "Comma seperated list of Sweeper Tests to run")
var sweeperFuncs map[string]*Sweeper

// type SweeperFunc is a signature for a function that acts as a sweeper. It
// accepts a string for the region that the sweeper is to be ran in. This
// function must be able to construct a valid client for that region.
type SweeperFunc func(r string) error

type Sweeper struct {
	// Name for sweeper. Must be unique to be ran by the Sweeper Runner
	Name string

	// Dependencies list the const names of other Sweeper functions that must be ran
	// prior to running this Sweeper. This is an ordered list that will be invoked
	// recursively at the helper/resource level
	Dependencies []string

	// Sweeper function that when invoked sweeps the Provider of specific
	// resources
	F SweeperFunc
}

func init() {
	sweeperFuncs = make(map[string]*Sweeper)
}

// AddTestSweepers function adds a given name and Sweeper configuration
// pair to the internal sweeperFuncs map. Invoke this function to register a
// resource sweeper to be available for running when the -sweep flag is used
// with `go test`. Sweeper names must be unique to help ensure a given sweeper
// is only ran once per run.
func AddTestSweepers(name string, s *Sweeper) {
	if _, ok := sweeperFuncs[name]; ok {
		log.Fatalf("[ERR] Error adding (%s) to sweeperFuncs: function already exists in map", name)
	}

	sweeperFuncs[name] = s
}

func TestMain(m *testing.M) {
	flag.Parse()
	if *flagSweep != "" {
		// parse flagSweep contents for regions to run
		regions := strings.Split(*flagSweep, ",")

		// get filtered list of sweepers to run based on sweep-run flag
		sweepers := filterSweepers(*flagSweepRun, sweeperFuncs)

		if _, err := runSweepers(regions, sweepers, *flagSweepAllowFailures); err != nil {
			os.Exit(1)
		}
	} else {
		os.Exit(m.Run())
	}
}

func runSweepers(regions []string, sweepers map[string]*Sweeper, allowFailures bool) (map[string]map[string]error, error) {
	var sweeperErrorFound bool
	sweeperRunList := make(map[string]map[string]error)

	for _, region := range regions {
		region = strings.TrimSpace(region)

		var regionSweeperErrorFound bool
		regionSweeperRunList := make(map[string]error)

		log.Printf("[DEBUG] Running Sweepers for region (%s):\n", region)
		for _, sweeper := range sweepers {
			if err := runSweeperWithRegion(region, sweeper, sweepers, regionSweeperRunList, allowFailures); err != nil {
				if allowFailures {
					continue
				}

				sweeperRunList[region] = regionSweeperRunList
				return sweeperRunList, fmt.Errorf("sweeper (%s) for region (%s) failed: %s", sweeper.Name, region, err)
			}
		}

		log.Printf("Sweeper Tests ran successfully:\n")
		for sweeper, sweeperErr := range regionSweeperRunList {
			if sweeperErr == nil {
				fmt.Printf("\t- %s\n", sweeper)
			} else {
				regionSweeperErrorFound = true
			}
		}

		if regionSweeperErrorFound {
			sweeperErrorFound = true
			log.Printf("Sweeper Tests ran unsuccessfully:\n")
			for sweeper, sweeperErr := range regionSweeperRunList {
				if sweeperErr != nil {
					fmt.Printf("\t- %s: %s\n", sweeper, sweeperErr)
				}
			}
		}

		sweeperRunList[region] = regionSweeperRunList
	}

	if sweeperErrorFound {
		return sweeperRunList, errors.New("at least one sweeper failed")
	}

	return sweeperRunList, nil
}

// filterSweepers takes a comma seperated string listing the names of sweepers
// to be ran, and returns a filtered set from the list of all of sweepers to
// run based on the names given.
func filterSweepers(f string, source map[string]*Sweeper) map[string]*Sweeper {
	filterSlice := strings.Split(strings.ToLower(f), ",")
	if len(filterSlice) == 1 && filterSlice[0] == "" {
		// if the filter slice is a single element of "" then no sweeper list was
		// given, so just return the full list
		return source
	}

	sweepers := make(map[string]*Sweeper)
	for name, sweeper := range source {
		for _, s := range filterSlice {
			if strings.Contains(strings.ToLower(name), s) {
				sweepers[name] = sweeper
			}
		}
	}
	return sweepers
}

// runSweeperWithRegion recieves a sweeper and a region, and recursively calls
// itself with that region for every dependency found for that sweeper. If there
// are no dependencies, invoke the contained sweeper fun with the region, and
// add the success/fail status to the sweeperRunList.
func runSweeperWithRegion(region string, s *Sweeper, sweepers map[string]*Sweeper, sweeperRunList map[string]error, allowFailures bool) error {
	for _, dep := range s.Dependencies {
		if depSweeper, ok := sweepers[dep]; ok {
			log.Printf("[DEBUG] Sweeper (%s) has dependency (%s), running..", s.Name, dep)
			err := runSweeperWithRegion(region, depSweeper, sweepers, sweeperRunList, allowFailures)

			if err != nil {
				if allowFailures {
					log.Printf("[ERROR] Error running Sweeper (%s) in region (%s): %s", depSweeper.Name, region, err)
					continue
				}

				return err
			}
		} else {
			log.Printf("[DEBUG] Sweeper (%s) has dependency (%s), but that sweeper was not found", s.Name, dep)
		}
	}

	if _, ok := sweeperRunList[s.Name]; ok {
		log.Printf("[DEBUG] Sweeper (%s) already ran in region (%s)", s.Name, region)
		return nil
	}

	log.Printf("[DEBUG] Running Sweeper (%s) in region (%s)", s.Name, region)

	runE := s.F(region)

	sweeperRunList[s.Name] = runE

	if runE != nil {
		log.Printf("[ERROR] Error running Sweeper (%s) in region (%s): %s", s.Name, region, runE)
	}

	return runE
}

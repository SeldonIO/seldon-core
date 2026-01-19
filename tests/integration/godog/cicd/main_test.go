package cicd__test

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/seldonio/seldon-core/tests/integration/godog/suite"
	"github.com/spf13/pflag"
)

const cmdOptPrefix = "godog."

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty",
}

func init() {
	godog.BindCommandLineFlags(cmdOptPrefix, &opts)
}

type suiteCfg struct {
	Name        string
	Path        string
	Tags        string
	Concurrency int
}

func runOne(name string, o godog.Options) int {
	s := godog.TestSuite{
		Name:                 name,
		TestSuiteInitializer: suite.InitializeTestSuite,
		ScenarioInitializer:  suite.InitializeScenario,
		Options:              &o,
	}
	return s.Run()
}

func discoverFeatureSuites(featuresRoot string, concurrency int) ([]suiteCfg, error) {
	entries, err := os.ReadDir(featuresRoot)
	if err != nil {
		return nil, fmt.Errorf("read features root %q: %w", featuresRoot, err)
	}

	// Collect subdirectories (each becomes a "suite group")
	var dirs []string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		name := e.Name()

		// Skipping server dir
		if strings.HasPrefix(name, ".") || name == "server" {
			continue
		}
		dirs = append(dirs, name)
	}

	sort.Strings(dirs) // stable ordering across machines

	var suites []suiteCfg
	for _, d := range dirs {
		path := filepath.Join(featuresRoot, d)

		// Setup then run
		suites = append(suites,
			suiteCfg{Name: d + "-setup", Path: path, Tags: "@ServerSetup", Concurrency: concurrency},
			suiteCfg{Name: d + "-run", Path: path, Tags: "~@ServerSetup", Concurrency: concurrency},
		)
	}

	return suites, nil
}

func TestMain(m *testing.M) {
	flagSet := pflag.CommandLine
	flagSet.StringSliceVar(&opts.Paths, fmt.Sprintf("%s%s", cmdOptPrefix, "paths"), []string{}, "paths to feature files")
	pflag.Parse()

	custom := pflag.CommandLine.Changed(cmdOptPrefix+"paths") || pflag.CommandLine.Changed(cmdOptPrefix+"tags")

	godogStatus := 0

	if custom {
		// Custom: run exactly what the user requested
		godogStatus = runOne("godog-normal-run", opts)
	} else {
		// Aggregated: auto-discover feature suites
		featuresRoot := "../features"
		concurrency := 3 // or make this a flag/env if you want

		suites, err := discoverFeatureSuites(featuresRoot, concurrency)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to discover suites: %v\n", err)
			os.Exit(1)
		}

		failed := 0
		for _, s := range suites {
			o := opts
			o.Paths = []string{s.Path}
			o.Tags = s.Tags
			o.Concurrency = s.Concurrency

			fmt.Printf("\n=== Running Godog suite: %s path=%s tags=%q ===\n", s.Name, s.Path, s.Tags)
			if st := runOne(s.Name, o); st != 0 {
				failed++
			}
		}

		fmt.Printf("\n=== Godog overall: %d/%d suites failed ===\n", failed, len(suites))
		if failed > 0 {
			godogStatus = 1
		}
	}

	// Run any regular Go tests in this package too
	testStatus := m.Run()
	if testStatus > godogStatus {
		godogStatus = testStatus
	}

	os.Exit(godogStatus)
}

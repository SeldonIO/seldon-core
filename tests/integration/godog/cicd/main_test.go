package cicd__test

import (
	"fmt"
	"os"
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

// TestMain todo: this is the future version of running all test cases for the test suite with defined suites and config
// todo: in the future we could read the file from config instead of having it defined
// todo: we need to add custom stats for the end summarizing all suits run and have possible retries in the test suite
// At the moment if it receives any flags such as paths or tags it runs only one test suite
func TestMain(m *testing.M) {
	// Ensure --godog.paths works even if your godog version doesn't bind it automatically
	pflag.CommandLine.StringSliceVar(
		&opts.Paths,
		fmt.Sprintf("%s%s", cmdOptPrefix, "paths"),
		[]string{},
		"paths to feature files",
	)

	// Parse flags once (godog binds into pflag)
	pflag.Parse()

	// Decide mode:
	// - If user explicitly provided --godog.paths OR --godog.tags, treat it as a custom run.
	//   (You can expand this condition if you want other flags to trigger “custom mode”.)
	custom := pflag.CommandLine.Changed(cmdOptPrefix+"paths") || pflag.CommandLine.Changed(cmdOptPrefix+"tags")

	godogStatus := 0

	if custom {
		// Custom: run exactly what the user requested (one suite)
		godogStatus = runOne("godog-normal-run", opts)
	} else {
		// Aggregated default: run all suites (Option A)
		suites := []suiteCfg{
			{Name: "model-setup", Path: "../features/model", Tags: "@ServerSetup", Concurrency: 3},
			{Name: "model-run", Path: "../features/model", Tags: "~@ServerSetup", Concurrency: 3},

			{Name: "autoscaling-setup", Path: "../features/autoscaling", Tags: "@ServerSetup", Concurrency: 3},
			{Name: "autoscaling-run", Path: "../features/autoscaling", Tags: "~@ServerSetup", Concurrency: 3},

			{Name: "experiment-setup", Path: "../features/experiment", Tags: "@ServerSetup", Concurrency: 3},
			{Name: "experiment-run", Path: "../features/experiment", Tags: "~@ServerSetup", Concurrency: 3},

			{Name: "pipeline-setup", Path: "../features/pipeline", Tags: "@ServerSetup", Concurrency: 3},
			{Name: "pipeline-run", Path: "../features/pipeline", Tags: "~@ServerSetup", Concurrency: 3},
		}

		failed := 0
		for _, s := range suites {
			o := opts // copy base opts (format/output/etc)
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

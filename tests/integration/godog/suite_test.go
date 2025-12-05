package main

import (
	"context"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/seldonio/seldon-core/godog/steps"
	"github.com/spf13/pflag"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
	// world struct will contain all dependencies needed to run scenarios

	var world *steps.World
	//todo: init world
	world = steps.NewWorld("seldon-mesh")

	// Before prep the state of world however clients need to be long live could delete all crs that match a test label before running a test
	ctx.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
		return ctx, nil
	})

	// After needs to return the cluster state if there were any changes to the cluster state such as
	// bringing back deployments to starting state and deleting models and servers, could also be a flag if its done on before step
	ctx.After(func(ctx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
		return ctx, nil
	})

	steps.LoadDomSteps(ctx, world)
	//todo load other steps such as pipeline, experiment steps etc
}

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "progress", // can define default values
}

func init() {
	godog.BindCommandLineFlags("godog.", &opts) // godog v0.11.0 and later
}

// todo: the execution of the test will need further bootstrap of dependencies and clients as well as retrieving config
func TestMain(m *testing.M) {
	pflag.Parse()
	opts.Paths = pflag.Args()

	status := godog.TestSuite{
		Name:                "seldon-godog",
		ScenarioInitializer: InitializeScenario,
		Options:             &opts,
	}.Run()

	if st := m.Run(); st > status {
		status = st
	}

	os.Exit(status)
}

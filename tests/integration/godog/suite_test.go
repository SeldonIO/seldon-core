package suite

import (
	"context"
	"testing"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/godog/steps"
)

func InitializeScenario(ctx *godog.ScenarioContext) {
	// world struct will contain all dependencies needed to run scenarios

	var world *steps.World
	//todo: init world
	world = &steps.World{}

	// Before prep the state of world however clients need to be long live could delete all crs that match a test label before running a test
	ctx.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
		return ctx, nil
	})

	// After needs to return the cluster state if there were any changes to the cluster state such as
	// bringing back deployments to starting state and deleting models and servers, could also be a flag if its done on before step
	ctx.After(func(ctx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
		return ctx, nil
	})

	steps.LoadModelSteps(ctx, world)
	//todo load other steps such as pipeline, experiment steps etc
}

// todo: the execution of the test will need further bootstrap of dependencies and clients as well as retrieving config
func TestFeatures(t *testing.T) {
	format := "progress"
	if testing.Verbose() {
		format = "pretty"
	}

	opts := &godog.Options{
		Format:   format,
		Paths:    []string{"features"},
		TestingT: t,
	}

	suite := godog.TestSuite{
		Name:                "seldon-godog",
		ScenarioInitializer: InitializeScenario,
		Options:             opts,
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

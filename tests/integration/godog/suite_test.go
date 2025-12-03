package suite

import (
	"context"
	"fmt"
	"testing"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/godog/k8sclient"
	"github.com/seldonio/seldon-core/godog/steps"
)

func InitializeScenario(k8sClient *k8sclient.K8sClient) func(*godog.ScenarioContext) {
	return func(scenarioCtx *godog.ScenarioContext) {
		// Create the world with long-lived deps once per scenario context
		world := &steps.World{
			KubeClient: k8sClient,
			// initialise any other long-lived deps here, e.g. loggers, config, etc.
			Models: make(map[string]*steps.Model),
		}

		// Before: reset state and prep cluster before each scenario
		scenarioCtx.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
			if err := world.KubeClient.DeleteGodogTestModels(); err != nil {
				return ctx, fmt.Errorf("error when deleting models on before steps: %w", err)
			}

			// Reset scenario-level state
			world.CurrentModel = steps.NewModel()
			world.Models = make(map[string]*steps.Model)

			return ctx, nil
		})

		// After: optional cleanup / rollback
		scenarioCtx.After(func(ctx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
			// e.g. clean up any resources if needed
			// if cleanupErr := world.KubeClient.DeleteGodogTestModels(); cleanupErr != nil && err == nil {
			//     err = cleanupErr
			// }
			return ctx, err
		})

		// Register step definitions with access to world + k8sClient
		steps.LoadModelSteps(scenarioCtx, world)
		// TODO: load other steps, e.g. pipeline, experiment, etc.
	}
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

	k8sClient, err := k8sclient.New("seldon-mesh")
	if err != nil {
		t.Fatal(err.Error())
	}

	suite := godog.TestSuite{
		Name:                "seldon-godog",
		ScenarioInitializer: InitializeScenario(k8sClient),
		Options:             opts,
	}

	if suite.Run() != 0 {
		t.Fatal("non-zero status returned, failed to run feature tests")
	}
}

package scenario

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/godog/k8sclient"
	"github.com/seldonio/seldon-core/godog/steps"
)

type SuiteDeps struct {
	K8sClient    *k8sclient.K8sClient
	WatcherStore k8sclient.WatcherStorage
}

// might have to pass the suit struct and other config with closures to avoid having global vars
// todo: do this once the config and overall layout is better defined
//
//	func () {
//		status := godog.TestSuite{
//			Name: "godogs",
//			TestSuiteInitializer: func(ts *godog.TestSuiteContext) {
//				scenario.InitializeTestSuite(ts, deps)
//			},
//			ScenarioInitializer: func(sc *godog.ScenarioContext) {
//				scenario.InitializeScenario(sc, deps)
//			},
//			Options: &opts,
//		}.Run()
//	}
var suiteDeps SuiteDeps

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	// Create long-lived deps here
	k8sClient, err := k8sclient.New("seldon-mesh")
	if err != nil {
		// decide how hard you want to fail here:
		panic(fmt.Errorf("failed to create k8s client: %w", err))
	}

	suiteDeps.K8sClient = k8sClient

	ctx.BeforeSuite(func() {
		// e.g. create namespace, apply CRDs, etc.
	})

	ctx.AfterSuite(func() {
		// e.g. clean namespace, close clients if needed
	})
}

func InitializeScenario(scenarioCtx *godog.ScenarioContext) {
	// Create the world with long-lived deps once per scenario context
	world := &steps.World{
		KubeClient: suiteDeps.K8sClient,
		// initialise any other long-lived deps here, e.g. loggers, config, etc.
	}

	world.CurrentModel = steps.NewModel(world)

	// Before: reset state and prep cluster before each scenario
	scenarioCtx.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
		if err := world.KubeClient.DeleteGodogTestModels(ctx); err != nil {
			return ctx, fmt.Errorf("error when deleting models on before steps: %w", err)
		}

		// Create a fresh model for THIS scenario
		world.CurrentModel.Reset(world)

		// Reset scenario-level state

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
	steps.LoadModelSteps(scenarioCtx, world.CurrentModel)
	// TODO: load other steps, e.g. pipeline, experiment, etc.

}

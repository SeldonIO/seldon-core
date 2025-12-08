package scenario

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/godog/k8sclient"
	"github.com/seldonio/seldon-core/godog/steps"
	"github.com/sirupsen/logrus"
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

// todo: think about how we can drive server config from a file
// - have a default server for test
// - but also have the posibility of specifying the servers deployed in the test suite

// we create a server config or multiple servers
// server 1 caps mlserver
// default modelCapDeployment =  mlserver

// we can set the default model capabilities for models

// if we have an scenario and it doesn't specify the server capabilities that it is deployed to
// it will deploy into the default server capability that might be in our server dpeloyment def

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	// Create long-lived deps here
	k8sClient, err := k8sclient.New("seldon-mesh")
	if err != nil {
		panic(fmt.Errorf("failed to create k8s client: %w", err))
	}

	watchStore, err := k8sclient.NewWatcherStore("seldon-mesh", k8sclient.CRDLabels, k8sClient.KubeClient)
	if err != nil {
		panic(fmt.Errorf("failed to create k8s watch store: %w", err))
	}

	suiteDeps.K8sClient = k8sClient
	suiteDeps.WatcherStore = watchStore

	ctx.BeforeSuite(func() {
		suiteDeps.WatcherStore.Start()
		// e.g. create namespace, apply CRDs, etc.
	})

	ctx.AfterSuite(func() {
		suiteDeps.WatcherStore.Stop()
		// e.g. clean namespace, close clients if needed
	})
}

func InitializeScenario(scenarioCtx *godog.ScenarioContext) {
	// Create the world with long-lived deps once per scenario context
	world := steps.NewWorld(steps.Config{
		Namespace:      "seldon-mesh", //TODO configurable
		Logger:         logrus.New().WithField("test_type", "godog"),
		KubeClient:     suiteDeps.K8sClient,
		WatcherStorage: suiteDeps.WatcherStore,
		IngressHost:    "localhost", //TODO configurable
		HTTPPort:       9000,        //TODO configurable
		GRPCPort:       9000,        //TODO configurable
	})

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

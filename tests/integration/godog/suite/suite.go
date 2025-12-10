/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package suite

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	v "github.com/seldonio/seldon-core/operator/v2/pkg/generated/clientset/versioned"
	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	"github.com/seldonio/seldon-core/tests/integration/godog/steps"
	"github.com/sirupsen/logrus"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

type SuiteDeps struct {
	k8sClient    *k8sclient.K8sClient
	mlopsClient  *v.Clientset
	watcherStore k8sclient.WatcherStorage
	Config       *GodogConfig
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
	// todo: we should bootstrap config here
	// Load configuration from JSON file
	config, err := LoadConfig()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	// Create long-lived deps here
	k8sClient, err := k8sclient.New(config.Namespace)
	if err != nil {
		panic(fmt.Errorf("failed to create k8s client: %w", err))
	}

	clientSet, err := v.NewForConfig(controllerruntime.GetConfigOrDie())
	if err != nil {
		panic(fmt.Errorf("failed to mlops client: %w", err))
	}

	watchStore, err := k8sclient.NewWatcherStore(config.Namespace, k8sclient.DefaultCRDLabel, clientSet.MlopsV1alpha1())
	if err != nil {
		panic(fmt.Errorf("failed to create k8s watch store: %w", err))
	}

	suiteDeps.k8sClient = k8sClient
	suiteDeps.mlopsClient = clientSet // todo: this clientSet might get use for get requests or for the mlops interface and could be passed to the world might be split up by type
	suiteDeps.watcherStore = watchStore
	suiteDeps.Config = config

	ctx.BeforeSuite(func() {
		suiteDeps.watcherStore.Start()
		// e.g. create namespace, apply CRDs, etc.
	})

	ctx.AfterSuite(func() {
		suiteDeps.watcherStore.Stop()
		// e.g. clean namespace, close clients if needed
	})
}

func InitializeScenario(scenarioCtx *godog.ScenarioContext) {
	log := logrus.New()
	if suiteDeps.Config.LogLevel != "" {
		logLevel, err := logrus.ParseLevel(suiteDeps.Config.LogLevel)
		if err != nil {
			panic(fmt.Errorf("failed to parse log level %s: %w", logLevel, err))
		}
		log.SetLevel(logLevel)
	}

	// Create the world with long-lived deps once per scenario context
	world, err := steps.NewWorld(steps.Config{
		Namespace:      suiteDeps.Config.Namespace,
		Logger:         log.WithField("test_type", "godog"),
		KubeClient:     suiteDeps.k8sClient,
		WatcherStorage: suiteDeps.watcherStore,
		IngressHost:    suiteDeps.Config.Inference.Host,
		HTTPPort:       suiteDeps.Config.Inference.HTTPPort,
		GRPCPort:       suiteDeps.Config.Inference.GRPCPort,
		SSL:            suiteDeps.Config.Inference.SSL,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create world: %w", err))
	}

	world.CurrentModel = steps.NewModel()

	// Before: reset state and prep cluster before each scenario
	scenarioCtx.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
		//world.CurrentModel.Reset(world)

		// Reset scenario-level state

		return ctx, nil
	})

	// After: optional cleanup / rollback
	scenarioCtx.After(func(ctx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
		if suiteDeps.Config.SkipCleanup {
			log.WithField("scenario", scenario.Name).Debug("Skipping cleanup")
			return ctx, nil
		}

		if err := world.KubeClient.DeleteScenarioResources(ctx, world.Label); err != nil {
			return ctx, fmt.Errorf("error when deleting models on before steps: %w", err)
		}

		return ctx, err
	})

	// Register step definitions with access to world + k8sClient
	steps.LoadModelSteps(scenarioCtx, world)
	steps.LoadExplicitModelSteps(scenarioCtx, world)
	steps.LoadInferenceSteps(scenarioCtx, world)
	// TODO: load other steps, e.g. pipeline, experiment, etc.
}

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
	"crypto/tls"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	v2dataplane "github.com/seldonio/seldon-core/apis/go/v2/mlops/v2_dataplane"
	v "github.com/seldonio/seldon-core/operator/v2/pkg/generated/clientset/versioned"
	"github.com/seldonio/seldon-core/tests/integration/godog/components"
	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	"github.com/seldonio/seldon-core/tests/integration/godog/steps"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	controllerruntime "sigs.k8s.io/controller-runtime"
)

type dependencies struct {
	logger          *logrus.Logger
	k8sClient       *k8sclient.K8sClient
	mlopsClient     *v.Clientset
	watcherStore    k8sclient.WatcherStorage
	inferenceClient v2dataplane.GRPCInferenceServiceClient
	infraManager    *components.EnvManager
	Config          *GodogConfig
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
var suiteDeps dependencies

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	realContext, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// Load configuration from JSON file
	config, err := LoadConfig()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	log := logrus.New()
	if config.LogLevel != "" {
		logLevel, err := logrus.ParseLevel(config.LogLevel)
		if err != nil {
			panic(fmt.Errorf("failed to parse log level %s: %w", logLevel, err))
		}
		log.SetLevel(logLevel)
	}

	// Create long-lived deps here
	k8sClient, err := k8sclient.New(config.Namespace, log)
	if err != nil {
		panic(fmt.Errorf("failed to create k8s client: %w", err))
	}

	clientSet, err := v.NewForConfig(controllerruntime.GetConfigOrDie())
	if err != nil {
		panic(fmt.Errorf("failed to mlops client: %w", err))
	}

	watchStore, err := k8sclient.NewWatcherStore(config.Namespace, k8sclient.DefaultCRDTestSuiteLabel, clientSet.MlopsV1alpha1(), log)
	if err != nil {
		panic(fmt.Errorf("failed to create k8s watch store: %w", err))
	}

	creds := insecure.NewCredentials()
	if config.Inference.SSL {
		creds = credentials.NewTLS(&tls.Config{
			InsecureSkipVerify: true,
		})
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	conn, err := grpc.NewClient(fmt.Sprintf("%s:%d", config.Inference.Host, config.Inference.GRPCPort), opts...)
	if err != nil {
		panic(fmt.Errorf("could not create grpc client: %w", err))
	}
	grpcClient := v2dataplane.NewGRPCInferenceServiceClient(conn)

	infraManger, err := components.StartComponents(k8sClient, config.Namespace)
	if err != nil {
		panic(err)
	}

	// Snapshot baseline state once per suite
	if err := infraManger.SnapshotAll(realContext); err != nil {
		panic(fmt.Errorf("failed to snapshot environment: %w", err))
	}

	suiteDeps.logger = log
	suiteDeps.k8sClient = k8sClient
	suiteDeps.mlopsClient = clientSet // todo: this clientSet might get use for get requests or for the mlops interface and could be passed to the world might be split up by type
	suiteDeps.watcherStore = watchStore
	suiteDeps.Config = config
	suiteDeps.inferenceClient = grpcClient
	suiteDeps.infraManager = infraManger

	ctx.BeforeSuite(func() {
		suiteDeps.watcherStore.Start()
		if err := suiteDeps.k8sClient.DeleteScenarioResources(context.Background(), k8sclient.DefaultCRDTestSuiteLabelMap); err != nil {
			suiteDeps.logger.Errorf("error when deleting models on before steps: %v", err)
		}
		// e.g. create namespace, apply CRDs, etc.
	})

	ctx.AfterSuite(func() {
		suiteDeps.watcherStore.Stop()

		// e.g. clean namespace, close clients if needed delete servers
	})
}

func InitializeScenario(scenarioCtx *godog.ScenarioContext) {
	// Create the world with long-lived deps once per scenario context
	log := suiteDeps.logger.WithField("test_type", "godog")
	world, err := steps.NewWorld(steps.Config{
		Namespace:      suiteDeps.Config.Namespace,
		Logger:         log,
		KubeClient:     suiteDeps.k8sClient,
		K8sClient:      suiteDeps.mlopsClient,
		WatcherStorage: suiteDeps.watcherStore,
		Env:            suiteDeps.infraManager,
		IngressHost:    suiteDeps.Config.Inference.Host,
		HTTPPort:       suiteDeps.Config.Inference.HTTPPort,
		GRPCPort:       suiteDeps.Config.Inference.GRPCPort,
		SSL:            suiteDeps.Config.Inference.SSL,
		GRPC:           suiteDeps.inferenceClient,
	})
	if err != nil {
		panic(fmt.Errorf("failed to create world: %w", err))
	}

	var slowForThis bool
	// Before: reset state and prep cluster before each scenario
	scenarioCtx.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
		// Enable slow mode only for @slow scenarios recommended only for local testing
		slowForThis = false
		for _, t := range scenario.Tags {
			if t.Name == "@slow" {
				slowForThis = true
				log.WithField("scenario", scenario.Name).Debugf("set to run slow scenario at %s per step", time.Duration(suiteDeps.Config.ScenarioStepDelay))
				break
			}
		}
		return ctx, nil
	})

	// After: optional cleanup / rollback
	scenarioCtx.After(func(ctx context.Context, scenario *godog.Scenario, err error) (context.Context, error) {
		err = suiteDeps.infraManager.RestoreAll(context.Background())
		if err != nil {
			return ctx, err
		}
		if err != nil && suiteDeps.Config.SkipCleanUpOnError {
			log.WithField("scenario", scenario.Name).Debugf("Skipping cleanup of resources for scenario with err %v", err)
			// don't clean up resources for scenarios that fail
			return ctx, nil
		}
		if suiteDeps.Config.SkipCleanup {
			log.WithField("scenario", scenario.Name).Debug("Skipping cleanup")
			return ctx, nil
		}

		if err := suiteDeps.k8sClient.DeleteScenarioResources(ctx, world.Label); err != nil {
			return ctx, fmt.Errorf("error when deleting models on after scenario steps: %w", err)
		}

		return ctx, nil
	})

	// --- Step hook: delay after every step ---
	scenarioCtx.StepContext().After(func(ctx context.Context, st *godog.Step, status godog.StepResultStatus, err error) (context.Context, error) {
		// set delay for scenario steps when @slow tag is present
		if slowForThis {
			time.Sleep(time.Duration(suiteDeps.Config.ScenarioStepDelay))
		}
		return ctx, nil
	})

	// Register step definitions with access to world + k8sClient
	steps.LoadTemplateModelSteps(scenarioCtx, world)
	steps.LoadCustomModelSteps(scenarioCtx, world)
	steps.LoadInferenceSteps(scenarioCtx, world)
	steps.LoadServerSteps(scenarioCtx, world)
	steps.LoadCustomPipelineSteps(scenarioCtx, world)
	steps.LoadExperimentSteps(scenarioCtx, world)
	steps.LoadUtilSteps(scenarioCtx, world)
	steps.LoadInfrastructureSteps(scenarioCtx, world)
}

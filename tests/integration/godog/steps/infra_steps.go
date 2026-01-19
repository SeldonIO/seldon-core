/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/tests/integration/godog/components"
	"github.com/sirupsen/logrus"
)

// Infrastructure todo: add attributes as we need
type Infrastructure struct {
	env *components.EnvManager
	log logrus.FieldLogger
}

func newInfrastructure(env *components.EnvManager, log logrus.FieldLogger) *Infrastructure {
	return &Infrastructure{env: env, log: log}
}

func LoadInfrastructureSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^(kafka-nodepool) is unavailable for Core 2 with timeout "([^"]+)"`, func(kind, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			switch kind {
			case "kafka-nodepool":
				k, err := w.env.Component(components.KafkaNodePool)
				if err != nil {
					return err
				}

				return k.MakeUnavailable(ctx)
			default:
				return fmt.Errorf("unknown target type: %s", kind)
			}
		})
	})

	scenario.Step(`^(kafka-nodepool) is available for Core 2 with timeout "([^"]+)"`, func(kind, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			switch kind {
			case "kafka-nodepool":
				k, err := w.env.Component(components.KafkaNodePool)
				if err != nil {
					return err
				}

				return k.MakeAvailable(ctx)
			default:
				return fmt.Errorf("unknown target type: %s", kind)
			}
		})
	})
	scenario.Step(`^I restart (scheduler|dataflow-engine|model-gw|pipeline-gw) with timeout "([^"]+)"$`,
		func(kind, timeout string) error {
			return withTimeoutCtx(timeout, func(ctx context.Context) error {
				runtime := w.env.Runtime()
				if runtime == nil {
					return fmt.Errorf("runtime not defined")
				}

				svc, err := w.infra.runtimeServiceFromStepKind(kind)
				if err != nil {
					return err
				}

				return runtime.RestartService(ctx, svc)
			})
		},
	)
	scenario.Step(`^I set replicas of (scheduler|dataflow-engine|model-gw|pipeline-gw) to "([^"]+)" with timeout "([^"]+)"$`,
		func(kind string, replicaCount int32, timeout string) error {
			return withTimeoutCtx(timeout, func(ctx context.Context) error {
				runtime := w.env.Runtime()
				if runtime == nil {
					return fmt.Errorf("runtime not defined")
				}

				svc, err := w.infra.runtimeServiceFromStepKind(kind)
				if err != nil {
					return err
				}

				return runtime.SetReplicas(ctx, svc, replicaCount)
			})
		},
	)
	scenario.Step(`^I wait for (scheduler|dataflow-engine|model-gw|pipeline-gw) to be ready with timeout "([^"]+)"$`,
		func(kind, timeout string) error {
			return withTimeoutCtx(timeout, func(ctx context.Context) error {
				runtime := w.env.Runtime()
				if runtime == nil {
					return fmt.Errorf("runtime not defined")
				}

				svc, err := w.infra.runtimeServiceFromStepKind(kind)
				if err != nil {
					return err
				}

				return runtime.WaitServiceReady(ctx, svc)
			})
		},
	)
	scenario.Step(`^I scale down (scheduler|dataflow-engine|model-gw|pipeline-gw) with timeout "([^"]+)"$`,
		func(kind, timeout string) error {
			return withTimeoutCtx(timeout, func(ctx context.Context) error {
				runtime := w.env.Runtime()
				if runtime == nil {
					return fmt.Errorf("runtime not defined")
				}

				svc, err := w.infra.runtimeServiceFromStepKind(kind)
				if err != nil {
					return err
				}

				return runtime.ScaleDown(ctx, svc)
			})
		},
	)
	scenario.Step(`^I scale up to baseline (scheduler|dataflow-engine|model-gw|pipeline-gw) with timeout "([^"]+)"$`,
		func(kind, timeout string) error {
			return withTimeoutCtx(timeout, func(ctx context.Context) error {
				runtime := w.env.Runtime()
				if runtime == nil {
					return fmt.Errorf("runtime not defined")
				}

				svc, err := w.infra.runtimeServiceFromStepKind(kind)
				if err != nil {
					return err
				}

				return runtime.ScaleUpToBaseline(ctx, svc)
			})
		},
	)
}

func (i *Infrastructure) runtimeServiceFromStepKind(kind string) (components.SeldonRuntimeService, error) {
	switch kind {
	case "scheduler":
		return components.ServiceScheduler, nil
	case "dataflow-engine":
		return components.ServiceDataflowEngine, nil
	case "model-gw":
		return components.ServiceModelGateway, nil
	case "pipeline-gw":
		return components.ServicePipelineGateway, nil
	default:
		return "", fmt.Errorf("unknown target type: %s", kind)
	}
}

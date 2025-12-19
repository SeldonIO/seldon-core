package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/tests/integration/godog/components"
)

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

				svc, err := runtimeServiceFromStepKind(kind)
				if err != nil {
					return err
				}

				return runtime.RestartService(ctx, svc)
			})
		},
	)
}

func runtimeServiceFromStepKind(kind string) (components.SeldonRuntimeService, error) {
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

package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/tests/integration/godog/components"
)

func LoadInfrastructureSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^(Kafka|kafka-nodepool) is unavailable for Core 2 with timeout "([^"]+)"`, func(kind, timeout string) error {
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

	scenario.Step(`^(Kafka|kafka-nodepool) is available for Core 2 with timeout "([^"]+)"`, func(kind, timeout string) error {
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
}

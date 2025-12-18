package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/tests/integration/godog/components"
)

type Env struct {
	env *components.EnvManager
}

func NewEnv(env *components.EnvManager) *Env {
	return &Env{env: env}
}

func LoadInfrastructureSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^(Kafka|kafka-nodepool) is unavailable for Core 2 with timeout "([^"]+)"`, func(kind, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			switch kind {
			case "Kafka":
				k := w.env.Kafka()
				if k != nil {
					return fmt.Errorf("kafka component not configured in env")
				}

				return k.MakeUnavailable(ctx)
			case "kafka-nodepool":
				k := w.env.KafkaNodePool()
				if k != nil {
					return fmt.Errorf("kafka node pool component not configured in env")
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
			case "Kafka":
				k := w.env.Kafka()
				if k != nil {
					return fmt.Errorf("kafka component not configured in env")
				}

				return k.MakeAvailable(ctx)
			case "kafka-nodepool":
				k := w.env.KafkaNodePool()
				if k != nil {
					return fmt.Errorf("kafka node pool component not configured in env")
				}

				return k.MakeAvailable(ctx)
			default:
				return fmt.Errorf("unknown target type: %s", kind)
			}
		})
	})
}

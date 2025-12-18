package steps

import (
	"context"
	"fmt"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
)

type Env struct {
	env *k8sclient.EnvManager
}

func LoadEnvSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^(Kafka) is unavailable for Core 2 with timeout "([^"]+)"`, func(kind, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			switch kind {
			case "Kafka":
				k := w.env.Kafka()
				if k != nil {
					return nil
				}

				return k.MakeUnavailable(ctx)
			default:
				return fmt.Errorf("unknown target type: %s", kind)
			}
		})
	})

	scenario.Step(`^(Kafka) is available for Core 2 with timeout "([^"]+)"`, func(kind, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			switch kind {
			case "Kafka":
				k := w.env.Kafka()
				if k != nil {
					return nil
				}

				return k.MakeAvailable(ctx)
			default:
				return fmt.Errorf("unknown target type: %s", kind)
			}
		})
	})
}

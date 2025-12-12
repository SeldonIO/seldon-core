package steps

import (
	"context"

	"github.com/cucumber/godog"
)

func LoadServerSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^I deploy server spec with timeout "([^"]+)":$`, func(timeout string, spec *godog.DocString) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.currentModel.deployModelSpec(ctx, spec)
		})
	})
}

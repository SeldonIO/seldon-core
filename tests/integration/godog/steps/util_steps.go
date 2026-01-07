package steps

import (
	"fmt"
	"time"

	"github.com/cucumber/godog"
)

func LoadUtilSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^I wait for "([^"]+)"`, func(wait string) error {
		d, err := time.ParseDuration(wait)
		if err != nil {
			return fmt.Errorf("invalid wait duration %s: %w", wait, err)
		}
		time.Sleep(d)
		return nil
	})
}

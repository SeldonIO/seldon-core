/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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

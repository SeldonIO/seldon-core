/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package main__test

import (
	"fmt"
	"os"
	"testing"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
	"github.com/seldonio/seldon-core/tests/integration/godog/suite"
	"github.com/spf13/pflag" // godog v0.11.0 and later
)

const cmdOptPrefix = "godog."

var opts = godog.Options{
	Output: colors.Colored(os.Stdout),
	Format: "pretty", // can define default values
}

func init() {
	godog.BindCommandLineFlags(cmdOptPrefix, &opts) // godog v0.11.0 and later
}

func TestMain(m *testing.M) {
	flagSet := pflag.CommandLine
	flagSet.StringSliceVar(&opts.Paths, fmt.Sprintf("%s%s", cmdOptPrefix, "paths"), []string{}, "paths to feature files")
	pflag.Parse()

	status := godog.TestSuite{
		Name:                 "godogs",
		TestSuiteInitializer: suite.InitializeTestSuite,
		ScenarioInitializer:  suite.InitializeScenario,
		Options:              &opts,
	}.Run()

	// Optional: Run `testing` package's logic besides godog.
	if st := m.Run(); st > status {
		status = st
	}

	os.Exit(status)
}

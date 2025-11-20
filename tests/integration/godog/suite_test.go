package godogtests

import (
	"testing"

	"github.com/cucumber/godog"
)

func TestFeatures(t *testing.T) {
	format := "progress"
	if testing.Verbose() {
		format = "pretty"
	}

	opts := &godog.Options{
		Format:   format,
		Paths:    []string{"features"},
		TestingT: t,
	}

	suite := godog.TestSuite{
		Name:                "seldon-godog",
		ScenarioInitializer: InitializeScenario,
		Options:             opts,
	}

	if suite.Run() != 0 {
		t.Fatal("godog scenarios failed")
	}
}

/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package state_machine

import log "github.com/sirupsen/logrus"

type StateMachine struct {
	Model      *ModelStateMachine
	Pipeline   *PipelineStateMachine
	Experiment *ExperimentStateMachine
}

// NewStateMachine todo: is this the best way to pass the logger?
// NewStateMachine creates a composite state machine with shared config
func NewStateMachine(config *Config, logger log.FieldLogger) *StateMachine {
	if config == nil {
		config = DefaultConfig()
	}

	logger = logger.WithField("component", "state_machine")
	return &StateMachine{
		Model:      NewModelStateMachine(config, logger),
		Pipeline:   NewPipelineStateMachine(config, logger),
		Experiment: NewExperimentStateMachine(config, logger),
	}
}

type ModelStateMachine struct {
	config *Config
	logger log.FieldLogger
}

type PipelineStateMachine struct {
	config *Config
	logger log.FieldLogger
}

type ExperimentStateMachine struct {
	config *Config
	logger log.FieldLogger
}

func NewModelStateMachine(config *Config, logger log.FieldLogger) *ModelStateMachine {
	if config == nil {
		config = DefaultConfig()
	}
	return &ModelStateMachine{config: config, logger: logger}
}

func NewPipelineStateMachine(config *Config, logger log.FieldLogger) *PipelineStateMachine {
	if config == nil {
		config = DefaultConfig()
	}
	return &PipelineStateMachine{config: config, logger: logger}
}

func NewExperimentStateMachine(config *Config, logger log.FieldLogger) *ExperimentStateMachine {
	if config == nil {
		config = DefaultConfig()
	}
	return &ExperimentStateMachine{config: config, logger: logger}
}

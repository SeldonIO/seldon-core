/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package state_machine

type StateMachine struct {
	Model      *ModelStateMachine
	Pipeline   *PipelineStateMachine
	Experiment *ExperimentStateMachine
}

// NewStateMachine creates a composite state machine with shared config
func NewStateMachine(config *Config) *StateMachine {
	if config == nil {
		config = DefaultConfig()
	}
	return &StateMachine{
		Model:      NewModelStateMachine(config),
		Pipeline:   NewPipelineStateMachine(config),
		Experiment: NewExperimentStateMachine(config),
	}
}

type ModelStateMachine struct {
	config *Config
}

type PipelineStateMachine struct {
	config *Config
}

type ExperimentStateMachine struct {
	config *Config
}

func NewModelStateMachine(config *Config) *ModelStateMachine {
	if config == nil {
		config = DefaultConfig()
	}
	return &ModelStateMachine{config: config}
}

func NewPipelineStateMachine(config *Config) *PipelineStateMachine {
	if config == nil {
		config = DefaultConfig()
	}
	return &PipelineStateMachine{config: config}
}

func NewExperimentStateMachine(config *Config) *ExperimentStateMachine {
	if config == nil {
		config = DefaultConfig()
	}
	return &ExperimentStateMachine{config: config}
}

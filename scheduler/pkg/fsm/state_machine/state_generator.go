/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package state_machine

// todo: state machine will need config options for calculating some state regarding servers
type StateMachine struct {
}

func NewStateMachine() *StateMachine {
	return &StateMachine{}
}

func NewModelStateGenerator() Model {
	return &StateMachine{}
}

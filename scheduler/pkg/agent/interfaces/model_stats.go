/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package interfaces

// TODO: use generics
// TODO: define more logic operations

type LogicOperation int

const (
	Gte LogicOperation = iota
)

type ModelStatsKV struct {
	Value     uint32
	Key       string
	ModelName string
}

type ModelStatsKeeper interface {
	ModelInferEnter(modelName, requestId string) error
	ModelInferExit(modelName, requestId string) error
	Add(modelName string) error
	Delete(modelName string) error
	Get(modelName string) (statValue uint32, err error)
	GetAll(threshold uint32, op LogicOperation, reset bool) ([]*ModelStatsKV, error)
}

type ModelStats interface {
	Enter(requestId string) error
	Exit(requestId string) error
	Get() uint32
	Reset() error
}

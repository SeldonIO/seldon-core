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

type ModelScalingStats interface {
	Inc(string, uint32) error
	IncDefault(string) error
	Dec(string, uint32) error
	DecDefault(string) error
	Reset(string) error
	Set(string, uint32) error
	Get(string) (uint32, error)
	GetAll(uint32, LogicOperation, bool) ([]*ModelStatsKV, error)
	Info() string
	Delete(string) error
	Add(string) error
}

/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
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

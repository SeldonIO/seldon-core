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

type CacheManager interface {
	// add a new node with specific id and priority/value
	Add(string, int64) error
	// add a new node with specific id and default priority/value
	AddDefault(string) error
	// update value for given id, which would reflect in order
	Update(id string, value int64) error
	// default bump value for given id, which would reflect in order
	UpdateDefault(string) error
	// check if value exists
	Exists(string) bool
	// get value/priority of given id
	Get(string) (int64, error)
	// delete item with id from cache
	Delete(id string) error
	// get a list of all keys / values
	GetItems() ([]string, []int64)
	// peek top of queue (no evict)
	Peek() (string, int64, error)
	// evict
	Evict() (string, int64, error)
}

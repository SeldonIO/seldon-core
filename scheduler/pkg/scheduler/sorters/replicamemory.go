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

package sorters

import (
	"math"
)

// Rationale - put models on replicas with more memory, including the models that are currently loading
// note that we can double count here as available memory (returned by the agent) could include memory
// that is allocated while the model is being loaded.
type AvailableMemorySorter struct{}

func (m AvailableMemorySorter) Name() string {
	return "AvailableMemorySorter"
}

func (m AvailableMemorySorter) IsLess(i *CandidateReplica, j *CandidateReplica) bool {
	iMem := math.Max(0, float64(i.Replica.GetAvailableMemory()-i.Replica.GetReservedMemory()))
	jMem := math.Max(0, float64(j.Replica.GetAvailableMemory()-j.Replica.GetReservedMemory()))
	return iMem > jMem
}

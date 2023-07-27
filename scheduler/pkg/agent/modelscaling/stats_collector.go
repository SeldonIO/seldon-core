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

package modelscaling

import (
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

type DataPlaneStatsCollector struct {
	StatKeepers []interfaces.ModelStatsKeeper
}

func NewDataPlaneStatsCollector(statKeepers []interfaces.ModelStatsKeeper) *DataPlaneStatsCollector {
	return &DataPlaneStatsCollector{
		StatKeepers: statKeepers,
	}
}

func (c *DataPlaneStatsCollector) ModelInferEnter(internalModelName, requestId string) error {
	var err error
	for _, stat := range c.StatKeepers {
		err = stat.ModelInferEnter(internalModelName, requestId)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *DataPlaneStatsCollector) ModelInferExit(internalModelName, requestId string) error {
	var err error
	for _, stat := range c.StatKeepers {
		err = stat.ModelInferExit(internalModelName, requestId)
		if err != nil {
			return err
		}
	}
	return nil
}

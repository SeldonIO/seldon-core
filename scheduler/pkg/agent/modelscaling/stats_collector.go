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
	"sync"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

// TODO: this has tight-coupling with specific metrics, how can we do it more generally?
type DataPlaneStatsCollector struct {
	ModelLagStats      interfaces.ModelScalingStats
	ModelLastUsedStats interfaces.ModelScalingStats
}

func NewDataPlaneStatsCollector(modelLagStats, modelLastUsedStats interfaces.ModelScalingStats) *DataPlaneStatsCollector {
	return &DataPlaneStatsCollector{
		ModelLagStats:      modelLagStats,
		ModelLastUsedStats: modelLastUsedStats,
	}
}

func (c *DataPlaneStatsCollector) ScalingMetricsSetup(wg *sync.WaitGroup, internalModelName string) error {

	err := c.ModelLagStats.IncDefault(internalModelName)
	wg.Done()
	if err != nil {
		return err
	}
	return c.ModelLastUsedStats.IncDefault(internalModelName)
}

func (c *DataPlaneStatsCollector) ScalingMetricsTearDown(wg *sync.WaitGroup, internalModelName string) error {
	wg.Wait() // make sure that Inc is called first
	return c.ModelLagStats.DecDefault(internalModelName)
}

/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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

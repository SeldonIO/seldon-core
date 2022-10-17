package modelscaling

import (
	"sync"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent/interfaces"
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

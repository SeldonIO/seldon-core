package modelscaling

import (
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

func TestStatsCollectorSmoke(t *testing.T) {
	g := NewGomegaWithT(t)
	dummyModel := "model_0"

	lags := NewModelReplicaLagsKeeper()
	lastUsed := NewModelReplicaLastUsedKeeper()

	collector := NewDataPlaneStatsCollector(lags, lastUsed)

	var wg sync.WaitGroup
	wg.Add(1)

	err := collector.ScalingMetricsSetup(&wg, dummyModel)
	g.Expect(err).To(BeNil())

	lagCount, _ := collector.ModelLagStats.Get(dummyModel)
	lastUsedCount, _ := collector.ModelLastUsedStats.Get(dummyModel)

	g.Expect(lagCount).To(Equal(uint32(1)))
	g.Expect(lastUsedCount).Should(BeNumerically("<=", time.Now().Unix()))

	err = collector.ScalingMetricsTearDown(&wg, dummyModel)
	g.Expect(err).To(BeNil())

	lagCount, _ = collector.ModelLagStats.Get(dummyModel)
	g.Expect(lagCount).To(Equal(uint32(0)))

}

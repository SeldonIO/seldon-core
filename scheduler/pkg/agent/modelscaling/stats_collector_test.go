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

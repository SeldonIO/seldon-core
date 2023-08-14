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
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

func TestStatsCollectorSmoke(t *testing.T) {
	g := NewGomegaWithT(t)
	dummyModel := "model_0"
	dummyRequestId := "request_id_0"

	lags := NewModelReplicaLagsKeeper()
	lastUsed := NewModelReplicaLastUsedKeeper()

	collector := NewDataPlaneStatsCollector([]interfaces.ModelStatsKeeper{lags, lastUsed}, nil)

	var wg sync.WaitGroup
	wg.Add(1)

	err := collector.ModelInferEnter(dummyModel, dummyRequestId)
	g.Expect(err).To(BeNil())

	lagCount, _ := lags.Get(dummyModel)
	lastUsedCount, _ := lastUsed.Get(dummyModel)

	g.Expect(lagCount).To(Equal(uint32(1)))
	g.Expect(lastUsedCount).Should(BeNumerically("<=", time.Now().Unix()))

	err = collector.ModelInferExit(dummyModel, dummyRequestId)
	g.Expect(err).To(BeNil())

	lagCount, _ = lags.Get(dummyModel)
	g.Expect(lagCount).To(Equal(uint32(0)))

}

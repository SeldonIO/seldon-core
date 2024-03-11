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
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	log "github.com/sirupsen/logrus"
)

func TestStatsCollectorSmoke(t *testing.T) {
	g := NewGomegaWithT(t)
	dummyModel := "model_0"
	dummyRequestId := "request_id_0"

	lags := NewModelReplicaLagsKeeper()
	lastUsed := NewModelReplicaLastUsedKeeper()

	collector := NewDataPlaneStatsCollector([]interfaces.ModelStatsKeeper{lags, lastUsed}, log.New())

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

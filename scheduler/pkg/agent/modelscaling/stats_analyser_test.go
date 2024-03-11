/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package modelscaling

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

const (
	statsPeriodSecondsDefault       = 5
	lagThresholdDefault             = 30
	lastUsedThresholdSecondsDefault = 30
)

func scalingMetricsSetup(
	internalModelName string, requestId string, modelScalingStatsCollector *DataPlaneStatsCollector) error {
	return modelScalingStatsCollector.ModelInferEnter(internalModelName, requestId)
}

func scalingMetricsTearDown(internalModelName string, requestId string,
	jobsWg *sync.WaitGroup, modelScalingStatsCollector *DataPlaneStatsCollector) error {
	err := modelScalingStatsCollector.ModelInferExit(internalModelName, requestId)
	jobsWg.Done()
	return err
}

func TestStatsAnalyserSmoke(t *testing.T) {
	g := NewGomegaWithT(t)
	dummyModelPrefix := "model_"

	t.Logf("Start!")

	lags := NewModelReplicaLagsKeeper()
	lastUsed := NewModelReplicaLastUsedKeeper()
	service := NewStatsAnalyserService(
		[]ModelScalingStatsWrapper{
			{
				StatsKeeper: lags,
				Operator:    interfaces.Gte,
				Threshold:   lagThresholdDefault,
				Reset:       true,
				EventType:   ScaleUpEvent,
			},
			{
				StatsKeeper: lastUsed,
				Operator:    interfaces.Gte,
				Threshold:   lastUsedThresholdSecondsDefault,
				Reset:       false,
				EventType:   ScaleDownEvent,
			},
		},
		log.New(),
		statsPeriodSecondsDefault,
	)

	err := service.Start()

	time.Sleep(time.Millisecond * 100) // for the service to actually start

	g.Expect(err).To(BeNil())
	g.Expect(service.isReady).To(BeTrue())

	ch := service.GetEventChannel()

	t.Logf("Test lags")

	// add the models, note only 0,1,3
	err = service.AddModel(dummyModelPrefix + "0")
	g.Expect(err).To(BeNil())
	err = service.AddModel(dummyModelPrefix + "1")
	g.Expect(err).To(BeNil())
	err = service.AddModel(dummyModelPrefix + "3")
	g.Expect(err).To(BeNil())
	lags.ModelInferEnter(dummyModelPrefix+"0", "")
	lags.stats[dummyModelPrefix+"1"].(*lagStats).lag = lagThresholdDefault + 1

	event := <-ch
	g.Expect(event.StatsData.ModelName).To(Equal(dummyModelPrefix + "1"))
	g.Expect(event.StatsData.Value).To(Equal(uint32(lagThresholdDefault + 1)))
	g.Expect(event.EventType).To(Equal(ScaleUpEvent))

	t.Logf("Test last used")
	err = setLastUsed(lastUsed, dummyModelPrefix+"3", uint32(time.Now().Unix())-lastUsedThresholdSecondsDefault)
	g.Expect(err).To(BeNil())
	event = <-ch
	g.Expect(event.StatsData.ModelName).To(Equal(dummyModelPrefix + "3"))
	g.Expect(event.EventType).To(Equal(ScaleDownEvent))

	_ = service.Stop()

	time.Sleep(time.Millisecond * 100) // for the service to actually stop

	g.Expect(service.isReady).To(BeFalse())

	t.Logf("Done!")
}

func TestStatsAnalyserEarlyStop(t *testing.T) {
	g := NewGomegaWithT(t)

	lags := NewModelReplicaLagsKeeper()
	lastUsed := NewModelReplicaLastUsedKeeper()
	service := NewStatsAnalyserService(
		[]ModelScalingStatsWrapper{
			{
				StatsKeeper: lags,
				Operator:    interfaces.Gte,
				Threshold:   lagThresholdDefault,
				Reset:       true,
				EventType:   ScaleUpEvent,
			},
			{
				StatsKeeper: lastUsed,
				Operator:    interfaces.Gte,
				Threshold:   lastUsedThresholdSecondsDefault,
				Reset:       false,
				EventType:   ScaleDownEvent,
			},
		},
		log.New(),
		statsPeriodSecondsDefault,
	)

	err := service.Stop()
	g.Expect(err).To(BeNil())
	g.Expect(service.isReady).To(BeFalse())
}

func TestStatsAnalyserSoak(t *testing.T) {
	numberIterations := 1000
	numberModels := 100

	g := NewGomegaWithT(t)
	dummyModelPrefix := "model_"

	t.Logf("Start!")

	lags := NewModelReplicaLagsKeeper()
	lastUsed := NewModelReplicaLastUsedKeeper()
	logger := log.New()
	modelScalingStatsCollector := NewDataPlaneStatsCollector([]interfaces.ModelStatsKeeper{lags, lastUsed}, logger)
	service := NewStatsAnalyserService(
		[]ModelScalingStatsWrapper{
			{
				StatsKeeper: lags,
				Operator:    interfaces.Gte,
				Threshold:   lagThresholdDefault,
				Reset:       true,
				EventType:   ScaleUpEvent,
			},
			{
				StatsKeeper: lastUsed,
				Operator:    interfaces.Gte,
				Threshold:   lastUsedThresholdSecondsDefault,
				Reset:       false,
				EventType:   ScaleDownEvent,
			},
		},
		logger,
		statsPeriodSecondsDefault,
	)

	err := service.Start()

	time.Sleep(time.Millisecond * 100) // for the service to actually start

	g.Expect(err).To(BeNil())
	g.Expect(service.isReady).To(BeTrue())

	for j := 0; j < numberModels; j++ {
		err := service.AddModel(dummyModelPrefix + strconv.Itoa(j))
		g.Expect(err).To(BeNil())
	}

	ch := service.GetEventChannel()

	var jobsWg sync.WaitGroup
	jobsWg.Add(numberIterations * numberModels)

	for i := 0; i < numberIterations; i++ {
		for j := 0; j < numberModels; j++ {
			requestId := fmt.Sprintf("request_id_%d_%d", i, j)
			var wg sync.WaitGroup
			wg.Add(1)

			setupFn := func(x int) {
				err := scalingMetricsSetup(dummyModelPrefix+strconv.Itoa(x), requestId, modelScalingStatsCollector)
				wg.Done()
				g.Expect(err).To(BeNil())
			}
			teardownFn := func(x int) {
				wg.Wait()
				err := scalingMetricsTearDown(dummyModelPrefix+strconv.Itoa(x), requestId, &jobsWg, modelScalingStatsCollector)
				g.Expect(err).To(BeNil())
			}
			go setupFn(j)
			go teardownFn(j)
		}
	}
	go func() {
		// dump messages on the floor
		<-ch
	}()
	jobsWg.Wait()

	// delete
	for j := 0; j < numberModels; j++ {
		err := service.DeleteModel(dummyModelPrefix + strconv.Itoa(j))
		g.Expect(err).To(BeNil())
	}

	_ = service.Stop()

	t.Logf("Done!")
}

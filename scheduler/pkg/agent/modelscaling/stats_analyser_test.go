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
	wg *sync.WaitGroup, internalModelName string, modelScalingStatsCollector *DataPlaneStatsCollector) error {
	return modelScalingStatsCollector.ScalingMetricsSetup(wg, internalModelName)
}

func scalingMetricsTearDown(wg *sync.WaitGroup, internalModelName string,
	jobsWg *sync.WaitGroup, modelScalingStatsCollector *DataPlaneStatsCollector) error {
	err := modelScalingStatsCollector.ScalingMetricsTearDown(wg, internalModelName)
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
				Stats:     lags,
				Operator:  interfaces.Gte,
				Threshold: lagThresholdDefault,
				Reset:     true,
				EventType: ScaleUpEvent,
			},
			{
				Stats:     lastUsed,
				Operator:  interfaces.Gte,
				Threshold: lastUsedThresholdSecondsDefault,
				Reset:     false,
				EventType: ScaleDownEvent,
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

	err = lags.Set(dummyModelPrefix+"0", lagThresholdDefault-1)
	g.Expect(err).To(BeNil())
	err = lags.Set(dummyModelPrefix+"1", lagThresholdDefault+1)
	g.Expect(err).To(BeNil())
	err = lags.Set(dummyModelPrefix+"2", lagThresholdDefault+1) //  model 2 not added so will not get returned to ch
	g.Expect(err).To(BeNil())
	event := <-ch
	g.Expect(event.StatsData.ModelName).To(Equal(dummyModelPrefix + "1"))
	g.Expect(event.StatsData.Value).To(Equal(uint32(lagThresholdDefault + 1)))
	g.Expect(event.EventType).To(Equal(ScaleUpEvent))

	t.Logf("Test last used")
	err = lastUsed.Set(dummyModelPrefix+"3", uint32(time.Now().Unix())-lastUsedThresholdSecondsDefault)
	g.Expect(err).To(BeNil())
	err = lastUsed.Set(dummyModelPrefix+"4", uint32(time.Now().Unix())-lastUsedThresholdSecondsDefault) // model 4 not added
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
				Stats:     lags,
				Operator:  interfaces.Gte,
				Threshold: lagThresholdDefault,
				Reset:     true,
				EventType: ScaleUpEvent,
			},
			{
				Stats:     lastUsed,
				Operator:  interfaces.Gte,
				Threshold: lastUsedThresholdSecondsDefault,
				Reset:     false,
				EventType: ScaleDownEvent,
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
	modelScalingStatsCollector := NewDataPlaneStatsCollector(lags, lastUsed)
	service := NewStatsAnalyserService(
		[]ModelScalingStatsWrapper{
			{
				Stats:     lags,
				Operator:  interfaces.Gte,
				Threshold: lagThresholdDefault,
				Reset:     true,
				EventType: ScaleUpEvent,
			},
			{
				Stats:     lastUsed,
				Operator:  interfaces.Gte,
				Threshold: lastUsedThresholdSecondsDefault,
				Reset:     false,
				EventType: ScaleDownEvent,
			},
		},
		log.New(),
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
			var wg sync.WaitGroup
			wg.Add(1)
			setupFn := func(x int) {
				err := scalingMetricsSetup(&wg, dummyModelPrefix+strconv.Itoa(x), modelScalingStatsCollector)
				g.Expect(err).To(BeNil())
			}
			teardownFn := func(x int) {
				err := scalingMetricsTearDown(&wg, dummyModelPrefix+strconv.Itoa(x), &jobsWg, modelScalingStatsCollector)
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

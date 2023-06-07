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

package metrics

import (
	"strconv"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	log "github.com/sirupsen/logrus"
)

const (
	serverName      = "dummy_server"
	serverIdx       = 0
	modelNamePrefix = "dummy_model_"
)

func createTestPrometheusMetrics() (*PrometheusMetrics, error) {
	return NewPrometheusModelMetrics(serverName, serverIdx, log.New())
}

func TestLoadedModelMetrics(t *testing.T) {
	g := NewGomegaWithT(t)

	memBytes := uint64(10)

	type test struct {
		name                  string
		isLoad                bool
		isSoft                bool
		expectedEvict         int
		expectedMiss          int
		expectedLoad          int
		expectedUnload        int
		expectedMemory        uint64
		expectedEvictedMemory uint64
	}
	tests := []test{
		{
			name:                  "evict",
			isLoad:                false,
			isSoft:                true,
			expectedEvict:         1,
			expectedMiss:          0,
			expectedLoad:          0,
			expectedUnload:        0,
			expectedMemory:        0,
			expectedEvictedMemory: memBytes,
		},
		{
			name:                  "real load",
			isLoad:                true,
			isSoft:                false,
			expectedEvict:         0,
			expectedMiss:          0,
			expectedLoad:          1,
			expectedUnload:        0,
			expectedMemory:        memBytes,
			expectedEvictedMemory: 0,
		},
		{
			name:                  "real unload",
			isLoad:                false,
			isSoft:                false,
			expectedEvict:         0,
			expectedMiss:          0,
			expectedLoad:          0,
			expectedUnload:        1,
			expectedMemory:        0,
			expectedEvictedMemory: 0,
		},
		{
			name:                  "reload",
			isLoad:                true,
			isSoft:                true,
			expectedEvict:         0,
			expectedMiss:          1,
			expectedLoad:          0,
			expectedUnload:        0,
			expectedMemory:        memBytes,
			expectedEvictedMemory: 0,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			promMetrics, err := createTestPrometheusMetrics()
			g.Expect(err).To(BeNil())

			modelName := modelNamePrefix + "0"
			promMetrics.AddLoadedModelMetrics(modelName, memBytes, test.isLoad, test.isSoft)

			actualVal := testutil.ToFloat64(
				promMetrics.loadedModelGauge.With(prometheus.Labels{
					SeldonInternalModelMetric: modelName,
					SeldonServerMetric:        serverName,
					SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
				}))
			g.Expect(float64(test.expectedLoad)).To(Equal(actualVal))

			actualVal = testutil.ToFloat64(
				promMetrics.loadModelCounter.With(prometheus.Labels{
					SeldonServerMetric:        serverName,
					SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
				}))
			g.Expect(float64(test.expectedLoad)).To(Equal(actualVal))

			actualVal = testutil.ToFloat64(
				promMetrics.unloadModelCounter.With(prometheus.Labels{
					SeldonServerMetric:        serverName,
					SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
				}))
			g.Expect(float64(test.expectedUnload)).To(Equal(actualVal))

			actualVal = testutil.ToFloat64(
				promMetrics.loadedModelMemoryGauge.With(prometheus.Labels{
					SeldonInternalModelMetric: modelName,
					SeldonServerMetric:        serverName,
					SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
				}))
			g.Expect(float64(test.expectedMemory)).To(Equal(actualVal))

			actualVal = testutil.ToFloat64(
				promMetrics.evictedModelMemoryGauge.With(prometheus.Labels{
					SeldonInternalModelMetric: modelName,
					SeldonServerMetric:        serverName,
					SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
				}))
			g.Expect(float64(test.expectedEvictedMemory)).To(Equal(actualVal))

			actualVal = testutil.ToFloat64(
				promMetrics.cacheEvictCounter.With(prometheus.Labels{
					SeldonServerMetric:        serverName,
					SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
				}))
			g.Expect(float64(test.expectedEvict)).To(Equal(actualVal))

			actualVal = testutil.ToFloat64(
				promMetrics.cacheMissCounter.With(prometheus.Labels{
					SeldonServerMetric:        serverName,
					SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
				}))
			g.Expect(float64(test.expectedMiss)).To(Equal(actualVal))

			promMetrics.cacheEvictCounter.Reset()
			promMetrics.cacheMissCounter.Reset()
			promMetrics.loadModelCounter.Reset()
			promMetrics.unloadModelCounter.Reset()
			promMetrics.evictedModelMemoryGauge.Reset()
			promMetrics.loadedModelMemoryGauge.Reset()

		})
	}
}

func TestInferModelMetrics(t *testing.T) {
	g := NewGomegaWithT(t)

	promMetrics, err := createTestPrometheusMetrics()
	g.Expect(err).To(BeNil())

	modelName := modelNamePrefix + "0"
	latency := float64(20)
	method := "rest"
	promMetrics.AddModelInferMetrics(modelName, modelName, method, latency, "200")

	actualVal := testutil.ToFloat64(
		promMetrics.modelInferLatencyCounter.With(prometheus.Labels{
			SeldonModelMetric:         modelName,
			SeldonInternalModelMetric: modelName,
			MethodTypeMetric:          method,
			SeldonServerMetric:        serverName,
			SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
			CodeMetric:                "200",
		}))
	g.Expect(latency).To(Equal(actualVal))

	actualVal = testutil.ToFloat64(
		promMetrics.modelInferCounter.With(prometheus.Labels{
			SeldonModelMetric:         modelName,
			SeldonInternalModelMetric: modelName,
			MethodTypeMetric:          method,
			SeldonServerMetric:        serverName,
			SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
			CodeMetric:                "200",
		}))
	g.Expect(float64(1)).To(Equal(actualVal))

	actualVal = testutil.ToFloat64(
		promMetrics.modelAggregateInferLatencyCounter.With(prometheus.Labels{
			MethodTypeMetric:          method,
			SeldonServerMetric:        serverName,
			SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
		}))
	g.Expect(latency).To(Equal(actualVal))

	actualVal = testutil.ToFloat64(
		promMetrics.modelAggregateInferCounter.With(prometheus.Labels{
			MethodTypeMetric:          method,
			SeldonServerMetric:        serverName,
			SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
		}))
	g.Expect(float64(1)).To(Equal(actualVal))

	promMetrics.modelInferLatencyCounter.Reset()
	promMetrics.modelInferCounter.Reset()
	promMetrics.modelAggregateInferLatencyCounter.Reset()
	promMetrics.modelAggregateInferCounter.Reset()
}

func TestServerEarlyStop(t *testing.T) {
	g := NewGomegaWithT(t)

	service, _ := createTestPrometheusMetrics()
	err := service.Stop()
	g.Expect(err).To(BeNil())
}

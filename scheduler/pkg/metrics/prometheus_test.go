package metrics

import (
	"strconv"
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	. "github.com/onsi/gomega"
)

const (
	serverName      = "dummy_server"
	serverIdx       = 0
	namesapce       = "namespace"
	modelNamePrefix = "dummy_model_"
)

func createTestPrometheusMetrics() (*PrometheusMetrics, error) {
	return NewPrometheusMetrics(serverName, serverIdx, namesapce, log.New())
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
					SeldonInternalModelMetric: modelName,
					SeldonServerMetric:        serverName,
					SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
				}))
			g.Expect(float64(test.expectedEvict)).To(Equal(actualVal))

			actualVal = testutil.ToFloat64(
				promMetrics.cacheMissCounter.With(prometheus.Labels{
					SeldonInternalModelMetric: modelName,
					SeldonServerMetric:        serverName,
					SeldonServerReplicaMetric: strconv.Itoa(serverIdx),
				}))
			g.Expect(float64(test.expectedMiss)).To(Equal(actualVal))

			promMetrics.cacheEvictCounter.Reset()
			promMetrics.cacheMissCounter.Reset()
			promMetrics.evictedModelMemoryGauge.Reset()
			promMetrics.loadedModelMemoryGauge.Reset()

		})
	}

}

package modelscaling

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/interfaces"
	log "github.com/sirupsen/logrus"
)

const (
	statsPeriodSecondsDefault       = 5
	lagThresholdDefault             = 30
	lastUsedThresholdSecondsDefault = 30
)

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
	err = lags.Set(dummyModelPrefix+"0", lagThresholdDefault-1)
	g.Expect(err).To(BeNil())
	err = lags.Set(dummyModelPrefix+"1", lagThresholdDefault+1)
	g.Expect(err).To(BeNil())
	event := <-ch
	g.Expect(event.StatsData.ModelName).To(Equal(dummyModelPrefix + "1"))
	g.Expect(event.StatsData.Value).To(Equal(uint32(lagThresholdDefault + 1)))
	g.Expect(event.EventType).To(Equal(ScaleUpEvent))

	t.Logf("Test last used")
	err = lastUsed.Set(dummyModelPrefix+"3", uint32(time.Now().Unix())-lastUsedThresholdSecondsDefault)
	g.Expect(err).To(BeNil())
	event = <-ch
	g.Expect(event.StatsData.ModelName).To(Equal(dummyModelPrefix + "3"))
	g.Expect(event.EventType).To(Equal(ScaleDownEvent))

	_ = service.Stop()

	time.Sleep(time.Millisecond * 100) // for the service to actually stop

	g.Expect(service.isReady).To(BeFalse())

	t.Logf("Done!")
}

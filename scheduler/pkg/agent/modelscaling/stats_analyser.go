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
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
)

const (
	channelBufferSize         = 100
	channelSendTimeoutSeconds = 1
)

type ModelScalingEventType int

const (
	ScaleUpEvent ModelScalingEventType = iota
	ScaleDownEvent
)

type ModelScalingEvent struct {
	EventType ModelScalingEventType
	StatsData *interfaces.ModelStatsKV
}

type ModelScalingStatsWrapper struct {
	StatsKeeper interfaces.ModelStatsKeeper
	Operator    interfaces.LogicOperation
	Threshold   uint
	Reset       bool
	EventType   ModelScalingEventType
}

type StatsAnalyserService struct {
	statsList     []ModelScalingStatsWrapper
	periodSeconds uint
	done          chan bool
	mu            sync.RWMutex
	isReady       bool
	events        chan *ModelScalingEvent
	logger        log.FieldLogger
	modelsEnabled sync.Map
}

func NewStatsAnalyserService(
	statsList []ModelScalingStatsWrapper,
	logger log.FieldLogger,
	statsPeriodSeconds uint,
) *StatsAnalyserService {
	return &StatsAnalyserService{
		statsList:     statsList,
		done:          make(chan bool),
		periodSeconds: statsPeriodSeconds,
		mu:            sync.RWMutex{},
		isReady:       false,
		events:        make(chan *ModelScalingEvent, channelBufferSize),
		logger:        logger.WithField("source", "StatsAnalyzerService"),
		modelsEnabled: sync.Map{},
	}
}

func (ss *StatsAnalyserService) SetState(state interface{}) {
	// TODO: this is a violation  of Liskov Substitution Principle (LSP) :(
}

func (ss *StatsAnalyserService) Start() error {
	go func() {
		err := ss.start()
		if err != nil {
			ss.logger.WithError(err).Warnf("Stats analyser failed")
		}
	}()
	return nil
}

func (ss *StatsAnalyserService) Ready() bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.isReady
}

func (ss *StatsAnalyserService) Stop() error {
	ss.logger.Info("Start graceful shutdown")
	close(ss.done)
	ss.logger.Info("Finished graceful shutdown")
	return nil
}

func (ss *StatsAnalyserService) GetEventChannel() chan *ModelScalingEvent {
	return ss.events
}

func (ss *StatsAnalyserService) DeleteModel(modelName string) error {
	for _, stats := range ss.statsList {
		// delete from list of models to report back
		ss.modelsEnabled.Delete(modelName)

		if err := stats.StatsKeeper.Delete(modelName); err != nil {
			return err
		}
	}
	return nil
}

// we allow allow models loaded on agent to have metrics collected for scaling purposes:
// it is hard to shortcut individual models without causing extra performance overheads
// we might need in the future to enable scaling after the model has been deployed and therefore
// keeping metrics for all models is useful
func (ss *StatsAnalyserService) AddModel(modelName string) error {
	for _, stats := range ss.statsList {
		// add to list pf models to report back
		ss.modelsEnabled.Store(modelName, struct{}{})

		if err := stats.StatsKeeper.Add(modelName); err != nil {
			return err
		}
	}
	return nil
}

func (ss *StatsAnalyserService) Name() string {
	return "Model Replica Scaling Trigger"
}

func (ss *StatsAnalyserService) start() error {
	ticker := time.NewTicker(time.Second * time.Duration(ss.periodSeconds))
	defer ticker.Stop()

	ss.mu.Lock()
	ss.isReady = true
	ss.mu.Unlock()

	defer func() {
		ss.mu.Lock()
		ss.isReady = false
		ss.mu.Unlock()
	}()

	for {
		select {
		case <-ss.done:
			return nil
		case <-ticker.C:
			if err := ss.process(); err != nil {
				return err
			}
		}
	}
}

func (ss *StatsAnalyserService) process() error {
	timeout := time.NewTimer(time.Duration(channelSendTimeoutSeconds) * time.Second)
	defer timeout.Stop()
	for _, stats := range ss.statsList {
		if err := ss.processImpl(stats, timeout); err != nil {
			// do not continue this batch of events.
			ss.logger.Warn("Stop sending more scaling events for this round")
			break
		}
	}
	return nil
}

func (ss *StatsAnalyserService) processImpl(statsWrapper ModelScalingStatsWrapper, timer *time.Timer) error {
	modelsToScale, err := statsWrapper.StatsKeeper.GetAll(
		uint32(statsWrapper.Threshold), statsWrapper.Operator, statsWrapper.Reset)
	if err != nil {
		return err
	}
	for _, modelStatsData := range modelsToScale {
		_, enabled := ss.modelsEnabled.Load(modelStatsData.ModelName)
		if !enabled {
			continue
		}

		select {
		case ss.events <- &ModelScalingEvent{
			EventType: statsWrapper.EventType,
			StatsData: modelStatsData,
		}:
			ss.logger.Debugf("Produced event %d on channel for model %s", statsWrapper.EventType, modelStatsData.ModelName)
		case <-timer.C:
			msg := "Timeout sending trigger scaling event on channel"
			ss.logger.Error(msg)
			// invalidate any further events in this round as well.
			return fmt.Errorf(msg)
		}
	}
	return nil
}

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
	"fmt"
	"sync"
	"time"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	log "github.com/sirupsen/logrus"
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
	Stats     interfaces.ModelScalingStats
	Operator  interfaces.LogicOperation
	Threshold uint
	Reset     bool
	EventType ModelScalingEventType
}

type StatsAnalyserService struct {
	statsList     []ModelScalingStatsWrapper
	periodSeconds uint
	done          chan bool
	mu            sync.RWMutex
	isReady       bool
	events        chan *ModelScalingEvent
	logger        log.FieldLogger
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
		if err := stats.Stats.Delete(modelName); err != nil {
			return err
		}
	}
	return nil
}

func (ss *StatsAnalyserService) AddModel(modelName string) error {
	for _, stats := range ss.statsList {
		if err := stats.Stats.Add(modelName); err != nil {
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
	modelsToScale, err := statsWrapper.Stats.GetAll(
		uint32(statsWrapper.Threshold), statsWrapper.Operator, statsWrapper.Reset)
	if err != nil {
		return err
	}
	for _, modelStatsData := range modelsToScale {
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

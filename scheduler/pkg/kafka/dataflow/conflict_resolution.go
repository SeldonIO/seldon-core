/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package dataflow

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

type ConflictResolutioner struct {
	vectorClock          map[string]uint64                             // key: pipeline name, value: vector clock timestamp
	vectorResponseStatus map[string]map[string]pipeline.PipelineStatus // key: pipeline name, value: map of dataflow name to PipelineStatus
	logger               log.FieldLogger
}

func NewConflictResolution(logger log.FieldLogger) *ConflictResolutioner {
	return &ConflictResolutioner{
		vectorClock:          make(map[string]uint64),
		vectorResponseStatus: make(map[string]map[string]pipeline.PipelineStatus),
		logger:               logger.WithField("source", "dataflow-conflict-resolution"),
	}
}

func (cr *ConflictResolutioner) DeletePipeline(pipelineName string) {
	delete(cr.vectorClock, pipelineName)
	delete(cr.vectorResponseStatus, pipelineName)
}

func (cr *ConflictResolutioner) UpdatePipelineStatus(pipelineName string, stream string, status pipeline.PipelineStatus) {
	logger := cr.logger.WithField("func", "UpdatePipelineStatus")
	logger.Debugf("Updating pipeline %s stream %s status to %s", pipelineName, stream, status)
	cr.vectorResponseStatus[pipelineName][stream] = status
}

func (cr *ConflictResolutioner) IsMessageOutdated(message *chainer.PipelineUpdateStatusMessage) bool {
	logger := cr.logger.WithField("func", "IsMessageOutdated")
	timestamp := message.Update.Timestamp
	pipelineName := message.Update.Pipeline
	stream := message.Update.Stream

	if timestamp != cr.vectorClock[pipelineName] {
		logger.Debugf("Message timestamp %d does not match current vector clock timestamp %d for pipeline %s, ignoring message", timestamp, cr.vectorClock[pipelineName], pipelineName)
		return true
	}

	if _, ok := cr.vectorResponseStatus[pipelineName][stream]; !ok {
		logger.Debugf("Stream %s not found in vector response status for pipeline %s, ignoring message", stream, pipelineName)
		return true
	}

	return false
}

func (cr *ConflictResolutioner) CreateNewIteration(pipelineName string, servers []string) {
	cr.vectorClock[pipelineName]++
	cr.vectorResponseStatus[pipelineName] = make(map[string]pipeline.PipelineStatus)

	for _, server := range servers {
		cr.vectorResponseStatus[pipelineName][server] = pipeline.PipelineStatusUnknown
	}
}

func (cr *ConflictResolutioner) GetCountPipelineWithStatus(pipelineName string, status pipeline.PipelineStatus) int {
	count := 0
	for _, streamStatus := range cr.vectorResponseStatus[pipelineName] {
		if streamStatus == status {
			count++
		}
	}
	return count
}

func (cr *ConflictResolutioner) GetPipelineStatus(pipelineName string, message *chainer.PipelineUpdateStatusMessage) (pipeline.PipelineStatus, string) {
	logger := cr.logger.WithField("func", "GetPipelineStatus")
	streams := cr.vectorResponseStatus[pipelineName]
	unknownCount := cr.GetCountPipelineWithStatus(pipelineName, pipeline.PipelineStatusUnknown)

	if message.Update.Op == chainer.PipelineUpdateMessage_Create {
		readyCount := cr.GetCountPipelineWithStatus(pipelineName, pipeline.PipelineReady)
		failedCount := len(streams) - readyCount - unknownCount
		message := fmt.Sprintf(
			"%d/%d streams are ready, %d/%d still creating, %d/%d streams failed",
			readyCount, len(streams),
			unknownCount, len(streams),
			failedCount, len(streams),
		)
		// We log info this cause the reason doesn't not display in case of
		// success in the message column of k9s
		//
		// TODO: Implement something similar to models to display the numbers
		// of available replicas
		logger.Infof("Pipeline %s status message: %s", pipelineName, message)
		if failedCount == len(streams) {
			return pipeline.PipelineFailed, message
		}
		if readyCount > 0 && unknownCount == 0 {
			return pipeline.PipelineReady, message
		}
		return pipeline.PipelineCreating, message
	}

	if message.Update.Op == chainer.PipelineUpdateMessage_Delete {
		terminatedCount := cr.GetCountPipelineWithStatus(pipelineName, pipeline.PipelineTerminated)
		failedCount := len(streams) - terminatedCount - unknownCount
		message := fmt.Sprintf(
			"%d/%d streams terminated, %d/%d still terminating, %d/%d streams failed to terminate",
			terminatedCount, len(streams),
			unknownCount, len(streams),
			failedCount, len(streams),
		)
		logger.Infof("Pipeline %s status message: %s", pipelineName, message)
		if failedCount > 0 {
			return pipeline.PipelineFailed, message
		}
		if terminatedCount == len(streams) {
			return pipeline.PipelineTerminated, message
		}
		return pipeline.PipelineTerminating, message
	}

	return pipeline.PipelineStatusUnknown, "Unknown operation or status"
}

func (cr *ConflictResolutioner) GetTimestamp(pipelineName string) uint64 {
	if timestamp, ok := cr.vectorClock[pipelineName]; ok {
		return timestamp
	}
	cr.logger.Warnf("Timestamp for pipeline %s not found, returning 0", pipelineName)
	return 0
}

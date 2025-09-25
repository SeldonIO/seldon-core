/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package conflict_resolution

import (
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/chainer"
	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/pipeline"
)

type ConflictResolutioner[Status comparable] struct {
	VectorClock          map[string]uint64
	VectorResponseStatus map[string]map[string]Status
	logger               log.FieldLogger
}

func NewConflictResolution[Status comparable](logger log.FieldLogger) *ConflictResolutioner[Status] {
	return &ConflictResolutioner[Status]{
		VectorClock:          make(map[string]uint64),
		VectorResponseStatus: make(map[string]map[string]Status),
		logger:               logger.WithField("source", "dataflow-conflict-resolution"),
	}
}

func (cr *ConflictResolutioner[Status]) Delete(name string) {
	delete(cr.VectorClock, name)
	delete(cr.VectorResponseStatus, name)
}

func (cr *ConflictResolutioner[Status]) UpdateStatus(name string, stream string, status Status) {
	logger := cr.logger.WithField("func", "UpdatePipelineStatus")
	logger.Debugf("Updating %s stream %s status to %v", name, stream, status)
	cr.VectorResponseStatus[name][stream] = status
}

func (cr *ConflictResolutioner[Status]) IsMessageOutdated(
	timestamp uint64, name string, stream string,
) bool {
	logger := cr.logger.WithField("func", "IsMessageOutdated")
	if timestamp != cr.VectorClock[name] {
		logger.Debugf("Message timestamp %d does not match current vector clock timestamp %d for %s, ignoring message", timestamp, cr.VectorClock[name], name)
		return true
	}

	if _, ok := cr.VectorResponseStatus[name][stream]; !ok {
		logger.Debugf("Stream %s not found in vector response status for pipeline %s, ignoring message", stream, name)
		return true
	}

	return false
}

func (cr *ConflictResolutioner[Status]) CreateNewIteration(name string, servers []string, status Status) {
	cr.VectorClock[name]++
	cr.VectorResponseStatus[name] = make(map[string]Status)

	for _, server := range servers {
		cr.VectorResponseStatus[name][server] = status
	}
}

func (cr *ConflictResolutioner[Status]) GetCountWithStatus(name string, status Status) int {
	count := 0
	for _, streamStatus := range cr.VectorResponseStatus[name] {
		if streamStatus == status {
			count++
		}
	}
	return count
}

func (cr *ConflictResolutioner[Status]) GetTimestamp(name string) uint64 {
	if timestamp, ok := cr.VectorClock[name]; ok {
		return timestamp
	}
	cr.logger.Warnf("Timestamp for %s not found, returning 0", name)
	return 0
}

// --------------------
// Pipeline-specific
// --------------------

func CreateNewPipelineIteration(
	cr *ConflictResolutioner[pipeline.PipelineStatus],
	pipelineName string,
	servers []string,
) {
	cr.CreateNewIteration(pipelineName, servers, pipeline.PipelineStatusUnknown)
}

func GetCountPipelineWithStatus(cr *ConflictResolutioner[pipeline.PipelineStatus], pipelineName string, status pipeline.PipelineStatus) int {
	count := 0
	for _, streamStatus := range cr.VectorResponseStatus[pipelineName] {
		if streamStatus == status {
			count++
		}
	}
	return count
}

func GetPipelineStatus(
	cr *ConflictResolutioner[pipeline.PipelineStatus],
	pipelineName string,
	message *chainer.PipelineUpdateStatusMessage,
) (pipeline.PipelineStatus, string) {
	logger := cr.logger.WithField("func", "GetPipelineStatus")
	streams := cr.VectorResponseStatus[pipelineName]

	var messageStr = ""
	readyCount := GetCountPipelineWithStatus(cr, pipelineName, pipeline.PipelineReady)
	if readyCount > 0 {
		messageStr += fmt.Sprintf("%d/%d ready ", readyCount, len(streams))
	}

	terminatedCount := GetCountPipelineWithStatus(cr, pipelineName, pipeline.PipelineTerminated)
	if terminatedCount > 0 {
		messageStr += fmt.Sprintf("%d/%d terminated ", terminatedCount, len(streams))
	}

	failedCount := GetCountPipelineWithStatus(cr, pipelineName, pipeline.PipelineFailed)
	if failedCount > 0 {
		messageStr += fmt.Sprintf("%d/%d failed ", failedCount, len(streams))
	}

	rebalancingCount := GetCountPipelineWithStatus(cr, pipelineName, pipeline.PipelineRebalancing)
	if rebalancingCount > 0 {
		messageStr += fmt.Sprintf("%d/%d rebalancing ", rebalancingCount, len(streams))
	}

	unknownCount := GetCountPipelineWithStatus(cr, pipelineName, pipeline.PipelineStatusUnknown)
	logger.Infof("Pipeline %s status counts: %s", pipelineName, messageStr)

	if message.Update.Op == chainer.PipelineUpdateMessage_Create {
		// We log info this cause the reason doesn't not display in case of
		// success in the message column of k9s
		//
		// TODO: Implement something similar to models to display the numbers
		// of available replicas
		if failedCount == len(streams) {
			return pipeline.PipelineFailed, messageStr
		}
		if readyCount > 0 && unknownCount == 0 {
			return pipeline.PipelineReady, messageStr
		}
		return pipeline.PipelineCreating, messageStr
	}

	if message.Update.Op == chainer.PipelineUpdateMessage_Delete {
		if failedCount > 0 {
			return pipeline.PipelineFailed, messageStr
		}
		if terminatedCount == len(streams) {
			return pipeline.PipelineTerminated, messageStr
		}
		return pipeline.PipelineTerminating, messageStr
	}

	if message.Update.Op == chainer.PipelineUpdateMessage_Rebalance || message.Update.Op == chainer.PipelineUpdateMessage_Ready {
		if failedCount == len(streams) {
			return pipeline.PipelineFailed, messageStr
		}
		if readyCount > 0 && rebalancingCount == 0 {
			return pipeline.PipelineReady, messageStr
		}
		return pipeline.PipelineRebalancing, messageStr
	}

	return pipeline.PipelineStatusUnknown, "Unknown operation or status"
}

func IsPipelineMessageOutdated(
	cr *ConflictResolutioner[pipeline.PipelineStatus],
	message *chainer.PipelineUpdateStatusMessage,
) bool {
	timestamp := message.Update.Timestamp
	pipelineName := message.Update.Pipeline
	stream := message.Update.Stream
	return cr.IsMessageOutdated(timestamp, pipelineName, stream)
}

// --------------------
// Model-specific
// --------------------

func CreateNewModelIteration(
	cr *ConflictResolutioner[store.ModelState],
	modelName string,
	servers []string,
) {
	cr.CreateNewIteration(modelName, servers, store.ModelStateUnknown)
}

func GetModelStatus(
	cr *ConflictResolutioner[store.ModelState],
	modelName string,
	message *pb.ModelUpdateStatusMessage,
) (store.ModelState, string) {
	logger := cr.logger.WithField("func", "GetModelStatus")
	streams := cr.VectorResponseStatus[modelName]
	unknownCount := cr.GetCountWithStatus(modelName, store.ModelStateUnknown)

	if message.Update.Op == pb.ModelUpdateMessage_Create {
		readyCount := cr.GetCountWithStatus(modelName, store.ModelAvailable)
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
		logger.Infof("Model %s status message: %s", modelName, message)
		if failedCount == len(streams) {
			return store.ModelFailed, message
		}
		if readyCount > 0 && unknownCount == 0 {
			return store.ModelAvailable, message
		}
		return store.ModelProgressing, message
	}

	if message.Update.Op == pb.ModelUpdateMessage_Delete {
		terminatedCount := cr.GetCountWithStatus(modelName, store.ModelTerminated)
		failedCount := len(streams) - terminatedCount - unknownCount
		message := fmt.Sprintf(
			"%d/%d streams terminated, %d/%d still terminating, %d/%d streams failed to terminate",
			terminatedCount, len(streams),
			unknownCount, len(streams),
			failedCount, len(streams),
		)
		logger.Infof("Model %s status message: %s", modelName, message)
		if failedCount > 0 {
			return store.ModelTerminateFailed, message
		}
		if terminatedCount == len(streams) {
			return store.ModelTerminated, message
		}
		return store.ModelTerminating, message
	}

	return store.ModelStateUnknown, "Unknown operation or status"
}

func IsModelMessageOutdated(
	cr *ConflictResolutioner[store.ModelState],
	message *pb.ModelUpdateStatusMessage,
) bool {
	timestamp := message.Update.Timestamp
	modelName := message.Update.Model
	stream := message.Update.Stream
	return cr.IsMessageOutdated(timestamp, modelName, stream)
}

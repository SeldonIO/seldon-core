/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package status

import (
	"context"
	"errors"
)

type PipelineReadyChecker interface {
	CheckPipelineReady(ctx context.Context, pipelineName string, requestId string) (bool, error)
}

type ModelReadyCaller interface {
	CheckModelReady(ctx context.Context, modelName string, requestId string) (bool, error)
}

type SimpleReadyChecker struct {
	pipelineStatusGetter PipelineStatusProvider
	modelReadyCaller     ModelReadyCaller
}

func NewSimpleReadyChecker(pipelineStatusProvider PipelineStatusProvider, modelReadyCaller ModelReadyCaller) *SimpleReadyChecker {
	return &SimpleReadyChecker{
		pipelineStatusGetter: pipelineStatusProvider,
		modelReadyCaller:     modelReadyCaller,
	}
}

var PipelineNotFoundErr = errors.New("Pipeline not found")

func (s *SimpleReadyChecker) CheckPipelineReady(ctx context.Context, pipelineName string, requestId string) (bool, error) {
	pipeline := s.pipelineStatusGetter.Get(pipelineName)
	if pipeline == nil {
		return false, PipelineNotFoundErr
	}
	for modelName := range pipeline.Steps {
		ready, err := s.modelReadyCaller.CheckModelReady(ctx, modelName, requestId)
		if err != nil || !ready {
			return ready, err
		}
	}
	return true, nil
}

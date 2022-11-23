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

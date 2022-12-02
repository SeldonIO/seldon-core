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

package experiment

import (
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func CreateExperimentFromRequest(request *scheduler.Experiment) *Experiment {
	var candidates []*Candidate
	var mirror *Mirror
	var config *Config
	var k8sMeta *KubernetesMeta
	for _, reqCandidate := range request.Candidates {
		candidates = append(candidates, &Candidate{
			Name:   reqCandidate.Name,
			Weight: reqCandidate.Weight,
		})
	}
	if request.Mirror != nil {
		mirror = &Mirror{
			Name:    request.Mirror.Name,
			Percent: request.Mirror.Percent,
		}
	}
	if request.Config != nil {
		config = &Config{
			StickySessions: request.Config.StickySessions,
		}
	}
	if request.KubernetesMeta != nil {
		k8sMeta = &KubernetesMeta{
			Namespace:  request.KubernetesMeta.Namespace,
			Generation: request.KubernetesMeta.Generation,
		}
	}
	var resourceType ResourceType
	switch request.ResourceType {
	case scheduler.ResourceType_PIPELINE:
		resourceType = PipelineResourceType
	case scheduler.ResourceType_MODEL:
		resourceType = ModelResourceType
	}
	return &Experiment{
		Name:           request.Name,
		Default:        request.Default,
		ResourceType:   resourceType,
		Active:         false,
		Candidates:     candidates,
		Mirror:         mirror,
		Config:         config,
		KubernetesMeta: k8sMeta,
	}
}

func CreateExperimentProto(experiment *Experiment) *scheduler.Experiment {
	var candidates []*scheduler.ExperimentCandidate
	for _, candidate := range experiment.Candidates {
		candidates = append(candidates, &scheduler.ExperimentCandidate{
			Name:   candidate.Name,
			Weight: candidate.Weight,
		})
	}
	var mirror *scheduler.ExperimentMirror
	if experiment.Mirror != nil {
		mirror = &scheduler.ExperimentMirror{
			Name:    experiment.Mirror.Name,
			Percent: experiment.Mirror.Percent,
		}
	}
	var config *scheduler.ExperimentConfig
	if experiment.Config != nil {
		config = &scheduler.ExperimentConfig{
			StickySessions: experiment.Config.StickySessions,
		}
	}
	var k8sMeta *scheduler.KubernetesMeta
	if experiment.KubernetesMeta != nil {
		k8sMeta = &scheduler.KubernetesMeta{
			Namespace:  experiment.KubernetesMeta.Namespace,
			Generation: experiment.KubernetesMeta.Generation,
		}
	}
	var resourceType scheduler.ResourceType
	switch experiment.ResourceType {
	case PipelineResourceType:
		resourceType = scheduler.ResourceType_PIPELINE
	case ModelResourceType:
		resourceType = scheduler.ResourceType_MODEL
	}
	return &scheduler.Experiment{
		Name:           experiment.Name,
		Default:        experiment.Default,
		ResourceType:   resourceType,
		Candidates:     candidates,
		Mirror:         mirror,
		Config:         config,
		KubernetesMeta: k8sMeta,
	}
}

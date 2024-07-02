/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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
		Active:         false, // this is always false when creating from a request
		Deleted:        false,
		Candidates:     candidates,
		Mirror:         mirror,
		Config:         config,
		KubernetesMeta: k8sMeta,
	}
}

func CreateExperimentFromSnapshot(request *scheduler.ExperimentSnapshot) *Experiment {
	experiment := CreateExperimentFromRequest(request.Experiment)
	experiment.Deleted = request.Deleted
	return experiment
}

func CreateExperimentSnapshotProto(experiment *Experiment) *scheduler.ExperimentSnapshot {
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
	return &scheduler.ExperimentSnapshot{
		Experiment: &scheduler.Experiment{
			Name:           experiment.Name,
			Default:        experiment.Default,
			ResourceType:   resourceType,
			Candidates:     candidates,
			Mirror:         mirror,
			Config:         config,
			KubernetesMeta: k8sMeta,
		},
		Deleted: experiment.Deleted,
	}
}

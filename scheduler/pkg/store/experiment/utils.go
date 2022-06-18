package experiment

import (
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

func CreateExperimentFromRequest(request *scheduler.Experiment) *Experiment {
	var candidates []*Candidate
	var mirror *Mirror
	var config *Config
	var k8sMeta *KubernetesMeta
	for _, reqCandidate := range request.Candidates {
		candidates = append(candidates, &Candidate{
			ModelName: reqCandidate.ModelName,
			Weight:    reqCandidate.Weight,
		})
	}
	if request.Mirror != nil {
		mirror = &Mirror{
			ModelName: request.Mirror.ModelName,
			Percent:   request.Mirror.Percent,
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
	return &Experiment{
		Name:           request.Name,
		DefaultModel:   request.DefaultModel,
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
			ModelName: candidate.ModelName,
			Weight:    candidate.Weight,
		})
	}
	var mirror *scheduler.ExperimentMirror
	if experiment.Mirror != nil {
		mirror = &scheduler.ExperimentMirror{
			ModelName: experiment.Mirror.ModelName,
			Percent:   experiment.Mirror.Percent,
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
	return &scheduler.Experiment{
		Name:           experiment.Name,
		DefaultModel:   experiment.DefaultModel,
		Candidates:     candidates,
		Mirror:         mirror,
		Config:         config,
		KubernetesMeta: k8sMeta,
	}
}

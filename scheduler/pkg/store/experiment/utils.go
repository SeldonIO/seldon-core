package experiment

import "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"

func CreateExperimentFromRequest(request *scheduler.StartExperimentRequest) *Experiment {
	var candidates []*Candidate
	var baseline *Candidate
	var mirror *Mirror
	var config *Config
	for _, reqCandidate := range request.Candidates {
		candidates = append(candidates, &Candidate{
			ModelName: reqCandidate.ModelName,
			Weight:    reqCandidate.Weight,
		})
	}
	if request.Baseline != nil {
		baseline = &Candidate{
			ModelName: request.Baseline.ModelName,
			Weight:    request.Baseline.Weight,
		}
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
	return &Experiment{
		Name:       request.Name,
		Active:     false,
		Candidates: candidates,
		Baseline:   baseline,
		Mirror:     mirror,
		Config:     config,
	}
}

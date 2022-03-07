package experiment

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

func TestLoadModel(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name       string
		proto      *scheduler.StartExperimentRequest
		experiment *Experiment
	}

	tests := []test{
		{
			name: "basic",
			proto: &scheduler.StartExperimentRequest{
				Name: "foo",
				Baseline: &scheduler.ExperimentCandidate{
					ModelName: "model1",
					Weight:    60,
				},
				Candidates: []*scheduler.ExperimentCandidate{
					{
						ModelName: "model2",
						Weight:    20,
					},
					{
						ModelName: "model3",
						Weight:    20,
					},
				},
				Mirror: &scheduler.ExperimentMirror{
					ModelName: "model4",
					Percent:   80,
				},
				Config: &scheduler.ExperimentConfig{
					StickySessions: true,
				},
			},
			experiment: &Experiment{
				Name:   "foo",
				Active: false,
				Baseline: &Candidate{
					ModelName: "model1",
					Weight:    60,
				},
				Candidates: []*Candidate{
					{
						ModelName: "model2",
						Weight:    20,
					},
					{
						ModelName: "model3",
						Weight:    20,
					},
				},
				Mirror: &Mirror{
					ModelName: "model4",
					Percent:   80,
				},
				Config: &Config{
					StickySessions: true,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			experiment := CreateExperimentFromRequest(test.proto)
			g.Expect(experiment).To(Equal(test.experiment))
		})
	}

}

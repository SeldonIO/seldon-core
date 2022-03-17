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
		proto      *scheduler.Experiment
		experiment *Experiment
	}

	getStrPtr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "basic",
			proto: &scheduler.Experiment{
				Name:         "foo",
				DefaultModel: getStrPtr("model1"),
				Candidates: []*scheduler.ExperimentCandidate{
					{
						ModelName: "model1",
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
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
				},
			},
			experiment: &Experiment{
				Name:         "foo",
				Active:       false,
				DefaultModel: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						ModelName: "model1",
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
				KubernetesMeta: &KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
				},
			},
		},
		{
			name: "candidates",
			proto: &scheduler.Experiment{
				Name: "foo",
				Candidates: []*scheduler.ExperimentCandidate{
					{
						ModelName: "model1",
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
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
				},
			},
			experiment: &Experiment{
				Name:   "foo",
				Active: false,
				Candidates: []*Candidate{
					{
						ModelName: "model1",
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
				KubernetesMeta: &KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
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

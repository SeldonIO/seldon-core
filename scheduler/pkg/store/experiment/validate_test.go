package experiment

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestValidateExperiment(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		store      *ExperimentStore
		experiment *Experiment
		err        error
	}

	tests := []test{
		{
			name: "valid",
			store: &ExperimentStore{
				baselines:   map[string]*Experiment{},
				experiments: map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name: "a",
				Baseline: &Candidate{
					ModelName: "model1",
				},
				Candidates: []*Candidate{
					{
						ModelName: "model2",
					},
				},
			},
		},
		{
			name: "baseline already exists",
			store: &ExperimentStore{
				baselines:   map[string]*Experiment{"model1": {}},
				experiments: map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name: "a",
				Baseline: &Candidate{
					ModelName: "model1",
				},
				Candidates: []*Candidate{
					{
						ModelName: "model2",
					},
				},
			},
			err: &ExperimentBaselineExists{experimentName: "a", modelName: "model1"},
		},
		{
			name: "No candidates in experiment",
			store: &ExperimentStore{
				baselines:   map[string]*Experiment{},
				experiments: map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name: "a",
				Baseline: &Candidate{
					ModelName: "model1",
				},
			},
			err: &ExperimentNoCandidates{experimentName: "a"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.store.validate(test.experiment)
			if test.err != nil {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

package experiment

import (
	"errors"
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

	var baselineError *ExperimentBaselineExists
	var noCandidatesError *ExperimentNoCandidates
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
			err: baselineError,
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
			err: noCandidatesError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.store.validate(test.experiment)
			if test.err != nil {
				g.Expect(errors.As(err, &test.err)).To(BeTrue())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

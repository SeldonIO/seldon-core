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

	getStrPtr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "valid",
			store: &ExperimentStore{
				baselines:   map[string]*Experiment{},
				experiments: map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name:         "a",
				DefaultModel: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						ModelName: "model1",
					},
					{
						ModelName: "model2",
					},
				},
			},
		},
		{
			name: "baseline already exists",
			store: &ExperimentStore{
				baselines:   map[string]*Experiment{"model1": {Name: "b"}},
				experiments: map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name:         "a",
				DefaultModel: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						ModelName: "model1",
					},
					{
						ModelName: "model2",
					},
				},
			},
			err: &ExperimentBaselineExists{experimentName: "a", modelName: "model1"},
		},
		{
			name: "baseline already exists but its this model so ignore",
			store: &ExperimentStore{
				baselines:   map[string]*Experiment{"model1": {Name: "a"}},
				experiments: map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name:         "a",
				DefaultModel: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						ModelName: "model1",
					},
					{
						ModelName: "model2",
					},
				},
			},
		},
		{
			name: "No Canidadates",
			store: &ExperimentStore{
				baselines:   map[string]*Experiment{},
				experiments: map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name: "a",
			},
			err: &ExperimentNoCandidates{experimentName: "a"},
		},
		{
			name: "Default model is not candidate",
			store: &ExperimentStore{
				baselines:   map[string]*Experiment{},
				experiments: map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name:         "a",
				DefaultModel: getStrPtr("model1"),
			},
			err: &ExperimentDefaultModelNotFound{experimentName: "a", defaultModel: "model1"},
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

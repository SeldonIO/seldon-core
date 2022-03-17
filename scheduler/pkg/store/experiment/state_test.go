package experiment

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestAddModelReference(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		store      *ExperimentStore
		modelName  string
		experiment *Experiment
	}

	getStrPtr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "empty store add experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{},
			},
			modelName: "model1",
			experiment: &Experiment{
				Name:         "a",
				DefaultModel: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						ModelName: "model1",
						Weight:    100,
					},
				},
			},
		},
		{
			name: "existing reference and adding experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{
					"model1": {"a": nil},
				},
			},
			modelName: "model1",
			experiment: &Experiment{
				Name:         "a",
				DefaultModel: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						ModelName: "model1",
						Weight:    100,
					},
				},
			},
		},
		{
			name: "existing reference to another experiment and adding experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{
					"model1": {"b": nil},
				},
			},
			modelName: "model1",
			experiment: &Experiment{
				Name:         "a",
				DefaultModel: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						ModelName: "model1",
						Weight:    100,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.store.addModelReference(test.modelName, test.experiment)
			g.Expect(test.store.modelReferences[test.modelName]).ToNot(BeNil())
			g.Expect(test.store.modelReferences[test.modelName][test.experiment.Name]).ToNot(BeNil())
			g.Expect(test.store.modelReferences[test.modelName][test.experiment.Name]).To(Equal(test.experiment))
		})
	}
}

func TestRemoveModelReference(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		store              *ExperimentStore
		modelName          string
		experiment         *Experiment
		expectedReferences int
	}

	getStrPtr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "empty store remove experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{},
			},
			modelName: "model1",
			experiment: &Experiment{
				Name:         "a",
				DefaultModel: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						ModelName: "model1",
						Weight:    100,
					},
				},
			},
			expectedReferences: 0,
		},
		{
			name: "existing reference and remove experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{
					"model1": {"a": nil},
				},
			},
			modelName: "model1",
			experiment: &Experiment{
				Name:         "a",
				DefaultModel: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						ModelName: "model1",
						Weight:    100,
					},
				},
			},
			expectedReferences: 0,
		},
		{
			name: "existing reference to another experiment and remove experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{
					"model1": {"b": nil, "a": nil},
				},
			},
			modelName: "model1",
			experiment: &Experiment{
				Name:         "a",
				DefaultModel: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						ModelName: "model1",
						Weight:    100,
					},
				},
			},
			expectedReferences: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.store.removeModelReference(test.modelName, test.experiment)
			g.Expect(len(test.store.modelReferences[test.modelName])).To(Equal(test.expectedReferences))
		})
	}
}

func TestCleanExperimentState(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		server             *ExperimentStore
		experiment         *Experiment
		expectedReferences int
	}

	getStrPtr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "empty server",
			server: &ExperimentStore{
				logger:          logrus.New(),
				baselines:       map[string]*Experiment{},
				modelReferences: map[string]map[string]*Experiment{},
			},
			experiment:         &Experiment{Name: "a"},
			expectedReferences: 0,
		},
		{
			name: "existing experiment",
			server: &ExperimentStore{
				logger:          logrus.New(),
				baselines:       map[string]*Experiment{"model1": nil},
				modelReferences: map[string]map[string]*Experiment{"model2": {"a": nil}, "model1": {"a": nil}},
				experiments: map[string]*Experiment{"a": {
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
				}},
			},
			experiment: &Experiment{
				Name: "a",
			},
			expectedReferences: 0,
		},
		{
			name: "other experiments",
			server: &ExperimentStore{
				logger:          logrus.New(),
				baselines:       map[string]*Experiment{"model1": nil},
				modelReferences: map[string]map[string]*Experiment{"model2": {"a": nil, "b": nil}, "model1": {"a": nil, "b": nil}},
				experiments: map[string]*Experiment{"a": {
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
				}},
			},
			experiment: &Experiment{
				Name: "a",
			},
			expectedReferences: 2,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.server.cleanExperimentState(test.experiment)
			g.Expect(test.server.baselines[test.experiment.Name]).To(BeNil())
			g.Expect(test.server.getTotalModelReferences()).To(Equal(test.expectedReferences))
		})
	}
}

func TestUpdateExperimentState(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		server             *ExperimentStore
		experiment         *Experiment
		expectedReferences int
	}

	getStrPtr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "existing experiment but empty store",
			server: &ExperimentStore{
				logger:          logrus.New(),
				baselines:       map[string]*Experiment{},
				modelReferences: map[string]map[string]*Experiment{},
				experiments:     map[string]*Experiment{},
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
			expectedReferences: 2,
		},
		{
			name: "existing experiment and model in store",
			server: &ExperimentStore{
				logger:          logrus.New(),
				baselines:       map[string]*Experiment{"model1": {Name: "a"}},
				modelReferences: map[string]map[string]*Experiment{"model2": {"a": nil}, "model1": {"a": nil}},
				experiments:     map[string]*Experiment{},
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
			expectedReferences: 2,
		},
		{
			name: "other experiments",
			server: &ExperimentStore{
				logger:          logrus.New(),
				baselines:       map[string]*Experiment{"model1": {Name: "a"}},
				modelReferences: map[string]map[string]*Experiment{"model2": {"b": nil}, "model1": {"b": nil}},
				experiments:     map[string]*Experiment{},
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
			expectedReferences: 4,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.server.validate(test.experiment)
			g.Expect(err).To(BeNil())
			test.server.updateExperimentState(test.experiment)
			g.Expect(test.server.getTotalModelReferences()).To(Equal(test.expectedReferences))
		})
	}
}

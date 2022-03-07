package experiment

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"
	"github.com/sirupsen/logrus"
)

func TestStartExperiment(t *testing.T) {
	g := NewGomegaWithT(t)

	type experimentAddition struct {
		experiment *Experiment
		fail       bool
	}

	type test struct {
		name           string
		experiments    []*experimentAddition
		expectedNumExp int
	}

	tests := []test{
		{
			name: "add one",
			experiments: []*experimentAddition{
				{
					experiment: &Experiment{
						Name: "a",
						Candidates: []*Candidate{
							{
								ModelName: "model1",
							},
						},
					},
				},
			},
			expectedNumExp: 1,
		},
		{
			name: "add two",
			experiments: []*experimentAddition{
				{
					experiment: &Experiment{
						Name: "a",
						Candidates: []*Candidate{
							{
								ModelName: "model1",
							},
						},
					},
				},
				{
					experiment: &Experiment{
						Name: "b",
						Candidates: []*Candidate{
							{
								ModelName: "model1",
							},
						},
					},
				},
			},
			expectedNumExp: 2,
		},
		{
			name: "add duplicates",
			experiments: []*experimentAddition{
				{
					experiment: &Experiment{
						Name: "a",
						Candidates: []*Candidate{
							{
								ModelName: "model1",
							},
						},
					},
				},
				{
					experiment: &Experiment{
						Name: "a",
						Candidates: []*Candidate{
							{
								ModelName: "model1",
							},
						},
					},
				},
			},
			expectedNumExp: 1,
		},
		{
			name: "add baseline experiment but no model exists",
			experiments: []*experimentAddition{
				{
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
			},
			expectedNumExp: 1,
		},
		{
			name: "add baseline experiment",
			experiments: []*experimentAddition{
				{
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
			},
			expectedNumExp: 1,
		},
		{
			name: "add baseline experiment twice to same model - not allowed",
			experiments: []*experimentAddition{
				{
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
					experiment: &Experiment{
						Name: "b",
						Baseline: &Candidate{
							ModelName: "model1",
						},
						Candidates: []*Candidate{
							{
								ModelName: "model2",
							},
						},
					},
					fail: true,
				},
			},
			expectedNumExp: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrus.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			server := NewExperimentServer(logger, eventHub)
			for _, ea := range test.experiments {
				err := server.StartExperiment(ea.experiment)
				if ea.fail {
					g.Expect(err).ToNot(BeNil())
				} else {
					g.Expect(err).To(BeNil())
				}
			}
			g.Expect(len(server.experiments)).To(Equal(test.expectedNumExp))
		})
	}
}

func TestStopExperiment(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name           string
		store          *ExperimentStore
		experimentName string
		err            error
	}

	tests := []test{
		{
			name: "remove existing",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"a": {},
				},
			},
			experimentName: "a",
		},
		{
			name: "remove not existing",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"b": {},
				},
			},
			experimentName: "a",
			err:            &ExperimentNotFound{experimentName: "a"},
		},
		{
			name: "remove existing multiple",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"a": {},
					"b": {},
				},
			},
			experimentName: "a",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.store.StopExperiment(test.experimentName)
			if test.err != nil {
				_, ok := err.(*ExperimentNotFound)
				g.Expect(ok).To(BeTrue())
			} else {
				g.Expect(err).To(BeNil())
				experiment, err := test.store.GetExperiment(test.experimentName)
				g.Expect(err).To(BeNil())
				g.Expect(experiment.Deleted).To(BeTrue())
			}
		})
	}
}

func TestGetExperiment(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name           string
		store          *ExperimentStore
		experimentName string
		err            error
	}
	var experimentNotFoundErr *ExperimentNotFound
	tests := []test{
		{
			name: "experiment found",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"a": {},
				},
			},
			experimentName: "a",
		},
		{
			name: "experiment not found",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"b": {},
				},
			},
			experimentName: "a",
			err:            experimentNotFoundErr,
		},
		{
			name: "deleted",
			store: &ExperimentStore{
				logger: logrus.New(),
				experiments: map[string]*Experiment{
					"a": {Deleted: true},
				},
			},
			experimentName: "a",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			experiment, err := test.store.GetExperiment(test.experimentName)
			if test.err != nil {
				g.Expect(errors.As(err, &test.err)).To(BeTrue())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(experiment).To(Equal(test.store.experiments[test.experimentName]))
				// Change store experiment and check its a deep copy
				newName := "123"
				test.store.experiments[test.experimentName].Name = newName
				g.Expect(experiment.Name).ToNot(Equal(newName))
			}
		})
	}
}

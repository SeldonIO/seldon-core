package experiment

import (
	"testing"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	"github.com/seldonio/seldon-core/scheduler/pkg/store/pipeline"

	"github.com/seldonio/seldon-core/scheduler/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/pkg/store"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

func TestAddReference(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		store        *ExperimentStore
		resourceName string
		experiment   *Experiment
	}

	getStrPtr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "empty store add model experiment",
			store: &ExperimentStore{
				modelReferences:    map[string]map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{},
			},
			resourceName: "model1",
			experiment: &Experiment{
				Name:    "a",
				Default: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						Name:   "model1",
						Weight: 100,
					},
				},
			},
		},
		{
			name: "empty store add pipeline experiment",
			store: &ExperimentStore{
				modelReferences:    map[string]map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{},
			},
			resourceName: "pipeline1",
			experiment: &Experiment{
				Name:         "a",
				Default:      getStrPtr("pipeline1"),
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name:   "pipeline1",
						Weight: 100,
					},
				},
			},
		},
		{
			name: "existing reference and adding model experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{
					"model1": {"a": nil},
				},
			},
			resourceName: "model1",
			experiment: &Experiment{
				Name:    "a",
				Default: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						Name:   "model1",
						Weight: 100,
					},
				},
			},
		},
		{
			name: "existing reference and adding pipeline experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{
					"pipeline1": {"a": nil},
				},
			},
			resourceName: "pipeline1",
			experiment: &Experiment{
				Name:         "a",
				Default:      getStrPtr("pipeline1"),
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name:   "pipeline1",
						Weight: 100,
					},
				},
			},
		},
		{
			name: "existing reference to another experiment and adding model experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{
					"model1": {"b": nil},
				},
			},
			resourceName: "model1",
			experiment: &Experiment{
				Name:    "a",
				Default: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						Name:   "model1",
						Weight: 100,
					},
				},
			},
		},
		{
			name: "existing reference to another experiment and adding pipeline experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{
					"pipeline1": {"b": nil},
				},
			},
			resourceName: "pipeline1",
			experiment: &Experiment{
				Name:         "a",
				Default:      getStrPtr("pipeline1"),
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name:   "pipeline1",
						Weight: 100,
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.store.addReference(test.resourceName, test.experiment)
			switch test.experiment.ResourceType {
			case PipelineResourceType:
				g.Expect(test.store.pipelineReferences[test.resourceName]).ToNot(BeNil())
				g.Expect(test.store.pipelineReferences[test.resourceName][test.experiment.Name]).ToNot(BeNil())
				g.Expect(test.store.pipelineReferences[test.resourceName][test.experiment.Name]).To(Equal(test.experiment))
			default:
				g.Expect(test.store.modelReferences[test.resourceName]).ToNot(BeNil())
				g.Expect(test.store.modelReferences[test.resourceName][test.experiment.Name]).ToNot(BeNil())
				g.Expect(test.store.modelReferences[test.resourceName][test.experiment.Name]).To(Equal(test.experiment))
			}
		})
	}
}

func TestRemoveReference(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		store              *ExperimentStore
		resourceName       string
		experiment         *Experiment
		expectedReferences int
	}

	getStrPtr := func(val string) *string { return &val }
	tests := []test{
		{
			name: "empty store remove model experiment",
			store: &ExperimentStore{
				modelReferences:    map[string]map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{},
			},
			resourceName: "model1",
			experiment: &Experiment{
				Name:    "a",
				Default: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						Name:   "model1",
						Weight: 100,
					},
				},
			},
			expectedReferences: 0,
		},
		{
			name: "empty store remove pipeline experiment",
			store: &ExperimentStore{
				modelReferences:    map[string]map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{},
			},
			resourceName: "pipeline1",
			experiment: &Experiment{
				Name:         "a",
				Default:      getStrPtr("pipeline1"),
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name:   "pipeline1",
						Weight: 100,
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
				pipelineReferences: map[string]map[string]*Experiment{},
			},
			resourceName: "model1",
			experiment: &Experiment{
				Name:    "a",
				Default: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						Name:   "model1",
						Weight: 100,
					},
				},
			},
			expectedReferences: 0,
		},
		{
			name: "existing reference and remove experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{
					"pipeline1": {"a": nil},
				},
			},
			resourceName: "pipeline1",
			experiment: &Experiment{
				Name:         "a",
				Default:      getStrPtr("pipeline1"),
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name:   "pipeline1",
						Weight: 100,
					},
				},
			},
			expectedReferences: 0,
		},
		{
			name: "existing reference to another experiment and remove model experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{
					"model1": {"b": nil, "a": nil},
				},
				pipelineReferences: map[string]map[string]*Experiment{},
			},
			resourceName: "model1",
			experiment: &Experiment{
				Name:    "a",
				Default: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						Name:   "model1",
						Weight: 100,
					},
				},
			},
			expectedReferences: 1,
		},
		{
			name: "existing reference to another experiment and remove pipeline experiment",
			store: &ExperimentStore{
				modelReferences: map[string]map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{
					"pipeline1": {"b": nil, "a": nil},
				},
			},
			resourceName: "pipeline1",
			experiment: &Experiment{
				Name:         "a",
				Default:      getStrPtr("pipeline1"),
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name:   "pipeline1",
						Weight: 100,
					},
				},
			},
			expectedReferences: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.store.removeReference(test.resourceName, test.experiment)
			switch test.experiment.ResourceType {
			case PipelineResourceType:
				g.Expect(len(test.store.pipelineReferences[test.resourceName])).To(Equal(test.expectedReferences))
			default:
				g.Expect(len(test.store.modelReferences[test.resourceName])).To(Equal(test.expectedReferences))
			}
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
				logger:             logrus.New(),
				modelBaselines:     map[string]*Experiment{},
				modelReferences:    map[string]map[string]*Experiment{},
				pipelineBaselines:  map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{},
			},
			experiment:         &Experiment{Name: "a"},
			expectedReferences: 0,
		},
		{
			name: "existing model experiment",
			server: &ExperimentStore{
				logger:             logrus.New(),
				modelBaselines:     map[string]*Experiment{"model1": nil},
				modelReferences:    map[string]map[string]*Experiment{"model2": {"a": nil}, "model1": {"a": nil}},
				pipelineBaselines:  map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{},
				experiments: map[string]*Experiment{"a": {
					Name:         "a",
					ResourceType: ModelResourceType,
					Default:      getStrPtr("model1"),
					Candidates: []*Candidate{
						{
							Name: "model1",
						},
						{
							Name: "model2",
						},
					},
				}},
			},
			experiment: &Experiment{
				Name:         "a",
				ResourceType: ModelResourceType,
			},
			expectedReferences: 0,
		},
		{
			name: "existing pipeline experiment",
			server: &ExperimentStore{
				logger:             logrus.New(),
				modelBaselines:     map[string]*Experiment{},
				modelReferences:    map[string]map[string]*Experiment{},
				pipelineBaselines:  map[string]*Experiment{"pipeline1": nil},
				pipelineReferences: map[string]map[string]*Experiment{"pipeline2": {"a": nil}, "pipeline1": {"a": nil}},
				experiments: map[string]*Experiment{"a": {
					Name:         "a",
					ResourceType: PipelineResourceType,
					Default:      getStrPtr("pipeline1"),
					Candidates: []*Candidate{
						{
							Name: "pipeline1",
						},
						{
							Name: "pipeline2",
						},
					},
				}},
			},
			experiment: &Experiment{
				Name:         "a",
				ResourceType: PipelineResourceType,
			},
			expectedReferences: 0,
		},
		{
			name: "other experiments",
			server: &ExperimentStore{
				logger:             logrus.New(),
				modelBaselines:     map[string]*Experiment{"model1": nil},
				modelReferences:    map[string]map[string]*Experiment{"model2": {"a": nil, "b": nil}, "model1": {"a": nil, "b": nil}},
				pipelineBaselines:  map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{},
				experiments: map[string]*Experiment{"a": {
					Name:    "a",
					Default: getStrPtr("model1"),
					Candidates: []*Candidate{
						{
							Name: "model1",
						},
						{
							Name: "model2",
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
			switch test.experiment.ResourceType {
			case PipelineResourceType:
				g.Expect(test.server.pipelineBaselines[test.experiment.Name]).To(BeNil())
				g.Expect(test.server.getTotalPipelineReferences()).To(Equal(test.expectedReferences))
			default:
				g.Expect(test.server.modelBaselines[test.experiment.Name]).To(BeNil())
				g.Expect(test.server.getTotalModelReferences()).To(Equal(test.expectedReferences))
			}
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
			name: "existing model experiment but empty store",
			server: &ExperimentStore{
				logger:             logrus.New(),
				modelBaselines:     map[string]*Experiment{},
				modelReferences:    map[string]map[string]*Experiment{},
				pipelineBaselines:  map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{},
				experiments:        map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name:    "a",
				Default: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
					{
						Name: "model2",
					},
				},
			},
			expectedReferences: 2,
		},
		{
			name: "existing pipeline experiment but empty store",
			server: &ExperimentStore{
				logger:             logrus.New(),
				modelBaselines:     map[string]*Experiment{},
				modelReferences:    map[string]map[string]*Experiment{},
				pipelineBaselines:  map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{},
				experiments:        map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name:         "a",
				Default:      getStrPtr("pipeline1"),
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name: "pipeline1",
					},
					{
						Name: "pipeline2",
					},
				},
			},
			expectedReferences: 2,
		},
		{
			name: "existing pipeline experiment and pipeline in store",
			server: &ExperimentStore{
				logger:             logrus.New(),
				modelBaselines:     map[string]*Experiment{},
				modelReferences:    map[string]map[string]*Experiment{},
				pipelineBaselines:  map[string]*Experiment{"pipeline1": {Name: "a"}},
				pipelineReferences: map[string]map[string]*Experiment{"pipeline2": {"a": nil}, "pipeline1": {"a": nil}},
				experiments:        map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name:         "a",
				Default:      getStrPtr("pipeline1"),
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name: "pipeline1",
					},
					{
						Name: "pipeline2",
					},
				},
			},
			expectedReferences: 2,
		},
		{
			name: "existing model experiment and model in store",
			server: &ExperimentStore{
				logger:             logrus.New(),
				modelBaselines:     map[string]*Experiment{"model1": {Name: "a"}},
				modelReferences:    map[string]map[string]*Experiment{"model2": {"a": nil}, "model1": {"a": nil}},
				pipelineBaselines:  map[string]*Experiment{},
				pipelineReferences: map[string]map[string]*Experiment{},
				experiments:        map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name:    "a",
				Default: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
					{
						Name: "model2",
					},
				},
			},
			expectedReferences: 2,
		},
		{
			name: "other experiments",
			server: &ExperimentStore{
				logger:          logrus.New(),
				modelBaselines:  map[string]*Experiment{"model1": {Name: "a"}},
				modelReferences: map[string]map[string]*Experiment{"model2": {"b": nil}, "model1": {"b": nil}},
				experiments:     map[string]*Experiment{},
			},
			experiment: &Experiment{
				Name:    "a",
				Default: getStrPtr("model1"),
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
					{
						Name: "model2",
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
			switch test.experiment.ResourceType {
			case PipelineResourceType:
				g.Expect(test.server.getTotalPipelineReferences()).To(Equal(test.expectedReferences))
			default:
				g.Expect(test.server.getTotalModelReferences()).To(Equal(test.expectedReferences))
			}
		})
	}
}

func TestSetCandidateAndMirrorModelReadiness(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                    string
		experiment              *Experiment
		modelStates             map[string]store.ModelState
		expectedCandidatesReady bool
		expectedMirrorReady     bool
	}

	tests := []test{
		{
			name: "candidate ready as model is ready",
			experiment: &Experiment{
				Name: "a",
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
				},
			},
			modelStates:             map[string]store.ModelState{"model1": store.ModelAvailable},
			expectedCandidatesReady: true,
			expectedMirrorReady:     true,
		},
		{
			name: "candidates not ready as model is not ready",
			experiment: &Experiment{
				Name: "a",
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
				},
			},
			modelStates:             map[string]store.ModelState{"model1": store.ModelFailed},
			expectedCandidatesReady: false,
			expectedMirrorReady:     true,
		},
		{
			name: "multiple candidates only one ready",
			experiment: &Experiment{
				Name: "a",
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
					{
						Name: "model2",
					},
				},
			},
			modelStates:             map[string]store.ModelState{"model1": store.ModelAvailable},
			expectedCandidatesReady: false,
			expectedMirrorReady:     true,
		},
		{
			name: "multiple candidates all ready",
			experiment: &Experiment{
				Name: "a",
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
					{
						Name: "model2",
					},
				},
			},
			modelStates:             map[string]store.ModelState{"model1": store.ModelAvailable, "model2": store.ModelAvailable},
			expectedCandidatesReady: true,
			expectedMirrorReady:     true,
		},
		{
			name: "mirror and candidate ready as model is ready",
			experiment: &Experiment{
				Name: "a",
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
				},
				Mirror: &Mirror{
					Name: "model1",
				},
			},
			modelStates:             map[string]store.ModelState{"model1": store.ModelAvailable},
			expectedCandidatesReady: true,
			expectedMirrorReady:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrus.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			server := NewExperimentServer(logger, eventHub, fakeModelStore{status: test.modelStates}, nil)
			err = server.StartExperiment(test.experiment)
			g.Expect(err).To(BeNil())

			server.setCandidateAndMirrorReadiness(test.experiment)
			g.Expect(err).To(BeNil())
			g.Expect(test.experiment.AreCandidatesReady()).To(Equal(test.expectedCandidatesReady))
			g.Expect(test.experiment.IsMirrorReady()).To(Equal(test.expectedMirrorReady))
		})
	}
}

type fakePipelineStore struct {
	status map[string]pipeline.PipelineStatus
}

func (f fakePipelineStore) AddPipeline(pipeline *scheduler.Pipeline) error {
	panic("implement me")
}

func (f fakePipelineStore) RemovePipeline(name string) error {
	panic("implement me")
}

func (f fakePipelineStore) GetPipelineVersion(name string, version uint32, uid string) (*pipeline.PipelineVersion, error) {
	panic("implement me")
}

func (f fakePipelineStore) GetPipeline(name string) (*pipeline.Pipeline, error) {
	return &pipeline.Pipeline{
		Versions: []*pipeline.PipelineVersion{
			{
				State: &pipeline.PipelineState{
					Status: f.status[name],
				},
			},
		},
	}, nil
}

func (f fakePipelineStore) GetPipelines() ([]*pipeline.Pipeline, error) {
	panic("implement me")
}

func (f fakePipelineStore) SetPipelineState(name string, version uint32, uid string, state pipeline.PipelineStatus, reason string) error {
	panic("implement me")
}

func (f fakePipelineStore) GetAllRunningPipelineVersions() []coordinator.PipelineEventMsg {
	panic("implement me")
}

func TestSetCandidateAndMirrorPipelineReadiness(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                    string
		experiment              *Experiment
		pipelineStates          map[string]pipeline.PipelineStatus
		expectedCandidatesReady bool
		expectedMirrorReady     bool
	}

	tests := []test{
		{
			name: "candidate ready as pipeline is ready",
			experiment: &Experiment{
				Name:         "a",
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
				},
			},
			pipelineStates:          map[string]pipeline.PipelineStatus{"model1": pipeline.PipelineReady},
			expectedCandidatesReady: true,
			expectedMirrorReady:     true,
		},
		{
			name: "candidates not ready as pipeline is not ready",
			experiment: &Experiment{
				Name:         "a",
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name: "model1",
					},
				},
			},
			pipelineStates:          map[string]pipeline.PipelineStatus{"model1": pipeline.PipelineFailed},
			expectedCandidatesReady: false,
			expectedMirrorReady:     true,
		},
		{
			name: "multiple candidates only one ready",
			experiment: &Experiment{
				Name:         "a",
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name: "pipeline1",
					},
					{
						Name: "pipeline2",
					},
				},
			},
			pipelineStates:          map[string]pipeline.PipelineStatus{"pipeline1": pipeline.PipelineReady},
			expectedCandidatesReady: false,
			expectedMirrorReady:     true,
		},
		{
			name: "multiple candidates all ready",
			experiment: &Experiment{
				Name:         "a",
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name: "pipeline1",
					},
					{
						Name: "pipeline2",
					},
				},
			},
			pipelineStates:          map[string]pipeline.PipelineStatus{"pipeline1": pipeline.PipelineReady, "pipeline2": pipeline.PipelineReady},
			expectedCandidatesReady: true,
			expectedMirrorReady:     true,
		},
		{
			name: "mirror and candidate ready as pipeline is ready",
			experiment: &Experiment{
				Name:         "a",
				ResourceType: PipelineResourceType,
				Candidates: []*Candidate{
					{
						Name: "pipeline1",
					},
				},
				Mirror: &Mirror{
					Name: "pipeline1",
				},
			},
			pipelineStates:          map[string]pipeline.PipelineStatus{"pipeline1": pipeline.PipelineReady},
			expectedCandidatesReady: true,
			expectedMirrorReady:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrus.New()
			eventHub, err := coordinator.NewEventHub(logger)
			g.Expect(err).To(BeNil())
			server := NewExperimentServer(logger, eventHub, nil, fakePipelineStore{status: test.pipelineStates})
			err = server.StartExperiment(test.experiment)
			g.Expect(err).To(BeNil())

			server.setCandidateAndMirrorReadiness(test.experiment)
			g.Expect(err).To(BeNil())
			g.Expect(test.experiment.AreCandidatesReady()).To(Equal(test.expectedCandidatesReady))
			g.Expect(test.experiment.IsMirrorReady()).To(Equal(test.expectedMirrorReady))
		})
	}
}

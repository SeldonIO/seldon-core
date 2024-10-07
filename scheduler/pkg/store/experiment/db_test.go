/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package experiment

import (
	"fmt"
	"testing"
	"time"

	"github.com/dgraph-io/badger/v3"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/utils"
)

// this is legacy for migration testing
func createExperimentProto(experiment *Experiment) *scheduler.Experiment {
	var candidates []*scheduler.ExperimentCandidate
	for _, candidate := range experiment.Candidates {
		candidates = append(candidates, &scheduler.ExperimentCandidate{
			Name:   candidate.Name,
			Weight: candidate.Weight,
		})
	}
	var mirror *scheduler.ExperimentMirror
	if experiment.Mirror != nil {
		mirror = &scheduler.ExperimentMirror{
			Name:    experiment.Mirror.Name,
			Percent: experiment.Mirror.Percent,
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
	var resourceType scheduler.ResourceType
	switch experiment.ResourceType {
	case PipelineResourceType:
		resourceType = scheduler.ResourceType_PIPELINE
	case ModelResourceType:
		resourceType = scheduler.ResourceType_MODEL
	}
	return &scheduler.Experiment{
		Name:           experiment.Name,
		Default:        experiment.Default,
		ResourceType:   resourceType,
		Candidates:     candidates,
		Mirror:         mirror,
		Config:         config,
		KubernetesMeta: k8sMeta,
	}
}

func TestSaveWithTTL(t *testing.T) {
	g := NewGomegaWithT(t)

	experiment := &Experiment{
		Name: "test1",
		Candidates: []*Candidate{
			{
				Name:   "model1",
				Weight: 50,
			},
			{
				Name:   "model2",
				Weight: 50,
			},
		},
		Mirror: &Mirror{
			Name:    "model3",
			Percent: 90,
		},
		Config: &Config{
			StickySessions: true,
		},
		KubernetesMeta: &KubernetesMeta{
			Namespace:  "default",
			Generation: 2,
		},
		Deleted: true,
	}

	ttl := time.Duration(time.Second)

	path := fmt.Sprintf("%s/db", t.TempDir())
	logger := log.New()
	db, err := newExperimentDbManager(getExperimentDbFolder(path), logger)
	g.Expect(err).To(BeNil())
	experiment.DeletedAt = time.Now().Add(ttl - utils.DeletedResourceTTL)
	err = db.save(experiment)
	g.Expect(err).To(BeNil())

	persistedExp, err := db.get(experiment.Name)
	g.Expect(err).To(BeNil())
	g.Expect(persistedExp).NotTo(BeNil())

	time.Sleep(ttl * 2)

	persistedExp, err = db.get(experiment.Name)
	g.Expect(err).ToNot(BeNil())
	g.Expect(persistedExp).To(BeNil())

	err = db.Stop()
	g.Expect(err).To(BeNil())
}

func TestSaveAndRestore(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name        string
		experiments []*Experiment
		errors      []bool
	}

	getStrPtr := func(val string) *string { return &val }

	tests := []test{
		{
			name: "basic model experiment",
			experiments: []*Experiment{
				{
					Name: "test1",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			errors: []bool{false},
		},
		{
			name: "basic 2 model experiment",
			experiments: []*Experiment{
				{
					Name: "test1",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
				{
					Name: "test2",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			errors: []bool{false, false},
		},
		{
			name: "basic pipeline experiment",
			experiments: []*Experiment{
				{
					Name:         "test1",
					ResourceType: PipelineResourceType,
					Candidates: []*Candidate{
						{
							Name:   "pipeline1",
							Weight: 50,
						},
						{
							Name:   "pipeline2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "pipeline3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			errors: []bool{false},
		},
		{
			name: "faulty experiment",
			experiments: []*Experiment{
				{
					Name:         "test1",
					ResourceType: ModelResourceType,
					Default:      getStrPtr("model1"),
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
				{ // duplicate default
					Name:         "test2",
					ResourceType: ModelResourceType,
					Default:      getStrPtr("model1"),
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model3",
							Weight: 50,
						},
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			errors: []bool{false, true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())
			logger := log.New()
			db, err := newExperimentDbManager(getExperimentDbFolder(path), logger)
			g.Expect(err).To(BeNil())
			for _, p := range test.experiments {
				err := db.save(p)
				g.Expect(err).To(BeNil())
			}
			err = db.Stop()
			g.Expect(err).To(BeNil())

			es := NewExperimentServer(log.New(), nil, nil, nil)
			err = es.InitialiseOrRestoreDB(path)
			g.Expect(err).To(BeNil())
			for idx, p := range test.experiments {
				if !test.errors[idx] {
					g.Expect(cmp.Equal(p, es.experiments[p.Name])).To(BeTrue())
				}
			}
		})
	}
}

func TestSaveAndRestoreDeletedExperiments(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name       string
		experiment Experiment
		withTTL    bool
	}

	getStrPtr := func(val string) *string { return &val }

	createDeletedExperiment := func(name string) Experiment {
		return Experiment{
			Name:         name,
			ResourceType: ModelResourceType,
			Default:      getStrPtr("model1"),
			Candidates: []*Candidate{
				{
					Name:   "model1",
					Weight: 50,
				},
				{
					Name:   "model2",
					Weight: 50,
				},
			},
			KubernetesMeta: &KubernetesMeta{
				Namespace:  "default",
				Generation: 2,
			},
			Deleted: true,
		}
	}

	tests := []test{
		{
			name:       "deleted experiment with ttl does not exist after restore",
			experiment: createDeletedExperiment("with-ttl"),
			withTTL:    true,
		},
		{
			name:       "deleted experiment without ttl does exist after restore",
			experiment: createDeletedExperiment("without-ttl"),
			withTTL:    false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(test.experiment.Deleted).To(BeTrue(), "this is a test for deleted experiments")
			path := fmt.Sprintf("%s/db", t.TempDir())
			logger := log.New()
			edb, err := newExperimentDbManager(getExperimentDbFolder(path), logger)
			g.Expect(err).To(BeNil())
			if !test.withTTL {
				err := edb.save(&test.experiment)
				g.Expect(err).To(BeNil())
			} else {
				test.experiment.DeletedAt = time.Now().Add(-utils.DeletedResourceTTL)
				err := edb.save(&test.experiment)
				g.Expect(err).To(BeNil())
			}
			err = edb.Stop()
			g.Expect(err).To(BeNil())

			es := NewExperimentServer(log.New(), nil, nil, nil)
			err = es.InitialiseOrRestoreDB(path)
			g.Expect(err).To(BeNil())

			if !test.withTTL {
				var item *badger.Item
				err = es.db.db.View(func(txn *badger.Txn) error {
					item, err = txn.Get(([]byte(test.experiment.Name)))
					return err
				})
				g.Expect(err).To(BeNil())
				g.Expect(item.ExpiresAt()).ToNot(BeZero())
				g.Expect(es.experiments[test.experiment.Name]).ToNot(BeNil())
			} else {
				g.Expect(es.experiments[test.experiment.Name]).To(BeNil())
			}
		})
	}
}

func TestGetExperimentFromDB(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name           string
		experiments    []*Experiment
		experimentName string
		isErr          bool
	}

	tests := []test{
		{
			name: "basic 2 model experiment",
			experiments: []*Experiment{
				{
					Name: "test1",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
				{
					Name: "test2",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			experimentName: "test2",
			isErr:          false,
		},
		{
			name: "Experiment not found",
			experiments: []*Experiment{
				{
					Name:         "test1",
					ResourceType: ModelResourceType,
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
				{
					Name:         "test2",
					ResourceType: ModelResourceType,
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model3",
							Weight: 50,
						},
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			experimentName: "test3",
			isErr:          true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())
			logger := log.New()
			db, err := newExperimentDbManager(getExperimentDbFolder(path), logger)
			g.Expect(err).To(BeNil())
			for _, p := range test.experiments {
				err := db.save(p)
				g.Expect(err).To(BeNil())
			}

			actualExperiment, err := db.get(test.experimentName)
			if test.isErr {
				g.Expect(err).ToNot(BeNil())
				g.Expect(actualExperiment).To(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(actualExperiment).ToNot(BeNil())
				expFound := false
				for _, exp := range test.experiments {
					if exp.Name == test.experimentName {
						g.Expect(cmp.Equal(exp, actualExperiment)).To(BeTrue())
						expFound = true
					}
				}
				g.Expect(expFound).To(BeTrue())
			}
			err = db.Stop()
			g.Expect(err).To(BeNil())
		})
	}
}

func TestDeleteExperimentFromDB(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name           string
		experiments    []*Experiment
		experimentName string
	}

	tests := []test{
		{
			name: "basic 2 model experiment",
			experiments: []*Experiment{
				{
					Name: "test1",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
				{
					Name: "test2",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			experimentName: "test2",
		},
		{
			name: "Experiment not found",
			experiments: []*Experiment{
				{
					Name:         "test1",
					ResourceType: ModelResourceType,
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
				{
					Name:         "test2",
					ResourceType: ModelResourceType,
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model3",
							Weight: 50,
						},
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
			experimentName: "test3",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())
			logger := log.New()
			db, err := newExperimentDbManager(getExperimentDbFolder(path), logger)
			g.Expect(err).To(BeNil())
			for _, p := range test.experiments {
				err := db.save(p)
				g.Expect(err).To(BeNil())
			}

			err = db.delete(test.experimentName)
			g.Expect(err).To(BeNil())

			_, err = db.get(test.experimentName)
			g.Expect(err).ToNot(BeNil()) // key not found

			err = db.Stop()
			g.Expect(err).To(BeNil())
		})
	}
}

func TestMigrateFromV1ToV2(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name        string
		experiments []*Experiment
	}

	tests := []test{
		{
			name: "basic model experiment",
			experiments: []*Experiment{
				{
					Name: "test1",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
		},
		{
			name: "basic 2 model experiment",
			experiments: []*Experiment{
				{
					Name: "test1",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
				{
					Name: "test2",
					Candidates: []*Candidate{
						{
							Name:   "model1",
							Weight: 50,
						},
						{
							Name:   "model2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "model3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
		},
		{
			name: "basic pipeline experiment",
			experiments: []*Experiment{
				{
					Name:         "test1",
					ResourceType: PipelineResourceType,
					Candidates: []*Candidate{
						{
							Name:   "pipeline1",
							Weight: 50,
						},
						{
							Name:   "pipeline2",
							Weight: 50,
						},
					},
					Mirror: &Mirror{
						Name:    "pipeline3",
						Percent: 90,
					},
					Config: &Config{
						StickySessions: true,
					},
					KubernetesMeta: &KubernetesMeta{
						Namespace:  "default",
						Generation: 2,
					},
				},
			},
		},
	}

	saveLegacyFn := func(experiment *Experiment, db *badger.DB) error {
		experimentProto := createExperimentProto(experiment)
		experimentBytes, err := proto.Marshal(experimentProto)
		if err != nil {
			return err
		}
		return db.Update(func(txn *badger.Txn) error {
			err = txn.Set([]byte(experiment.Name), experimentBytes)
			return err
		})
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())
			logger := log.New()
			db, err := utils.Open(getExperimentDbFolder(path), logger, "experimentDb")
			g.Expect(err).To(BeNil())
			for _, p := range test.experiments {
				err := saveLegacyFn(p, db)
				g.Expect(err).To(BeNil())
			}
			_ = db.Close()

			// migrate
			edb, err := newExperimentDbManager(getExperimentDbFolder(path), logger)
			g.Expect(err).To(BeNil())

			g.Expect(err).To(BeNil())
			version, err := edb.getVersion()
			g.Expect(err).To(BeNil())
			g.Expect(version).To(Equal(currentExperimentSnapshotVersion))
			err = edb.Stop()
			g.Expect(err).To(BeNil())

			// check that we have no experiments in the db format
			es := NewExperimentServer(log.New(), nil, nil, nil)
			err = es.InitialiseOrRestoreDB(path)
			g.Expect(err).To(BeNil())
			g.Expect(len(es.experiments)).To(Equal(0))
		})
	}
}

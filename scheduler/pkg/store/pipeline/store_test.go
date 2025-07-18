/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package pipeline

import (
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

func TestAddPipeline(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name            string
		proto           *scheduler.Pipeline
		store           *PipelineStore
		expectedVersion uint32
		err             error
	}

	tests := []test{
		{
			name: "add pipeline none existing",
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "step1",
						Inputs: []string{},
					},
				},
				Output: &scheduler.PipelineOutput{
					Steps: []string{"step1.outputs"},
				},
			},
			store: &PipelineStore{
				logger:    logrus.New(),
				pipelines: map[string]*Pipeline{},
				modelStatusHandler: ModelStatusHandler{
					modelReferences: map[string]map[string]void{},
					store:           fakeModelStore{status: map[string]store.ModelState{}},
				},
			},
			expectedVersion: 1,
		},
		{
			name: "add pipeline none existing with k8s meta",
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "step1",
						Inputs: []string{},
					},
				},
				Output: &scheduler.PipelineOutput{
					Steps: []string{"step1.outputs"},
				},
				KubernetesMeta: &scheduler.KubernetesMeta{
					Namespace:  "default",
					Generation: 1,
				},
			},
			store: &PipelineStore{
				logger:    logrus.New(),
				pipelines: map[string]*Pipeline{},
				modelStatusHandler: ModelStatusHandler{
					modelReferences: map[string]map[string]void{},
					store:           fakeModelStore{status: map[string]store.ModelState{}},
				},
			},
			expectedVersion: 1,
		},
		{
			name: "version added",
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "step1",
						Inputs: []string{},
					},
				},
				Output: &scheduler.PipelineOutput{
					Steps: []string{"step1.outputs"},
				},
			},
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"pipeline": {
						Name:        "pipeline",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "pipeline",
								Version: 1,
								State: &PipelineState{
									Status: PipelineCreate,
								},
							},
						},
					},
				},
				modelStatusHandler: ModelStatusHandler{
					modelReferences: map[string]map[string]void{},
					store:           fakeModelStore{status: map[string]store.ModelState{}},
				},
			},
			expectedVersion: 2,
		},
		{
			name: "version added when previous terminated",
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "step1",
						Inputs: []string{},
					},
				},
			},
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"pipeline": {
						Name:        "pipeline",
						LastVersion: 1,
						Deleted:     true,
						Versions: []*PipelineVersion{
							1: {
								Name:    "pipeline",
								Version: 1,
								State: &PipelineState{
									Status: PipelineTerminated,
								},
							},
						},
					},
				},
				modelStatusHandler: ModelStatusHandler{
					modelReferences: map[string]map[string]void{},
					store:           fakeModelStore{status: map[string]store.ModelState{}},
				},
			},
			expectedVersion: 1,
		},
		{
			name: "version added when previous terminating",
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "step1",
						Inputs: []string{},
					},
				},
			},
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"pipeline": {
						Name:        "pipeline",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							1: {
								Name:    "pipeline",
								Version: 1,
								State: &PipelineState{
									Status: PipelineTerminating,
								},
							},
						},
					},
				},
				modelStatusHandler: ModelStatusHandler{
					modelReferences: map[string]map[string]void{},
					store:           fakeModelStore{status: map[string]store.ModelState{}},
				},
			},
			expectedVersion: 1,
		},
		{
			name: "version ok when previous terminate state",
			proto: &scheduler.Pipeline{
				Name: "pipeline",
				Steps: []*scheduler.PipelineStep{
					{
						Name:   "step1",
						Inputs: []string{},
					},
				},
			},
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"pipeline": {
						Name:        "pipeline",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							1: {
								Name:    "pipeline",
								Version: 1,
								State: &PipelineState{
									Status: PipelineTerminate,
								},
							},
						},
					},
				},
				modelStatusHandler: ModelStatusHandler{
					modelReferences: map[string]map[string]void{},
					store:           fakeModelStore{status: map[string]store.ModelState{}},
				},
			},
			expectedVersion: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrus.New()
			path := fmt.Sprintf("%s/db", t.TempDir())
			db, _ := newPipelineDbManager(getPipelineDbFolder(path), logger, 10)
			test.store.db = db
			err := test.store.AddPipeline(test.proto)
			if test.err == nil {
				p := test.store.pipelines[test.proto.Name]
				g.Expect(p).ToNot(BeNil())
				g.Expect(p.LastVersion).To(Equal(test.expectedVersion))
				pv := p.GetLatestPipelineVersion()
				g.Expect(pv).ToNot(BeNil())
				g.Expect(pv.Version).To(Equal(test.expectedVersion))
				g.Expect(pv.UID).ToNot(Equal(""))
				g.Expect(pv.Name).To(Equal(test.proto.Name))
				g.Expect(pv.State.Status).To(Equal(PipelineCreate))

				// check db
				pipelineFromDB, _ := test.store.db.get(test.proto.Name)
				g.Expect(pipelineFromDB.Deleted).To(Equal(p.Deleted))
				g.Expect(pipelineFromDB.LastVersion).To(Equal(p.LastVersion))
				g.Expect(len(pipelineFromDB.Versions)).To(Equal(len(p.Versions)))
				g.Expect(len(pipelineFromDB.Versions)).To(Equal(len(p.Versions)))
			} else {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
				pipelineFromDB, _ := test.store.db.get(test.proto.Name)
				g.Expect(pipelineFromDB).To(BeNil())
			}
		})
	}
}

func TestRemovePipeline(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name         string
		pipelineName string
		store        *PipelineStore
		err          error
	}

	tests := []test{
		{
			name:         "not found err",
			pipelineName: "pipeline",
			store: &PipelineStore{
				logger:    logrus.New(),
				pipelines: map[string]*Pipeline{},
			},
			err: &PipelineNotFoundErr{pipeline: "pipeline"},
		},
		{
			name:         "pipeline terminating err",
			pipelineName: "pipeline",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"pipeline": {
						Name:        "pipeline",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "pipeline",
								Version: 1,
								State: &PipelineState{
									Status: PipelineTerminating,
								},
							},
						},
					},
				},
			},
			err: &PipelineTerminatingErr{pipeline: "pipeline"},
		},
		{
			name:         "pipeline terminated err",
			pipelineName: "pipeline",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"pipeline": {
						Name:        "pipeline",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "pipeline",
								Version: 1,
								State: &PipelineState{
									Status: PipelineTerminated,
								},
							},
						},
					},
				},
			},
			err: &PipelineAlreadyTerminatedErr{pipeline: "pipeline"},
		},
		{
			name:         "deleted ok",
			pipelineName: "pipeline",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"pipeline": {
						Name:        "pipeline",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "pipeline",
								Version: 1,
								State: &PipelineState{
									Status: PipelineReady,
								},
							},
						},
					},
				},
			},
		},
		{
			name:         "deleted ok",
			pipelineName: "pipeline",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"pipeline": {
						Name:        "pipeline",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "pipeline",
								Version: 1,
								State: &PipelineState{
									Status: PipelineCreating,
								},
							},
							{
								Name:    "pipeline",
								Version: 2,
								State: &PipelineState{
									Status: PipelineCreating,
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrus.New()
			path := fmt.Sprintf("%s/db", t.TempDir())
			db, _ := newPipelineDbManager(getPipelineDbFolder(path), logger, 1)
			test.store.db = db
			err := test.store.RemovePipeline(test.pipelineName)
			if test.err == nil {
				p := test.store.pipelines[test.pipelineName]
				g.Expect(p).ToNot(BeNil())
				pv := p.GetLatestPipelineVersion()
				g.Expect(pv).ToNot(BeNil())
				g.Expect(pv.State.Status).To(Equal(PipelineTerminate))

				// check db contains pipeline
				pipelines := map[string]*Pipeline{}
				restoreCb := func(pipeline *Pipeline) {
					pipelines[pipeline.Name] = pipeline
				}
				_ = test.store.db.restore(restoreCb)
				actualPipeline := pipelines[test.pipelineName]
				expectedPipeline := test.store.pipelines[test.pipelineName]
				// TODO: check all fields
				g.Expect(actualPipeline.Deleted).To(Equal(expectedPipeline.Deleted))
				g.Expect(actualPipeline.LastVersion).To(Equal(expectedPipeline.LastVersion))
				g.Expect(len(actualPipeline.Versions)).To(Equal(len(expectedPipeline.Versions)))
				g.Expect(len(actualPipeline.Versions)).To(Equal(len(expectedPipeline.Versions)))
				time.Sleep(1 * time.Second)
				test.store.cleanupDeletedPipelines()
				g.Expect(test.store.pipelines[test.pipelineName]).To(BeNil())
			} else {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
			}
			_ = test.store.db.Stop()
		})
	}
}

func TestGetPipelineVersion(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name            string
		pipelineName    string
		pipelineVersion uint32
		uid             string
		store           *PipelineStore
		err             error
	}

	tests := []test{
		{
			name:            "ok",
			pipelineName:    "p",
			pipelineVersion: 1,
			uid:             "uid",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"p": {
						Name:        "p",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "p",
								Version: 1,
								UID:     "uid",
								State: &PipelineState{
									Status: PipelineReady,
								},
							},
						},
					},
				},
			},
		},
		{
			name:            "name not found",
			pipelineName:    "p2",
			pipelineVersion: 1,
			uid:             "uid",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"p": {
						Name:        "p",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "p",
								Version: 1,
								UID:     "uid",
								State: &PipelineState{
									Status: PipelineReady,
								},
							},
						},
					},
				},
			},
			err: &PipelineNotFoundErr{pipeline: "p2"},
		},
		{
			name:            "version not found",
			pipelineName:    "p",
			pipelineVersion: 2,
			uid:             "uid",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"p": {
						Name:        "p",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "p",
								Version: 1,
								UID:     "uid",
								State: &PipelineState{
									Status: PipelineReady,
								},
							},
						},
					},
				},
			},
			err: &PipelineVersionNotFoundErr{pipeline: "p", version: 2},
		},
		{
			name:            "uid not found",
			pipelineName:    "p",
			pipelineVersion: 1,
			uid:             "uid2",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"p": {
						Name:        "p",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "p",
								Version: 1,
								UID:     "uid",
								State: &PipelineState{
									Status: PipelineReady,
								},
							},
						},
					},
				},
			},
			err: &PipelineVersionUidMismatchErr{pipeline: "p", version: 1, uidExpected: "uid2", uidActual: "uid"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			pv, err := test.store.GetPipelineVersion(test.pipelineName, test.pipelineVersion, test.uid)
			if test.err == nil {
				g.Expect(err).To(BeNil())
				g.Expect(pv).ToNot(BeNil())
				g.Expect(pv.Name).To(Equal(test.pipelineName))
				g.Expect(pv.Version).To(Equal(test.pipelineVersion))
			} else {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
			}
		})
	}
}

func TestSetPipelineState(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name                         string
		pipelineName                 string
		pipelineVersion              uint32
		uid                          string
		status                       PipelineStatus
		reason                       string
		store                        *PipelineStore
		err                          error
		expectedPipelineVersionStats []PipelineStatus
	}

	tests := []test{
		{
			name:            "ok",
			pipelineName:    "p",
			pipelineVersion: 1,
			uid:             "uid",
			status:          PipelineFailed,
			reason:          "failed",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"p": {
						Name:        "p",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "p",
								Version: 1,
								UID:     "uid",
								State: &PipelineState{
									Status: PipelineReady,
								},
							},
						},
					},
				},
			},
			expectedPipelineVersionStats: []PipelineStatus{PipelineFailed},
		},
		{
			name:            "ready with previous pipeline",
			pipelineName:    "p",
			pipelineVersion: 3,
			uid:             "uid3",
			status:          PipelineReady,
			reason:          "",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"p": {
						Name:        "p",
						LastVersion: 3,
						Versions: []*PipelineVersion{
							{
								Name:    "p",
								Version: 1,
								UID:     "uid1",
								State: &PipelineState{
									Status: PipelineFailed,
								},
							},
							{
								Name:    "p",
								Version: 2,
								UID:     "uid2",
								State: &PipelineState{
									Status: PipelineReady,
								},
							},
							{
								Name:    "p",
								Version: 3,
								UID:     "uid3",
								State: &PipelineState{
									Status: PipelineCreate,
								},
							},
						},
					},
				},
			},
			expectedPipelineVersionStats: []PipelineStatus{PipelineTerminate, PipelineTerminate, PipelineReady},
		},
		{
			name:            "ready with previous pipeline terminating",
			pipelineName:    "p",
			pipelineVersion: 2,
			uid:             "uid2",
			status:          PipelineReady,
			reason:          "",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"p": {
						Name:        "p",
						LastVersion: 2,
						Versions: []*PipelineVersion{
							{
								Name:    "p",
								Version: 1,
								UID:     "uid1",
								State: &PipelineState{
									Status: PipelineTerminating,
								},
							},
							{
								Name:    "p",
								Version: 2,
								UID:     "uid2",
								State: &PipelineState{
									Status: PipelineCreate,
								},
							},
						},
					},
				},
			},
			expectedPipelineVersionStats: []PipelineStatus{PipelineTerminating, PipelineReady},
		},
		{
			name:            "name not found",
			pipelineName:    "p2",
			pipelineVersion: 1,
			uid:             "uid",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"p": {
						Name:        "p",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "p",
								Version: 1,
								UID:     "uid",
								State: &PipelineState{
									Status: PipelineReady,
								},
							},
						},
					},
				},
			},
			err: &PipelineNotFoundErr{pipeline: "p2"},
		},
		{
			name:            "version not found",
			pipelineName:    "p",
			pipelineVersion: 2,
			uid:             "uid",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"p": {
						Name:        "p",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "p",
								Version: 1,
								UID:     "uid",
								State: &PipelineState{
									Status: PipelineReady,
								},
							},
						},
					},
				},
			},
			err: &PipelineVersionNotFoundErr{pipeline: "p", version: 2},
		},
		{
			name:            "uid mismatch",
			pipelineName:    "p",
			pipelineVersion: 1,
			uid:             "uid",
			store: &PipelineStore{
				logger: logrus.New(),
				pipelines: map[string]*Pipeline{
					"p": {
						Name:        "p",
						LastVersion: 1,
						Versions: []*PipelineVersion{
							{
								Name:    "p",
								Version: 1,
								UID:     "uid2",
								State: &PipelineState{
									Status: PipelineReady,
								},
							},
						},
					},
				},
			},
			err: &PipelineVersionUidMismatchErr{pipeline: "p", version: 1, uidActual: "uid2", uidExpected: "uid"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.store.SetPipelineState(test.pipelineName, test.pipelineVersion, test.uid, test.status, test.reason, "")
			if test.err == nil {
				g.Expect(err).To(BeNil())
				pv, err := test.store.GetPipelineVersion(test.pipelineName, test.pipelineVersion, test.uid)
				g.Expect(err).To(BeNil())
				g.Expect(pv).ToNot(BeNil())
				g.Expect(pv.Name).To(Equal(test.pipelineName))
				g.Expect(pv.Version).To(Equal(test.pipelineVersion))
				g.Expect(pv.UID).To(Equal(test.uid))
				g.Expect(pv.State.Status).To(Equal(test.status))
				g.Expect(pv.State.Reason).To(Equal(test.reason))
				for idx, pv := range test.store.pipelines[test.pipelineName].Versions {
					g.Expect(test.expectedPipelineVersionStats[idx]).To(Equal(pv.State.Status))
				}
			} else {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
			}
		})
	}
}

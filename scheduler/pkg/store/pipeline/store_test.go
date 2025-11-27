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

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/coordinator"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

func TestGetPipelinesPipelineGwStatus(t *testing.T) {
	tests := []struct {
		name             string
		pipelines        map[string]*Pipeline
		queryStatus      PipelineStatus
		expectedCount    int
		expectedNames    []string
		expectedVersions []uint32
		expectedUIDs     []string
		validate         func(g *WithT, events []coordinator.PipelineEventMsg)
	}{
		{
			name:          "empty pipelines map",
			pipelines:     make(map[string]*Pipeline),
			queryStatus:   PipelineReady,
			expectedCount: 0,
		},
		{
			name: "no matching status",
			pipelines: map[string]*Pipeline{
				"test-pipeline": {
					Name:        "test-pipeline",
					LastVersion: 1,
					Versions: []*PipelineVersion{
						{
							Name:    "test-pipeline",
							Version: 1,
							UID:     "uid-1",
							State: &PipelineState{
								Status:           PipelineReady,
								PipelineGwStatus: PipelineCreating,
							},
						},
					},
				},
			},
			queryStatus:   PipelineReady,
			expectedCount: 0,
		},
		{
			name: "single matching pipeline",
			pipelines: map[string]*Pipeline{
				"test-pipeline": {
					Name:        "test-pipeline",
					LastVersion: 1,
					Versions: []*PipelineVersion{
						{
							Name:    "test-pipeline",
							Version: 1,
							UID:     "uid-1",
							State: &PipelineState{
								Status:           PipelineReady,
								PipelineGwStatus: PipelineReady,
							},
						},
					},
				},
			},
			queryStatus:      PipelineReady,
			expectedCount:    1,
			expectedNames:    []string{"test-pipeline"},
			expectedVersions: []uint32{1},
			expectedUIDs:     []string{"uid-1"},
		},
		{
			name: "multiple matching pipelines",
			pipelines: map[string]*Pipeline{
				"pipeline-1": {
					Name:        "pipeline-1",
					LastVersion: 1,
					Versions: []*PipelineVersion{
						{
							Name:    "pipeline-1",
							Version: 1,
							UID:     "uid-1",
							State: &PipelineState{
								Status:           PipelineReady,
								PipelineGwStatus: PipelineCreating,
							},
						},
					},
				},
				"pipeline-2": {
					Name:        "pipeline-2",
					LastVersion: 1,
					Versions: []*PipelineVersion{
						{
							Name:    "pipeline-2",
							Version: 1,
							UID:     "uid-2",
							State: &PipelineState{
								Status:           PipelineReady,
								PipelineGwStatus: PipelineCreating,
							},
						},
					},
				},
			},
			queryStatus:   PipelineCreating,
			expectedCount: 2,
			validate: func(g *WithT, events []coordinator.PipelineEventMsg) {
				pipelineNames := []string{events[0].PipelineName, events[1].PipelineName}
				g.Expect(pipelineNames).To(ContainElement("pipeline-1"))
				g.Expect(pipelineNames).To(ContainElement("pipeline-2"))
			},
		},
		{
			name: "mixed statuses - only return matching",
			pipelines: map[string]*Pipeline{
				"ready-pipeline": {
					Name:        "ready-pipeline",
					LastVersion: 1,
					Versions: []*PipelineVersion{
						{
							Name:    "ready-pipeline",
							Version: 1,
							UID:     "uid-ready",
							State: &PipelineState{
								Status:           PipelineReady,
								PipelineGwStatus: PipelineReady,
							},
						},
					},
				},
				"creating-pipeline": {
					Name:        "creating-pipeline",
					LastVersion: 1,
					Versions: []*PipelineVersion{
						{
							Name:    "creating-pipeline",
							Version: 1,
							UID:     "uid-creating",
							State: &PipelineState{
								Status:           PipelineCreating,
								PipelineGwStatus: PipelineCreating,
							},
						},
					},
				},
				"terminating-pipeline": {
					Name:        "terminating-pipeline",
					LastVersion: 1,
					Versions: []*PipelineVersion{
						{
							Name:    "terminating-pipeline",
							Version: 1,
							UID:     "uid-terminating",
							State: &PipelineState{
								Status:           PipelineTerminating,
								PipelineGwStatus: PipelineTerminating,
							},
						},
					},
				},
			},
			queryStatus:      PipelineReady,
			expectedCount:    1,
			expectedNames:    []string{"ready-pipeline"},
			expectedVersions: []uint32{1},
			expectedUIDs:     []string{"uid-ready"},
		},
		{
			name: "multiple versions - return latest",
			pipelines: map[string]*Pipeline{
				"test-pipeline": {
					Name:        "test-pipeline",
					LastVersion: 3,
					Versions: []*PipelineVersion{
						{
							Name:    "test-pipeline",
							Version: 1,
							UID:     "uid-1",
							State: &PipelineState{
								Status:           PipelineTerminated,
								PipelineGwStatus: PipelineTerminated,
							},
						},
						{
							Name:    "test-pipeline",
							Version: 2,
							UID:     "uid-2",
							State: &PipelineState{
								Status:           PipelineTerminated,
								PipelineGwStatus: PipelineTerminated,
							},
						},
						{
							Name:    "test-pipeline",
							Version: 3,
							UID:     "uid-3",
							State: &PipelineState{
								Status:           PipelineReady,
								PipelineGwStatus: PipelineReady,
							},
						},
					},
				},
			},
			queryStatus:      PipelineReady,
			expectedCount:    1,
			expectedNames:    []string{"test-pipeline"},
			expectedVersions: []uint32{3},
			expectedUIDs:     []string{"uid-3"},
		},
		{
			name: "check PipelineGwStatus not Status",
			pipelines: map[string]*Pipeline{
				"test-pipeline": {
					Name:        "test-pipeline",
					LastVersion: 1,
					Versions: []*PipelineVersion{
						{
							Name:    "test-pipeline",
							Version: 1,
							UID:     "uid-1",
							State: &PipelineState{
								Status:           PipelineReady,
								PipelineGwStatus: PipelineCreating,
							},
						},
					},
				},
			},
			queryStatus:      PipelineCreating,
			expectedCount:    1,
			expectedNames:    []string{"test-pipeline"},
			expectedVersions: []uint32{1},
			expectedUIDs:     []string{"uid-1"},
		},
		{
			name: "pipeline with no versions",
			pipelines: map[string]*Pipeline{
				"test-pipeline": {
					Name:        "test-pipeline",
					LastVersion: 0,
					Versions:    []*PipelineVersion{},
				},
			},
			queryStatus:   PipelineReady,
			expectedCount: 0,
		},
		{
			name: "status matches but PipelineGwStatus doesn't",
			pipelines: map[string]*Pipeline{
				"pipeline-match-both": {
					Name:        "pipeline-match-both",
					LastVersion: 1,
					Versions: []*PipelineVersion{
						{
							Name:    "pipeline-match-both",
							Version: 1,
							UID:     "uid-match",
							State: &PipelineState{
								Status:           PipelineReady,
								PipelineGwStatus: PipelineReady,
							},
						},
					},
				},
				"pipeline-match-status-only": {
					Name:        "pipeline-match-status-only",
					LastVersion: 1,
					Versions: []*PipelineVersion{
						{
							Name:    "pipeline-match-status-only",
							Version: 1,
							UID:     "uid-no-match",
							State: &PipelineState{
								Status:           PipelineReady,
								PipelineGwStatus: PipelineCreating,
							},
						},
					},
				},
			},
			queryStatus:      PipelineReady,
			expectedCount:    1,
			expectedNames:    []string{"pipeline-match-both"},
			expectedVersions: []uint32{1},
			expectedUIDs:     []string{"uid-match"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			s := &PipelineStore{
				logger:    logrus.New(),
				pipelines: tt.pipelines,
			}

			events := s.GetPipelinesPipelineGwStatus(tt.queryStatus)

			g.Expect(events).To(HaveLen(tt.expectedCount))

			if tt.validate != nil {
				tt.validate(g, events)
			} else if tt.expectedCount > 0 {
				// Default validation for single expected result
				for i := 0; i < tt.expectedCount && i < len(tt.expectedNames); i++ {
					g.Expect(events[i].PipelineName).To(Equal(tt.expectedNames[i]))
					if len(tt.expectedVersions) > i {
						g.Expect(events[i].PipelineVersion).To(Equal(tt.expectedVersions[i]))
					}
					if len(tt.expectedUIDs) > i {
						g.Expect(events[i].UID).To(Equal(tt.expectedUIDs[i]))
					}
				}
			}
		})
	}
}

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
			name:         "pipeline terminating err - (terminating, terminating)",
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
									Status:           PipelineTerminating,
									PipelineGwStatus: PipelineTerminating,
								},
							},
						},
					},
				},
			},
			err: &PipelineTerminatingErr{pipeline: "pipeline"},
		},
		{
			name:         "pipeline terminating err - (terminating, terminate)",
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
									Status:           PipelineTerminating,
									PipelineGwStatus: PipelineTerminate,
								},
							},
						},
					},
				},
			},
			err: &PipelineTerminatingErr{pipeline: "pipeline"},
		},
		{
			name:         "pipeline terminating err - (terminating, terminated)",
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
									Status:           PipelineTerminating,
									PipelineGwStatus: PipelineTerminated,
								},
							},
						},
					},
				},
			},
			err: &PipelineTerminatingErr{pipeline: "pipeline"},
		},
		{
			name:         "pipeline terminating err - (terminate, terminating)",
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
									Status:           PipelineTerminate,
									PipelineGwStatus: PipelineTerminating,
								},
							},
						},
					},
				},
			},
			err: &PipelineTerminatingErr{pipeline: "pipeline"},
		},
		{
			name:         "pipeline terminating err - (terminated, terminating)",
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
									Status:           PipelineTerminated,
									PipelineGwStatus: PipelineTerminating,
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
									Status:           PipelineTerminated,
									PipelineGwStatus: PipelineTerminated,
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
									Status:           PipelineReady,
									PipelineGwStatus: PipelineReady,
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
									Status:           PipelineCreating,
									PipelineGwStatus: PipelineCreating,
								},
							},
							{
								Name:    "pipeline",
								Version: 2,
								State: &PipelineState{
									Status:           PipelineCreating,
									PipelineGwStatus: PipelineCreating,
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

func TestSetPipelineGwPipelineState(t *testing.T) {
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
									PipelineGwStatus: PipelineReady,
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
									PipelineGwStatus: PipelineFailed,
								},
							},
							{
								Name:    "p",
								Version: 2,
								UID:     "uid2",
								State: &PipelineState{
									PipelineGwStatus: PipelineReady,
								},
							},
							{
								Name:    "p",
								Version: 3,
								UID:     "uid3",
								State: &PipelineState{
									PipelineGwStatus: PipelineCreate,
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
									PipelineGwStatus: PipelineTerminating,
								},
							},
							{
								Name:    "p",
								Version: 2,
								UID:     "uid2",
								State: &PipelineState{
									PipelineGwStatus: PipelineCreate,
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
									PipelineGwStatus: PipelineReady,
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
									PipelineGwStatus: PipelineReady,
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
									PipelineGwStatus: PipelineReady,
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
			err := test.store.SetPipelineGwPipelineState(test.pipelineName, test.pipelineVersion, test.uid, test.status, test.reason, "")
			if test.err == nil {
				g.Expect(err).To(BeNil())
				pv, err := test.store.GetPipelineVersion(test.pipelineName, test.pipelineVersion, test.uid)
				g.Expect(err).To(BeNil())
				g.Expect(pv).ToNot(BeNil())
				g.Expect(pv.Name).To(Equal(test.pipelineName))
				g.Expect(pv.Version).To(Equal(test.pipelineVersion))
				g.Expect(pv.UID).To(Equal(test.uid))
				g.Expect(pv.State.PipelineGwStatus).To(Equal(test.status))
				g.Expect(pv.State.PipelineGwReason).To(Equal(test.reason))
				for idx, pv := range test.store.pipelines[test.pipelineName].Versions {
					g.Expect(test.expectedPipelineVersionStats[idx]).To(Equal(pv.State.PipelineGwStatus))
				}
			} else {
				g.Expect(err.Error()).To(Equal(test.err.Error()))
			}
		})
	}
}

func TestPipelineStatusAfterDBRestart(t *testing.T) {
	g := NewGomegaWithT(t)

	type testStep struct {
		name string
		call func(handler PipelineHandler) error
	}

	type test struct {
		name                            string
		pipelineName                    string
		store                           *PipelineStore
		sequence                        []testStep
		wantEndPipelineStatus           PipelineStatus
		wantEndPipelineGwPipelineStatus PipelineStatus
	}

	logger := logrus.New()

	tests := []test{
		{
			name:         "Load pipeline flow happy path",
			pipelineName: "test-pipeline",
			store:        NewPipelineStore(logger, nil, nil),
			sequence: []testStep{
				{
					name: "add pipeline",
					call: func(handler PipelineHandler) error {
						p := &scheduler.Pipeline{
							Name: "test-pipeline",
							Steps: []*scheduler.PipelineStep{
								{
									Name:   "step1",
									Inputs: []string{},
								},
							},
						}
						return handler.AddPipeline(p)
					},
				},
				{
					name: "creating pipeline",
					call: func(handler PipelineHandler) error {
						pipeline, err := handler.GetPipeline("test-pipeline")
						if err != nil {
							return err
						}

						pipelineLatestVersion := pipeline.GetLatestPipelineVersion()

						if err := handler.SetPipelineState(
							"test-pipeline", pipelineLatestVersion.Version, pipelineLatestVersion.UID, PipelineCreating, "", "test-pipeline",
						); err != nil {
							return err
						}
						return handler.SetPipelineGwPipelineState(
							"test-pipeline", pipelineLatestVersion.Version, pipelineLatestVersion.UID, PipelineCreating, "", "test-pipeline",
						)
					},
				},
				{
					name: "ready pipeline",
					call: func(handler PipelineHandler) error {
						pipeline, err := handler.GetPipeline("test-pipeline")
						if err != nil {
							return err
						}

						pipelineLatestVersion := pipeline.GetLatestPipelineVersion()

						if err := handler.SetPipelineState(
							"test-pipeline", pipelineLatestVersion.Version, pipelineLatestVersion.UID, PipelineReady, "", "test-pipeline",
						); err != nil {
							return err
						}
						return handler.SetPipelineGwPipelineState(
							"test-pipeline", pipelineLatestVersion.Version, pipelineLatestVersion.UID, PipelineReady, "", "test-pipeline",
						)
					},
				},
			},
			wantEndPipelineStatus:           PipelineReady,
			wantEndPipelineGwPipelineStatus: PipelineReady,
		},
		{
			name:         "Load pipeline restarts after pipeline create",
			pipelineName: "test-pipeline",
			store:        NewPipelineStore(logger, nil, nil),
			sequence: []testStep{
				{
					name: "add pipeline",
					call: func(handler PipelineHandler) error {
						p := &scheduler.Pipeline{
							Name: "test-pipeline",
							Steps: []*scheduler.PipelineStep{
								{
									Name:   "step1",
									Inputs: []string{},
								},
							},
						}
						return handler.AddPipeline(p)
					},
				},
			},
			wantEndPipelineStatus:           PipelineCreate,
			wantEndPipelineGwPipelineStatus: PipelineCreate,
		},
		{
			name:         "Load pipeline restarts after pipeline creating",
			pipelineName: "test-pipeline",
			store:        NewPipelineStore(logger, nil, nil),
			sequence: []testStep{
				{
					name: "add pipeline",
					call: func(handler PipelineHandler) error {
						p := &scheduler.Pipeline{
							Name: "test-pipeline",
							Steps: []*scheduler.PipelineStep{
								{
									Name:   "step1",
									Inputs: []string{},
								},
							},
						}
						return handler.AddPipeline(p)
					},
				},
				{
					name: "creating pipeline",
					call: func(handler PipelineHandler) error {
						pipeline, err := handler.GetPipeline("test-pipeline")
						if err != nil {
							return err
						}

						pipelineLatestVersion := pipeline.GetLatestPipelineVersion()

						if err := handler.SetPipelineState(
							"test-pipeline", pipelineLatestVersion.Version, pipelineLatestVersion.UID, PipelineCreating, "", "test-pipeline",
						); err != nil {
							return err
						}
						return handler.SetPipelineGwPipelineState(
							"test-pipeline", pipelineLatestVersion.Version, pipelineLatestVersion.UID, PipelineCreating, "", "test-pipeline",
						)
					},
				},
			},
			wantEndPipelineStatus:           PipelineCreating,
			wantEndPipelineGwPipelineStatus: PipelineCreating,
		},
		{
			name:         "Unload pipeline flow",
			pipelineName: "test-pipeline",
			store:        NewPipelineStore(logger, nil, nil),
			sequence: []testStep{
				{
					name: "AddPipeline",
					call: func(handler PipelineHandler) error {
						p := &scheduler.Pipeline{
							Name: "test-pipeline",
							Steps: []*scheduler.PipelineStep{
								{
									Name:   "step1",
									Inputs: []string{},
								},
							},
						}
						return handler.AddPipeline(p)
					},
				},
				{
					name: "RemovePipeline",
					call: func(handler PipelineHandler) error {
						return handler.RemovePipeline("test-pipeline")
					},
				},
				{
					name: "Update Pipeline to terminating",
					call: func(handler PipelineHandler) error {
						pipeline, err := handler.GetPipeline("test-pipeline")
						if err != nil {
							return err
						}

						pipelineLatestVersion := pipeline.GetLatestPipelineVersion()

						if err := handler.SetPipelineState(
							"test-pipeline", pipelineLatestVersion.Version, pipelineLatestVersion.UID, PipelineTerminating, "", "test-pipeline",
						); err != nil {
							return err
						}
						return handler.SetPipelineGwPipelineState(
							"test-pipeline", pipelineLatestVersion.Version, pipelineLatestVersion.UID, PipelineTerminating, "", "test-pipeline",
						)
					},
				},
				{
					name: "Update Pipeline to terminated",
					call: func(handler PipelineHandler) error {
						pipeline, err := handler.GetPipeline("test-pipeline")
						if err != nil {
							return err
						}

						pipelineLatestVersion := pipeline.GetLatestPipelineVersion()

						if err := handler.SetPipelineState(
							"test-pipeline", pipelineLatestVersion.Version, pipelineLatestVersion.UID, PipelineTerminated, "", "test-pipeline",
						); err != nil {
							return err
						}
						return handler.SetPipelineGwPipelineState(
							"test-pipeline", pipelineLatestVersion.Version, pipelineLatestVersion.UID, PipelineTerminated, "", "test-pipeline",
						)
					},
				},
			},
			wantEndPipelineStatus:           PipelineTerminated,
			wantEndPipelineGwPipelineStatus: PipelineTerminated,
		},
		{
			name:         "Unload pipeline restarts after pipeline remove",
			pipelineName: "test-pipeline",
			store:        NewPipelineStore(logger, nil, nil),
			sequence: []testStep{
				{
					name: "AddPipeline",
					call: func(handler PipelineHandler) error {
						p := &scheduler.Pipeline{
							Name: "test-pipeline",
							Steps: []*scheduler.PipelineStep{
								{
									Name:   "step1",
									Inputs: []string{},
								},
							},
						}
						return handler.AddPipeline(p)
					},
				},
				{
					name: "RemovePipeline",
					call: func(handler PipelineHandler) error {
						return handler.RemovePipeline("test-pipeline")
					},
				},
			},
			wantEndPipelineStatus:           PipelineTerminate,
			wantEndPipelineGwPipelineStatus: PipelineTerminate,
		},
		{
			name:         "Unload pipeline restarts after pipeline terminating",
			pipelineName: "test-pipeline",
			store:        NewPipelineStore(logger, nil, nil),
			sequence: []testStep{
				{
					name: "AddPipeline",
					call: func(handler PipelineHandler) error {
						p := &scheduler.Pipeline{
							Name: "test-pipeline",
							Steps: []*scheduler.PipelineStep{
								{
									Name:   "step1",
									Inputs: []string{},
								},
							},
						}
						return handler.AddPipeline(p)
					},
				},
				{
					name: "RemovePipeline",
					call: func(handler PipelineHandler) error {
						return handler.RemovePipeline("test-pipeline")
					},
				},
				{
					name: "Update Pipeline to terminating",
					call: func(handler PipelineHandler) error {
						pipeline, err := handler.GetPipeline("test-pipeline")
						if err != nil {
							return err
						}

						pipelineLatestVersion := pipeline.GetLatestPipelineVersion()

						if err := handler.SetPipelineState(
							"test-pipeline", pipelineLatestVersion.Version, pipelineLatestVersion.UID, PipelineTerminating, "", "test-pipeline",
						); err != nil {
							return err
						}
						return handler.SetPipelineGwPipelineState(
							"test-pipeline", pipelineLatestVersion.Version, pipelineLatestVersion.UID, PipelineTerminating, "", "test-pipeline",
						)
					},
				},
			},
			wantEndPipelineStatus:           PipelineTerminating,
			wantEndPipelineGwPipelineStatus: PipelineTerminating,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())
			db, _ := newPipelineDbManager(getPipelineDbFolder(path), logger, 1000000)

			test.store.db = db

			for _, step := range test.sequence {
				err := step.call(test.store)
				if err != nil {
					t.Errorf("unexpected error in step %s: %v", step.name, err)
				}
			}

			pipelines, err := test.store.GetPipelines()
			if err != nil {
				t.Errorf("unexpected error in GetPipelines: %v", err)
				return
			}

			for _, pipeline := range pipelines {
				if pipeline.Name == test.pipelineName {
					g.Expect(pipeline.GetLatestPipelineVersion().State.Status.String()).To(Equal(test.wantEndPipelineStatus.String()))
				}
			}

			// stop first db
			_ = db.Stop()

			// simulate restart and restore pipeline
			pipelineStore := NewPipelineStore(logrus.New(), nil, nil)
			dbAfterRestart, _ := newPipelineDbManager(getPipelineDbFolder(path), logger, 1000000)
			pipelineStore.db = dbAfterRestart

			err = pipelineStore.db.restore(pipelineStore.restorePipeline)
			if err != nil {
				t.Errorf("unexpected error in restore: %v", err)
				return
			}

			// check status of pipeline is the same as the last time saved
			pipelines, err = pipelineStore.GetPipelines()
			if err != nil {
				t.Errorf("unexpected error in GetPipelines: %v", err)
				return
			}

			for _, pipeline := range pipelines {
				if pipeline.Name == test.pipelineName {
					g.Expect(pipeline.GetLatestPipelineVersion().State.Status.String()).To(Equal(test.wantEndPipelineStatus.String()))
					g.Expect(pipeline.GetLatestPipelineVersion().State.PipelineGwStatus.String()).To(Equal(test.wantEndPipelineGwPipelineStatus.String()))
				}
			}
		})
	}
}

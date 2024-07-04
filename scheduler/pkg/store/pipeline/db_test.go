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

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/utils"
)

func TestSaveAndRestore(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name      string
		pipelines []*Pipeline
	}

	tests := []test{
		{
			name: "test single pipeline",
			pipelines: []*Pipeline{
				{
					Name:        "test",
					LastVersion: 0,
					Versions: []*PipelineVersion{
						{
							Name:    "p1",
							Version: 0,
							UID:     "x",
							Steps: map[string]*PipelineStep{
								"a": {Name: "a"},
							},
							State: &PipelineState{
								Status:    PipelineReady,
								Reason:    "deployed",
								Timestamp: time.Now(),
							},
							Output: &PipelineOutput{},
							KubernetesMeta: &KubernetesMeta{
								Namespace: "default",
							},
						},
					},
					Deleted: false,
				},
			},
		},
		{
			name:      "no pipelines",
			pipelines: []*Pipeline{},
		},
		{
			name: "test multiple pipelines",
			pipelines: []*Pipeline{
				{
					Name:        "test1",
					LastVersion: 0,
					Versions: []*PipelineVersion{
						{
							Name:    "p1",
							Version: 0,
							UID:     "x",
							Steps: map[string]*PipelineStep{
								"a": {Name: "a"},
							},
							State: &PipelineState{
								Status:    PipelineReady,
								Reason:    "deployed",
								Timestamp: time.Now(),
							},
							Output: &PipelineOutput{},
							KubernetesMeta: &KubernetesMeta{
								Namespace: "default",
							},
						},
					},
					Deleted: false,
				},
				{
					Name:        "test2",
					LastVersion: 0,
					Versions: []*PipelineVersion{
						{
							Name:    "p1",
							Version: 0,
							UID:     "x",
							Steps: map[string]*PipelineStep{
								"b": {Name: "b"},
							},
							State: &PipelineState{
								Status:    PipelineTerminating,
								Reason:    "deployed",
								Timestamp: time.Now(),
							},
							Output: &PipelineOutput{},
							KubernetesMeta: &KubernetesMeta{
								Namespace: "default",
							},
						},
					},
					Deleted: false,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())
			logger := log.New()
			db, err := newPipelineDbManager(getPipelineDbFolder(path), logger)
			g.Expect(err).To(BeNil())
			for _, p := range test.pipelines {
				err := db.save(p)
				g.Expect(err).To(BeNil())
			}
			err = db.Stop()
			g.Expect(err).To(BeNil())

			ps := NewPipelineStore(log.New(), nil, fakeModelStore{status: map[string]store.ModelState{}})
			err = ps.InitialiseOrRestoreDB(path)
			g.Expect(err).To(BeNil())
			for _, p := range test.pipelines {
				g.Expect(cmp.Equal(p, ps.pipelines[p.Name])).To(BeTrue())
			}
		})
	}
}

func TestGetPipelineFromDB(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name         string
		pipelines    []*Pipeline
		pipelineName string
		isErr        bool
	}

	tests := []test{
		{
			name: "test single pipeline",
			pipelines: []*Pipeline{
				{
					Name:        "test",
					LastVersion: 0,
					Versions: []*PipelineVersion{
						{
							Name:    "p1",
							Version: 0,
							UID:     "x",
							Steps: map[string]*PipelineStep{
								"a": {Name: "a"},
							},
							State: &PipelineState{
								Status:    PipelineReady,
								Reason:    "deployed",
								Timestamp: time.Now(),
							},
							Output: &PipelineOutput{},
							KubernetesMeta: &KubernetesMeta{
								Namespace: "default",
							},
						},
					},
					Deleted: false,
				},
			},
			pipelineName: "test",
			isErr:        false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())
			logger := log.New()
			db, err := newPipelineDbManager(getPipelineDbFolder(path), logger)
			g.Expect(err).To(BeNil())
			for _, p := range test.pipelines {
				err := db.save(p)
				g.Expect(err).To(BeNil())
			}
			g.Expect(err).To(BeNil())

			actualPipeline, err := db.get(test.pipelineName)
			if test.isErr {
				g.Expect(err).To(BeNil())
				g.Expect(actualPipeline).To(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(actualPipeline).ToNot(BeNil())
				pipeFound := false
				for _, pipe := range test.pipelines {
					if pipe.Name == test.pipelineName {
						g.Expect(cmp.Equal(pipe, actualPipeline)).To(BeTrue())
						pipeFound = true
					}
				}
				g.Expect(pipeFound).To(BeTrue())
			}
			err = db.Stop()
			g.Expect(err).To(BeNil())
		})
	}
}

func TestDeletePipelineFromDB(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name         string
		pipelines    []*Pipeline
		pipelineName string
	}

	tests := []test{
		{
			name: "test single pipeline",
			pipelines: []*Pipeline{
				{
					Name:        "test",
					LastVersion: 0,
					Versions: []*PipelineVersion{
						{
							Name:    "p1",
							Version: 0,
							UID:     "x",
							Steps: map[string]*PipelineStep{
								"a": {Name: "a"},
							},
							State: &PipelineState{
								Status:    PipelineReady,
								Reason:    "deployed",
								Timestamp: time.Now(),
							},
							Output: &PipelineOutput{},
							KubernetesMeta: &KubernetesMeta{
								Namespace: "default",
							},
						},
					},
					Deleted: false,
				},
			},
			pipelineName: "test",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())
			logger := log.New()
			db, err := newPipelineDbManager(getPipelineDbFolder(path), logger)
			g.Expect(err).To(BeNil())
			for _, p := range test.pipelines {
				err := db.save(p)
				g.Expect(err).To(BeNil())
			}
			g.Expect(err).To(BeNil())

			err = db.delete(test.pipelineName)
			g.Expect(err).To(BeNil())

			_, err = db.get(test.pipelineName)
			g.Expect(err).ToNot(BeNil()) // expect error as pipeline should be deleted

			err = db.Stop()
			g.Expect(err).To(BeNil())
		})
	}
}

func TestMigrateFromV1ToV2(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name      string
		pipelines []*Pipeline
	}

	tests := []test{
		{
			name: "test single pipeline",
			pipelines: []*Pipeline{
				{
					Name:        "test",
					LastVersion: 0,
					Versions: []*PipelineVersion{
						{
							Name:    "p1",
							Version: 0,
							UID:     "x",
							Steps: map[string]*PipelineStep{
								"a": {Name: "a"},
							},
							State: &PipelineState{
								Status:    PipelineReady,
								Reason:    "deployed",
								Timestamp: time.Now(),
							},
							Output: &PipelineOutput{},
							KubernetesMeta: &KubernetesMeta{
								Namespace: "default",
							},
						},
					},
					Deleted: false,
				},
			},
		},
		{
			name:      "no pipelines",
			pipelines: []*Pipeline{},
		},
		{
			name: "test multiple pipelines",
			pipelines: []*Pipeline{
				{
					Name:        "test1",
					LastVersion: 0,
					Versions: []*PipelineVersion{
						{
							Name:    "p1",
							Version: 0,
							UID:     "x",
							Steps: map[string]*PipelineStep{
								"a": {Name: "a"},
							},
							State: &PipelineState{
								Status:    PipelineReady,
								Reason:    "deployed",
								Timestamp: time.Now(),
							},
							Output: &PipelineOutput{},
							KubernetesMeta: &KubernetesMeta{
								Namespace: "default",
							},
						},
					},
					Deleted: false,
				},
				{
					Name:        "test2",
					LastVersion: 0,
					Versions: []*PipelineVersion{
						{
							Name:    "p1",
							Version: 0,
							UID:     "x",
							Steps: map[string]*PipelineStep{
								"b": {Name: "b"},
							},
							State: &PipelineState{
								Status:    PipelineTerminating,
								Reason:    "deployed",
								Timestamp: time.Now(),
							},
							Output: &PipelineOutput{},
							KubernetesMeta: &KubernetesMeta{
								Namespace: "default",
							},
						},
					},
					Deleted: false,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := fmt.Sprintf("%s/db", t.TempDir())
			logger := log.New()
			db, err := utils.Open(getPipelineDbFolder(path), logger, "pipelineDb")
			g.Expect(err).To(BeNil())
			for _, p := range test.pipelines {
				err := save(p, db)
				g.Expect(err).To(BeNil())
			}
			err = db.Close()
			g.Expect(err).To(BeNil())

			ps := NewPipelineStore(log.New(), nil, fakeModelStore{status: map[string]store.ModelState{}})
			err = ps.InitialiseOrRestoreDB(path)
			g.Expect(err).To(BeNil())

			// make sure we still have the pipelines
			for _, p := range test.pipelines {
				g.Expect(cmp.Equal(p, ps.pipelines[p.Name])).To(BeTrue())
			}

			// make sure we have the correct version
			version, err := ps.db.getVersion()
			g.Expect(err).To(BeNil())
			g.Expect(version).To(Equal(currentPipelineSnapshotVersion))

		})
	}
}

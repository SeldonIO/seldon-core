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

	"github.com/dgraph-io/badger/v3"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store/utils"
)

func TestSaveWithTTL(t *testing.T) {
	g := NewGomegaWithT(t)
	pipeline := &Pipeline{
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
		Deleted: true,
	}
	ttl := time.Duration(time.Second)
	pipeline.DeletedAt = time.Now().Add(-utils.DeletedResourceTTL).Add(ttl)

	path := fmt.Sprintf("%s/db", t.TempDir())
	logger := log.New()
	db, err := newPipelineDbManager(getPipelineDbFolder(path), logger)
	g.Expect(err).To(BeNil())
	err = db.save(pipeline)
	g.Expect(err).To(BeNil())

	persistedExp, err := db.get(pipeline.Name)
	g.Expect(err).To(BeNil())
	g.Expect(persistedExp).NotTo(BeNil())

	time.Sleep(ttl * 2)

	persistedExp, err = db.get(pipeline.Name)
	g.Expect(err).ToNot(BeNil())
	g.Expect(persistedExp).To(BeNil())

	err = db.Stop()
	g.Expect(err).To(BeNil())
}

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

func TestSaveAndRestoreDeletedPipelines(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name     string
		pipeline Pipeline
		withTTL  bool
	}

	createdDeletedPipeline := func(name string) Pipeline {
		return Pipeline{
			Name:        "test-with",
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
			Deleted: true,
		}
	}

	tests := []test{
		{
			name:     "test deleted pipeline with TTL",
			pipeline: createdDeletedPipeline("with-ttl"),
			withTTL:  true,
		},
		{
			name:     "test deleted pipeline without TTL",
			pipeline: createdDeletedPipeline("without-ttl"),
			withTTL:  false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g.Expect(test.pipeline.Deleted).To(BeTrue(), "this is a test for deleted pipelines")
			path := fmt.Sprintf("%s/db", t.TempDir())
			logger := log.New()
			pdb, err := newPipelineDbManager(getPipelineDbFolder(path), logger)
			g.Expect(err).To(BeNil())
			if !test.withTTL {
				err = pdb.save(&test.pipeline)
			} else {
				test.pipeline.DeletedAt = time.Now().Add(-utils.DeletedResourceTTL)
				err = pdb.save(&test.pipeline)
			}
			g.Expect(err).To(BeNil())
			err = pdb.Stop()
			g.Expect(err).To(BeNil())

			ps := NewPipelineStore(log.New(), nil, fakeModelStore{status: map[string]store.ModelState{}})
			err = ps.InitialiseOrRestoreDB(path)
			g.Expect(err).To(BeNil())
			if !test.withTTL {
				var item *badger.Item
				err = ps.db.db.View(func(txn *badger.Txn) error {
					item, err = txn.Get(([]byte(test.pipeline.Name)))
					return err
				})
				g.Expect(err).To(BeNil())
				g.Expect(item.ExpiresAt()).ToNot(BeZero())
				g.Expect(ps.pipelines[test.pipeline.Name]).ToNot(BeNil())
			} else {
				g.Expect(ps.pipelines[test.pipeline.Name]).To(BeNil())
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

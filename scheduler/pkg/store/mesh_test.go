/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package store

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"google.golang.org/protobuf/proto"
	"knative.dev/pkg/ptr"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestModelVersion_DeepCopy(t *testing.T) {
	uInt32Ptr := func(i uint32) *uint32 {
		return &i
	}
	uInt64Ptr := func(i uint64) *uint64 {
		return &i
	}

	tests := []struct {
		name     string
		setupMV  func() *ModelVersion
		validate func(t *testing.T, original, copied *ModelVersion)
	}{
		{
			name: "empty model version",
			setupMV: func() *ModelVersion {
				return &ModelVersion{}
			},
			validate: func(t *testing.T, original, copied *ModelVersion) {
				RegisterTestingT(t)
				Expect(copied).ToNot(BeNil())
				Expect(copied.version).To(Equal(uint32(0)))
				Expect(copied.server).To(Equal(""))
				Expect(copied.state).To(Equal(ModelStatus{}))
				Expect(copied.modelDefn).To(BeNil())
				Expect(copied.replicas).To(BeNil())
			},
		},
		{
			name: "basic fields only",
			setupMV: func() *ModelVersion {
				return &ModelVersion{
					version: 123,
					server:  "test-server",
					state: ModelStatus{
						State:               ModelAvailable,
						Reason:              "some reason",
						AvailableReplicas:   1,
						UnavailableReplicas: 2,
						DrainingReplicas:    3,
						Timestamp:           time.Now(),
					},
				}
			},
			validate: func(t *testing.T, original, copied *ModelVersion) {
				RegisterTestingT(t)
				Expect(copied.version).To(Equal(uint32(123)))
				Expect(copied.server).To(Equal("test-server"))
				Expect(copied.state).To(Equal(original.state))
				Expect(copied.modelDefn).To(BeNil())
				Expect(copied.replicas).To(BeNil())
			},
		},
		{
			name: "with model definition",
			setupMV: func() *ModelVersion {
				return &ModelVersion{
					version: 456,
					server:  "model-server",
					replicas: map[int]ReplicaStatus{
						1: {
							State:     Available,
							Reason:    "some reason",
							Timestamp: time.Now(),
						},
						2: {
							State:     LoadedUnavailable,
							Reason:    "some other reason",
							Timestamp: time.Now(),
						},
					},
					state: ModelStatus{
						State:               ModelAvailable,
						Reason:              "some reason",
						AvailableReplicas:   1,
						UnavailableReplicas: 2,
						DrainingReplicas:    3,
						Timestamp:           time.Now(),
					},
					modelDefn: &pb.Model{
						Meta: &pb.MetaData{
							Name:    "some name",
							Kind:    ptr.String("some kind"),
							Version: ptr.String("some version"),
							KubernetesMeta: &pb.KubernetesMeta{
								Namespace:  "some namespace",
								Generation: 1,
							},
						},
						ModelSpec: &pb.ModelSpec{
							Uri:             "some/url",
							ArtifactVersion: uInt32Ptr(1),
							Requirements:    []string{"some requirements"},
							MemoryBytes:     uInt64Ptr(2),
							Server:          ptr.String("some server"),
							Parameters: []*pb.ParameterSpec{{
								Name:  "some name",
								Value: "some value",
							}},
							ModelRuntimeInfo: &pb.ModelRuntimeInfo{
								ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{
									Mlserver: &pb.MLServerModelSettings{ParallelWorkers: 2},
								},
							},
							ModelSpec: &pb.ModelSpec_Explainer{
								Explainer: &pb.ExplainerSpec{
									Type:        "some type",
									ModelRef:    ptr.String("some model ref"),
									PipelineRef: ptr.String("some pipeline ref"),
								},
							},
						},
					},
				}
			},
			validate: func(t *testing.T, original, copied *ModelVersion) {
				RegisterTestingT(t)
				Expect(copied.modelDefn).ToNot(BeNil())

				// Verify it's a deep copy (different pointers)
				Expect(copied.modelDefn).ToNot(BeIdenticalTo(original.modelDefn))

				// Verify proto equality
				Expect(proto.Equal(original.modelDefn, copied.modelDefn)).To(BeTrue())
				Expect(copied.version).To(Equal(original.version))
				Expect(copied.server).To(Equal(original.server))
				Expect(copied.state).To(Equal(original.state))
				Expect(copied.replicas).To(Equal(original.replicas))
			},
		},
		{
			name: "with empty replicas map",
			setupMV: func() *ModelVersion {
				return &ModelVersion{
					version: 100,
					server:  "empty-replica-server",
					state: ModelStatus{
						State:               ModelAvailable,
						Reason:              "some reason",
						AvailableReplicas:   1,
						UnavailableReplicas: 2,
						DrainingReplicas:    3,
						Timestamp:           time.Now(),
					},
					replicas: make(map[int]ReplicaStatus),
				}
			},
			validate: func(t *testing.T, original, copied *ModelVersion) {
				RegisterTestingT(t)
				Expect(copied.replicas).ToNot(BeNil())
				Expect(len(copied.replicas)).To(Equal(0))
				Expect(copied.state).To(Equal(original.state))
				Expect(&copied.replicas).ToNot(BeIdenticalTo(&original.replicas))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := tt.setupMV()
			copied := original.DeepCopy()
			tt.validate(t, original, copied)
		})
	}
}

func TestReplicaStateToString(t *testing.T) {
	for _, state := range replicaStates {
		_ = state.String()
	}
}

func TestCleanCapabilities(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		in       []string
		expected []string
	}

	tests := []test{
		{
			name:     "misc",
			in:       []string{"mlserver", " foo ", " bar", "bar   "},
			expected: []string{"mlserver", "foo", "bar", "bar"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			out := cleanCapabilities(test.in)
			g.Expect(out).To(Equal(test.expected))
		})
	}
}

func TestCreateSnapshot(t *testing.T) {
	g := NewGomegaWithT(t)

	server := &Server{
		name: "test",
		replicas: map[int]*ServerReplica{
			0: {
				inferenceSvc: "svc",
				loadedModels: map[ModelVersionID]bool{
					{Name: "model1", Version: 1}: true,
					{Name: "model2", Version: 2}: true,
				},
				loadingModels: map[ModelVersionID]bool{
					{Name: "model10", Version: 1}: true,
					{Name: "model20", Version: 2}: true,
				},
			},
		},
		kubernetesMeta: &pb.KubernetesMeta{Namespace: "default"},
	}

	snapshot := server.CreateSnapshot(false, true)

	g.Expect(snapshot.Replicas[0].loadedModels).To(Equal(
		map[ModelVersionID]bool{
			{Name: "model1", Version: 1}: true,
			{Name: "model2", Version: 2}: true,
		},
	))

	g.Expect(snapshot.Replicas[0].loadingModels).To(Equal(
		map[ModelVersionID]bool{
			{Name: "model10", Version: 1}: true,
			{Name: "model20", Version: 2}: true,
		},
	))

	server.replicas[1] = &ServerReplica{
		inferenceSvc: "svc",
		loadedModels: map[ModelVersionID]bool{
			{Name: "model3", Version: 1}: true,
			{Name: "model4", Version: 2}: true,
		},
	}
	server.name = "foo"
	server.kubernetesMeta.Namespace = "test"

	g.Expect(snapshot.Name).To(Equal("test"))
	g.Expect(len(snapshot.Replicas)).To(Equal(1))
	g.Expect(snapshot.KubernetesMeta.Namespace).To(Equal("default"))

}

func TestLoadedModel(t *testing.T) {
	g := NewGomegaWithT(t)

	const (
		add int = iota
		remove
	)

	type test struct {
		name                       string
		op                         int
		model                      string
		version                    int
		state                      ModelReplicaState
		loadedModels               map[ModelVersionID]bool
		uniqueLoadedModels         map[string]bool
		loadingModels              map[ModelVersionID]bool
		expectedLoadedModels       map[ModelVersionID]bool
		expectedLoadingModels      map[ModelVersionID]bool
		expectedUniqueLoadedModels map[string]bool
	}

	tests := []test{
		{
			name:         "add loading",
			op:           add,
			model:        "dummy",
			version:      1,
			state:        Loading,
			loadedModels: map[ModelVersionID]bool{},
			loadingModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true, // we should already have an entry from an earlier load request
			},
			uniqueLoadedModels:         map[string]bool{},
			expectedLoadedModels:       map[ModelVersionID]bool{},
			expectedUniqueLoadedModels: map[string]bool{},
			expectedLoadingModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
		},
		{
			name:         "add loaded",
			op:           add,
			model:        "dummy",
			version:      1,
			state:        Loaded,
			loadedModels: map[ModelVersionID]bool{},
			loadingModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true, // we should transition from loading to loaded
			},
			uniqueLoadedModels: map[string]bool{},
			expectedLoadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
			expectedUniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadingModels: map[ModelVersionID]bool{},
		},
		{
			name:    "add loading - new version",
			op:      add,
			model:   "dummy",
			version: 2,
			state:   Loading,
			loadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
			loadingModels: map[ModelVersionID]bool{},
			uniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
			expectedUniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadingModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 2}: true,
			},
		},
		{
			name:    "add loaded- new version",
			op:      add,
			model:   "dummy",
			version: 2,
			state:   Loaded,
			loadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
			loadingModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 2}: true, // we should transition from loading to loaded
			},
			uniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
				{Name: "dummy", Version: 2}: true,
			},
			expectedUniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadingModels: map[ModelVersionID]bool{},
		},
		{
			name:    "remove with early version",
			op:      remove,
			model:   "dummy",
			version: 2,
			loadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
				{Name: "dummy", Version: 2}: true,
			},
			loadingModels: map[ModelVersionID]bool{},
			uniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
			expectedUniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadingModels: map[ModelVersionID]bool{},
		},
		{
			name:    "remove",
			op:      remove,
			model:   "dummy",
			version: 1,
			loadedModels: map[ModelVersionID]bool{
				{Name: "dummy", Version: 1}: true,
			},
			loadingModels: map[ModelVersionID]bool{},
			uniqueLoadedModels: map[string]bool{
				"dummy": true,
			},
			expectedLoadedModels:       map[ModelVersionID]bool{},
			expectedUniqueLoadedModels: map[string]bool{},
			expectedLoadingModels:      map[ModelVersionID]bool{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := ServerReplica{
				inferenceSvc:       "svc",
				loadedModels:       test.loadedModels,
				loadingModels:      test.loadingModels,
				uniqueLoadedModels: test.uniqueLoadedModels,
			}

			if test.op == add {
				server.addModelVersion(test.model, uint32(test.version), test.state)
			} else {
				server.deleteModelVersion(test.model, uint32(test.version))
			}
			g.Expect(server.loadedModels).To(Equal(test.expectedLoadedModels))
			g.Expect(server.loadingModels).To(Equal(test.expectedLoadingModels))
			g.Expect(server.uniqueLoadedModels).To(Equal(test.expectedUniqueLoadedModels))
		})
	}
}

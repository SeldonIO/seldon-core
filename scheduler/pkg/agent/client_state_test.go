package agent

import (
	"testing"

	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pbs "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

func TestAddModelVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                   string
		state                  *ClientState
		modelVersion           *pb.ModelVersion
		success                bool
		expectedAvailableBytes uint64
	}

	getUint64Ptr := func(val uint64) *uint64 {
		return &val
	}

	tests := []test{
		{
			name: "ModelOK",
			state: NewClientState(&pb.ReplicaConfig{
				MemoryBytes: 1000,
			}),
			modelVersion: &pb.ModelVersion{
				Model: &pbs.Model{
					Meta: &pbs.MetaData{
						Name: "iris",
					},
					ModelSpec: &pbs.ModelSpec{
						MemoryBytes: getUint64Ptr(500),
					},
				},
				Version: 1,
			},
			success:                true,
			expectedAvailableBytes: 500,
		},
		{
			name: "ModelTooBig",
			state: NewClientState(&pb.ReplicaConfig{
				MemoryBytes: 1000,
			}),
			modelVersion: &pb.ModelVersion{
				Model: &pbs.Model{
					Meta: &pbs.MetaData{
						Name: "iris",
					},
					ModelSpec: &pbs.ModelSpec{
						MemoryBytes: getUint64Ptr(2000),
					},
				},
				Version: 1,
			},
			success:                false,
			expectedAvailableBytes: 1000,
		},
		{
			name: "ModelVersionTooBig",
			state: &ClientState{
				availableMemoryBytes: 100,
				loadedModels: map[string]*ModelVersions{
					"iris": {
						versions: map[uint32]*pb.ModelVersion{
							1: {
								Model: &pbs.Model{
									Meta: &pbs.MetaData{
										Name: "iris",
									},
									ModelSpec: &pbs.ModelSpec{
										MemoryBytes: getUint64Ptr(500),
									},
								},
							},
						},
						totalMemoryBytes: 500,
					},
				},
			},
			modelVersion: &pb.ModelVersion{
				Model: &pbs.Model{
					Meta: &pbs.MetaData{
						Name: "iris",
					},
					ModelSpec: &pbs.ModelSpec{
						MemoryBytes: getUint64Ptr(200),
					},
				},
				Version: 2,
			},
			success:                false,
			expectedAvailableBytes: 100,
		},
		{
			name: "VersionExists",
			state: &ClientState{
				availableMemoryBytes: 100,
				loadedModels: map[string]*ModelVersions{
					"iris": {
						versions: map[uint32]*pb.ModelVersion{
							1: {
								Model: &pbs.Model{
									Meta: &pbs.MetaData{
										Name: "iris",
									},
									ModelSpec: &pbs.ModelSpec{
										MemoryBytes: getUint64Ptr(500),
									},
								},
							},
						},
						totalMemoryBytes: 500,
					},
				},
			},
			modelVersion: &pb.ModelVersion{
				Model: &pbs.Model{
					Meta: &pbs.MetaData{
						Name: "iris",
					},
					ModelSpec: &pbs.ModelSpec{
						MemoryBytes: getUint64Ptr(10),
					},
				},
				Version: 1,
			},
			success:                false,
			expectedAvailableBytes: 100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.state.addModelVersion(test.modelVersion)
			if test.success {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
			}
			g.Expect(test.state.availableMemoryBytes).To(Equal(test.expectedAvailableBytes))
		})
	}
}

func TestRemoveModelVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                   string
		state                  *ClientState
		modelVersion           *pb.ModelVersion
		modelDeleted           bool
		expectedAvailableBytes uint64
	}

	getUint64Ptr := func(val uint64) *uint64 {
		return &val
	}

	tests := []test{
		{
			name: "ModelDeleted",
			state: &ClientState{
				availableMemoryBytes: 100,
				loadedModels: map[string]*ModelVersions{
					"iris": {
						versions: map[uint32]*pb.ModelVersion{
							1: {
								Model: &pbs.Model{
									Meta: &pbs.MetaData{
										Name: "iris",
									},
									ModelSpec: &pbs.ModelSpec{
										MemoryBytes: getUint64Ptr(500),
									},
								},
							},
						},
						totalMemoryBytes: 500,
					},
				},
			},
			modelVersion: &pb.ModelVersion{
				Model: &pbs.Model{
					Meta: &pbs.MetaData{
						Name: "iris",
					},
				},
				Version: 1,
			},
			modelDeleted:           true,
			expectedAvailableBytes: 600,
		},
		{
			name: "ModelNotDeleted",
			state: &ClientState{
				availableMemoryBytes: 100,
				loadedModels: map[string]*ModelVersions{
					"iris": {
						versions: map[uint32]*pb.ModelVersion{
							1: {
								Model: &pbs.Model{
									Meta: &pbs.MetaData{
										Name: "iris",
									},
									ModelSpec: &pbs.ModelSpec{
										MemoryBytes: getUint64Ptr(500),
									},
								},
							},
							2: {
								Model: &pbs.Model{
									Meta: &pbs.MetaData{
										Name: "iris",
									},
									ModelSpec: &pbs.ModelSpec{
										MemoryBytes: getUint64Ptr(500),
									},
								},
							},
						},
						totalMemoryBytes: 1000,
					},
				},
			},
			modelVersion: &pb.ModelVersion{
				Model: &pbs.Model{
					Meta: &pbs.MetaData{
						Name: "iris",
					},
				},
				Version: 1,
			},
			modelDeleted:           false,
			expectedAvailableBytes: 600,
		},
		{
			name: "ModelNotExist",
			state: &ClientState{
				availableMemoryBytes: 100,
				loadedModels:         map[string]*ModelVersions{},
			},
			modelVersion: &pb.ModelVersion{
				Model: &pbs.Model{
					Meta: &pbs.MetaData{
						Name: "iris",
					},
				},
				Version: 1,
			},
			modelDeleted:           true,
			expectedAvailableBytes: 100,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			modelDeleted := test.state.removeModelVersion(test.modelVersion)
			g.Expect(modelDeleted).To(Equal(test.modelDeleted))
			g.Expect(test.state.availableMemoryBytes).To(Equal(test.expectedAvailableBytes))
		})
	}
}

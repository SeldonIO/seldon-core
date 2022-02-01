package agent

import (
	"testing"

	. "github.com/onsi/gomega"
	pb "github.com/seldonio/seldon-core/scheduler/apis/mlops/agent"
	pbs "github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
)

func getUint64Ptr(val uint64) *uint64 {
	return &val
}

func TestAddModelVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                    string
		state                   *ModelState
		modelVersion            *pb.ModelVersion
		versionAdded            bool
		expectedTotalModelBytes uint64
	}

	tests := []test{
		{
			name:  "NewModel",
			state: NewModelState(),
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
			versionAdded:            true,
			expectedTotalModelBytes: 500,
		},
		{
			name: "NewModel (Another Model Exsits)",
			state: &ModelState{
				loadedModels: map[string]*ModelVersions{
					"mnist": {
						versions: map[uint32]*pb.ModelVersion{
							1: {
								Model: &pbs.Model{
									Meta: &pbs.MetaData{
										Name: "mnist",
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
						MemoryBytes: getUint64Ptr(500),
					},
				},
				Version: 1,
			},
			versionAdded:            true,
			expectedTotalModelBytes: 500,
		},
		{
			name: "NewVersion",
			state: &ModelState{
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
						MemoryBytes: getUint64Ptr(500),
					},
				},
				Version: 2,
			},
			versionAdded:            true,
			expectedTotalModelBytes: 1000,
		},
		{
			name: "VersionExists",
			state: &ModelState{
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
						MemoryBytes: getUint64Ptr(500),
					},
				},
				Version: 1,
			},
			versionAdded:            false,
			expectedTotalModelBytes: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			versionAdded := test.state.addModelVersion(test.modelVersion)
			g.Expect(versionAdded).To(Equal(test.versionAdded))
			//check version exists
			g.Expect(test.state.versionExists("iris", test.modelVersion.GetVersion())).To(Equal(true))
			g.Expect(test.state.getModelTotalMemoryBytes("iris")).To(Equal(test.expectedTotalModelBytes))
		})
	}
}

func TestRemoveModelVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                    string
		state                   *ModelState
		modelVersion            *pb.ModelVersion
		modelDeleted            bool
		expectedTotalModelBytes uint64
		numModels               int
	}

	tests := []test{
		{
			name: "ModelDeleted",
			state: &ModelState{
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
			modelDeleted:            true,
			expectedTotalModelBytes: 0,
			numModels:               0,
		},
		{
			name: "ModelNotDeleted",
			state: &ModelState{
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
			modelDeleted:            false,
			expectedTotalModelBytes: 500,
			numModels:               1,
		},
		{
			name: "ModelNotDeleted With Another Model Existing",
			state: &ModelState{
				loadedModels: map[string]*ModelVersions{
					"mnist": {
						versions: map[uint32]*pb.ModelVersion{
							1: {
								Model: &pbs.Model{
									Meta: &pbs.MetaData{
										Name: "mnist",
									},
									ModelSpec: &pbs.ModelSpec{
										MemoryBytes: getUint64Ptr(500),
									},
								},
							},
						},
					},
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
			modelDeleted:            false,
			expectedTotalModelBytes: 500,
			numModels:               2,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			modelDeleted := test.state.removeModelVersion(test.modelVersion)
			g.Expect(modelDeleted).To(Equal(test.modelDeleted))
			//check version not exists
			g.Expect(test.state.versionExists("iris", test.modelVersion.GetVersion())).To(Equal(false))
			if !modelDeleted {
				g.Expect(test.state.getModelTotalMemoryBytes("iris")).To(Equal(test.expectedTotalModelBytes))
			}
			g.Expect(test.state.numModels()).To(Equal(test.numModels))
		})
	}
}

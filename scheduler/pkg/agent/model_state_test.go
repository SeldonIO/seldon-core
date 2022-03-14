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
		name               string
		state              *ModelState
		modelVersion       *pb.ModelVersion
		versionAdded       bool
		expectedModelBytes uint64
		expectedTotalBytes uint64
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
			versionAdded:       true,
			expectedModelBytes: 500,
			expectedTotalBytes: 500,
		},
		{
			name: "NewModel (Another Model Exsits)",
			state: &ModelState{
				loadedModels: map[string]*modelVersion{
					"mnist": {
						&pb.ModelVersion{
							Model: &pbs.Model{
								Meta: &pbs.MetaData{
									Name: "mnist",
								},
								ModelSpec: &pbs.ModelSpec{
									MemoryBytes: getUint64Ptr(500),
								},
							},
							Version: 1,
						},
					},
				},
				totalMemoryForAllModels: 500,
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
			versionAdded:       true,
			expectedModelBytes: 500,
			expectedTotalBytes: 1000,
		},
		{
			name: "NewVersion",
			state: &ModelState{
				loadedModels: map[string]*modelVersion{
					"iris": {
						&pb.ModelVersion{
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
					},
				},
				totalMemoryForAllModels: 500,
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
			versionAdded:       false,
			expectedModelBytes: 500,
			expectedTotalBytes: 500,
		},
		{
			name: "VersionExists",
			state: &ModelState{
				loadedModels: map[string]*modelVersion{
					"iris": {
						&pb.ModelVersion{
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
					},
				},
				totalMemoryForAllModels: 500,
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
			versionAdded:       false,
			expectedModelBytes: 500,
			expectedTotalBytes: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			versionAdded, err := test.state.addModelVersion(test.modelVersion)
			g.Expect(versionAdded).To(Equal(test.versionAdded))
			//check version exists
			if versionAdded {
				g.Expect(test.state.versionExists("iris", test.modelVersion.GetVersion())).To(Equal(true))
			} else if err != nil {
				g.Expect(test.state.versionExists("iris", test.modelVersion.GetVersion())).To(Equal(false))
			}
			g.Expect(test.state.getModelMemoryBytes("iris")).To(Equal(test.expectedModelBytes))
			g.Expect(test.state.getTotalMemoryBytesForAllModels()).To(Equal(test.expectedTotalBytes))
		})
	}
}

func TestRemoveModelVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		state              *ModelState
		modelVersion       *pb.ModelVersion
		modelDeleted       bool
		expectedModelBytes uint64
		expectedTotalBytes uint64
		numModels          int
	}

	tests := []test{
		{
			name: "ModelDeleted",
			state: &ModelState{
				loadedModels: map[string]*modelVersion{
					"iris": {
						&pb.ModelVersion{
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
					},
				},
				totalMemoryForAllModels: 500,
			},
			modelVersion: &pb.ModelVersion{
				Model: &pbs.Model{
					Meta: &pbs.MetaData{
						Name: "iris",
					},
				},
				Version: 1,
			},
			modelDeleted:       true,
			expectedModelBytes: 0,
			numModels:          0,
			expectedTotalBytes: 0,
		},
		{
			name: "ModelNotDeleted",
			state: &ModelState{
				loadedModels: map[string]*modelVersion{
					"iris": {
						&pb.ModelVersion{
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
					},
				},
				totalMemoryForAllModels: 500,
			},
			modelVersion: &pb.ModelVersion{
				Model: &pbs.Model{
					Meta: &pbs.MetaData{
						Name: "iris",
					},
				},
				Version: 2,
			},
			modelDeleted:       false,
			expectedModelBytes: 500,
			numModels:          1,
			expectedTotalBytes: 500,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			modelDeleted, _ := test.state.removeModelVersion(test.modelVersion)
			g.Expect(modelDeleted).To(Equal(test.modelDeleted))
			//check version not exists
			g.Expect(test.state.versionExists("iris", test.modelVersion.GetVersion())).To(Equal(false))
			if !modelDeleted {
				g.Expect(test.state.getModelMemoryBytes("iris")).To(Equal(test.expectedModelBytes))
			}
			g.Expect(test.state.numModels()).To(Equal(test.numModels))
			g.Expect(test.state.getTotalMemoryBytesForAllModels()).To(Equal(test.expectedTotalBytes))
		})
	}
}

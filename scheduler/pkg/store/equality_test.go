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

	. "github.com/onsi/gomega"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestModelEquality(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name   string
		model1 *pb.Model
		model2 *pb.Model
		answer ModelEquality
	}

	someMemory := uint64(10_000_000)
	someServer := "server"
	rcloneConfig1 := "abc"
	tests := []test{
		{
			name: "Equal",
			model1: &pb.Model{
				Meta: &pb.MetaData{
					Name: "foo",
				},
				ModelSpec: &pb.ModelSpec{
					Uri:           "gs://mymodels/iris",
					StorageConfig: &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
					Requirements:  []string{"sklearn", "gpu"},
					MemoryBytes:   &someMemory,
					Server:        &someServer,
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 1,
					MaxReplicas: 1,
				},
			},
			model2: &pb.Model{
				Meta: &pb.MetaData{
					Name: "foo",
				},
				ModelSpec: &pb.ModelSpec{
					Uri:           "gs://mymodels/iris",
					StorageConfig: &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
					Requirements:  []string{"sklearn", "gpu"},
					MemoryBytes:   &someMemory,
					Server:        &someServer,
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 1,
					MaxReplicas: 1,
				},
			},
			answer: ModelEquality{Equal: true},
		},
		{
			name: "Equal - model runtime info is ignored",
			model1: &pb.Model{
				Meta: &pb.MetaData{
					Name: "foo",
				},
				ModelSpec: &pb.ModelSpec{
					Uri:              "gs://mymodels/iris",
					StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
					Requirements:     []string{"sklearn", "gpu"},
					MemoryBytes:      &someMemory,
					Server:           &someServer,
					ModelRuntimeInfo: &pb.ModelRuntimeInfo{},
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 1,
					MaxReplicas: 1,
				},
			},
			model2: &pb.Model{
				Meta: &pb.MetaData{
					Name: "foo",
				},
				ModelSpec: &pb.ModelSpec{
					Uri:           "gs://mymodels/iris",
					StorageConfig: &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
					Requirements:  []string{"sklearn", "gpu"},
					MemoryBytes:   &someMemory,
					Server:        &someServer,
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 1,
					MaxReplicas: 1,
				},
			},
			answer: ModelEquality{Equal: true},
		},
		{
			name: "DeploymentsDiffer",
			model1: &pb.Model{
				Meta: &pb.MetaData{
					Name: "foo",
				},
				ModelSpec: &pb.ModelSpec{
					Uri:           "gs://mymodels/iris",
					StorageConfig: &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
					Requirements:  []string{"sklearn", "gpu"},
					MemoryBytes:   &someMemory,
					Server:        &someServer,
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 1,
					MaxReplicas: 1,
				},
			},
			model2: &pb.Model{
				Meta: &pb.MetaData{
					Name: "foo",
				},
				ModelSpec: &pb.ModelSpec{
					Uri:           "gs://mymodels/iris",
					StorageConfig: &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
					Requirements:  []string{"sklearn", "gpu"},
					MemoryBytes:   &someMemory,
					Server:        &someServer,
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    2,
					MinReplicas: 1,
					MaxReplicas: 1,
				},
			},
			answer: ModelEquality{DeploymentSpecDiffers: true},
		},
		{
			name: "ModelsDiffer",
			model1: &pb.Model{
				Meta: &pb.MetaData{
					Name: "foo",
				},
				ModelSpec: &pb.ModelSpec{
					Uri:           "gs://mymodels/iris2",
					StorageConfig: &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
					Requirements:  []string{"sklearn", "gpu"},
					MemoryBytes:   &someMemory,
					Server:        &someServer,
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 1,
					MaxReplicas: 1,
				},
			},
			model2: &pb.Model{
				Meta: &pb.MetaData{
					Name: "foo",
				},
				ModelSpec: &pb.ModelSpec{
					Uri:           "gs://mymodels/iris",
					StorageConfig: &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
					Requirements:  []string{"sklearn", "gpu"},
					MemoryBytes:   &someMemory,
					Server:        &someServer,
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 1,
					MaxReplicas: 1,
				},
			},
			answer: ModelEquality{ModelSpecDiffers: true},
		},
		{
			name: "MetaDiffers",
			model1: &pb.Model{
				Meta: &pb.MetaData{
					Name: "foo2",
				},
				ModelSpec: &pb.ModelSpec{
					Uri:           "gs://mymodels/iris",
					StorageConfig: &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
					Requirements:  []string{"sklearn", "gpu"},
					MemoryBytes:   &someMemory,
					Server:        &someServer,
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 1,
					MaxReplicas: 1,
				},
			},
			model2: &pb.Model{
				Meta: &pb.MetaData{
					Name: "foo",
				},
				ModelSpec: &pb.ModelSpec{
					Uri:           "gs://mymodels/iris",
					StorageConfig: &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
					Requirements:  []string{"sklearn", "gpu"},
					MemoryBytes:   &someMemory,
					Server:        &someServer,
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 1,
					MaxReplicas: 1,
				},
			},
			answer: ModelEquality{MetaDiffers: true},
		},
		{
			name: "ModelsAndDeploymentsDiffer",
			model1: &pb.Model{
				Meta: &pb.MetaData{
					Name: "foo",
				},
				ModelSpec: &pb.ModelSpec{
					Uri:           "gs://mymodels/iris2",
					StorageConfig: &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
					Requirements:  []string{"sklearn", "gpu"},
					MemoryBytes:   &someMemory,
					Server:        &someServer,
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    1,
					MinReplicas: 1,
					MaxReplicas: 1,
				},
			},
			model2: &pb.Model{
				Meta: &pb.MetaData{
					Name: "foo",
				},
				ModelSpec: &pb.ModelSpec{
					Uri:           "gs://mymodels/iris",
					StorageConfig: &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
					Requirements:  []string{"sklearn", "gpu"},
					MemoryBytes:   &someMemory,
					Server:        &someServer,
				},
				DeploymentSpec: &pb.DeploymentSpec{
					Replicas:    2,
					MinReplicas: 1,
					MaxReplicas: 1,
				},
			},
			answer: ModelEquality{ModelSpecDiffers: true, DeploymentSpecDiffers: true},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			answer := ModelEqualityCheck(test.model1, test.model2)
			g.Expect(answer).To(Equal(test.answer))
		})
	}
}

func TestModelSpecEqual(t *testing.T) {
	g := NewGomegaWithT(t)

	version := uint32(1)
	someMemory := uint64(10_000_000)
	someServer := "server"
	rcloneConfig1 := "abc"
	tests := []struct {
		name           string
		modelA         *pb.ModelSpec
		modelB         *pb.ModelSpec
		expectedResult bool
	}{
		{
			name:           "both are empty",
			modelA:         &pb.ModelSpec{},
			modelB:         &pb.ModelSpec{},
			expectedResult: true,
		},
		{
			name:           "both are nil",
			modelA:         nil,
			modelB:         nil,
			expectedResult: true,
		},
		{
			name:           "A is nil",
			modelA:         nil,
			modelB:         &pb.ModelSpec{},
			expectedResult: false,
		},
		{
			name:           "B is nil",
			modelA:         &pb.ModelSpec{},
			modelB:         nil,
			expectedResult: false,
		},
		{
			name: "everything is equal",
			modelA: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			modelB: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			expectedResult: true,
		},
		{
			name: "uri is not equal",
			modelA: &pb.ModelSpec{
				Uri:              "ur",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			modelB: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			expectedResult: false,
		},
		{
			name: "artifact version is not equal",
			modelA: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  nil,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			modelB: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			expectedResult: false,
		},
		{
			name: "storage config is not equal",
			modelA: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			modelB: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			expectedResult: false,
		},
		{
			name: "requirements are not equal",
			modelA: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			modelB: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			expectedResult: false,
		},
		{
			name: "memory bytes is not equal",
			modelA: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			modelB: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      nil,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			expectedResult: false,
		},
		{
			name: "server is not equal",
			modelA: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			modelB: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           nil,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			expectedResult: false,
		},
		{
			name: "parameters are not equal",
			modelA: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			modelB: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value1"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			expectedResult: false,
		},
		{
			name: "model runtime info is ignored",
			modelA: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			modelB: &pb.ModelSpec{
				Uri:              "uri",
				ArtifactVersion:  &version,
				StorageConfig:    &pb.StorageConfig{Config: &pb.StorageConfig_StorageRcloneConfig{StorageRcloneConfig: rcloneConfig1}},
				Requirements:     []string{"sklearn", "gpu"},
				MemoryBytes:      &someMemory,
				Server:           &someServer,
				Parameters:       []*pb.ParameterSpec{{Name: "name", Value: "value"}},
				ModelRuntimeInfo: &pb.ModelRuntimeInfo{ModelRuntimeInfo: &pb.ModelRuntimeInfo_Mlserver{Mlserver: &pb.MLServerModelSettings{ParallelWorkers: uint32(1)}}},
			},
			expectedResult: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := modelSpecEqual(test.modelA, test.modelB)
			g.Expect(result).To(Equal(test.expectedResult))
		})
	}
}

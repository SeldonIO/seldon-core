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

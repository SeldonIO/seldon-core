package controllers

import (
	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"testing"
)

func createTestSeldonDeployment() *machinelearningv1.SeldonDeployment {
	var modelType = machinelearningv1.MODEL
	key := types.NamespacedName{
		Name:      "dep",
		Namespace: "default",
	}
	return &machinelearningv1.SeldonDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      key.Name,
			Namespace: key.Namespace,
		},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Name: "mydep",
			Predictors: []machinelearningv1.PredictorSpec{
				{
					Name: "p1",
					ComponentSpecs: []*machinelearningv1.SeldonPodSpec{
						{
							Spec: v1.PodSpec{
								Containers: []v1.Container{
									{
										Image: "seldonio/mock_classifier:1.0",
										Name:  "classifier",
									},
								},
							},
						},
					},
					Graph: &machinelearningv1.PredictiveUnit{
						Name: "classifier",
						Type: &modelType,
						Endpoint: &machinelearningv1.Endpoint{
							Type: machinelearningv1.GRPC,
						},
					},
				},
			},
		},
	}
}

func TestExecutorCreateNoEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	envExecutorImage = ""
	envExecutorImageRelated = ""
	mlDep := createTestSeldonDeployment()
	_, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).ToNot(BeNil())
}

func TestExecutorCreateEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	imageName := "myimage"
	envExecutorImage = imageName
	envExecutorImageRelated = ""
	mlDep := createTestSeldonDeployment()
	con, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	g.Expect(con.Image).To(Equal(imageName))
}

func TestExecutorCreateEnvRelated(t *testing.T) {
	g := NewGomegaWithT(t)
	imageName := "myimage"
	imageNameRelated := "myimage2"
	envExecutorImage = imageName
	envExecutorImageRelated = imageNameRelated
	mlDep := createTestSeldonDeployment()
	con, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	g.Expect(con.Image).To(Equal(imageNameRelated))
}

func TestEngineCreateNoEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	envEngineImage = ""
	envEngineImageRelated = ""
	mlDep := createTestSeldonDeployment()
	_, err := createEngineContainerSpec(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).ToNot(BeNil())
}

func TestEngineCreateEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	imageName := "myimage"
	envEngineImage = imageName
	envEngineImageRelated = ""
	mlDep := createTestSeldonDeployment()
	con, err := createEngineContainerSpec(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	g.Expect(con.Image).To(Equal(imageName))
}

func TestEngineCreateEnvRelated(t *testing.T) {
	g := NewGomegaWithT(t)
	imageName := "myimage"
	imageNameRelated := "myimage2"
	envEngineImage = imageName
	envEngineImageRelated = imageNameRelated
	mlDep := createTestSeldonDeployment()
	con, err := createEngineContainerSpec(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	g.Expect(con.Image).To(Equal(imageNameRelated))
}

package controllers

import (
	"testing"

	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"github.com/seldonio/seldon-core/operator/constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func createTestSeldonDeployment() *machinelearningv1.SeldonDeployment {
	var modelType = machinelearningv1.MODEL
	key := types.NamespacedName{
		Name:      "dep",
		Namespace: "default",
	}
	return &machinelearningv1.SeldonDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        key.Name,
			Namespace:   key.Namespace,
			Annotations: make(map[string]string),
		},
		Spec: machinelearningv1.SeldonDeploymentSpec{
			Name:        "mydep",
			Annotations: make(map[string]string),
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
					Graph: machinelearningv1.PredictiveUnit{
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

func cleanEnvImagesExecutor() {
	envUseExecutor = ""
	envExecutorImage = ""
	envExecutorImageRelated = ""
}

func TestExecutorCreateNoEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImagesExecutor()
	envExecutorImage = ""
	envExecutorImageRelated = ""
	mlDep := createTestSeldonDeployment()
	_, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).ToNot(BeNil())
	cleanEnvImagesExecutor()
}

func TestExecutorCreateEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImagesExecutor()
	imageName := "myimage"
	envExecutorImage = imageName
	envExecutorImageRelated = ""
	mlDep := createTestSeldonDeployment()
	con, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	g.Expect(con.Image).To(Equal(imageName))
	cleanEnvImagesExecutor()
}

func TestExecutorCreateEnvRelated(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImagesExecutor()
	imageName := "myimage"
	imageNameRelated := "myimage2"
	envExecutorImage = imageName
	envExecutorImageRelated = imageNameRelated
	mlDep := createTestSeldonDeployment()
	con, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	g.Expect(con.Image).To(Equal(imageNameRelated))
	cleanEnvImagesExecutor()
}

func TestExecutorCreateKafka(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImagesExecutor()
	mlDep := createTestSeldonDeployment()
	mlDep.Spec.ServerType = machinelearningv1.ServerKafka
	_, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).ToNot(BeNil())
	cleanEnvImagesExecutor()
}

func TestEngineCreateLoggerParams(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImagesExecutor()
	envExecutorImage = "executor"
	mlDep := createTestSeldonDeployment()
	con, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	for idx, arg := range con.Args {
		if arg == "--log_work_buffer_size" {
			g.Expect(con.Args[idx+1]).To(Equal(constants.DefaultExecutorReqLoggerWorkQueueSize))
		}
		if arg == "--log_write_timeout_ms" {
			g.Expect(con.Args[idx+1]).To(Equal(constants.DefaultExecutorReqLoggerWriteTimeoutMs))
		}
	}
	cleanEnvImagesExecutor()
}

func TestEngineCreateLoggerParamsEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImagesExecutor()
	envExecutorImage = "executor"
	executorReqLoggerWorkQueueSize = "1"
	executorReqLoggerWriteTimeoutMs = "1"
	mlDep := createTestSeldonDeployment()
	con, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	for idx, arg := range con.Args {
		if arg == "--log_work_buffer_size" {
			g.Expect(con.Args[idx+1]).To(Equal("1"))
		}
		if arg == "--log_write_timeout_ms" {
			g.Expect(con.Args[idx+1]).To(Equal("1"))
		}
	}
	cleanEnvImagesExecutor()
	executorReqLoggerWorkQueueSize = constants.DefaultExecutorReqLoggerWorkQueueSize
	executorReqLoggerWriteTimeoutMs = constants.DefaultExecutorReqLoggerWriteTimeoutMs
}

func TestEngineCreateLoggerAnnotation(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImagesExecutor()
	envExecutorImage = "executor"
	mlDep := createTestSeldonDeployment()
	mlDep.Annotations[machinelearningv1.ANNOTATION_LOGGER_WORK_QUEUE_SIZE] = "22"
	mlDep.Annotations[machinelearningv1.ANNOTATION_LOGGER_WRITE_TIMEOUT_MS] = "5"
	con, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	for idx, arg := range con.Args {
		if arg == "--log_work_buffer_size" {
			g.Expect(con.Args[idx+1]).To(Equal("22"))
		}
		if arg == "--log_write_timeout_ms" {
			g.Expect(con.Args[idx+1]).To(Equal("5"))
		}
	}
	cleanEnvImagesExecutor()
}

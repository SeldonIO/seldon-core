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

func cleanEnvImages() {
	envUseExecutor = ""
	envExecutorImage = ""
	envExecutorImageRelated = ""
	envEngineImage = ""
	envEngineImageRelated = ""
}

func setUseExecutorAnnotation(mlDep *machinelearningv1.SeldonDeployment, useExecutor string) {
	mlDep.Spec.Annotations[machinelearningv1.ANNOTATION_EXECUTOR] = useExecutor
}

func conductExecutorUsageTest(t *testing.T, testEnvUseExecutor, testAnnotationUseExecutor string, expectedExecutorEnabled bool) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	mlDep := createTestSeldonDeployment()
	if testAnnotationUseExecutor != "" {
		setUseExecutorAnnotation(mlDep, testAnnotationUseExecutor)
	}
	if testEnvUseExecutor != "" {
		envUseExecutor = testEnvUseExecutor
	}
	executorEnabled := isExecutorEnabled(mlDep)
	g.Expect(executorEnabled).To(Equal(expectedExecutorEnabled))
	cleanEnvImages()
}

func TestUseExecutor(t *testing.T) {
	conductExecutorUsageTest(t, "", "", false)

	// environment variable works as default if annotation is not set
	conductExecutorUsageTest(t, "false", "", false)
	conductExecutorUsageTest(t, "true", "", true)

	// annotation always takes priority
	conductExecutorUsageTest(t, "true", "true", true)
	conductExecutorUsageTest(t, "true", "false", false)
	conductExecutorUsageTest(t, "false", "true", true)
	conductExecutorUsageTest(t, "false", "false", false)
}

func TestExecutorCreateNoEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	envExecutorImage = ""
	envExecutorImageRelated = ""
	mlDep := createTestSeldonDeployment()
	_, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).ToNot(BeNil())
	cleanEnvImages()
}

func TestExecutorCreateEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	imageName := "myimage"
	envExecutorImage = imageName
	envExecutorImageRelated = ""
	mlDep := createTestSeldonDeployment()
	con, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	g.Expect(con.Image).To(Equal(imageName))
	cleanEnvImages()
}

func TestExecutorCreateEnvRelated(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	imageName := "myimage"
	imageNameRelated := "myimage2"
	envExecutorImage = imageName
	envExecutorImageRelated = imageNameRelated
	mlDep := createTestSeldonDeployment()
	con, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	g.Expect(con.Image).To(Equal(imageNameRelated))
	cleanEnvImages()
}

func TestEngineCreateNoEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	envEngineImage = ""
	envEngineImageRelated = ""
	mlDep := createTestSeldonDeployment()
	_, err := createEngineContainerSpec(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).ToNot(BeNil())
	cleanEnvImages()
}

func TestEngineCreateEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	imageName := "myimage"
	envEngineImage = imageName
	envEngineImageRelated = ""
	mlDep := createTestSeldonDeployment()
	con, err := createEngineContainerSpec(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	g.Expect(con.Image).To(Equal(imageName))
	cleanEnvImages()
}

func TestEngineCreateEnvRelated(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	imageName := "myimage"
	imageNameRelated := "myimage2"
	envEngineImage = imageName
	envEngineImageRelated = imageNameRelated
	mlDep := createTestSeldonDeployment()
	con, err := createEngineContainerSpec(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	g.Expect(con.Image).To(Equal(imageNameRelated))
	cleanEnvImages()
}

func TestExecutorCreateKafka(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	mlDep := createTestSeldonDeployment()
	mlDep.Spec.ServerType = machinelearningv1.ServerKafka
	_, err := createExecutorContainer(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).ToNot(BeNil())
	cleanEnvImages()
}

func TestEngineCreateLoggerParams(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
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
	cleanEnvImages()
}

func TestEngineCreateLoggerParamsEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
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
	cleanEnvImages()
	executorReqLoggerWorkQueueSize = constants.DefaultExecutorReqLoggerWorkQueueSize
	executorReqLoggerWriteTimeoutMs = constants.DefaultExecutorReqLoggerWriteTimeoutMs
}

func TestEngineCreateLoggerAnnotation(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
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
	cleanEnvImages()
}

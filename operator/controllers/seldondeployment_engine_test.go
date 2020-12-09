package controllers

import (
	"testing"

	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func createTestSeldonDeployment(svcOrchSpec machinelearningv1.SvcOrchSpec) *machinelearningv1.SeldonDeployment {
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
					SvcOrchSpec: svcOrchSpec,
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
	mlDep := createTestSeldonDeployment(machinelearningv1.SvcOrchSpec{})
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
	mlDep := createTestSeldonDeployment(machinelearningv1.SvcOrchSpec{})
	_, err := createExecutorContainerSpec(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).ToNot(BeNil())
	cleanEnvImages()
}

func TestExecutorCreateEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	imageName := "myimage"
	envExecutorImage = imageName
	envExecutorImageRelated = ""
	mlDep := createTestSeldonDeployment(machinelearningv1.SvcOrchSpec{})
	con, err := createExecutorContainerSpec(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
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
	mlDep := createTestSeldonDeployment(machinelearningv1.SvcOrchSpec{})
	con, err := createExecutorContainerSpec(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	g.Expect(con.Image).To(Equal(imageNameRelated))
	cleanEnvImages()
}

func TestExecutorWithoutSvcOrchSpec(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	imageName := "myimage"
	envExecutorImage = imageName
	envExecutorImageRelated = "x"
	mlDep := createTestSeldonDeployment(machinelearningv1.SvcOrchSpec{})
	setUseExecutorAnnotation(mlDep, "true")
	engineDep, err := createEngineDeployment(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2)
	g.Expect(err).To(BeNil())
	// Default annotations for executor
	g.Expect(engineDep.Spec.Template.ObjectMeta.Annotations).To(Equal(map[string]string{
		"prometheus.io/path":   getPrometheusPath(mlDep),
		"prometheus.io/scrape": "true",
	}))

	cleanEnvImages()
}

func TestExecutorWithSvcOrchSpec(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	imageName := "myimage"
	envExecutorImage = imageName
	envExecutorImageRelated = "x"
	svcOrchSpec := machinelearningv1.SvcOrchSpec{
		Annotations: map[string]string{
			"prometheus.io/scrape": "false",
			"custom/annotation1":   "value1",
			"custom/annotation2":   "value2",
		},
		Tolerations: []v1.Toleration{
			{
				Key:      "spotInstance",
				Operator: "Exists",
				Effect:   "PreferNoSchedule",
			},
		},
	}
	mlDep := createTestSeldonDeployment(svcOrchSpec)
	setUseExecutorAnnotation(mlDep, "true")
	engineDep, err := createEngineDeployment(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2)
	g.Expect(err).To(BeNil())
	// fmt.Printf("%+v", engineDep.Spec.Template.ObjectMeta)
	g.Expect(engineDep.Spec.Template.ObjectMeta.Annotations).To(Equal(map[string]string{
		"prometheus.io/path":   getPrometheusPath(mlDep),
		"prometheus.io/scrape": "false",
		"custom/annotation1":   "value1",
		"custom/annotation2":   "value2",
	}))
	g.Expect(engineDep.Spec.Template.Spec.Tolerations).To(Equal([]v1.Toleration{
		{
			Key:      "spotInstance",
			Operator: "Exists",
			Effect:   "PreferNoSchedule",
		},
	}))

	cleanEnvImages()
}

func TestEngineCreateNoEnv(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	envEngineImage = ""
	envEngineImageRelated = ""
	mlDep := createTestSeldonDeployment(machinelearningv1.SvcOrchSpec{})
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
	mlDep := createTestSeldonDeployment(machinelearningv1.SvcOrchSpec{})
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
	mlDep := createTestSeldonDeployment(machinelearningv1.SvcOrchSpec{})
	con, err := createEngineContainerSpec(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).To(BeNil())
	g.Expect(con.Image).To(Equal(imageNameRelated))
	cleanEnvImages()
}

func TestExecutorCreateKafka(t *testing.T) {
	g := NewGomegaWithT(t)
	cleanEnvImages()
	mlDep := createTestSeldonDeployment(machinelearningv1.SvcOrchSpec{})
	mlDep.Spec.ServerType = machinelearningv1.ServerKafka
	_, err := createExecutorContainerSpec(mlDep, &mlDep.Spec.Predictors[0], "", 1, 2, &v1.ResourceRequirements{})
	g.Expect(err).ToNot(BeNil())
	cleanEnvImages()
}

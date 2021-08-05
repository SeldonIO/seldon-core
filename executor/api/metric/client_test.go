package metric

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	v12 "k8s.io/api/core/v1"
)

func TestNewFromtMeta(t *testing.T) {
	g := NewGomegaWithT(t)

	const imageName = "image"
	const imageVersion = "1.2"
	const modelName = "classifier"
	const deploymentName = "dep"
	predictor := v1.PredictorSpec{
		Name: "",
		Graph: v1.PredictiveUnit{
			Name: modelName,
		},
		ComponentSpecs: []*v1.SeldonPodSpec{
			&v1.SeldonPodSpec{
				Metadata: v1.ObjectMeta{},
				Spec: v12.PodSpec{
					Containers: []v12.Container{
						v12.Container{
							Name:  modelName,
							Image: imageName + ":" + imageVersion,
						},
					},
				},
				HpaSpec: nil,
			},
		},
	}

	metrics := NewClientMetrics(&predictor, deploymentName, modelName)

	g.Expect(metrics.ImageName).To(Equal(imageName))
	g.Expect(metrics.ImageVersion).To(Equal(imageVersion))
	g.Expect(metrics.ModelName).To(Equal(modelName))
	g.Expect(metrics.DeploymentName).To(Equal(deploymentName))
}

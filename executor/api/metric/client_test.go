package metric

import (
	"github.com/onsi/gomega"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning/v1"
	v12 "k8s.io/api/core/v1"
	v1meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"testing"
)

func TestNewFromtMeta(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	const imageName = "image"
	const imageVersion = "1.2"
	const modelName = "classifier"
	const deploymentName = "dep"
	predictor := v1.PredictorSpec{
		Name: "",
		Graph: &v1.PredictiveUnit{
			Name: modelName,
		},
		ComponentSpecs: []*v1.SeldonPodSpec{
			&v1.SeldonPodSpec{
				Metadata: v1meta.ObjectMeta{},
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

	g.Expect(metrics.ImageName).To(gomega.Equal(imageName))
	g.Expect(metrics.ImageVersion).To(gomega.Equal(imageVersion))
	g.Expect(metrics.ModelName).To(gomega.Equal(modelName))
	g.Expect(metrics.DeploymentName).To(gomega.Equal(deploymentName))
}

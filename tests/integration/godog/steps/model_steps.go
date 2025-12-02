package steps

import (
	"github.com/cucumber/godog"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Model struct {
	model *mlopsv1alpha1.Model
}

type TestModelConfig struct {
	Name       string
	StorageURI string
	Runtime    string // "mlserver", etc
	// maybe: resources, env, etc
}

var TestModels = map[string]TestModelConfig{
	"iris": {
		Name:       "iris-model",
		StorageURI: "s3://my-bucket/models/iris/",
		Runtime:    "mlserver",
	},
	"fraud-detector": {
		Name:       "fraud-detector",
		StorageURI: "gs://other-bucket/models/fraud/",
		Runtime:    "mlserver",
	},
}

func LoadModelSteps(ctx *godog.ScenarioContext, w *World) {
	ctx.Step(`I have a model "([^"]+)"`, w.CurrentModel.IHaveAModel)
}

func (m *Model) IHaveAModel(model string) error {

	m.model = &mlopsv1alpha1.Model{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec:       mlopsv1alpha1.ModelSpec{},
		Status:     mlopsv1alpha1.ModelStatus{},
	}

	return nil
}

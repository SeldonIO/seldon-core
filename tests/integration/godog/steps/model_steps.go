package steps

import (
	"fmt"

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

var testModels = map[string]TestModelConfig{
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
	// Model Operations
	ctx.Step(`^I have an? "([^"]+)" model$`, w.CurrentModel.IHaveAModel)
	ctx.Step(`^the model has "(\d+)" min replicas$`, w.CurrentModel.SetMinReplicas)
	ctx.Step(`^the model has "(\d+)" max replicas$`, w.CurrentModel.SetMaxReplicas)
	ctx.Step(`^the model has "(\d+)" replicas$`, w.CurrentModel.SetReplicas)
	// Model Deployments
	ctx.Step(`^the model is applied$`, w.ApplyModel)
	// Model Assertions
	ctx.Step(`^the model should be Ready$`)
	ctx.Step(`^the model status message should be "([^"]+)"$`)

}

func (m *Model) IHaveAModel(model string) error {
	testModel, ok := testModels[model]
	if !ok {
		return fmt.Errorf("model %s not found", model)
	}

	modelName := fmt.Sprintf("%s-%s", testModel.Name, randomSuffix(3))

	m.model = &mlopsv1alpha1.Model{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Model",
			APIVersion: "mlops.seldon.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:                       modelName,
			GenerateName:               "",
			Namespace:                  "",
			SelfLink:                   "",
			UID:                        "",
			ResourceVersion:            "",
			Generation:                 0,
			CreationTimestamp:          metav1.Time{},
			DeletionTimestamp:          nil,
			DeletionGracePeriodSeconds: nil,
			Labels:                     nil,
			Annotations:                nil,
			OwnerReferences:            nil,
			Finalizers:                 nil,
			ManagedFields:              nil,
		},
		Spec:   mlopsv1alpha1.ModelSpec{},
		Status: mlopsv1alpha1.ModelStatus{},
	}

	return nil
}

func NewModel() *Model {
	return &Model{model: &mlopsv1alpha1.Model{}}
}

func (m *Model) SetMinReplicas(replicas int) {

}

func (m *Model) SetMaxReplicas(replicas int) {}

func (m *Model) SetReplicas(replicas int) {}

func (w *World) ApplyModel() error {

	// retrieve current model and apply in k8s
	err := w.kubeClient.ApplyModel(w.CurrentModel.model)

	if err != nil {
		return err
	}

	// add the model to track and undo model in scenario
	w.Models[w.CurrentModel.model.Name] = w.CurrentModel
	return nil

}

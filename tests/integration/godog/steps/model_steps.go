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

// map to have all common testing model definitions for testing popular models
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

func LoadDomSteps(ctx *godog.ScenarioContext, w *World) {
	// Model Operations
	ctx.Step(`^I deploy model spec:$`, w.deployModelSpec)
	ctx.Step(`^the model "([^"]+)" should eventually become Ready with timeout "([^"]+)"$`, w.waitForModelReady)
	ctx.Step(`^send inference request with timeout "([^"]+)" to model "([^"]+)" with payload:$`, w.sendHTTPInferenceRequest)
	ctx.Step(`^expect http response status code "([^"]*)"$`, w.httpRespCheckStatus)
	ctx.Step(`^expect http response body to contain JSON:$`, w.httpRespCheckBodyContainsJSON)
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

// ApplyModel model is aware of namespace and testsuite config and it might add extra information to the cr that the step hasn't added like namespace
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

func (w *World) ModelReady() error {
	return nil
}

func (w *World) AssertModelStatus(status string) error {
	return nil
}

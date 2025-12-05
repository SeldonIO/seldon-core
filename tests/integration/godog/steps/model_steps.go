package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/godog/scenario/assertions"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Model struct {
	model *mlopsv1alpha1.Model
	world *World
}

type TestModelConfig struct {
	Name         string
	StorageURI   string
	Requirements []string // requirements might have to be applied on the applied of k8s
}

// map to have all common testing model definitions for testing popular models
// todo: this requirements might have to be empty and automatically selected by the applier based on config if they aren't explicitly added by the scenario
var testModels = map[string]TestModelConfig{
	"iris": {
		Name:         "iris",
		StorageURI:   "gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn",
		Requirements: []string{"sklearn"},
	},
	"fraud-detector": {
		Name:         "fraud-detector",
		StorageURI:   "gs://other-bucket/models/fraud/",
		Requirements: []string{"sklearn"},
	},
}

func LoadModelSteps(ctx *godog.ScenarioContext, w *Model) {
	// Model Operations
	ctx.Step(`^I have an? "([^"]+)" model$`, w.IHaveAModel)
	ctx.Step(`^the model has "(\d+)" min replicas$`, w.SetMinReplicas)
	ctx.Step(`^the model has "(\d+)" max replicas$`, w.SetMaxReplicas)
	ctx.Step(`^the model has "(\d+)" replicas$`, w.SetReplicas)
	// Model Deployments
	ctx.Step(`^the model is applied$`, w.ApplyModel)
	// Model Assertions
	ctx.Step(`^the model (?:should )?eventually become(?:s)? Ready$`, w.ModelReady)
	ctx.Step(`^the model status message should be "([^"]+)"$`, w.AssertModelStatus)

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
		Spec: mlopsv1alpha1.ModelSpec{
			InferenceArtifactSpec: mlopsv1alpha1.InferenceArtifactSpec{
				ModelType:       nil,
				StorageURI:      testModel.StorageURI,
				ArtifactVersion: nil,
				SchemaURI:       nil,
				SecretName:      nil,
			},
			Requirements: testModel.Requirements,
			Memory:       nil,
			ScalingSpec:  mlopsv1alpha1.ScalingSpec{},
			Server:       nil,
			PreLoaded:    false,
			Dedicated:    false,
			Logger:       nil,
			Explainer:    nil,
			Parameters:   nil,
			Llm:          nil,
			Dataflow:     nil,
		},
		Status: mlopsv1alpha1.ModelStatus{},
	}

	return nil
}
func NewModel(world *World) *Model {
	return &Model{model: &mlopsv1alpha1.Model{}, world: world}
}

func (m *Model) Reset(world *World) {
	m.world.CurrentModel.model = &mlopsv1alpha1.Model{}
	m.world.CurrentModel.world = world
}

func (m *Model) SetMinReplicas(replicas int) {

}

func (m *Model) SetMaxReplicas(replicas int) {}

func (m *Model) SetReplicas(replicas int) {}

// ApplyModel model is aware of namespace and testsuite config and it might add extra information to the cr that the step hasn't added like namespace
func (m *Model) ApplyModel() error {

	// retrieve current model and apply in k8s
	err := m.world.KubeClient.ApplyModel(m.world.CurrentModel.model)

	if err != nil {
		return err
	}

	return nil

}

func (m *Model) ModelReady() error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// m.world.CurrentModel.model is assumed to be *mlopsv1alpha1.Model
	// which implements runtime.Object, so no cast needed.
	return m.world.WatcherStorage.WaitFor(
		ctx,
		m.world.CurrentModel.model,
		assertions.ModelReady,
	)
}

func (m *Model) AssertModelStatus(status string) error {

	return nil
}

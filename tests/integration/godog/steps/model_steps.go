/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package steps

import (
	"context"
	"fmt"
	"time"

	"github.com/cucumber/godog"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	"github.com/seldonio/seldon-core/tests/integration/godog/steps/assertions"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Model struct {
	model *mlopsv1alpha1.Model
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

func LoadModelSteps(scenario *godog.ScenarioContext, w *World) {
	// Model Operations
	scenario.Step(`^I have an? "([^"]+)" model$`, func(modelName string) error {
		return w.CurrentModel.IHaveAModel(modelName, w.Label)
	})
	scenario.Step(`^the model has "(\d+)" min replicas$`, w.CurrentModel.SetMinReplicas)
	scenario.Step(`^the model has "(\d+)" max replicas$`, w.CurrentModel.SetMaxReplicas)
	scenario.Step(`^the model has "(\d+)" replicas$`, w.CurrentModel.SetReplicas)
	// Model Deployments
	scenario.Step(`^the model is applied$`, func() error {
		return w.CurrentModel.ApplyModel(w.KubeClient)
	})
	// Model Assertions
	scenario.Step(`^the model (?:should )?eventually become(?:s)? Ready$`, func() error {
		return w.CurrentModel.ModelReady(w.WatcherStorage)
	})
	scenario.Step(`^the model status message should be "([^"]+)"$`, w.CurrentModel.AssertModelStatus)
}

func LoadExplicitModelSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^I deploy model spec:$`, func(spec *godog.DocString) error {
		return w.CurrentModel.deployModelSpec(spec, w.namespace, w.KubeClient)
	})
	scenario.Step(`^the model "([^"]+)" should eventually become Ready with timeout "([^"]+)"$`, func(model, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.CurrentModel.waitForModelReady(ctx, model)
		})
	})
}

func LoadInferenceSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^send HTTP inference request with timeout "([^"]+)" to model "([^"]+)" with payload:$`, func(timeout, model string, payload *godog.DocString) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.infer.sendHTTPModelInferenceRequest(ctx, model, payload)
		})
	})
	scenario.Step(`^send gRPC inference request with timeout "([^"]+)" to model "([^"]+)" with payload:$`, func(timeout, model string, payload *godog.DocString) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.infer.sendGRPCModelInferenceRequest(ctx, model, payload)
		})
	})
	scenario.Step(`^expect http response status code "([^"]*)"$`, w.infer.httpRespCheckStatus)
	scenario.Step(`^expect http response body to contain JSON:$`, w.infer.httpRespCheckBodyContainsJSON)
	scenario.Step(`^expect gRPC response body to contain JSON:$`, w.infer.gRPCRespCheckBodyContainsJSON)
}

func (m *Model) deployModelSpec(spec *godog.DocString, namespace string, _ *k8sclient.K8sClient) error {
	modelSpec := &mlopsv1alpha1.Model{}
	if err := yaml.Unmarshal([]byte(spec.Content), &modelSpec); err != nil {
		return fmt.Errorf("failed unmarshalling model spec: %w", err)
	}
	modelSpec.Namespace = namespace
	// TODO: uncomment when auto-gen k8s client merged
	//if _, err := w.k8sClient.MlopsV1alpha1().Models(w.namespace).Create(context.TODO(), modelSpec, metav1.CreateOptions{}); err != nil {
	//	return fmt.Errorf("failed creating model: %w", err)
	//}
	return nil
}

func (m *Model) IHaveAModel(model string, label map[string]string) error {
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
			Labels:                     label,
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
func NewModel() *Model {
	return &Model{model: &mlopsv1alpha1.Model{}}
}

func (m *Model) Reset(world *World) {
	world.CurrentModel.model = &mlopsv1alpha1.Model{}
}

func (m *Model) SetMinReplicas(replicas int) {

}

func (m *Model) SetMaxReplicas(replicas int) {}

func (m *Model) SetReplicas(replicas int) {}

// ApplyModel model is aware of namespace and testsuite config and it might add extra information to the cr that the step hasn't added like namespace
func (m *Model) ApplyModel(k *k8sclient.K8sClient) error {

	// retrieve current model and apply in k8s
	err := k.ApplyModel(m.model)

	if err != nil {
		return err
	}

	return nil
}

func (m *Model) ModelReady(w k8sclient.WatcherStorage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// m.world.CurrentModel.model is assumed to be *mlopsv1alpha1.Model
	// which implements runtime.Object, so no cast needed.
	return w.WaitFor(
		ctx,
		m.model,
		assertions.ModelReady,
	)
}

func (m *Model) AssertModelStatus(status string) error {

	return nil
}

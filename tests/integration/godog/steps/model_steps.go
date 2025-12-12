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
	"maps"
	"time"

	"github.com/cucumber/godog"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/generated/clientset/versioned"
	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	"github.com/seldonio/seldon-core/tests/integration/godog/steps/assertions"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Model struct {
	label          map[string]string
	namespace      string
	model          *mlopsv1alpha1.Model
	modelName      string
	modelType      string
	k8sClient      versioned.Interface
	watcherStorage k8sclient.WatcherStorage
	log            logrus.FieldLogger
}

type TestModelConfig struct {
	Name                  string
	StorageURI            string
	Requirements          []string // requirements might have to be applied on the applied of k8s
	ValidInferenceRequest string
	ValidJSONResponse     string
}

// map to have all common testing model definitions for testing popular models
// todo: this requirements might have to be empty and automatically selected by the applier based on config if they aren't explicitly added by the scenario
var testModels = map[string]TestModelConfig{
	"iris": {
		Name:                  "iris",
		StorageURI:            "gs://seldon-models/scv2/samples/mlserver_1.3.5/iris-sklearn",
		Requirements:          []string{"sklearn"},
		ValidInferenceRequest: `{"inputs": [{"name": "predict", "shape": [1, 4], "datatype": "FP32", "data": [[1, 2, 3, 4]]}]}`,
	},
	"income-xgb": {
		Name:                  "income-xgb",
		StorageURI:            "gs://seldon-models/scv2/samples/mlserver_1.3.5/income-xgb",
		Requirements:          []string{"xgboost"},
		ValidInferenceRequest: `{ "parameters": {"content_type": "pd"}, "inputs": [{"name": "Age", "shape": [1, 1], "datatype": "INT64", "data": [47]},{"name": "Workclass", "shape": [1, 1], "datatype": "INT64", "data": [4]},{"name": "Education", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Marital Status", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Occupation", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Relationship", "shape": [1, 1], "datatype": "INT64", "data": [3]},{"name": "Race", "shape": [1, 1], "datatype": "INT64", "data": [4]},{"name": "Sex", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Capital Gain", "shape": [1, 1], "datatype": "INT64", "data": [0]},{"name": "Capital Loss", "shape": [1, 1], "datatype": "INT64", "data": [0]},{"name": "Hours per week", "shape": [1, 1], "datatype": "INT64", "data": [40]},{"name": "Country", "shape": [1, 1], "datatype": "INT64", "data": [9]}]}`,
	},
	"mnist-onnx": {
		Name:                  "mnist-onnx",
		StorageURI:            "gs://seldon-models/scv2/samples/triton_23-03/mnist-onnx",
		Requirements:          []string{"onnx"},
		ValidInferenceRequest: `{"inputs":[{"name":"Input3","data":[],"datatype":"FP32","shape":[]}]}`,
	},
	"income-lgb": {
		Name:                  "income-lgb",
		StorageURI:            "gs://seldon-models/scv2/samples/mlserver_1.3.5/income-lgb",
		Requirements:          []string{"lightgbm"},
		ValidInferenceRequest: `{"inputs": [{"name": "Age", "shape": [1, 1], "datatype": "INT64", "data": [47]},{"name": "Workclass", "shape": [1, 1], "datatype": "INT64", "data": [4]},{"name": "Education", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Marital Status", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Occupation", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Relationship", "shape": [1, 1], "datatype": "INT64", "data": [3]},{"name": "Race", "shape": [1, 1], "datatype": "INT64", "data": [4]},{"name": "Sex", "shape": [1, 1], "datatype": "INT64", "data": [1]},{"name": "Capital Gain", "shape": [1, 1], "datatype": "INT64", "data": [0]},{"name": "Capital Loss", "shape": [1, 1], "datatype": "INT64", "data": [0]},{"name": "Hours per week", "shape": [1, 1], "datatype": "INT64", "data": [40]},{"name": "Country", "shape": [1, 1], "datatype": "INT64", "data": [9]}]}`,
	},
	"wine": {
		Name:                  "wine",
		StorageURI:            "gs://seldon-models/scv2/samples/mlserver_1.3.5/wine-mlflow",
		Requirements:          []string{"mlflow"},
		ValidInferenceRequest: `{ "inputs": [ { "name": "fixed acidity", "shape": [1], "datatype": "FP32", "data": [7.4] }, { "name": "volatile acidity", "shape": [1], "datatype": "FP32", "data": [0.7000] }, { "name": "citric acid", "shape": [1], "datatype": "FP32", "data": [0] }, { "name": "residual sugar", "shape": [1], "datatype": "FP32", "data": [1.9] }, { "name": "chlorides", "shape": [1], "datatype": "FP32", "data": [0.076] }, { "name": "free sulfur dioxide", "shape": [1], "datatype": "FP32", "data": [11] }, { "name": "total sulfur dioxide", "shape": [1], "datatype": "FP32", "data": [34] }, { "name": "density", "shape": [1], "datatype": "FP32", "data": [0.9978] }, { "name": "pH", "shape": [1], "datatype": "FP32", "data": [3.51] }, { "name": "sulphates", "shape": [1], "datatype": "FP32", "data": [0.56] }, { "name": "alcohol", "shape": [1], "datatype": "FP32", "data": [9.4] } ] }`,
	},
	"mnist-pytorch": {
		Name:                  "mnist-pytorch",
		StorageURI:            "gs://seldon-models/scv2/samples/triton_23-03/mnist-pytorch",
		Requirements:          []string{"pytorch"},
		ValidInferenceRequest: `{'inputs': [{'name': 'x__0', 'data': [0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.3294117748737335, 0.7254902124404907, 0.6235294342041016, 0.5921568870544434, 0.23529411852359772, 0.1411764770746231, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.8705882430076599, 0.9960784316062927, 0.9960784316062927, 0.9960784316062927, 0.9960784316062927, 0.9450980424880981, 0.7764706015586853, 0.7764706015586853, 0.7764706015586853, 0.7764706015586853, 0.7764706015586853, 0.7764706015586853, 0.7764706015586853, 0.7764706015586853, 0.6666666865348816, 0.20392157137393951, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.26274511218070984, 0.4470588266849518, 0.2823529541492462, 0.4470588266849518, 0.6392157077789307, 0.8901960849761963, 0.9960784316062927, 0.8823529481887817, 0.9960784316062927, 0.9960784316062927, 0.9960784316062927, 0.9803921580314636, 0.8980392217636108, 0.9960784316062927, 0.9960784316062927, 0.5490196347236633, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.06666667014360428, 0.25882354378700256, 0.054901961237192154, 0.26274511218070984, 0.26274511218070984, 0.26274511218070984, 0.23137255012989044, 0.08235294371843338, 0.9254902005195618, 0.9960784316062927, 0.4156862795352936, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.32549020648002625, 0.9921568632125854, 0.8196078538894653, 0.07058823853731155, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.08627451211214066, 0.9137254953384399, 1.0, 0.32549020648002625, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.5058823823928833, 0.9960784316062927, 0.9333333373069763, 0.1725490242242813, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.23137255012989044, 0.9764705896377563, 0.9960784316062927, 0.24313725531101227, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.5215686559677124, 0.9960784316062927, 0.7333333492279053, 0.019607843831181526, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.03529411926865578, 0.8039215803146362, 0.9725490212440491, 0.22745098173618317, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.4941176474094391, 0.9960784316062927, 0.7137255072593689, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.29411765933036804, 0.9843137264251709, 0.9411764740943909, 0.2235294133424759, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.07450980693101883, 0.8666666746139526, 0.9960784316062927, 0.6509804129600525, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0117647061124444, 0.7960784435272217, 0.9960784316062927, 0.8588235378265381, 0.13725490868091583, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.14901961386203766, 0.9960784316062927, 0.9960784316062927, 0.3019607961177826, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.12156862765550613, 0.8784313797950745, 0.9960784316062927, 0.45098039507865906, 0.003921568859368563, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.5215686559677124, 0.9960784316062927, 0.9960784316062927, 0.20392157137393951, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.239215686917305, 0.9490196108818054, 0.9960784316062927, 0.9960784316062927, 0.20392157137393951, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.4745098054409027, 0.9960784316062927, 0.9960784316062927, 0.8588235378265381, 0.1568627506494522, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.4745098054409027, 0.9960784316062927, 0.8117647171020508, 0.07058823853731155, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0], 'datatype': 'FP32', 'shape': [1, 1, 28, 28]}]}`,
	},
	"tfsimple1": {
		Name:                  "tfsimple1",
		StorageURI:            "gs://seldon-models/triton/simple",
		Requirements:          []string{"tensorflow"},
		ValidInferenceRequest: `{"inputs":[{"name":"INPUT0","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]},{"name":"INPUT1","data":[1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16],"datatype":"INT32","shape":[1,16]}]}`,
		ValidJSONResponse:     `[ { "name": "OUTPUT0", "datatype": "INT32", "shape": [ 1, 16 ], "data": [ 2, 4, 6, 8, 10, 12, 14, 16, 18, 20, 22, 24, 26, 28, 30, 32 ] }, { "name": "OUTPUT1", "datatype": "INT32", "shape": [ 1, 16 ], "data": [ 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ] } ]`,
	},
}

func LoadTemplateModelSteps(scenario *godog.ScenarioContext, w *World) {
	// Model Operations
	scenario.Step(`^I have an? "([^"]+)" model$`, func(modelName string) error {
		return w.currentModel.IHaveAModel(modelName)
	})
	scenario.Step(`^the model has "(\d+)" min replicas$`, w.currentModel.SetMinReplicas)
	scenario.Step(`^the model has "(\d+)" max replicas$`, w.currentModel.SetMaxReplicas)
	scenario.Step(`^the model has "(\d+)" replicas$`, w.currentModel.SetReplicas)
	// Model Deployments
	scenario.Step(`^the model is applied$`, func() error {
		return w.currentModel.ApplyModel(w.kubeClient)
	})
	// Model Assertions
	scenario.Step(`^the model (?:should )?eventually become(?:s)? Ready$`, func() error {
		// todo: maybe convert this to a flag
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()

		return w.currentModel.waitForModelReady(ctx)
	})
	scenario.Step(`^the model status message eventually should be "([^"]+)"$`, w.currentModel.AssertModelStatus)
}

func LoadCustomModelSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^I deploy model spec with timeout "([^"]+)":$`, func(timeout string, spec *godog.DocString) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.currentModel.deployModelSpec(ctx, spec)
		})
	})
	scenario.Step(`^the model "([^"]+)" should eventually become Ready with timeout "([^"]+)"$`, func(model, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.currentModel.waitForModelNameReady(ctx, model)
		})
	})
	scenario.Step(`^delete the model "([^"]+)" with timeout "([^"]+)"$`, func(model, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.currentModel.deleteModel(ctx, model)
		})
	})
}

func (m *Model) deployModelSpec(ctx context.Context, spec *godog.DocString) error {
	modelSpec := &mlopsv1alpha1.Model{}
	if err := yaml.Unmarshal([]byte(spec.Content), &modelSpec); err != nil {
		return fmt.Errorf("failed unmarshalling model spec: %w", err)
	}
	modelSpec.Namespace = m.namespace
	m.model = modelSpec
	m.applyScenarioLabel()
	if _, err := m.k8sClient.MlopsV1alpha1().Models(m.namespace).Create(ctx, m.model, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("failed creating model: %w", err)
	}
	return nil
}

func (m *Model) applyScenarioLabel() {
	if m.model.Labels == nil {
		m.model.Labels = make(map[string]string)
	}

	maps.Copy(m.model.Labels, m.label)

	// todo: change this approach
	for k, v := range k8sclient.DefaultCRDTestSuiteLabelMap {
		m.model.Labels[k] = v
	}
}

func (m *Model) IHaveAModel(model string) error {
	testModel, ok := testModels[model]
	if !ok {
		return fmt.Errorf("model %s not found", model)
	}

	modelName := fmt.Sprintf("%s-%s", testModel.Name, randomString(3))
	m.modelName = modelName
	m.modelType = model
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
			Labels:                     m.label,
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
func NewModel(label map[string]string, namespace string, k8sClient versioned.Interface, log logrus.FieldLogger, watcherStorage k8sclient.WatcherStorage) *Model {
	return &Model{label: label, model: &mlopsv1alpha1.Model{}, log: log, namespace: namespace, k8sClient: k8sClient, watcherStorage: watcherStorage}
}

func (m *Model) SetMinReplicas(replicas int) {

}

func (m *Model) SetMaxReplicas(replicas int) {}

func (m *Model) SetReplicas(replicas int) {}

// ApplyModel model is aware of namespace and testsuite config and it might add extra information to the cr that the step hasn't added like namespace
func (m *Model) ApplyModel(k *k8sclient.K8sClient) error {
	// retrieve current model and apply in k8s
	if err := k.ApplyModel(m.model); err != nil {
		return err
	}

	return nil
}

func (m *Model) waitForModelReady(ctx context.Context) error {
	return m.watcherStorage.WaitForObject(
		ctx,
		m.model,               // the k8s object being watched
		assertions.ModelReady, // predicate from steps/assertions
	)
}

func (m *Model) AssertModelStatus(status string) error {

	return nil
}

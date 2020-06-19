package predictor

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"testing"
)

func TestGraphMetadataSimple(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	model := v1.MODEL

	spec := &v1.PredictorSpec{
		Name: "predictor-name",
		Graph: &v1.PredictiveUnit{
			Name: "model-1",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	expectedGrahMetadata := GraphMetadata{
		Name: "predictor-name",
		Models: map[string]ModelMetadata{
			"model-1": {
				Name:     "model-1",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs: []MetadataTensor{
					{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
				},
				Outputs: []MetadataTensor{
					{Name: "output", DataType: "BYTES", Shape: []int{1, 3}},
				},
			},
		},
		GraphInputs: []MetadataTensor{
			{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
		},
		GraphOutputs: []MetadataTensor{
			{Name: "output", DataType: "BYTES", Shape: []int{1, 3}},
		},
	}

	graphMetadata, err := NewGraphMetadata(createPredictorProcess(t), spec)
	g.Expect(err).Should(BeNil())

	expectedJson, err := json.Marshal(expectedGrahMetadata)
	g.Expect(err).Should(BeNil())

	actualJson, err := json.Marshal(graphMetadata)
	g.Expect(err).Should(BeNil())

	g.Expect(actualJson).To(MatchJSON(expectedJson))
}

func TestGraphMetadataTwoLevel(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	model := v1.MODEL

	spec := &v1.PredictorSpec{
		Name: "predictor-name",
		Graph: &v1.PredictiveUnit{
			Name: "model-1",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
			Children: []v1.PredictiveUnit{
				{
					Name: "model-2",
					Type: &model,
					Endpoint: &v1.Endpoint{
						ServiceHost: "foo",
						ServicePort: 9001,
						Type:        v1.REST,
					},
				},
			},
		},
	}
	expectedGrahMetadata := GraphMetadata{
		Name: "predictor-name",
		Models: map[string]ModelMetadata{
			"model-1": {
				Name:     "model-1",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs: []MetadataTensor{
					{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
				},
				Outputs: []MetadataTensor{
					{Name: "output", DataType: "BYTES", Shape: []int{1, 3}},
				},
			},
			"model-2": {
				Name:     "model-2",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs: []MetadataTensor{
					{Name: "input", DataType: "BYTES", Shape: []int{1, 3}},
				},
				Outputs: []MetadataTensor{
					{Name: "output", DataType: "BYTES", Shape: []int{3}},
				},
			},
		},
		GraphInputs: []MetadataTensor{
			{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
		},
		GraphOutputs: []MetadataTensor{
			{Name: "output", DataType: "BYTES", Shape: []int{3}},
		},
	}

	graphMetadata, err := NewGraphMetadata(createPredictorProcess(t), spec)
	g.Expect(err).Should(BeNil())

	expectedJson, err := json.Marshal(expectedGrahMetadata)
	g.Expect(err).Should(BeNil())

	actualJson, err := json.Marshal(graphMetadata)
	g.Expect(err).Should(BeNil())

	g.Expect(actualJson).To(MatchJSON(expectedJson))
}

func TestGraphMetadataInputTransformer(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	model := v1.MODEL
	transformer := v1.TRANSFORMER

	spec := &v1.PredictorSpec{
		Name: "predictor-name",
		Graph: &v1.PredictiveUnit{
			Name: "model-1",
			Type: &transformer,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
			Children: []v1.PredictiveUnit{
				{
					Name: "model-2",
					Type: &model,
					Endpoint: &v1.Endpoint{
						ServiceHost: "foo",
						ServicePort: 9001,
						Type:        v1.REST,
					},
				},
			},
		},
	}
	expectedGrahMetadata := GraphMetadata{
		Name: "predictor-name",
		Models: map[string]ModelMetadata{
			"model-1": {
				Name:     "model-1",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs: []MetadataTensor{
					{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
				},
				Outputs: []MetadataTensor{
					{Name: "output", DataType: "BYTES", Shape: []int{1, 3}},
				},
			},
			"model-2": {
				Name:     "model-2",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs: []MetadataTensor{
					{Name: "input", DataType: "BYTES", Shape: []int{1, 3}},
				},
				Outputs: []MetadataTensor{
					{Name: "output", DataType: "BYTES", Shape: []int{3}},
				},
			},
		},
		GraphInputs: []MetadataTensor{
			{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
		},
		GraphOutputs: []MetadataTensor{
			{Name: "output", DataType: "BYTES", Shape: []int{3}},
		},
	}

	graphMetadata, err := NewGraphMetadata(createPredictorProcess(t), spec)
	g.Expect(err).Should(BeNil())

	expectedJson, err := json.Marshal(expectedGrahMetadata)
	g.Expect(err).Should(BeNil())

	actualJson, err := json.Marshal(graphMetadata)
	g.Expect(err).Should(BeNil())

	g.Expect(actualJson).To(MatchJSON(expectedJson))
}

func TestGraphMetadataOutputTransformer(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	model := v1.MODEL
	outputTransformer := v1.OUTPUT_TRANSFORMER

	spec := &v1.PredictorSpec{
		Name: "predictor-name",
		Graph: &v1.PredictiveUnit{
			Name: "model-2",
			Type: &outputTransformer,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
			Children: []v1.PredictiveUnit{
				{
					Name: "model-1",
					Type: &model,
					Endpoint: &v1.Endpoint{
						ServiceHost: "foo",
						ServicePort: 9001,
						Type:        v1.REST,
					},
				},
			},
		},
	}
	expectedGrahMetadata := GraphMetadata{
		Name: "predictor-name",
		Models: map[string]ModelMetadata{
			"model-1": {
				Name:     "model-1",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs: []MetadataTensor{
					{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
				},
				Outputs: []MetadataTensor{
					{Name: "output", DataType: "BYTES", Shape: []int{1, 3}},
				},
			},
			"model-2": {
				Name:     "model-2",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs: []MetadataTensor{
					{Name: "input", DataType: "BYTES", Shape: []int{1, 3}},
				},
				Outputs: []MetadataTensor{
					{Name: "output", DataType: "BYTES", Shape: []int{3}},
				},
			},
		},
		GraphInputs: []MetadataTensor{
			{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
		},
		GraphOutputs: []MetadataTensor{
			{Name: "output", DataType: "BYTES", Shape: []int{3}},
		},
	}

	graphMetadata, err := NewGraphMetadata(createPredictorProcess(t), spec)
	g.Expect(err).Should(BeNil())

	expectedJson, err := json.Marshal(expectedGrahMetadata)
	g.Expect(err).Should(BeNil())

	actualJson, err := json.Marshal(graphMetadata)
	g.Expect(err).Should(BeNil())

	g.Expect(actualJson).To(MatchJSON(expectedJson))
}

func TestGraphMetadataCombinerModel(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	combiner := v1.COMBINER

	spec := &v1.PredictorSpec{
		Name: "predictor-name",
		Graph: &v1.PredictiveUnit{
			Name: "model-combiner",
			Type: &combiner,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
			Children: []v1.PredictiveUnit{
				{
					Name: "model-a1",
					Type: &model,
					Endpoint: &v1.Endpoint{
						ServiceHost: "foo",
						ServicePort: 9001,
						Type:        v1.REST,
					},
				},
				{
					Name: "model-a2",
					Type: &model,
					Endpoint: &v1.Endpoint{
						ServiceHost: "foo",
						ServicePort: 9002,
						Type:        v1.REST,
					},
				},
			},
		},
	}

	expectedGrahMetadata := GraphMetadata{
		Name: "predictor-name",
		Models: map[string]ModelMetadata{
			"model-combiner": {
				Name:     "model-combiner",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs: []MetadataTensor{
					{Name: "input-1", DataType: "BYTES", Shape: []int{1, 10}},
					{Name: "input-2", DataType: "BYTES", Shape: []int{1, 20}},
				},
				Outputs: []MetadataTensor{
					{Name: "combined output", DataType: "BYTES", Shape: []int{3}},
				},
			},
			"model-a1": {
				Name:     "model-a1",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs: []MetadataTensor{
					{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
				},
				Outputs: []MetadataTensor{
					{Name: "output", DataType: "BYTES", Shape: []int{1, 10}},
				},
			},
			"model-a2": {
				Name:     "model-a2",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs: []MetadataTensor{
					{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
				},
				Outputs: []MetadataTensor{
					{Name: "output", DataType: "BYTES", Shape: []int{1, 20}},
				},
			},
		},
		GraphInputs: []MetadataTensor{
			{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
		},
		GraphOutputs: []MetadataTensor{
			{Name: "combined output", DataType: "BYTES", Shape: []int{3}},
		},
	}

	graphMetadata, err := NewGraphMetadata(createPredictorProcess(t), spec)
	g.Expect(err).Should(BeNil())

	expectedJson, err := json.Marshal(expectedGrahMetadata)
	g.Expect(err).Should(BeNil())

	actualJson, err := json.Marshal(graphMetadata)
	g.Expect(err).Should(BeNil())

	g.Expect(actualJson).To(MatchJSON(expectedJson))
}

func TestGraphMetadataRouter(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)

	model := v1.MODEL
	router := v1.ROUTER

	spec := &v1.PredictorSpec{
		Name: "predictor-name",
		Graph: &v1.PredictiveUnit{
			Name: "model-router",
			Type: &router,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
			Children: []v1.PredictiveUnit{
				{
					Name: "model-a1",
					Type: &model,
					Endpoint: &v1.Endpoint{
						ServiceHost: "foo",
						ServicePort: 9001,
						Type:        v1.REST,
					},
				},
				{
					Name: "model-b1",
					Type: &model,
					Endpoint: &v1.Endpoint{
						ServiceHost: "foo",
						ServicePort: 9002,
						Type:        v1.REST,
					},
				},
			},
		},
	}

	expectedGrahMetadata := GraphMetadata{
		Name: "predictor-name",
		Models: map[string]ModelMetadata{
			"model-router": {
				Name:     "model-router",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs:   nil,
				Outputs:  nil,
			},
			"model-a1": {
				Name:     "model-a1",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs: []MetadataTensor{
					{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
				},
				Outputs: []MetadataTensor{
					{Name: "output", DataType: "BYTES", Shape: []int{1, 10}},
				},
			},
			"model-b1": {
				Name:     "model-b1",
				Platform: "platform-name",
				Versions: []string{"model-version"},
				Inputs: []MetadataTensor{
					{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
				},
				Outputs: []MetadataTensor{
					{Name: "output", DataType: "BYTES", Shape: []int{1, 10}},
				},
			},
		},
		GraphInputs: []MetadataTensor{
			{Name: "input", DataType: "BYTES", Shape: []int{1, 5}},
		},
		GraphOutputs: []MetadataTensor{
			{Name: "output", DataType: "BYTES", Shape: []int{1, 10}},
		},
	}

	graphMetadata, err := NewGraphMetadata(createPredictorProcess(t), spec)
	g.Expect(err).Should(BeNil())

	expectedJson, err := json.Marshal(expectedGrahMetadata)
	g.Expect(err).Should(BeNil())

	actualJson, err := json.Marshal(graphMetadata)
	g.Expect(err).Should(BeNil())

	g.Expect(actualJson).To(MatchJSON(expectedJson))
}

func TestGraphV1Array(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	model := v1.MODEL

	spec := &v1.PredictorSpec{
		Name: "predictor-name",
		Graph: &v1.PredictiveUnit{
			Name: "model-v1-array",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	expectedGrahMetadata := GraphMetadata{
		Name: "predictor-name",
		Models: map[string]ModelMetadata{
			"model-v1-array": {
				ApiVersion: "v1",
				Name:       "model-v1-array",
				Platform:   "platform-name",
				Versions:   []string{"model-version"},
				Inputs: map[string]interface{}{
					"datatype": "array",
					"shape":    []int{2, 2},
				},
				Outputs: map[string]interface{}{
					"datatype": "array",
					"shape":    []int{1},
				},
			},
		},
		GraphInputs: map[string]interface{}{
			"datatype": "array",
			"shape":    []int{2, 2},
		},
		GraphOutputs: map[string]interface{}{
			"datatype": "array",
			"shape":    []int{1},
		},
	}

	graphMetadata, err := NewGraphMetadata(createPredictorProcess(t), spec)
	g.Expect(err).Should(BeNil())

	expectedJson, err := json.Marshal(expectedGrahMetadata)
	g.Expect(err).Should(BeNil())

	actualJson, err := json.Marshal(graphMetadata)
	g.Expect(err).Should(BeNil())

	g.Expect(actualJson).To(MatchJSON(expectedJson))
}

func TestGraphV1JsonData(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	model := v1.MODEL

	spec := &v1.PredictorSpec{
		Name: "predictor-name",
		Graph: &v1.PredictiveUnit{
			Name: "model-v1-jsondata",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	expectedGrahMetadata := GraphMetadata{
		Name: "predictor-name",
		Models: map[string]ModelMetadata{
			"model-v1-jsondata": {
				ApiVersion: "v1",
				Name:       "model-v1-jsondata",
				Platform:   "platform-name",
				Versions:   []string{"model-version"},
				Inputs: map[string]interface{}{
					"datatype": "jsonData",
				},
				Outputs: map[string]interface{}{
					"datatype": "jsonData",
					"schema": map[string]string{
						"custom": "definition",
					},
				},
			},
		},
		GraphInputs: map[string]interface{}{
			"datatype": "jsonData",
		},
		GraphOutputs: map[string]interface{}{
			"datatype": "jsonData",
			"schema": map[string]string{
				"custom": "definition",
			},
		},
	}

	graphMetadata, err := NewGraphMetadata(createPredictorProcess(t), spec)
	g.Expect(err).Should(BeNil())

	expectedJson, err := json.Marshal(expectedGrahMetadata)
	g.Expect(err).Should(BeNil())

	actualJson, err := json.Marshal(graphMetadata)
	g.Expect(err).Should(BeNil())

	g.Expect(actualJson).To(MatchJSON(expectedJson))
}

func TestGraphV1ArrayStringMix(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	model := v1.MODEL

	spec := &v1.PredictorSpec{
		Name: "predictor-name",
		Graph: &v1.PredictiveUnit{
			Name: "model-v1-array-string-mix",
			Type: &model,
			Endpoint: &v1.Endpoint{
				ServiceHost: "foo",
				ServicePort: 9000,
				Type:        v1.REST,
			},
		},
	}

	expectedGrahMetadata := GraphMetadata{
		Name: "predictor-name",
		Models: map[string]ModelMetadata{
			"model-v1-array-string-mix": {
				ApiVersion: "v1",
				Name:       "model-v1-array-string-mix",
				Platform:   "platform-name",
				Versions:   []string{"model-version"},
				Inputs: map[string]interface{}{
					"datatype": "array",
					"shape":    []int{2, 2},
				},
				Outputs: map[string]interface{}{
					"datatype": "strData",
				},
			},
		},
		GraphInputs: map[string]interface{}{
			"datatype": "array",
			"shape":    []int{2, 2},
		},
		GraphOutputs: map[string]interface{}{
			"datatype": "strData",
		},
	}

	graphMetadata, err := NewGraphMetadata(createPredictorProcess(t), spec)
	g.Expect(err).Should(BeNil())

	expectedJson, err := json.Marshal(expectedGrahMetadata)
	g.Expect(err).Should(BeNil())

	actualJson, err := json.Marshal(graphMetadata)
	g.Expect(err).Should(BeNil())

	g.Expect(actualJson).To(MatchJSON(expectedJson))
}

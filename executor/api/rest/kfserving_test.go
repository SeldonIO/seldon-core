package rest

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/payload"
)

func TestChainKFServingInputs(t *testing.T) {
	g := NewGomegaWithT(t)

	var dataStr = []byte(`{"inputs":[{"name":"features","datatype":"BYTES","shape":[1,1],"data":["hello"]}]}`)

	var input interface{}
	err := json.Unmarshal(dataStr, &input)
	g.Expect(err).To(BeNil())

	inputPayload := payload.BytesPayload{Msg: dataStr, ContentType: ContentTypeJSON}

	outputPayload, err := ChainKFserving(&inputPayload)
	g.Expect(err).To(BeNil())

	var output interface{}
	outputStr, err := outputPayload.GetBytes()
	g.Expect(err).To(BeNil())
	err = json.Unmarshal(outputStr, &output)

	inputMap := input.(map[string]interface{})
	outputMap := output.(map[string]interface{})

	g.Expect(inputMap["inputs"]).To(Equal(outputMap["inputs"]))
	g.Expect(outputMap["outputs"]).To(BeNil())
}

func TestChainKFServingOutputs(t *testing.T) {
	g := NewGomegaWithT(t)

	var dataStr = []byte(`{"outputs":[{"name":"features","datatype":"BYTES","shape":[1,1],"data":["hello"]}]}`)

	var input interface{}
	err := json.Unmarshal(dataStr, &input)
	g.Expect(err).To(BeNil())

	inputPayload := payload.BytesPayload{Msg: dataStr, ContentType: ContentTypeJSON}

	outputPayload, err := ChainKFserving(&inputPayload)
	g.Expect(err).To(BeNil())

	var output interface{}
	outputStr, err := outputPayload.GetBytes()
	g.Expect(err).To(BeNil())
	err = json.Unmarshal(outputStr, &output)

	inputMap := input.(map[string]interface{})
	outputMap := output.(map[string]interface{})

	g.Expect(inputMap["outputs"]).To(Equal(outputMap["inputs"]))
	g.Expect(outputMap["outputs"]).To(BeNil())
}

func TestChainKFServingParams(t *testing.T) {
	g := NewGomegaWithT(t)

	var dataStr = []byte(`{"parameters": {"key": "value"},"outputs":[{"name":"features","datatype":"BYTES","shape":[1,1],"data":["hello"]}]}`)

	var input interface{}
	err := json.Unmarshal(dataStr, &input)
	g.Expect(err).To(BeNil())

	inputPayload := payload.BytesPayload{Msg: dataStr, ContentType: ContentTypeJSON}

	outputPayload, err := ChainKFserving(&inputPayload)
	g.Expect(err).To(BeNil())

	var output interface{}
	outputStr, err := outputPayload.GetBytes()
	g.Expect(err).To(BeNil())
	err = json.Unmarshal(outputStr, &output)

	inputMap := input.(map[string]interface{})
	outputMap := output.(map[string]interface{})

	g.Expect(inputMap["parameters"]).To(Equal(outputMap["parameters"]))
	g.Expect(outputMap["outputs"]).To(BeNil())
}

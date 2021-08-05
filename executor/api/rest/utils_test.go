package rest

import (
	"encoding/json"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/util"
)

func TestCombineExtractMessages(t *testing.T) {
	g := NewGomegaWithT(t)
	sm1 := &payload.BytesPayload{Msg: []byte(`{"a":1}`)}
	sm2 := &payload.BytesPayload{Msg: []byte(`{"a":1}`)}
	arr := []payload.SeldonPayload{sm1, sm2}
	msg, err := CombineSeldonMessagesToJson(arr)
	g.Expect(err).To(BeNil())
	arr2, err := ExtractSeldonMessagesFromJson(msg)
	g.Expect(err).To(BeNil())
	g.Expect(arr2[0]).To(Equal(sm1))
	g.Expect(arr2[1]).To(Equal(sm2))
}

func TestConversions(t *testing.T) {
	g := NewGomegaWithT(t)
	val := "[1,2]"
	arr, err := util.ExtractRouteAsJsonArray([]byte(val))
	g.Expect(err).Should(BeNil())
	g.Expect(arr[0]).Should(Equal(1))
	g.Expect(arr[1]).Should(Equal(2))
}

func TestEmbedSeldonDeploymentValuesToSwaggerFile(t *testing.T) {
	g := NewGomegaWithT(t)
	testNamespace := "default"
	testDeployment := "seldon-deployment"
	var unmarshallErr error = nil
	var embedErr error = nil

	openapiTestStr1 := `{
		"servers": {},
		"paths": {
			"/seldon/{namespace}/{deployment}/api/v1.0/predictions": {
				"post": {
					"parameters": {}
				}
			},
			"/seldon/{namespace}/{deployment}/api/v1.0/feedback": {
				"post": {
					"parameters": {}
				}
			}
		}
	}`

	var openapiTestInterface1 interface{}
	unmarshallErr = json.Unmarshal([]byte(openapiTestStr1), &openapiTestInterface1)
	embedErr = embedSeldonDeploymentValuesInJson(testNamespace, testDeployment, &openapiTestInterface1)
	g.Expect(unmarshallErr).To(BeNil())
	g.Expect(embedErr).To(BeNil())

	openapiTestStr2 := `{
		"paths": {
			"/seldon/{namespace}/{deployment}/api/v1.0/predictions": {
				"post": {
					"parameters": {}
				}
			},
			"/seldon/{namespace}/{deployment}/api/v1.0/feedback": {
				"post": {
					"parameters": {}
				}
			}
		}
	}`

	var openapiTestInterface2 interface{}
	unmarshallErr = json.Unmarshal([]byte(openapiTestStr2), &openapiTestInterface2)
	embedErr = embedSeldonDeploymentValuesInJson(testNamespace, testDeployment, &openapiTestInterface2)
	g.Expect(unmarshallErr).To(BeNil())
	g.Expect(embedErr).To(BeNil())

	openapiTestStr3 := `{
		"servers": {},
		"pathss": {
			"/seldon/{namespace}/{deployment}/api/v1.0/predictions": {
				"post": {
					"parameters": {}
				}
			},
			"/seldon/{namespace}/{deployment}/api/v1.0/feedback": {
				"post": {
					"parameters": {}
				}
			}
		}
	}`

	var openapiTestInterface3 interface{}
	unmarshallErr = json.Unmarshal([]byte(openapiTestStr3), &openapiTestInterface3)
	embedErr = embedSeldonDeploymentValuesInJson(testNamespace, testDeployment, &openapiTestInterface3)
	g.Expect(unmarshallErr).To(BeNil())
	g.Expect(embedErr).ToNot(BeNil())

	openapiTestStr4 := `{
		"servers": {},
		"paths": {
			"/seldon/{namespace}/{deployment}/api/v1.0/predictions": {
				"postt": {
					"parameters": {}
				}
			},
			"/seldon/{namespace}/{deployment}/api/v1.0/feedback": {
				"post": {
					"parameters": {}
				}
			}
		}
	}`

	var openapiTestInterface4 interface{}
	unmarshallErr = json.Unmarshal([]byte(openapiTestStr4), &openapiTestInterface4)
	embedErr = embedSeldonDeploymentValuesInJson(testNamespace, testDeployment, &openapiTestInterface4)
	g.Expect(unmarshallErr).To(BeNil())
	g.Expect(embedErr).ToNot(BeNil())

	openapiTestStr5 := `{
		"servers": {},
		"paths": {
			"/seldon/{namespace}/{deployment}/api/v1.0/feedback": {
				"post": {
					"parameters": {}
				}
			}
		}
	}`

	var openapiTestInterface5 interface{}
	unmarshallErr = json.Unmarshal([]byte(openapiTestStr5), &openapiTestInterface5)
	embedErr = embedSeldonDeploymentValuesInJson(testNamespace, testDeployment, &openapiTestInterface5)
	g.Expect(unmarshallErr).To(BeNil())
	g.Expect(embedErr).ToNot(BeNil())
}

func TestIsJson(t *testing.T) {
	g := NewGomegaWithT(t)
	badJson := "ab"
	res := isJSON([]byte(badJson))
	g.Expect(res).To(Equal(false))

	goodJson := "{\"foo\":\"bar\"}"
	res = isJSON([]byte(goodJson))
	g.Expect(res).To(Equal(true))
}

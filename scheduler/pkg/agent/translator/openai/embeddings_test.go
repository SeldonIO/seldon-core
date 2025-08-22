package openai

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestEmbeddingsRequest(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name               string
		openAIContent      map[string]interface{}
		expectedOipContent map[string]interface{}
	}

	tests := []test{
		{
			name: "default",
			openAIContent: map[string]interface{}{
				"model":           "text-embedding-ada-002",
				"input":           "This is a test",
				"encoding_format": "float",
			},
			expectedOipContent: map[string]interface{}{
				"inputs": []map[string]interface{}{
					{
						"name":     "input",
						"shape":    []int{1},
						"datatype": "BYTES",
						"data":     []string{"This is a test"},
					},
				},
				"parameters": map[string]interface{}{
					"llm_parameters": map[string]interface{}{
						"encoding_format": "float",
					},
				},
			},
		},
	}

	openAITranslator := &OpenAIEmbeddingsTranslator{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			openAIReqBody, err := json.Marshal(test.openAIContent)
			g.Expect(err).To(BeNil(), "Error marshalling OpenAI request content")

			openAIReq := http.Request{
				Method: http.MethodPost,
				Body:   io.NopCloser(strings.NewReader(string(openAIReqBody))),
				URL: &url.URL{
					Scheme: "http",
					Host:   "localhost:9000",
					Path:   "/v2/models/text-embedding-ada-002_1/infer/embeddings",
				},
			}
			oipReq, err := openAITranslator.TranslateToOIP(&openAIReq)
			g.Expect(err).To(BeNil(), "Error translating OpenAI request to OIP format")

			oipReqBody, err := io.ReadAll(oipReq.Body)
			g.Expect(err).To(BeNil(), "Error reading OIP request body")

			expectedOipBody, err := json.Marshal(test.expectedOipContent)
			g.Expect(err).To(BeNil(), "Error marshalling expected OIP request content")

			g.Expect(oipReqBody).To(Equal(expectedOipBody), "OIP request body does not match expected format")
		})
	}
}

func TestEmbeddingRespose(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                   string
		oipResponse            map[string]interface{}
		expectedOpenAIResponse map[string]interface{}
	}

	tests := []test{
		{
			name: "api response",
			oipResponse: map[string]interface{}{
				"model_name": "text-embedding-ada-002",
				"outputs": []map[string]interface{}{
					{
						"name":     "embedding",
						"shape":    []int{1, 3},
						"datatype": "FP64",
						"data":     []interface{}{0.0023064255, -0.009327292, -0.0028842222},
					},
					{
						"name":     "output_all",
						"shape":    []int{1, 3},
						"datatype": "FP64",
						"data": []string{
							"{\n" +
								"  \"object\": \"list\",\n" +
								"  \"data\": [\n" +
								"    {\n" +
								"      \"object\": \"embedding\",\n" +
								"      \"embedding\": [\n" +
								"        0.0023064255,\n" +
								"        -0.009327292,\n" +
								"        -0.0028842222\n" +
								"      ],\n" +
								"      \"index\": 0\n" +
								"    }\n" +
								"  ],\n" +
								"  \"model\": \"text-embedding-ada-002\",\n" +
								"  \"usage\": {\n" +
								"    \"prompt_tokens\": 8,\n" +
								"    \"total_tokens\": 8\n" +
								"  }\n" +
								"}",
						},
					},
				},
				"parameters": map[string]interface{}{
					"content_type": "str",
				},
			},
			expectedOpenAIResponse: map[string]interface{}{
				"object": "list",
				"data": []map[string]interface{}{
					{
						"object":    "embedding",
						"embedding": []float64{0.0023064255, -0.009327292, -0.0028842222},
						"index":     0,
					},
				},
				"model": "text-embedding-ada-002",
				"usage": map[string]interface{}{
					"prompt_tokens": 8,
					"total_tokens":  8,
				},
			},
		},
		{
			name: "local response - single",
			oipResponse: map[string]interface{}{
				"model_name": "text-embedding-ada-002",
				"outputs": []map[string]interface{}{
					{
						"name":     "embedding",
						"shape":    []int{1, 3},
						"datatype": "FP64",
						"data":     []float64{0.0023064255, -0.009327292, -0.0028842222},
					},
				},
				"parameters": map[string]interface{}{
					"content_type": "str",
				},
			},
			expectedOpenAIResponse: map[string]interface{}{
				"object": "list",
				"data": []map[string]interface{}{
					{
						"object":    "embedding",
						"embedding": []float64{0.0023064255, -0.009327292, -0.0028842222},
						"index":     0,
					},
				},
				"model": "text-embedding-ada-002",
			},
		},
		{
			name: "local response - multiple",
			oipResponse: map[string]interface{}{
				"model_name": "text-embedding-ada-002",
				"outputs": []map[string]interface{}{
					{
						"name":     "embedding",
						"shape":    []int{2, 3},
						"datatype": "FP64",
						"data":     []float64{0.0023064255, -0.009327292, -0.0028842222, 0.0033064255, -0.008327292, -0.0018842222},
					},
				},
				"parameters": map[string]interface{}{
					"content_type": "str",
				},
			},
			expectedOpenAIResponse: map[string]interface{}{
				"object": "list",
				"data": []map[string]interface{}{
					{
						"object":    "embedding",
						"embedding": []float64{0.0023064255, -0.009327292, -0.0028842222},
						"index":     0,
					},
					{
						"object":    "embedding",
						"embedding": []float64{0.0033064255, -0.008327292, -0.0018842222},
						"index":     1,
					},
				},
				"model": "text-embedding-ada-002",
			},
		},
	}

	openAITranslator := &OpenAIEmbeddingsTranslator{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oipResponseBody, err := json.Marshal(test.oipResponse)
			g.Expect(err).To(BeNil(), "Error marshalling OIP response content")

			oipRes := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(string(oipResponseBody))),
			}

			openAIResp, err := openAITranslator.TranslateFromOIP(oipRes)
			g.Expect(err).To(BeNil(), "Error translating OIP response to OpenAI format")

			openAIBody, err := io.ReadAll(openAIResp.Body)
			g.Expect(err).To(BeNil(), "Error reading OpenAI response body")

			var openAIRespContent map[string]interface{}
			err = json.Unmarshal(openAIBody, &openAIRespContent)
			g.Expect(err).To(BeNil(), "Error unmarshalling OpenAI response content")

			openAIRespMarshal, err := json.Marshal(openAIRespContent)
			g.Expect(err).To(BeNil(), "Error marshalling OpenAI response content")

			// marshal expected response for comparison including null values
			expectedOpenAIResponseMarshal, err := json.Marshal(test.expectedOpenAIResponse)
			g.Expect(err).To(BeNil(), "Error marshalling expected OpenAI response content")
			g.Expect(openAIRespMarshal).To(Equal(expectedOpenAIResponseMarshal), "OpenAI response body does not match expected format")
		})
	}

}

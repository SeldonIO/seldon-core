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

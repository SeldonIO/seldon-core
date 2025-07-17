package openai

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func cleanBody(body string) string {
	// Remove whitespace, tabs, and newlines for comparison
	body = strings.ReplaceAll(body, "\n", "")
	body = strings.ReplaceAll(body, "\t", "")
	body = strings.ReplaceAll(body, " ", "")
	return body
}

func TestAddModelVersion(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name               string
		openAIContent      map[string]interface{}
		expectedOipContent map[string]interface{}
	}

	tests := []test{
		{
			name: "default-single-message",
			openAIContent: map[string]interface{}{
				"model": "gpt-4.1",
				"messages": []map[string]interface{}{
					{
						"role":    "user",
						"content": "Hello!",
					},
				},
			},
			expectedOipContent: map[string]interface{}{
				"inputs": []map[string]interface{}{
					{
						"name":     "role",
						"shape":    []int{1},
						"datatype": "BYTES",
						"data":     []string{"user"},
					},
					{
						"name":     "content",
						"shape":    []int{1},
						"datatype": "BYTES",
						"data":     []string{"Hello!"},
					},
					{
						"name":     "type",
						"shape":    []int{1},
						"datatype": "BYTES",
						"data":     []string{"text"},
					},
				},
				"parameters": map[string]interface{}{
					"llm_parameters": map[string]interface{}{},
				},
			},
		},
		{
			name: "default-multiple-messages",
			openAIContent: map[string]interface{}{
				"model": "gpt-4.1",
				"messages": []map[string]interface{}{
					{
						"role":    "developer",
						"content": "You are a helpful assistant.",
					},
					{
						"role":    "user",
						"content": "Hello!",
					},
				},
			},
			expectedOipContent: map[string]interface{}{
				"inputs": []map[string]interface{}{
					{
						"name":     "role",
						"shape":    []int{2},
						"datatype": "BYTES",
						"data":     []string{"developer", "user"},
					},
					{
						"name":     "content",
						"shape":    []int{2},
						"datatype": "BYTES",
						"data": []string{
							"[\"You are a helpful assistant.\"]",
							"[\"Hello!\"]",
						},
					},
					{
						"name":     "type",
						"shape":    []int{2},
						"datatype": "BYTES",
						"data": []string{
							"[\"text\"]",
							"[\"text\"]",
						},
					},
				},
				"parameters": map[string]interface{}{
					"llm_parameters": map[string]interface{}{},
				},
			},
		},
	}

	logger := log.New().WithField("Source", "HTTPProxy")
	openAITranslator := &OpenAIChatCompletionsTranslator{}

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
					Path:   "/v2/models/chatgpt/infer/chat/completions",
				},
			}
			oipReq, err := openAITranslator.TranslateToOIP(&openAIReq, logger)
			g.Expect(err).To(BeNil(), "Error translating OpenAI request to OIP format")

			oipReqBody, err := io.ReadAll(oipReq.Body)
			g.Expect(err).To(BeNil(), "Error reading OIP request body")

			expectedOipBody, err := json.Marshal(test.expectedOipContent)
			g.Expect(err).To(BeNil(), "Error marshalling expected OIP request content")

			g.Expect(oipReqBody).To(Equal(expectedOipBody), "OIP request body does not match expected format")
		})
	}
}

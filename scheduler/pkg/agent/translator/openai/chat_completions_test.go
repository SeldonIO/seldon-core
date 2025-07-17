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

func TestChatCompletionsRequest(t *testing.T) {
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
		{
			name: "tools",
			openAIContent: map[string]interface{}{
				"model": "gpt-4.1",
				"messages": []map[string]interface{}{
					{
						"role":    "user",
						"content": "Hello! What is the weather like in New York?",
					},
				},
				"tools": []map[string]interface{}{
					{
						"type":     "function",
						"function": "foo",
					},
					{
						"type":     "function",
						"function": "bar",
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
						"data":     []string{"Hello! What is the weather like in New York?"},
					},
					{
						"name":     "type",
						"shape":    []int{1},
						"datatype": "BYTES",
						"data":     []string{"text"},
					},
					{
						"name":     "tools",
						"shape":    []int{2},
						"datatype": "BYTES",
						"data": []string{
							"{\"function\":\"foo\",\"type\":\"function\"}",
							"{\"function\":\"bar\",\"type\":\"function\"}",
						},
					},
				},
				"parameters": map[string]interface{}{
					"llm_parameters": map[string]interface{}{},
				},
			},
		},
		{
			name: "tools-calls",
			openAIContent: map[string]interface{}{
				"model": "gpt-4.1",
				"messages": []map[string]interface{}{
					{
						"role":    "system",
						"content": "You are a helpful assistant",
					},
					{
						"role":    "user",
						"content": "What's the weather like in Paris today?",
					},
					{
						"role": "assistant",
						"tool_calls": []map[string]interface{}{
							{
								"function": "foo_function",
								"id":       "foo_id",
								"type":     "function",
							},
						},
					},
					{
						"role":         "tool",
						"tool_call_id": "foo_id",
						"content":      "foo_content",
					},
					{
						"role":    "user",
						"content": "What's the weather like in London today?",
					},
					{
						"role": "assistant",
						"tool_calls": []map[string]interface{}{
							{
								"function": "bar_function",
								"id":       "bar_id",
								"type":     "function",
							},
						},
					},
					{
						"role":         "tool",
						"tool_call_id": "bar_id",
						"content":      "bar_content",
					},
				},
				"tools": []map[string]interface{}{
					{
						"type":     "function",
						"function": "foo_function",
					},
					{
						"type":     "function",
						"function": "bar_function",
					},
				},
			},
			expectedOipContent: map[string]interface{}{
				"inputs": []map[string]interface{}{
					{
						"name":     "role",
						"shape":    []int{7},
						"datatype": "BYTES",
						"data": []string{
							"system",
							"user",
							"assistant",
							"tool",
							"user",
							"assistant",
							"tool",
						},
					},
					{
						"name":     "content",
						"shape":    []int{7},
						"datatype": "BYTES",
						"data": []string{
							"[\"You are a helpful assistant\"]",
							"[\"What's the weather like in Paris today?\"]",
							"[\"\"]",
							"[\"foo_content\"]",
							"[\"What's the weather like in London today?\"]",
							"[\"\"]",
							"[\"bar_content\"]",
						},
					},
					{
						"name":     "type",
						"shape":    []int{7},
						"datatype": "BYTES",
						"data": []string{
							"[\"text\"]",
							"[\"text\"]",
							"[\"text\"]",
							"[\"text\"]",
							"[\"text\"]",
							"[\"text\"]",
							"[\"text\"]",
						},
					},
					{
						"name":     "tool_calls",
						"shape":    []int{7},
						"datatype": "BYTES",
						"data": []string{
							"",
							"",
							"[\"{\\\"function\\\":\\\"foo_function\\\",\\\"id\\\":\\\"foo_id\\\",\\\"type\\\":\\\"function\\\"}\"]",
							"",
							"",
							"[\"{\\\"function\\\":\\\"bar_function\\\",\\\"id\\\":\\\"bar_id\\\",\\\"type\\\":\\\"function\\\"}\"]",
							"",
						},
					},
					{
						"name":     "tool_call_id",
						"shape":    []int{7},
						"datatype": "BYTES",
						"data": []string{
							"",
							"",
							"",
							"foo_id",
							"",
							"",
							"bar_id",
						},
					},
					{
						"name":     "tools",
						"shape":    []int{2},
						"datatype": "BYTES",
						"data": []string{
							"{\"function\":\"foo_function\",\"type\":\"function\"}",
							"{\"function\":\"bar_function\",\"type\":\"function\"}",
						},
					},
				},
				"parameters": map[string]interface{}{
					"llm_parameters": map[string]interface{}{},
				},
			},
		},
		{
			name: "image-input",
			openAIContent: map[string]interface{}{
				"model": "gpt-4.1",
				"messages": []map[string]interface{}{
					{
						"role": "user",
						"content": []map[string]interface{}{
							{
								"type": "text",
								"text": "What is in this image?",
							},
							{
								"type": "image_url",
								"image_url": map[string]interface{}{
									"url": "dummy_image_url",
								},
							},
						},
					},
				},
				"max_tokens": 300,
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
						"shape":    []int{2},
						"datatype": "BYTES",
						"data": []string{
							"What is in this image?",
							"{\"url\":\"dummy_image_url\"}",
						},
					},
					{
						"name":     "type",
						"shape":    []int{2},
						"datatype": "BYTES",
						"data": []string{
							"text",
							"image_url",
						},
					},
				},
				"parameters": map[string]interface{}{
					"llm_parameters": map[string]interface{}{
						"max_tokens": 300,
					},
				},
			},
		},
		{
			name: "streaming",
			openAIContent: map[string]interface{}{
				"model": "gpt-4.1",
				"messages": []map[string]interface{}{
					{
						"role":    "user",
						"content": "Hello!",
					},
				},
				"stream": true,
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
					"llm_parameters": map[string]interface{}{
						"stream": true,
					},
				},
			},
		},
		{
			name: "params",
			openAIContent: map[string]interface{}{
				"model": "gpt-4.1",
				"messages": []map[string]interface{}{
					{
						"role":    "user",
						"content": "Hello!",
					},
				},
				"logprobs":     true,
				"top_logprobs": 2,
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
					"llm_parameters": map[string]interface{}{
						"logprobs":     true,
						"top_logprobs": 2,
					},
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

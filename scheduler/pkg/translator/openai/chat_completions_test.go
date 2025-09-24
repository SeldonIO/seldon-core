/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package openai

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestChatCompletionsRequest(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name               string
		openAIContent      map[string]any
		expectedOipContent map[string]any
	}

	tests := []test{
		{
			name: "default-single-message",
			openAIContent: map[string]any{
				"model": "gpt-4.1",
				"messages": []map[string]any{
					{
						"role":    "user",
						"content": "Hello!",
					},
				},
			},
			expectedOipContent: map[string]any{
				"inputs": []map[string]any{
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
				"parameters": map[string]any{
					"llm_parameters": map[string]any{},
					"kwargs":         map[string]any{},
				},
			},
		},
		{
			name: "default-multiple-messages",
			openAIContent: map[string]any{
				"model": "gpt-4.1",
				"messages": []map[string]any{
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
			expectedOipContent: map[string]any{
				"inputs": []map[string]any{
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
				"parameters": map[string]any{
					"llm_parameters": map[string]any{},
					"kwargs":         map[string]any{},
				},
			},
		},
		{
			name: "tools",
			openAIContent: map[string]any{
				"model": "gpt-4.1",
				"messages": []map[string]any{
					{
						"role":    "user",
						"content": "Hello! What is the weather like in New York?",
					},
				},
				"tools": []map[string]any{
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
			expectedOipContent: map[string]any{
				"inputs": []map[string]any{
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
				"parameters": map[string]any{
					"llm_parameters": map[string]any{},
					"kwargs":         map[string]any{},
				},
			},
		},
		{
			name: "tools-calls",
			openAIContent: map[string]any{
				"model": "gpt-4.1",
				"messages": []map[string]any{
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
						"tool_calls": []map[string]any{
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
						"tool_calls": []map[string]any{
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
				"tools": []map[string]any{
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
			expectedOipContent: map[string]any{
				"inputs": []map[string]any{
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
				"parameters": map[string]any{
					"llm_parameters": map[string]any{},
					"kwargs":         map[string]any{},
				},
			},
		},
		{
			name: "image-input",
			openAIContent: map[string]any{
				"model": "gpt-4.1",
				"messages": []map[string]any{
					{
						"role": "user",
						"content": []map[string]any{
							{
								"type": "text",
								"text": "What is in this image?",
							},
							{
								"type": "image_url",
								"image_url": map[string]any{
									"url": "dummy_image_url",
								},
							},
						},
					},
				},
				"max_tokens": 300,
			},
			expectedOipContent: map[string]any{
				"inputs": []map[string]any{
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
				"parameters": map[string]any{
					"llm_parameters": map[string]any{
						"max_tokens": 300,
					},
					"kwargs": map[string]any{
						"max_tokens": 300,
					},
				},
			},
		},
		{
			name: "streaming",
			openAIContent: map[string]any{
				"model": "gpt-4.1",
				"messages": []map[string]any{
					{
						"role":    "user",
						"content": "Hello!",
					},
				},
				"stream": true,
			},
			expectedOipContent: map[string]any{
				"inputs": []map[string]any{
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
				"parameters": map[string]any{
					"llm_parameters": map[string]any{
						"stream": true,
					},
					"kwargs": map[string]any{
						"stream": true,
					},
				},
			},
		},
		{
			name: "params",
			openAIContent: map[string]any{
				"model": "gpt-4.1",
				"messages": []map[string]any{
					{
						"role":    "user",
						"content": "Hello!",
					},
				},
				"logprobs":     true,
				"top_logprobs": 2,
			},
			expectedOipContent: map[string]any{
				"inputs": []map[string]any{
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
				"parameters": map[string]any{
					"llm_parameters": map[string]any{
						"logprobs":     true,
						"top_logprobs": 2,
					},
					"kwargs": map[string]any{
						"logprobs":     true,
						"top_logprobs": 2,
					},
				},
			},
		},
		{
			name: "parallel-tool-calls",
			openAIContent: map[string]any{
				"model": "gpt-4.1",
				"messages": []map[string]any{
					{
						"role":    "user",
						"content": "Hello! What is the weather like in New York?",
					},
				},
				"tools": []map[string]any{
					{
						"type":     "function",
						"function": "foo",
					},
					{
						"type":     "function",
						"function": "bar",
					},
				},
				"parallel_tool_calls": true,
			},
			expectedOipContent: map[string]any{
				"inputs": []map[string]any{
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
					{
						"name":     "parallel_tool_calls",
						"shape":    []int{1},
						"datatype": "BYTES",
						"data":     []string{"true"},
					},
				},
				"parameters": map[string]any{
					"llm_parameters": map[string]any{},
					"kwargs":         map[string]any{},
				},
			},
		},
		{
			name: "tool-choice",
			openAIContent: map[string]any{
				"model": "gpt-4.1",
				"messages": []map[string]any{
					{
						"role":    "user",
						"content": "Hello! What is the weather like in New York?",
					},
				},
				"tools": []map[string]any{
					{
						"type":     "function",
						"function": "foo",
					},
					{
						"type":     "function",
						"function": "bar",
					},
				},
				"tool_choice": "foo",
			},
			expectedOipContent: map[string]any{
				"inputs": []map[string]any{
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
					{
						"name":     "tool_choice",
						"shape":    []int{1},
						"datatype": "BYTES",
						"data":     []string{"foo"},
					},
				},
				"parameters": map[string]any{
					"llm_parameters": map[string]any{},
					"kwargs":         map[string]any{},
				},
			},
		},
		{
			name: "tool-choice-complex",
			openAIContent: map[string]any{
				"model": "gpt-4.1",
				"messages": []map[string]any{
					{
						"role":    "user",
						"content": "Hello! What is the weather like in New York?",
					},
				},
				"tools": []map[string]any{
					{
						"type":     "function",
						"function": "foo",
					},
					{
						"type":     "function",
						"function": "bar",
					},
				},
				"tool_choice": map[string]any{
					"type":     "function",
					"function": "dummy_function",
				},
			},
			expectedOipContent: map[string]any{
				"inputs": []map[string]any{
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
					{
						"name":     "tool_choice",
						"shape":    []int{1},
						"datatype": "BYTES",
						"data":     []string{"{\"function\":\"dummy_function\",\"type\":\"function\"}"},
					},
				},
				"parameters": map[string]any{
					"llm_parameters": map[string]any{},
					"kwargs":         map[string]any{},
				},
			},
		},
	}

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
					Path:   "/v2/models/gpt-4.1_1/infer/chat/completions",
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

func TestChatCompletionsResponse(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                   string
		oipResponse            map[string]any
		expectedOpenAIResponse map[string]any
	}

	tests := []test{
		{
			name: "api-response",
			oipResponse: map[string]any{
				"id":         "aa65a0ac-6aea-4284-9e55-7c74e299390e",
				"model_name": "openai-chat-completions",
				"outputs": []map[string]any{
					{
						"name":     "role",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data":     []string{"assistant"},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
					{
						"name":     "content",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data":     []string{"Hello! How can I assist you today?"},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
					{
						"name":     "type",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data":     []string{"text"},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
					{
						"name":     "output_all",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data": []string{
							"{\n" +
								"  \"id\": \"chatcmpl-AYD7ygNL8rqda0t5mG691k1aXpGLC\",\n" +
								"  \"choices\": [\n" +
								"    {\n" +
								"      \"finish_reason\": \"stop\",\n" +
								"      \"index\": 0,\n" +
								"      \"logprobs\": null,\n" +
								"      \"message\": {\n" +
								"        \"content\": \"Hello! How can I assist you today?\",\n" +
								"        \"refusal\": null,\n" +
								"        \"role\": \"assistant\"\n" +
								"      }\n" +
								"    }\n" +
								"  ],\n" +
								"  \"created\": 1732716978,\n" +
								"  \"model\": \"gpt-3.5-turbo-0125\",\n" +
								"  \"object\": \"chat.completion\",\n" +
								"  \"system_fingerprint\": null,\n" +
								"  \"usage\": {\n" +
								"    \"completion_tokens\": 9,\n" +
								"    \"prompt_tokens\": 21,\n" +
								"    \"total_tokens\": 30,\n" +
								"    \"completion_tokens_details\": {\n" +
								"      \"reasoning_tokens\": 0,\n" +
								"      \"audio_tokens\": 0,\n" +
								"      \"accepted_prediction_tokens\": 0,\n" +
								"      \"rejected_prediction_tokens\": 0\n" +
								"    },\n" +
								"    \"prompt_tokens_details\": {\n" +
								"      \"cached_tokens\": 0,\n" +
								"      \"audio_tokens\": 0\n" +
								"    }\n" +
								"  }\n" +
								"}",
						},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
				},
				"parameters": map[string]any{},
			},
			expectedOpenAIResponse: map[string]any{
				"id":      "chatcmpl-AYD7ygNL8rqda0t5mG691k1aXpGLC",
				"model":   "gpt-3.5-turbo-0125",
				"created": 1732716978,
				"object":  "chat.completion",
				"choices": []map[string]any{
					{
						"index": 0,
						"message": map[string]any{
							"content": "Hello! How can I assist you today?",
							"role":    "assistant",
							"refusal": nil,
						},
						"finish_reason": "stop",
						"logprobs":      nil,
					},
				},
				"system_fingerprint": nil,
				"usage": map[string]any{
					"prompt_tokens":     21,
					"completion_tokens": 9,
					"total_tokens":      30,
					"prompt_tokens_details": map[string]any{
						"cached_tokens": 0,
						"audio_tokens":  0,
					},
					"completion_tokens_details": map[string]any{
						"reasoning_tokens":           0,
						"audio_tokens":               0,
						"accepted_prediction_tokens": 0,
						"rejected_prediction_tokens": 0,
					},
				},
			},
		},
		{
			name: "local-response",
			oipResponse: map[string]any{
				"id":         "aa65a0ac-6aea-4284-9e55-7c74e299390e",
				"model_name": "local-chat-completions",
				"outputs": []map[string]any{
					{
						"name":     "role",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data":     []string{"assistant"},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
					{
						"name":     "content",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data":     []string{"Hello! How can I assist you today?"},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
					{
						"name":     "type",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data":     []string{"text"},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
				},
			},
			expectedOpenAIResponse: map[string]any{
				"id":      "aa65a0ac-6aea-4284-9e55-7c74e299390e",
				"model":   "local-chat-completions",
				"created": 0, // Local responses do not have a created
				"object":  "chat.completion",
				"choices": []map[string]any{
					{
						"index": 0,
						"message": map[string]any{
							"content": "Hello! How can I assist you today?",
							"role":    "assistant",
						},
					},
				},
			},
		},
	}

	openAITranslator := &OpenAIChatCompletionsTranslator{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oipResponseBody, err := json.Marshal(test.oipResponse)
			g.Expect(err).To(BeNil(), "Error marshalling OIP response content")

			oipResp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(string(oipResponseBody))),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			}

			openAIResp, err := openAITranslator.TranslateFromOIP(oipResp)
			g.Expect(err).To(BeNil(), "Error translating OIP response to OpenAI format")

			openAIRespBody, err := io.ReadAll(openAIResp.Body)
			g.Expect(err).To(BeNil(), "Error reading OpenAI response body")

			// unmarsal and marshal to ensure formatting is correct
			var openAIRespContent map[string]any
			err = json.Unmarshal(openAIRespBody, &openAIRespContent)
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

func TestChatCompletionStreamResponse(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                   string
		oipResponse            map[string]any
		expectedOpenAIResponse map[string]any
	}

	tests := []test{
		{
			name: "api-streaming-response",
			oipResponse: map[string]any{
				"id":         "aa65a0ac-6aea-4284-9e55-7c74e299390e",
				"model_name": "openai-chat-completions",
				"outputs": []map[string]any{
					{
						"name":     "role",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data":     []string{"assistant"},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
					{
						"name":     "content",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data":     []string{"Hello!"},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
					{
						"name":     "type",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data":     []string{"text"},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
					{
						"name":     "output_all",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data": []string{
							"{\n" +
								"  \"id\": \"chatcmpl-AYD7ygNL8rqda0t5mG691k1aXpGLC\",\n" +
								"  \"choices\": [\n" +
								"    {\n" +
								"      \"finish_reason\": null,\n" +
								"      \"index\": 0,\n" +
								"      \"logprobs\": null,\n" +
								"      \"delta\": {\n" +
								"        \"role\": \"assistant\",\n" +
								"        \"content\": \"Hello!\"\n" +
								"      }\n" +
								"    }\n" +
								"  ],\n" +
								"  \"created\": 1732716978,\n" +
								"  \"model\": \"gpt-3.5-turbo-0125\",\n" +
								"  \"object\": \"chat.completion.chunk\",\n" +
								"  \"system_fingerprint\": null,\n" +
								"  \"usage\": {\n" +
								"    \"completion_tokens\": 9,\n" +
								"    \"prompt_tokens\": 21,\n" +
								"    \"total_tokens\": 30,\n" +
								"    \"completion_tokens_details\": {\n" +
								"      \"reasoning_tokens\": 0,\n" +
								"      \"audio_tokens\": 0,\n" +
								"      \"accepted_prediction_tokens\": 0,\n" +
								"      \"rejected_prediction_tokens\": 0\n" +
								"    },\n" +
								"    \"prompt_tokens_details\": {\n" +
								"      \"cached_tokens\": 0,\n" +
								"      \"audio_tokens\": 0\n" +
								"    }\n" +
								"  }\n" +
								"}",
						},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
				},
				"parameters": map[string]any{},
			},
			expectedOpenAIResponse: map[string]any{
				"id":      "chatcmpl-AYD7ygNL8rqda0t5mG691k1aXpGLC",
				"model":   "gpt-3.5-turbo-0125",
				"created": float64(1732716978), // JSON unmarshaling converts numbers to float64
				"object":  "chat.completion.chunk",
				"choices": []any{
					map[string]any{
						"index": float64(0),
						"delta": map[string]any{
							"role":    "assistant",
							"content": "Hello!",
						},
						"finish_reason": nil,
						"logprobs":      nil,
					},
				},
				"system_fingerprint": nil,
				"usage": map[string]any{
					"prompt_tokens":     float64(21),
					"completion_tokens": float64(9),
					"total_tokens":      float64(30),
					"prompt_tokens_details": map[string]any{
						"cached_tokens": float64(0),
						"audio_tokens":  float64(0),
					},
					"completion_tokens_details": map[string]any{
						"reasoning_tokens":           float64(0),
						"audio_tokens":               float64(0),
						"accepted_prediction_tokens": float64(0),
						"rejected_prediction_tokens": float64(0),
					},
				},
			},
		},
		{
			name: "local-streaming-response",
			oipResponse: map[string]any{
				"id":         "aa65a0ac-6aea-4284-9e55-7c74e299390e",
				"model_name": "local-chat-completions",
				"outputs": []map[string]any{
					{
						"name":     "role",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data":     []string{"assistant"},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
					{
						"name":     "content",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data":     []string{"Hello!"},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
					{
						"name":     "type",
						"datatype": "BYTES",
						"shape":    []int{1, 1},
						"data":     []string{"text"},
						"parameters": map[string]any{
							"content_type": "str",
						},
					},
				},
			},
			expectedOpenAIResponse: map[string]any{
				"id":      "aa65a0ac-6aea-4284-9e55-7c74e299390e",
				"model":   "local-chat-completions",
				"created": float64(0), // Local responses do not have a created timestamp
				"object":  "chat.completion.chunk",
				"choices": []any{
					map[string]any{
						"index": float64(0),
						"delta": map[string]any{
							"role":    "assistant",
							"content": "Hello!",
						},
					},
				},
			},
		},
	}

	openAITranslator := &OpenAIChatCompletionsTranslator{}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			oipResponseBody, err := json.Marshal(test.oipResponse)
			g.Expect(err).To(BeNil(), "Error marshalling OIP response content")

			// Proper SSE format: "data: " prefix, JSON payload, double newline
			sseBody := "data: " + string(oipResponseBody) + "\n\n"

			// Alternative formats you might need to test:
			// For multiple chunks: sseBody += "data: " + string(anotherChunk) + "\n\n"
			// For ending the stream: sseBody += "data: [DONE]\n\n"

			oipResp := &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(sseBody)),
				Header:     http.Header{"Content-Type": []string{"text/event-stream"}},
			}

			openAIResp, err := openAITranslator.TranslateFromOIP(oipResp)
			g.Expect(err).To(BeNil(), "Error translating OIP response to OpenAI format")

			openAIRespBody, err := io.ReadAll(openAIResp.Body)
			g.Expect(err).To(BeNil(), "Error reading OpenAI response body")

			// remove prefix and suffix to get the JSON payload
			openAIRespBody = bytes.TrimPrefix(openAIRespBody, []byte("data: "))
			openAIRespBody = bytes.TrimSuffix(openAIRespBody, []byte("\n\n"))

			var openAIRespContent map[string]any
			err = json.Unmarshal([]byte(openAIRespBody), &openAIRespContent)
			g.Expect(err).To(BeNil(), "Error unmarshalling OpenAI SSE response content")

			// Compare the actual response with expected
			g.Expect(openAIRespContent).To(Equal(test.expectedOpenAIResponse), "OpenAI response content does not match expected format")
		})
	}
}

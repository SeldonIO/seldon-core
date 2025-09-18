/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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

func TestImagesGenerationRequest(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name               string
		openAIContent      map[string]any
		expectedOipContent map[string]any
	}

	tests := []test{
		{
			name: "default",
			openAIContent: map[string]any{
				"model":  "gpt-image-1",
				"prompt": "A cute baby sea otter",
				"n":      1,
				"size":   "1024x1024",
			},
			expectedOipContent: map[string]any{
				"inputs": []map[string]any{
					{
						"name":     "prompt",
						"shape":    []int{1},
						"datatype": "BYTES",
						"data":     []string{"A cute baby sea otter"},
					},
				},
				"parameters": map[string]any{
					"llm_parameters": map[string]any{
						"n":    1,
						"size": "1024x1024",
					},
				},
			},
		},
	}

	openAITranslator := &OpenAIImagesGenerationsTranslator{}

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
					Path:   "/v2/models/gpt-image-1_1/infer/images/generations",
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

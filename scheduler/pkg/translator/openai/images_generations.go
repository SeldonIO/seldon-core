/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package openai

import (
	"fmt"
	"net/http"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/translator"
)

type OpenAIImagesGenerationsTranslator struct {
	translator.BaseTranslator
}

const (
	promptKey = "prompt"
)

func getImagesGenerationsParameters(jsonBody map[string]any) map[string]any {
	llmParameters := make(map[string]any)
	skipKeys := []string{modelKey, promptKey}
	for key, value := range jsonBody {
		if translator.Contains(skipKeys, key) {
			continue
		}
		llmParameters[key] = value
	}
	return llmParameters
}

func getPrompt(jsonBody map[string]any) ([]string, error) {
	prompt, ok := jsonBody[promptKey].(string)
	if !ok {
		return nil, fmt.Errorf("OpenAI request body does not contain '%s' field", promptKey)
	}

	delete(jsonBody, promptKey)
	return []string{prompt}, nil
}

func constructImagesGenerationsRequest(prompt []string, llmParams map[string]any) map[string]any {
	return map[string]any{
		inputsKey: []map[string]any{
			translator.ConstructStringTensor(promptKey, prompt),
		},
		parametersKey: map[string]any{
			llmParametersKey: llmParams,
		},
	}
}

func (t *OpenAIImagesGenerationsTranslator) TranslateToOIP(req *http.Request) (*http.Request, error) {
	// Convert OpenAI API request to JSON
	jsonBody, err := translator.ConvertRequestToJsonBody(req)
	if err != nil {
		return nil, err
	}

	// Check if model name matches the one in the request path
	err = translator.CheckModelsMatch(jsonBody, req.URL.Path)
	if err != nil {
		return nil, err
	}

	// Read the prompt field to be used for image generation
	prompt, err := getPrompt(jsonBody)
	if err != nil {
		return nil, err
	}

	// Perpare parameters
	llmParams := getImagesGenerationsParameters(jsonBody)

	// Construct the inference request
	inferenceRequest := constructImagesGenerationsRequest(prompt, llmParams)

	// Construct new request
	return translator.ConvertInferenceRequestToHttpRequest(inferenceRequest, req)
}

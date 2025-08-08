package openai

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/translator"
)

type OpenAIImagesGenerationsTranslator struct {
	translator.BaseTranslator
}

const (
	promptKey = "prompt"
)

func getImagesGenerationsParameters(jsonBody map[string]interface{}) map[string]interface{} {
	llmParameters := make(map[string]interface{})
	skipKeys := []string{modelKey, promptKey}
	for key, value := range jsonBody {
		if translator.Contains(skipKeys, key) {
			continue
		}
		llmParameters[key] = value
	}
	return llmParameters
}

func getPrompt(jsonBody map[string]interface{}) ([]string, error) {
	prompt, ok := jsonBody[promptKey].(string)
	if !ok {
		return nil, fmt.Errorf("OpenAI request body does not contain '%s' field", promptKey)
	}

	delete(jsonBody, promptKey)
	return []string{prompt}, nil
}

func constructImagesGenerationsRequest(prompt []string, llmParams map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		inputsKey: []map[string]interface{}{
			translator.ConstructStringTensor(promptKey, prompt),
		},
		parametersKey: map[string]interface{}{
			llmParametersKey: llmParams,
		},
	}
}

func (t *OpenAIImagesGenerationsTranslator) TranslateToOIP(req *http.Request, logger log.FieldLogger) (*http.Request, error) {
	// Convert OpenAI API request to JSON
	jsonBody, err := translator.ConvertRequestToJsonBody(req, logger)
	if err != nil {
		logger.WithError(err).Error("Failed to convert OpenAI API request to JSON body")
		return nil, err
	}

	// Check if model name matches the one in the request path
	err = translator.CheckModelsMatch(jsonBody, req.URL.Path, logger)
	if err != nil {
		logger.WithError(err).Error("Model name mismatch in OpenAI API request")
		return nil, err
	}

	// Read the prompt field to be used for image generation
	prompt, err := getPrompt(jsonBody)
	if err != nil {
		logger.WithError(err).Error("Failed to get input field from OpenAI request body")
		return nil, err
	}

	// Perpare parameters
	llmParams := getImagesGenerationsParameters(jsonBody)

	// Construct the inference request
	inferenceRequest := constructImagesGenerationsRequest(prompt, llmParams)

	// Construct new request
	return translator.ConvertInferenceRequestToHttpRequest(inferenceRequest, req, logger)
}

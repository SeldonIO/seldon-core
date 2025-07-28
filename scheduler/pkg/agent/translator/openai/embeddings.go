package openai

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/translator"
)

type OpenAIEmbeddingsTranslator struct {
	translator.BaseTranslator
}

const (
	inputKey = "input"
)

func getEmbeddingParameters(jsonBody map[string]interface{}) map[string]interface{} {
	llmParameters := make(map[string]interface{})
	skipKeys := []string{modelKey, inputKey}

	for key, value := range jsonBody {
		if translator.Contains(skipKeys, key) {
			continue
		}
		llmParameters[key] = value
	}
	return llmParameters
}

func getInput(jsonBody map[string]interface{}) ([]string, error) {
	input, ok := jsonBody[inputKey]
	if !ok {
		return nil, fmt.Errorf("OpenAI request body does not contain '%s' field", inputKey)
	}

	delete(jsonBody, inputKey)
	switch v := input.(type) {
	case string:
		return []string{v}, nil
	case []interface{}:
		strs := make([]string, len(v))
		for i, item := range v {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("OpenAI request body '%s' field contains non-string item: %v", inputKey, item)
			}
			strs[i] = str
		}
		return strs, nil
	default:
		return nil, fmt.Errorf("OpenAI request body '%s' field is not a string or an array of strings: %v", inputKey, input)
	}
}

func constructEmbeddingsInferenceRequest(input []string, llmParams map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		inputsKey: []map[string]interface{}{
			translator.ConstructStringTensor(inputKey, input),
		},
		parametersKey: map[string]interface{}{
			llmParametersKey: llmParams,
		},
	}
}

func (t *OpenAIEmbeddingsTranslator) TranslateToOIP(req *http.Request, logger log.FieldLogger) (*http.Request, error) {
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

	// Read the input field to be embedded
	input, err := getInput(jsonBody)
	if err != nil {
		logger.WithError(err).Error("Failed to get input field from OpenAI request body")
		return nil, err
	}

	// Perpare parameters
	llmParams := getEmbeddingParameters(jsonBody)

	// Construct the inference request
	inferenceRequest := constructEmbeddingsInferenceRequest(input, llmParams)

	// Construct new request
	return translator.ConvertInferenceRequestToHttpRequest(inferenceRequest, req, logger)
}

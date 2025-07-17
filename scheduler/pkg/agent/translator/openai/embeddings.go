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

func getEmbeddingParameters(jsonBody map[string]interface{}) map[string]interface{} {
	llmParameters := make(map[string]interface{})
	for key, value := range jsonBody {
		if key == "model" || key == "input" {
			continue
		}
		llmParameters[key] = value
	}
	return llmParameters
}

func getInput(jsonBody map[string]interface{}) ([]string, error) {
	input, ok := jsonBody["input"]
	if !ok {
		return nil, fmt.Errorf("OpenAI request body does not contain 'input' field")
	}

	delete(jsonBody, "input")
	switch v := input.(type) {
	case string:
		return []string{v}, nil
	case []interface{}:
		strs := make([]string, len(v))
		for i, item := range v {
			str, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("OpenAI request body 'input' field contains non-string item: %v", item)
			}
			strs[i] = str
		}
		return strs, nil
	default:
		return nil, fmt.Errorf("OpenAI request body 'input' field is not a string or an array of strings: %v", input)
	}
}

func constructEmbeddingsInferenceRequest(input []string, llmParams map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"inputs": []map[string]interface{}{
			translator.ConstructStringTensor("input", input),
		},
		"parameters": map[string]interface{}{
			"llm_parameters": llmParams,
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

	// Read model name. TODO: Check if the model name is in the request path
	_, _ = translator.GetModelName(jsonBody)

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

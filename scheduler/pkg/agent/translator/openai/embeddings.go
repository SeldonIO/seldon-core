package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
	body, err := translator.ReadRequestBody(req)
	if err != nil {
		logger.WithError(err).Error("Failed to read OpenAI request body")
		return nil, err
	}

	jsonBody, err := translator.GetJsonBody(body)
	logger.Info("Parsing OpenAI API request body %v", jsonBody)
	if err != nil {
		logger.WithError(err).Error("Failed to parse OpenAI API request body")
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

	// Marshal the inference request to JSON
	data, err := json.Marshal(inferenceRequest)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal OpenAI API request inputs")
		return nil, err
	}

	// Create a new request with the translated body
	newBody := io.NopCloser(bytes.NewBuffer(data))
	newReq, err := http.NewRequest(req.Method, req.URL.String(), newBody)
	if err != nil {
		logger.WithError(err).Error("Failed to create new HTTP request for OpenAI API")
		return nil, err
	}
	newReq.Header = req.Header.Clone()

	// OpenAI API clinet adds `chat/completions` to the path, we need to remove it
	err = translator.TrimPathAfterInfer(newReq)
	if err != nil {
		logger.WithError(err).Error("Failed to trim path after infer in OpenAI API request")
		return nil, err
	}
	return newReq, nil
}

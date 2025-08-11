package openai

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/translator"
)

type OpenAIEmbeddingsTranslator struct {
	translator.BaseTranslator
}

const (
	inputKey     = "input"
	embeddingKey = "embedding"
)

func (t *OpenAIEmbeddingsTranslator) TranslateToOIP(req *http.Request) (*http.Request, error) {
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

	// Read the input field to be embedded
	input, err := getInput(jsonBody)
	if err != nil {
		return nil, err
	}

	// Perpare parameters
	llmParams := getEmbeddingParameters(jsonBody)

	// Construct the inference request
	inferenceRequest := constructEmbeddingsInferenceRequest(input, llmParams)

	// Construct new request
	return translator.ConvertInferenceRequestToHttpRequest(inferenceRequest, req)
}

func (t *OpenAIEmbeddingsTranslator) TranslateFromOIP(res *http.Response) (*http.Response, error) {
	httpRespones, err := t.BaseTranslator.TranslateFromOIP(res)
	if err == nil {
		return httpRespones, nil
	}

	jsonBody, isGzipped, err := translator.DecompressIfNeededAndConvertToJSON(res)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress and parse the response: %w", err)
	}

	outputs, ok := jsonBody[translator.OutputsKey].([]interface{})
	if !ok {
		return nil, fmt.Errorf("`%s` field not found or not an array in the response", translator.OutputsKey)
	}

	content, err := parseOutputEmbeddings(outputs, jsonBody["id"].(string), jsonBody["model"].(string))
	if err != nil {
		return nil, fmt.Errorf("failed to parse chat completion response: %w", err)
	}
	return translator.CreateResponseFromContent(content, res.StatusCode, res.Header, isGzipped)
}

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

func extractEmbeddingVectorsFromResponse(outputs []interface{}) ([][]float64, error) {
	// Extract the embedding tensor from the inference response
	tensor, err := translator.ExtractTensorByName(outputs, embeddingKey)
	if err != nil {
		return nil, err
	}

	data, ok := tensor[translator.DataKey].([]interface{})
	if !ok {
		return nil, fmt.Errorf("`%s` field not found or not an array in output tensor %s", translator.DataKey, embeddingKey)
	}

	vectors := make([][]float64, len(data))
	for i, item := range data {
		vector, ok := item.([]interface{})
		if !ok {
			return nil, fmt.Errorf("item %d in `%s` field of output tensor %s is not an array", i, translator.DataKey, embeddingKey)
		}

		floatVector := make([]float64, len(vector))
		for j, value := range vector {
			floatValue, ok := value.(float64)
			if !ok {
				return nil, fmt.Errorf("item %d in `%s` field of output tensor %s is not a float64", j, translator.DataKey, embeddingKey)
			}
			floatVector[j] = floatValue
		}
		vectors[i] = floatVector
	}
	return vectors, nil
}

func parseOutputEmbeddings(outputs []interface{}, id string, modelName string) (string, error) {
	embeddings, err := extractEmbeddingVectorsFromResponse(outputs)
	if err != nil {
		return "", err
	}

	data := make([]map[string]interface{}, 0, len(outputs))
	for i, embedding := range embeddings {
		data = append(data, map[string]interface{}{
			"object":    "embedding",
			"embedding": embedding,
			"index":     i,
		})
	}

	response := map[string]interface{}{
		"object": "list",
		"data":   data,
		"model":  modelName,
	}

	// convert the response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("failed to marshal local embedding response: %w", err)
	}
	return string(jsonResponse), nil
}

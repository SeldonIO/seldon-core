package translator

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
	"knative.dev/pkg/webhook/json"
)

type Translator interface {
	TranslateToOIP(req *http.Request, logger log.FieldLogger) (*http.Request, error)
	TranslateFromOIP(res *http.Response, logger log.FieldLogger) (*http.Response, error)
}

type BaseTranslator struct{}

const (
	dataKey      = "data"
	roleKey      = "role"
	contentKey   = "content"
	typeKey      = "type"
	outputsKey   = "outputs"
	outputAllKey = "output_all"
)

func decompressIfNeededAndConvertToJSON(res *http.Response, logger log.FieldLogger) (map[string]interface{}, bool, error) {
	// Decompress the response if needed - gzip
	var isGzipped bool
	var err error

	if isGzipped, err = DecompressIfNeeded(res); err != nil {
		logger.WithError(err).Error("Failed to decompress OpenAI API response")
		return nil, isGzipped, err
	}

	// Read the response body
	body, err := ReadResponseBody(res)
	if err != nil {
		logger.WithError(err).Error("Failed to read OpenAI API response body")
		return nil, isGzipped, err
	}

	jsonBody, err := GetJsonBody(body)
	if err != nil {
		logger.WithError(err).Error("Failed to parse OpenAI API response body")
		return nil, isGzipped, err
	}

	return jsonBody, isGzipped, nil
}

func createResponseFromContent(
	content string, statusCode int, headers http.Header, isGzipped bool, logger log.FieldLogger,
) (*http.Response, error) {
	// Create a new response with the translated body
	newBody := io.NopCloser(bytes.NewBuffer([]byte(content)))
	newRes := http.Response{
		StatusCode: statusCode,
		Header:     headers.Clone(),
		Body:       newBody,
	}

	// compress the response body if needed
	if isGzipped {
		if err := Compress(&newRes); err != nil {
			logger.WithError(err).Error("Failed to compress OpenAI API response")
			return nil, err
		}
	}
	return &newRes, nil
}

func extractTensorContentFromResponse(outputs []interface{}, tensorName string, logger log.FieldLogger) (string, error) {
	// Extract the output_all tensor form the inference response. This contains the full response
	// OpenAI API response - only works for OpenAI runtime, since we return the original OpenAI API response
	tensor, err := ExtractTensorByName(outputs, tensorName)
	if err != nil {
		logger.WithError(err).Errorf("Failed to extract '%s' tensor from the response", outputAllKey)
		return "", err
	}

	return extractContentFromTensor(tensor, dataKey, logger)

}

func extractContentFromTensor(tensor map[string]interface{}, key string, logger log.FieldLogger) (string, error) {
	data, ok := tensor[dataKey].([]interface{})
	if !ok {
		logger.Errorf("`%s` field not found or not an array of strings in output tensor %s", dataKey, outputAllKey)
		return "", fmt.Errorf("`%s` field not found or not an array of strings in output tensor %s", dataKey, outputAllKey)
	}

	content, ok := data[0].(string)
	if !ok {
		logger.Errorf("`%s` field in output tensor %s is not a byte array", dataKey, outputAllKey)
		return "", fmt.Errorf("`%s` field in output tensor %s is not a byte array", dataKey, outputAllKey)
	}
	return content, nil
}

func parseOutputAll(outputs []interface{}, logger log.FieldLogger) (string, error) {
	return extractTensorContentFromResponse(outputs, outputAllKey, logger)
}

func parseOuputChatCompletion(outputs []interface{}, id string, modelName string, logger log.FieldLogger) (string, error) {
	role, err := extractTensorContentFromResponse(outputs, roleKey, logger)
	if err != nil {
		logger.WithError(err).Errorf("Failed to extract '%s' tensor from the response", roleKey)
		return "", err
	}

	content, err := extractTensorContentFromResponse(outputs, contentKey, logger)
	if err != nil {
		logger.WithError(err).Errorf("Failed to extract '%s' tensor from the response", contentKey)
		return "", err
	}

	// get current timestamp
	timestamp := time.Now().Format(time.RFC3339)

	// construct the OpenAI API response
	response := map[string]interface{}{
		"id":      id,
		"model":   modelName,
		"created": timestamp,
		"object":  "chat.completion",
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"content": content,
					"role":    role,
				},
			},
		},
	}

	// convert the response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		logger.WithError(err).Error("Failed to marshal OpenAI API response")
		return "", fmt.Errorf("failed to marshal OpenAI API response: %w", err)
	}
	return string(jsonResponse), nil
}

func (b *BaseTranslator) TranslateFromOIP(res *http.Response, logger log.FieldLogger) (*http.Response, error) {
	jsonBody, isGzipped, err := decompressIfNeededAndConvertToJSON(res, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress and parse the response: %w", err)
	}

	outputs, ok := jsonBody[outputsKey].([]interface{})
	if !ok {
		logger.Errorf("`%s` field not found or not an array in the response", outputsKey)
		return nil, fmt.Errorf("`%s` field not found or not an array in the response", outputsKey)
	}

	content := ""
	if _, err := ExtractTensorByName(outputs, outputAllKey); err == nil {
		content, err = parseOutputAll(outputs, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to parse output_all field: %w", err)
		}
	} else {
		id, ok := jsonBody["id"].(string)
		if !ok {
			logger.Error("`id` field not found or not a string in the response")
			return nil, fmt.Errorf("`id` field not found or not a string in the response")
		}

		modelName, ok := jsonBody["model_name"].(string)
		if !ok {
			logger.Error("`model_name` field not found or not a string in the response")
			return nil, fmt.Errorf("`model_name` field not found or not a string in the response")
		}

		content, err = parseOuputChatCompletion(outputs, id, modelName, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to parse chat completion response: %w", err)
		}
	}
	return createResponseFromContent(content, res.StatusCode, res.Header, isGzipped, logger)
}

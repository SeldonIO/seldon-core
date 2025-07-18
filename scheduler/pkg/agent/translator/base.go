package translator

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Translator interface {
	TranslateToOIP(req *http.Request, logger log.FieldLogger) (*http.Request, error)
	TranslateFromOIP(res *http.Response, logger log.FieldLogger) (*http.Response, error)
}

type BaseTranslator struct{}

func (b *BaseTranslator) TranslateFromOIP(res *http.Response, logger log.FieldLogger) (*http.Response, error) {
	// Decompress the response if needed - gzip
	var isGzipped bool
	var err error

	if isGzipped, err = DecompressIfNeeded(res); err != nil {
		logger.WithError(err).Error("Failed to decompress OpenAI API response")
		return nil, err
	}

	// Read the response body
	body, err := ReadResponseBody(res)
	if err != nil {
		logger.WithError(err).Error("Failed to read OpenAI API response body")
		return nil, err
	}

	jsonBody, err := GetJsonBody(body)
	if err != nil {
		logger.WithError(err).Error("Failed to parse OpenAI API response body")
		return nil, err
	}

	// Parse the response body
	outputs, ok := jsonBody["outputs"].([]interface{})
	if !ok {
		logger.Error("`outputs` field not found or not an array in OpenAI API response")
		return nil, fmt.Errorf("`outputs` field not found or not an array in OpenAI API response")
	}

	// Extract the output_all tensor form the inference response. This contains the full response
	// OpenAI API response - only works for OpenAI runtime, since we return the original OpenAI API response
	tensorName := "output_all"
	outputAll, err := ExtractTensorByName(outputs, tensorName)
	if err != nil {
		logger.WithError(err).Errorf("Failed to extract '%s' tensor from OpenAI API response", tensorName)
		return nil, err
	}

	data, ok := outputAll["data"].([]interface{})
	if !ok {
		logger.Errorf("`data` field not found or not an array of strings in output tensor %s", tensorName)
		return nil, fmt.Errorf("`data` field not found or not an array of strings in output tensor %s", tensorName)
	}

	content, ok := data[0].(string)
	if !ok {
		logger.Errorf("`data` field in output tensor %s is not a byte array", tensorName)
		return nil, fmt.Errorf("`data` field in output tensor %s is not a byte array", tensorName)
	}

	// Create a new response with the translated body
	newBody := io.NopCloser(bytes.NewBuffer([]byte(content)))
	newRes := http.Response{
		StatusCode: res.StatusCode,
		Header:     res.Header.Clone(),
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

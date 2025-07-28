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

const (
	dataKey      = "data"
	outputsKey   = "outputs"
	outputAllKey = "output_all"
)

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
	outputs, ok := jsonBody[outputsKey].([]interface{})
	if !ok {
		logger.Errorf("`%s` field not found or not an array in OpenAI API response", outputsKey)
		return nil, fmt.Errorf("`%s` field not found or not an array in OpenAI API response", outputsKey)
	}

	// Extract the output_all tensor form the inference response. This contains the full response
	// OpenAI API response - only works for OpenAI runtime, since we return the original OpenAI API response
	outputAll, err := ExtractTensorByName(outputs, outputAllKey)
	if err != nil {
		logger.WithError(err).Errorf("Failed to extract '%s' tensor from OpenAI API response", outputAllKey)
		return nil, err
	}

	data, ok := outputAll[dataKey].([]interface{})
	if !ok {
		logger.Errorf("`%s` field not found or not an array of strings in output tensor %s", dataKey, outputAllKey)
		return nil, fmt.Errorf("`%s` field not found or not an array of strings in output tensor %s", dataKey, outputAllKey)
	}

	content, ok := data[0].(string)
	if !ok {
		logger.Errorf("`%s` field in output tensor %s is not a byte array", dataKey, outputAllKey)
		return nil, fmt.Errorf("`%s` field in output tensor %s is not a byte array", dataKey, outputAllKey)
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

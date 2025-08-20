package translator

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type Translator interface {
	TranslateToOIP(req *http.Request) (*http.Request, error)
	TranslateFromOIP(res *http.Response) (*http.Response, error)
}

type BaseTranslator struct{}

const (
	DataKey      = "data"
	ShapeKey     = "shape"
	OutputsKey   = "outputs"
	OutputAllKey = "output_all"
)

func DecompressIfNeededAndConvertToJSON(res *http.Response) (map[string]interface{}, bool, error) {
	// Decompress the response if needed - gzip
	var err error
	var isGzipped bool
	var newRes *http.Response

	if newRes, isGzipped, err = DecompressIfNeeded(res); err != nil {
		return nil, isGzipped, err
	}

	// Read the response body
	body, err := ReadResponseBody(newRes)
	if err != nil {
		return nil, isGzipped, err
	}

	jsonBody, err := GetJsonBody(body)
	if err != nil {
		return nil, isGzipped, err
	}

	return jsonBody, isGzipped, nil
}

func CreateResponseFromContent(content string, statusCode int, headers http.Header, isGzipped bool) (*http.Response, error) {
	// Create a new response with the translated body
	newBody := io.NopCloser(bytes.NewBuffer([]byte(content)))
	if headers == nil {
		headers = http.Header{}
	}
	newRes := http.Response{
		StatusCode: statusCode,
		Header:     headers.Clone(),
		Body:       newBody,
	}

	// set content length
	newRes.Header.Set("Content-Length", fmt.Sprintf("%d", len(content)))

	// compress the response body if needed
	if isGzipped {
		if err := Compress(&newRes); err != nil {
			return nil, err
		}
	}
	return &newRes, nil
}

func ExtractTensorContentFromResponse(outputs []interface{}, tensorName string) (string, error) {
	// Extract the output_all tensor form the inference response. This contains the full response
	// OpenAI API response - only works for OpenAI runtime, since we return the original OpenAI API response
	tensor, err := ExtractTensorByName(outputs, tensorName)
	if err != nil {
		return "", err
	}

	return extractContentFromTensor(tensor, DataKey)

}

func extractContentFromTensor(tensor map[string]interface{}, key string) (string, error) {
	data, ok := tensor[DataKey].([]interface{})
	if !ok {
		return "", fmt.Errorf("`%s` field not found or not an array of strings in output tensor %s", DataKey, OutputAllKey)
	}

	content, ok := data[0].(string)
	if !ok {
		return "", fmt.Errorf("`%s` field in output tensor %s is not a byte array", DataKey, OutputAllKey)
	}
	return content, nil
}

func parseOutputAll(outputs []interface{}) (string, error) {
	return ExtractTensorContentFromResponse(outputs, OutputAllKey)
}

func (b *BaseTranslator) TranslateFromOIP(res *http.Response) (*http.Response, error) {
	jsonBody, isGzipped, err := DecompressIfNeededAndConvertToJSON(res)
	if err != nil {
		return nil, fmt.Errorf("failed to decompress and parse the response: %w", err)
	}

	outputs, ok := jsonBody[OutputsKey].([]interface{})
	if !ok {
		return nil, fmt.Errorf("`%s` field not found or not an array in the response", OutputsKey)
	}

	content, err := parseOutputAll(outputs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output_all field: %w", err)
	}

	return CreateResponseFromContent(content, res.StatusCode, res.Header, isGzipped)
}

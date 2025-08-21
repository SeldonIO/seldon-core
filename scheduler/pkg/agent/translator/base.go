package translator

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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
	SSEPrefix    = "data: "
	SSESuffix    = "\n\n"
	FirstLineKey = "First-Line"
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

func CreateResponseFromContent(content string, statusCode int, header http.Header, isGzipped bool) (*http.Response, error) {
	if header == nil {
		header = http.Header{}
	}

	newBody := io.NopCloser(bytes.NewBufferString(content))
	header.Set("Content-Length", fmt.Sprintf("%d", len(content)))

	if isGzipped {
		newRes := http.Response{
			StatusCode: statusCode,
			Header:     header.Clone(),
			Body:       newBody,
		}
		if err := Compress(&newRes); err != nil {
			return nil, err
		}
		return &newRes, nil
	}

	return &http.Response{
		StatusCode: statusCode,
		Header:     header.Clone(),
		Body:       newBody,
	}, nil
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
	if IsServerSentEvent(res) {
		return translateStreamFromOIP(res)
	}

	return translateFromOIP(res)
}

func translateFromOIP(res *http.Response) (*http.Response, error) {
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

func translateStreamFromOIP(res *http.Response) (*http.Response, error) {
	pr, pw := io.Pipe()

	// override the default split function
	scanner := bufio.NewScanner(res.Body)
	scanner.Split(SplitSSE)

	_ = scanner.Scan()
	line := scanner.Text()

	// Transform the first line to check if we have output_all key
	translated, err := translateLine(line)
	if err != nil {
		res.Header.Set(FirstLineKey, line)
		return nil, fmt.Errorf("failed to translate first line: %w", err)
	}

	// Start background goroutine to copy/transform as data arrives
	go ScanAndWriteSSE(res, &translated, scanner, pw, translateLine)

	// Return single streaming response
	return &http.Response{
		StatusCode: res.StatusCode,
		Header:     res.Header.Clone(),
		Body:       pr,
	}, nil
}

func translateLine(line string) (string, error) {
	line = strings.TrimPrefix(line, SSEPrefix)
	jsonLine, err := GetJsonBody([]byte(line))
	if err != nil {
		return "", fmt.Errorf("failed to parse SSE line: %w", err)
	}

	outputs, ok := jsonLine[OutputsKey].([]interface{})
	if !ok {
		return "", fmt.Errorf("`%s` field not found or not an array in the response", OutputsKey)
	}

	content, err := parseOutputAll(outputs)
	if err != nil {
		return "", fmt.Errorf("failed to parse %s field: %w", OutputAllKey, err)
	}

	// unmarshal and then marshal again to ensure proper formatting
	mapContent := map[string]interface{}{}
	err = json.Unmarshal([]byte(content), &mapContent)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal content: %w", err)
	}

	contentBytes, err := json.Marshal(mapContent)
	if err != nil {
		return "", fmt.Errorf("failed to marshal content: %w", err)
	}

	// Reconstruct the line with the original SSE format
	return fmt.Sprintf("%s%s%s", SSEPrefix, string(contentBytes), SSESuffix), nil
}

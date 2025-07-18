package translator

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetJsonBody(body []byte) (map[string]interface{}, error) {
	var jsonBody map[string]interface{}
	err := json.Unmarshal(body, &jsonBody)
	if err != nil {
		return nil, err
	}
	return jsonBody, nil
}

func GetModelName(jsonBody map[string]interface{}) (string, error) {
	modelName, ok := jsonBody["model"].(string)
	if !ok {
		return "", errors.New("model name not found in request body")
	}
	delete(jsonBody, "model")
	return modelName, nil
}

func ConstructStringTensor(name string, data []string) map[string]interface{} {
	return map[string]interface{}{
		"name":     name,
		"shape":    []int{len(data)},
		"datatype": "BYTES",
		"data":     data,
	}
}

func ExtractTensorByName(outputs []interface{}, name string) (map[string]interface{}, error) {
	for i, output := range outputs {
		outputMap, ok := output.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("failed to parse output tensor %d", i)
		}

		if outputMap["name"] == name {
			return outputMap, nil
		}
	}
	return nil, fmt.Errorf("output tensor with name %s not found", name)
}

func ReadRequestBody(req *http.Request) ([]byte, error) {
	if req.Body == nil {
		return nil, fmt.Errorf("request body is empty")
	}
	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body: %w", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("request body is empty")
	}

	return body, nil
}

func ReadResponseBody(res *http.Response) ([]byte, error) {
	if res.Body == nil {
		return nil, fmt.Errorf("response body is empty")
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if len(body) == 0 {
		return nil, fmt.Errorf("response body is empty")
	}

	return body, nil
}

func ConvertRequestToJsonBody(req *http.Request, logger log.FieldLogger) (map[string]interface{}, error) {
	body, err := ReadRequestBody(req)
	if err != nil {
		logger.WithError(err).Error("Failed to read OpenAI API request body")
		return nil, err
	}

	jsonBody, err := GetJsonBody(body)
	if err != nil {
		logger.WithError(err).Error("Failed to parse OpenAI API request body")
		return nil, err
	}

	return jsonBody, nil
}

func ConvertInferenceRequestToHttpRequest(inferenceRequest map[string]interface{}, req *http.Request, logger log.FieldLogger) (*http.Request, error) {
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
	err = TrimPathAfterInfer(newReq)
	if err != nil {
		logger.WithError(err).Error("Failed to trim path after infer in OpenAI API request")
		return nil, err
	}

	return newReq, nil
}

func MatchMarker(path string, marker string) (int, error) {
	if path == "" || marker == "" {
		return -1, fmt.Errorf("path or marker cannot be empty")
	}
	pos := strings.Index(path, marker)
	return pos, nil
}

func TrimPathAfterInfer(req *http.Request) error {
	posInfer, err := MatchMarker(req.URL.Path, "/infer/")
	if err != nil {
		return fmt.Errorf("error matching '/infer' in path: %w", err)
	}
	if posInfer != -1 {
		req.URL.Path = req.URL.Path[:posInfer+len("/infer")]
		return nil
	}

	posInferStream, err := MatchMarker(req.URL.Path, "/infer_stream/")
	if err != nil {
		return fmt.Errorf("error matching '/infer_stream' in path: %w", err)
	}
	if posInferStream != -1 {
		req.URL.Path = req.URL.Path[:posInferStream+len("/infer_stream")]
		return nil
	}
	return fmt.Errorf("neither '/infer' nor '/infer_stream' found in path")
}

func DecompressIfNeeded(res *http.Response) (bool, error) {
	if res.Header.Get("Content-Encoding") == "gzip" {
		gr, err := gzip.NewReader(res.Body)
		if err != nil {
			return false, err
		}

		res.Body = &gzipReadCloser{
			gzipReader:   gr,
			originalBody: res.Body,
		}

		res.Header.Del("Content-Length")
		res.Header.Del("Content-Encoding")
		return true, nil
	}
	return false, nil
}

type gzipReadCloser struct {
	gzipReader   *gzip.Reader
	originalBody io.ReadCloser
}

func (grc *gzipReadCloser) Read(p []byte) (int, error) {
	return grc.gzipReader.Read(p)
}

func (grc *gzipReadCloser) Close() error {
	gzipErr := grc.gzipReader.Close()
	bodyErr := grc.originalBody.Close()
	if gzipErr != nil {
		return gzipErr
	}
	return bodyErr
}

func Compress(res *http.Response) error {
	originalBody := res.Body
	defer originalBody.Close()

	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := io.Copy(gz, originalBody); err != nil {
		return err
	}

	if err := gz.Close(); err != nil {
		return err
	}

	res.Body = io.NopCloser(&buf)
	res.Header.Set("Content-Length", strconv.Itoa(buf.Len()))
	res.Header.Set("Content-Encoding", "gzip")
	return nil
}

func GetPathTermination(req *http.Request) (string, error) {
	const marker = "/infer"
	path := req.URL.Path
	pos := strings.Index(path, marker)
	if pos == -1 {
		return "", fmt.Errorf("'/infer' not found in path")
	}
	return path[pos+len(marker):], nil
}

func IsEmptySlice(slice []string) bool {
	for _, item := range slice {
		if item != "" {
			return false
		}
	}
	return true
}

func ExtractModelNameFromPath(p string) (string, error) {
	// Normalize the path to remove any trailing slashes
	p = path.Clean(p)

	parts := strings.Split(p, "/")
	for i := 0; i < len(parts)-2; i++ {
		if parts[i] == "v2" && parts[i+1] == "models" {
			return parts[i+2], nil
		}
	}

	return "", errors.New("model name not found in path")
}

func CheckModelsMatch(jsonBody map[string]interface{}, path string, logger log.FieldLogger) error {
	modelName, err := GetModelName(jsonBody)
	if err != nil {
		return err
	}

	pathModelName, err := ExtractModelNameFromPath(path)
	if err != nil {
		return err
	}

	if strings.Contains(pathModelName, "_") {
		pathModelName = strings.Split(pathModelName, "_")[0]
	}

	if modelName != pathModelName {
		return fmt.Errorf("model %s not loaded at endpoint %s", modelName, path)
	}
	return nil
}

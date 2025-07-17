package translator

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
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
		return "", nil
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

func TrimPathAfterInfer(req *http.Request) error {
	const marker = "/infer"
	path := req.URL.Path
	pos := strings.Index(path, marker)
	if pos == -1 {
		return fmt.Errorf("'/infer' not found in path")
	}
	req.URL.Path = path[:pos+len(marker)]
	return nil
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

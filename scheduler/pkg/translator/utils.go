/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package translator

import (
	"bufio"
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
)

func GetJsonBody(body []byte) (map[string]any, error) {
	var jsonBody map[string]any
	err := json.Unmarshal(body, &jsonBody)
	if err != nil {
		return nil, err
	}
	return jsonBody, nil
}

func GetJsonBodyStream(body []byte) (map[string]any, error) {
	idx := bytes.Index(body, []byte(SSESuffix))
	if idx == -1 {
		return nil, fmt.Errorf("no SSE event found")
	}
	ev := body[:idx]

	if !bytes.HasPrefix(ev, []byte(SSEPrefix)) {
		return nil, fmt.Errorf("invalid SSE event: missing 'data: ' prefix")
	}
	ev = ev[len(SSEPrefix):]
	return GetJsonBody(ev)
}

func GetModelName(jsonBody map[string]any) (string, error) {
	modelName, ok := jsonBody["model"].(string)
	if !ok {
		return "", errors.New("model name not found in request body")
	}
	delete(jsonBody, "model")
	return modelName, nil
}

func ConstructStringTensor(name string, data []string) map[string]any {
	return map[string]any{
		"name":     name,
		"shape":    []int{len(data)},
		"datatype": "BYTES",
		"data":     data,
	}
}

func ExtractTensorByName(outputs []any, name string) (map[string]any, error) {
	for i, output := range outputs {
		outputMap, ok := output.(map[string]any)
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
		return nil, fmt.Errorf("response body is nil")
	}

	body, err := io.ReadAll(res.Body)
	res.Body.Close()

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	res.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func ConvertRequestToJsonBody(req *http.Request) (map[string]any, error) {
	body, err := ReadRequestBody(req)
	if err != nil {
		return nil, err
	}

	jsonBody, err := GetJsonBody(body)
	if err != nil {
		return nil, err
	}

	return jsonBody, nil
}

func ConvertInferenceRequestToHttpRequest(inferenceRequest map[string]any, req *http.Request) (*http.Request, error) {
	data, err := json.Marshal(inferenceRequest)
	if err != nil {
		return nil, err
	}

	// Create a new request with the translated body
	newBody := io.NopCloser(bytes.NewReader(data))
	newReq, err := http.NewRequest(req.Method, req.URL.String(), newBody)
	if err != nil {
		return nil, err
	}
	newReq.Header = req.Header.Clone()

	// OpenAI API clinet adds `chat/completions` to the path, we need to remove it
	err = TrimPathAfterInfer(newReq)
	if err != nil {
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

func DecompressIfNeeded(res *http.Response) (*http.Response, bool, error) {
	if res.Header.Get("Content-Encoding") == "gzip" {
		body, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, false, fmt.Errorf("failed to read gzipped response body: %w", err)
		}

		// reset the body to the original state
		res.Body.Close()
		res.Body = io.NopCloser(bytes.NewReader(body))

		gr, err := gzip.NewReader(io.NopCloser(bytes.NewReader(body)))
		if err != nil {
			return res, false, err
		}

		newRes := http.Response{
			StatusCode: res.StatusCode,
			Header:     res.Header.Clone(),
			Body:       gr,
		}

		newRes.Header.Del("Content-Length")
		newRes.Header.Del("Content-Encoding")
		return &newRes, true, nil
	}
	return res, false, nil
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
	markers := []string{"/infer_stream", "/infer"}
	path := req.URL.Path

	for _, marker := range markers {
		if pos := strings.Index(path, marker); pos != -1 {
			return path[pos+len(marker):], nil
		}
	}

	return "", fmt.Errorf("no valid marker found in path")
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

func CheckModelsMatch(jsonBody map[string]any, path string) error {
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

func Contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func IsServerSentEvent(resp *http.Response) bool {
	return IsServerSentEventHeader(&resp.Header)
}

func IsServerSentEventHeader(header *http.Header) bool {
	contentType := header.Get("Content-Type")
	return strings.HasPrefix(contentType, "text/event-stream")
}

func SplitSSE(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if i := bytes.Index(data, []byte(SSESuffix)); i >= 0 {
		return i + 2, data[:i], nil
	}
	if atEOF && len(data) > 0 {
		return len(data), data, nil
	}
	return 0, nil, nil
}

func ScanAndWriteSSE(res *http.Response, firstTranslatedLine *string, scanner *bufio.Scanner, pw *io.PipeWriter, translateLineFunc func(string) (string, error)) {
	defer res.Body.Close()

	// Write the first line to the pipe
	if _, err := pw.Write([]byte(*firstTranslatedLine)); err != nil {
		pw.CloseWithError(err)
		return
	}

	// Process the rest of the lines
	for scanner.Scan() {
		line := scanner.Text()
		translated, err := translateLineFunc(line)
		if err != nil {
			pw.CloseWithError(err)
			return
		}

		if _, err := pw.Write([]byte(translated)); err != nil {
			return
		}
	}

	if err := scanner.Err(); err != nil {
		pw.CloseWithError(err)
	}

	pw.Close()
}

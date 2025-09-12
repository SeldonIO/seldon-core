package translator

import (
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
)

func TestTrimPathAfterInfer(t *testing.T) {
	g := NewGomegaWithT(t)
	tests := []struct {
		name      string
		inputPath string
		wantPath  string
		expectErr bool
	}{
		{
			name:      "Trim after /infer/",
			inputPath: "/v2/models/infer/suffix",
			wantPath:  "/v2/models/infer",
			expectErr: false,
		},
		{
			name:      "Trim after /infer_stream/",
			inputPath: "/v2/models/infer_stream/suffix",
			wantPath:  "/v2/models/infer_stream",
			expectErr: false,
		},
		{
			name:      "Path with neither /infer/ nor /infer_stream/",
			inputPath: "/v2/models/query",
			wantPath:  "/v2/models/query",
			expectErr: true,
		},
		{
			name:      "Empty path",
			inputPath: "",
			wantPath:  "",
			expectErr: true,
		},
		{
			name:      "Only /infer/ at end",
			inputPath: "/infer/",
			wantPath:  "/infer",
			expectErr: false,
		},
		{
			name:      "Only /infer_stream/ at end",
			inputPath: "/infer_stream/",
			wantPath:  "/infer_stream",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &http.Request{
				URL: &url.URL{
					Path: tt.inputPath,
				},
			}
			err := TrimPathAfterInfer(req)
			if tt.expectErr {
				g.Expect(err).NotTo(BeNil(), "expected an error but got none")
			} else {
				g.Expect(err).To(BeNil(), "expected no error but got one")
				g.Expect(req.URL.Path).To(Equal(tt.wantPath), "expected path to be trimmed correctly")
			}
		})
	}
}

func TestGzip(t *testing.T) {
	g := NewGomegaWithT(t)

	// create request
	content := "This is a test string"
	res := &http.Response{
		Body:   io.NopCloser(strings.NewReader(content)),
		Header: http.Header{},
	}

	// test Gzip compression
	err := Compress(res)
	g.Expect(err).To(BeNil(), "Error compressing request body with Gzip")
	g.Expect(res.Header.Get("Content-Encoding")).To(Equal("gzip"), "Content-Encoding header should be set to gzip")

	// test gzip decompression
	newRes, isGzipped, err := DecompressIfNeeded(res)
	g.Expect(err).To(BeNil(), "Error decompressing response body")
	g.Expect(isGzipped).To(BeTrue(), "Response body should be gzipped")

	// read decompressed body and check content
	decompressedContent, err := ReadResponseBody(newRes)
	g.Expect(err).To(BeNil(), "Error reading decompressed response body")
	g.Expect(string(decompressedContent)).To(Equal(content), "Decompressed content does not match original content")
}

func TestReadRequestBody(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name      string
		req       *http.Request
		expectErr bool
	}{
		{
			name:      "Valid request body",
			req:       &http.Request{Body: io.NopCloser(strings.NewReader("test body"))},
			expectErr: false,
		},
		{
			name:      "Empty request body",
			req:       &http.Request{Body: io.NopCloser(strings.NewReader(""))},
			expectErr: true,
		},
		{
			name:      "Nil request body",
			req:       &http.Request{Body: nil},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := ReadRequestBody(test.req)
			if test.expectErr {
				g.Expect(err).NotTo(BeNil(), "expected an error but got none")
			} else {
				g.Expect(err).To(BeNil(), "expected no error but got one")
				g.Expect(body).NotTo(BeNil(), "expected body to be non-nil")
			}
		})
	}
}

func TestReadResponseBody(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name      string
		res       *http.Response
		expectErr bool
	}{
		{
			name:      "Valid response body",
			res:       &http.Response{Body: io.NopCloser(strings.NewReader("test response"))},
			expectErr: false,
		},
		{
			name:      "Nil response body",
			res:       &http.Response{Body: nil},
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := ReadResponseBody(test.res)
			if test.expectErr {
				g.Expect(err).NotTo(BeNil(), "expected an error but got none")
			} else {
				g.Expect(err).To(BeNil(), "expected no error but got one")
				g.Expect(body).NotTo(BeNil(), "expected body to be non-nil")
			}
		})
	}
}

func TestConstructStringTensor(t *testing.T) {
	g := NewGomegaWithT(t)
	tests := []struct {
		name      string
		fieldName string
		fieldData []string
		expected  map[string]any
	}{
		{
			name:      "Valid string tensor",
			fieldName: "testField",
			fieldData: []string{"value1", "value2"},
			expected: map[string]any{
				"name":     "testField",
				"shape":    []int{2},
				"datatype": "BYTES",
				"data":     []string{"value1", "value2"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := ConstructStringTensor(test.fieldName, test.fieldData)
			g.Expect(result).To(Equal(test.expected), "Constructed tensor does not match expected result")
		})
	}
}

func TestExtractTensorByName(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name       string
		outputs    []any
		tensorName string
		expectErr  bool
		expected   map[string]any
	}{
		{
			name: "Extract existing tensor",
			outputs: []any{
				map[string]any{"name": "tensor1", "data": []string{"a", "b"}},
				map[string]any{"name": "tensor2", "data": []string{"c", "d"}},
			},
			tensorName: "tensor1",
			expectErr:  false,
			expected:   map[string]any{"name": "tensor1", "data": []string{"a", "b"}},
		},
		{
			name:       "Extract non-existing tensor",
			outputs:    []any{},
			tensorName: "tensor3",
			expectErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := ExtractTensorByName(test.outputs, test.tensorName)
			if test.expectErr {
				g.Expect(err).NotTo(BeNil(), "expected an error but got none")
			} else {
				g.Expect(err).To(BeNil(), "expected no error but got one")
				g.Expect(result).To(Equal(test.expected), "Extracted tensor does not match expected result")
			}
		})
	}
}

func TestExtractModelNameFromPath(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name      string
		path      string
		expectErr bool
		expected  string
	}{
		{
			name:      "Valid model path",
			path:      "/v2/models/my-model/infer",
			expectErr: false,
			expected:  "my-model",
		},
		{
			name:      "Invalid model path",
			path:      "/v2/some-other-path/infer",
			expectErr: true,
			expected:  "",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			modelName, err := ExtractModelNameFromPath(test.path)
			if test.expectErr {
				g.Expect(err).NotTo(BeNil(), "expected an error but got none")
				g.Expect(modelName).To(Equal(test.expected), "expected model name to be empty")
			} else {
				g.Expect(err).To(BeNil(), "expected no error but got one")
				g.Expect(modelName).To(Equal(test.expected), "expected model name to match")
			}
		})
	}
}

func TestCheckModelsMatch(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name      string
		jsonBody  map[string]any
		path      string
		expectErr bool
	}{
		{
			name: "Matching model names",
			jsonBody: map[string]any{
				"model": "my-model",
			},
			path:      "/v2/models/my-model/infer",
			expectErr: false,
		},
		{
			name: "Matching model names with underscore",
			jsonBody: map[string]any{
				"model": "my-model",
			},
			path:      "/v2/models/my-model_1/infer",
			expectErr: false,
		},
		{
			name: "Non-matching model names",
			jsonBody: map[string]any{
				"model": "my-model",
			},
			path:      "/v2/models/another-model_1/infer",
			expectErr: true,
		},
		{
			name:      "Missing model name in JSON body",
			jsonBody:  map[string]any{},
			path:      "/v2/models/my-model_1/infer",
			expectErr: true,
		},
		{
			name: "Invalid path format",
			jsonBody: map[string]any{
				"model": "my-model",
			},
			path:      "/v2/some-other-path/infer",
			expectErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := CheckModelsMatch(test.jsonBody, test.path)
			if test.expectErr {
				g.Expect(err).NotTo(BeNil(), "expected an error but got none")
			} else {
				g.Expect(err).To(BeNil(), "expected no error but got one")
			}
		})
	}
}

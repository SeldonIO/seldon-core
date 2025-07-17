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
	isGzipped, err := DecompressIfNeeded(res)
	g.Expect(err).To(BeNil(), "Error decompressing response body")
	g.Expect(isGzipped).To(BeTrue(), "Response body should be gzipped")

	// read decompressed body and check content
	decompressedContent, err := ReadResponseBody(res)
	g.Expect(err).To(BeNil(), "Error reading decompressed response body")
	g.Expect(string(decompressedContent)).To(Equal(content), "Decompressed content does not match original content")
}

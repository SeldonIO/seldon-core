package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
)

func TestXSSMiddleware(t *testing.T) {
	g := NewGomegaWithT(t)

	m := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrapped := xssMiddleware(m)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	headerVal := res.Header.Get(contentTypeOptsHeader)
	g.Expect(headerVal).To(Equal(contentTypeOptsValue))
}

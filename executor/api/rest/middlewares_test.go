package rest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/gomega"
)

func TestCORSHeadersGet(t *testing.T) {
	g := NewGomegaWithT(t)

	m := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrapped := corsHeaders(m)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	headerValAllowOrigin := res.Header.Get(corsAllowOriginHeader)
	g.Expect(headerValAllowOrigin).To(Equal(corsAllowOriginValue))

	headerValAllowHeaders := res.Header.Get(corsAllowHeadersHeader)
	g.Expect(headerValAllowHeaders).To(Equal(corsAllowHeadersValue))

	headerValAllowMethods := res.Header.Get(corsAllowMethodsHeader)
	g.Expect(headerValAllowMethods).To(Equal(corsAllowMethodsValue))
}

func TestCORSHeadersOptions(t *testing.T) {
	g := NewGomegaWithT(t)

	m := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrapped := corsHeaders(m)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	headerValAllowOrigin := res.Header.Get(corsAllowOriginHeader)
	g.Expect(headerValAllowOrigin).To(Equal(corsAllowOriginValue))

	headerValAllowHeaders := res.Header.Get(corsAllowHeadersHeader)
	g.Expect(headerValAllowHeaders).To(Equal(corsAllowHeadersValue))

	headerValAllowMethods := res.Header.Get(corsAllowMethodsHeader)
	g.Expect(headerValAllowMethods).To(Equal(corsAllowMethodsValue))

	statusCode := res.StatusCode
	g.Expect(statusCode).To(Equal(http.StatusOK))
}

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

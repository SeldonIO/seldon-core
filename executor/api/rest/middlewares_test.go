package rest

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	. "github.com/onsi/gomega"
)

func TestEnvVars(t *testing.T) {
	g := NewGomegaWithT(t)

	os.Setenv(corsAllowOriginEnvVar, "http://www.google.com")
	os.Setenv(corsAllowOriginHeadersVar, "Accept")
	defer os.Unsetenv(corsAllowOriginEnvVar)
	defer os.Unsetenv(corsAllowOriginHeadersVar)

	m := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrapped := handleCORSRequests(m)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	headerValAllowOrigin := res.Header.Get(corsAllowOriginHeader)
	g.Expect(headerValAllowOrigin).To(Equal("http://www.google.com"))

	headerValAllowHeaders := res.Header.Get(corsAllowHeadersHeader)
	g.Expect(headerValAllowHeaders).To(Equal("Accept"))
}

func TestCORSHeadersGetRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	m := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrapped := handleCORSRequests(m)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	headerValAllowOrigin := res.Header.Get(corsAllowOriginHeader)
	g.Expect(headerValAllowOrigin).To(Equal(corsAllowOriginValueAll))

	headerValAllowHeaders := res.Header.Get(corsAllowHeadersHeader)
	g.Expect(headerValAllowHeaders).To(Equal(corsAllowHeadersValueDefault))
}

func TestCORSHeadersOptionsRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	m := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
	wrapped := handleCORSRequests(m)

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	wrapped.ServeHTTP(w, req)

	res := w.Result()
	defer res.Body.Close()

	headerValAllowOrigin := res.Header.Get(corsAllowOriginHeader)
	g.Expect(headerValAllowOrigin).To(Equal(corsAllowOriginValueAll))

	headerValAllowHeaders := res.Header.Get(corsAllowHeadersHeader)
	g.Expect(headerValAllowHeaders).To(Equal(corsAllowHeadersValueDefault))

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

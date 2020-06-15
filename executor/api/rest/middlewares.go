package rest

import (
	"net/http"

	guuid "github.com/google/uuid"
	"github.com/seldonio/seldon-core/executor/api/payload"
)

const (
	CLOUDEVENTS_HEADER_ID_NAME             = "Ce-Id"
	CLOUDEVENTS_HEADER_SPECVERSION_NAME    = "Ce-Specversion"
	CLOUDEVENTS_HEADER_SOURCE_NAME         = "Ce-Source"
	CLOUDEVENTS_HEADER_TYPE_NAME           = "Ce-Type"
	CLOUDEVENTS_HEADER_PATH_NAME           = "Ce-Path"
	CLOUDEVENTS_HEADER_SPECVERSION_DEFAULT = "0.3"

	contentTypeOptsHeader = "X-Content-Type-Options"
	contentTypeOptsValue  = "nosniff"

	corsAllowOriginHeader  = "Access-Control-Allow-Origin"
	corsAllowOriginValue   = "*"
	corsAllowMethodsHeader = "Access-Control-Allow-Methods"
	corsAllowMethodsValue  = "GET, OPTIONS, POST"
	corsAllowHeadersHeader = "Access-Control-Allow-Headers"
	corsAllowHeadersValue  = "Accept, Accept-Encoding, Authorization, Content-Length, Content-Type, X-CSRF-Token"
)

type CloudeventHeaderMiddleware struct {
	deploymentName string
	namespace      string
}

func (h *CloudeventHeaderMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Checking if request is cloudevent based on specname being present
		if _, ok := r.Header[CLOUDEVENTS_HEADER_SPECVERSION_NAME]; ok {
			puid := r.Header.Get(payload.SeldonPUIDHeader)
			w.Header().Set(CLOUDEVENTS_HEADER_ID_NAME, puid)
			w.Header().Set(CLOUDEVENTS_HEADER_SPECVERSION_NAME, CLOUDEVENTS_HEADER_SPECVERSION_DEFAULT)
			w.Header().Set(CLOUDEVENTS_HEADER_PATH_NAME, r.URL.Path)
			w.Header().Set(CLOUDEVENTS_HEADER_TYPE_NAME, "seldon."+h.deploymentName+"."+h.namespace+".response")
			w.Header().Set(CLOUDEVENTS_HEADER_SOURCE_NAME, "seldon."+h.deploymentName)
		}

		next.ServeHTTP(w, r)
	})
}

// handleCORSRequests adds CORS-required headers, and during CORS Preflight
// requests, it will exit the request and the request status will be
// http.StatusOK
func handleCORSRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(corsAllowOriginHeader, corsAllowOriginValue)
		w.Header().Set(corsAllowMethodsHeader, corsAllowMethodsValue)
		w.Header().Set(corsAllowHeadersHeader, corsAllowHeadersValue)
		// Don't pass along OPTIONS (CORS Prefetch) Requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func puidHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		puid := r.Header.Get(payload.SeldonPUIDHeader)
		if len(puid) == 0 {
			puid = guuid.New().String()
			r.Header.Set(payload.SeldonPUIDHeader, puid)
		}
		if res_puid := w.Header().Get(payload.SeldonPUIDHeader); len(res_puid) == 0 {
			w.Header().Set(payload.SeldonPUIDHeader, puid)
		}

		next.ServeHTTP(w, r)
	})
}

func xssMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentTypeOptsHeader, contentTypeOptsValue)

		next.ServeHTTP(w, r)
	})
}

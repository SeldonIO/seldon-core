/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package health_probe

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

const (
	pathReadiness = "/ready"
	pathLiveness  = "/live"
	pathStartup   = "/startup"
)

type HTTPServer struct {
	srv     *http.Server
	rtr     *mux.Router
	log     *log.Logger
	manager Manager
}

func NewHTTPServer(port int, manager Manager, log *log.Logger) *HTTPServer {
	return &HTTPServer{
		manager: manager,
		log:     log,
		rtr:     mux.NewRouter(),
		srv: &http.Server{
			Addr: fmt.Sprintf(":%d", port),
		},
	}
}

// Start is blocking until server encounters an error and shuts down.
func (h *HTTPServer) Start() error {
	if h.manager.HasCallbacks(ProbeReadiness) {
		h.rtr.HandleFunc(pathReadiness, h.healthCheck(h.manager.CheckReadiness)).Methods(http.MethodGet)
	}

	if h.manager.HasCallbacks(ProbeLiveness) {
		h.rtr.HandleFunc(pathLiveness, h.healthCheck(h.manager.CheckLiveness)).Methods(http.MethodGet)
	}

	if h.manager.HasCallbacks(ProbeStartUp) {
		h.rtr.HandleFunc(pathStartup, h.healthCheck(h.manager.CheckStartup)).Methods(http.MethodGet)
	}

	h.srv.Handler = h.rtr
	return h.srv.ListenAndServe()
}

func (h *HTTPServer) Router() *mux.Router {
	return h.rtr
}

func (h *HTTPServer) healthCheck(fn func() error) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if err := fn(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			h.log.WithError(err).WithField("uri", req.RequestURI).Warn("Failed health probe")
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func (h *HTTPServer) Shutdown(ctx context.Context) error {
	return h.srv.Shutdown(ctx)
}

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
	log     *log.Logger
	manager Manager
}

func NewHTTPServer(port int, manager Manager, log *log.Logger) *HTTPServer {
	return &HTTPServer{
		manager: manager,
		log:     log,
		srv: &http.Server{
			Addr: fmt.Sprintf(":%d", port),
		},
	}
}

// Start is blocking until server encounters an error and shuts down.
func (h *HTTPServer) Start() error {
	rtr := mux.NewRouter()

	if h.manager.HasCallbacks(ProbeReadiness) {
		rtr.HandleFunc(pathReadiness, h.healthCheck(h.manager.CheckReadiness)).Methods(http.MethodGet)
	}

	if h.manager.HasCallbacks(ProbeLiveness) {
		rtr.HandleFunc(pathLiveness, h.healthCheck(h.manager.CheckLiveness)).Methods(http.MethodGet)
	}

	if h.manager.HasCallbacks(ProbeStartUp) {
		rtr.HandleFunc(pathStartup, h.healthCheck(h.manager.CheckStartup)).Methods(http.MethodGet)
	}

	h.srv.Handler = rtr
	return h.srv.ListenAndServe()
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

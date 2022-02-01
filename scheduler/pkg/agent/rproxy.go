package agent

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	ReverseProxyHTTPPort = 9999
)

type reverseHTTPProxy struct {
	stateManager *LocalStateManager
	logger       log.FieldLogger
	server       *http.Server
	serverReady  bool
	port         uint
}

// need to rewrite the host of the outbound request with the host of the incoming request
// (added by ReverseProxy)
func (rp *reverseHTTPProxy) rewriteHostHandler(r *http.Request) {
	r.Host = r.URL.Host
}

func (rp *reverseHTTPProxy) addHandlers(proxy http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rp.rewriteHostHandler(r)

		modelName, _ := ExtractModelNamefromPath(r.URL.Path)
		//TODO: what is best practice here for dealing with err?

		if err := rp.stateManager.EnsureLoadModel(modelName); err != nil {
			rp.logger.Errorf("Cannot load model in agent %s", modelName)
			http.NotFound(w, r)
		} else {
			proxy.ServeHTTP(w, r)
		}
	})
}

func (rp *reverseHTTPProxy) Start() error {
	if rp.stateManager == nil {
		rp.logger.Error("Set state before starting reverse proxy service")
		return fmt.Errorf("State not set, aborting")
	}

	backend := rp.stateManager.GetBackEndPath()
	proxy := httputil.NewSingleHostReverseProxy(backend)
	rp.logger.Infof("Start reverse proxy on port %d for %s", rp.port, backend)
	rp.server = &http.Server{Addr: ":" + strconv.Itoa(int(rp.port)), Handler: rp.addHandlers(proxy)}
	// TODO: check for errors? we rely for now on Ready
	go func() {
		rp.serverReady = true
		err := rp.server.ListenAndServe()
		rp.logger.Infof("HTTP/REST reverse proxy debug service stopped (%s)", err)
		rp.serverReady = false
	}()
	return nil
}

func (rp *reverseHTTPProxy) Stop() error {
	// Shutdown is graceful
	return rp.server.Shutdown(context.TODO())
}

func (rp *reverseHTTPProxy) Ready() error {
	if rp.serverReady {
		return nil
	} else {
		return fmt.Errorf("HTTP reverse proxy is not ready")
	}
}

func (rp *reverseHTTPProxy) SetState(stateManager *LocalStateManager) {
	rp.stateManager = stateManager
}

func NewReverseHTTPProxy(
	logger log.FieldLogger,
	port uint,
) *reverseHTTPProxy {

	rp := reverseHTTPProxy{
		logger: logger,
		port:   port,
	}

	return &rp
}

func ExtractModelNamefromPath(path string) (string, error) {
	re := regexp.MustCompile(`v2/models/(\w+)/`)
	matches := re.FindStringSubmatch(path)
	if len(matches) == 2 {
		return matches[1], nil
	}
	return "", fmt.Errorf("cannot extract model from path %s", path)
}

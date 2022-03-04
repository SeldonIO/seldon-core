package agent

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent/metrics"
	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"

	log "github.com/sirupsen/logrus"
)

const (
	ReverseProxyHTTPPort    = 9999
	maxIdleConnsHTTP        = 500
	maxIdleConnsPerHostHTTP = 250
	disableKeepAlivesHTTP   = false
	maxConnsPerHostHTTP     = 500
)

type reverseHTTPProxy struct {
	stateManager *LocalStateManager
	logger       log.FieldLogger
	server       *http.Server
	serverReady  bool
	port         uint
	mu           sync.RWMutex
	metrics      metrics.MetricsHandler
}

// need to rewrite the host of the outbound request with the host of the incoming request
// (added by ReverseProxy)
func (rp *reverseHTTPProxy) rewriteHostHandler(r *http.Request) {
	r.Host = r.URL.Host
}

func (rp *reverseHTTPProxy) addHandlers(proxy http.Handler) http.Handler {
	return rp.metrics.AddHistogramMetricsHandler(func(w http.ResponseWriter, r *http.Request) {
		rp.rewriteHostHandler(r)

		externalModelName := r.Header.Get(resources.SeldonModelHeader)
		internalModelName := r.Header.Get(resources.SeldonInternalModel)
		//TODO should we return a 404 if headers not found?
		if externalModelName == "" || internalModelName == "" {
			rp.logger.Warnf("Failed to extract model name %s:[%s] %s:[%s]", resources.SeldonInternalModel, internalModelName, resources.SeldonModelHeader, externalModelName)
			proxy.ServeHTTP(w, r)
			return
		} else {
			rp.logger.Debugf("Extracted model name %s:%s %s:%s", resources.SeldonInternalModel, internalModelName, resources.SeldonModelHeader, externalModelName)
		}

		if err := rp.stateManager.EnsureLoadModel(internalModelName); err != nil {
			rp.logger.Errorf("Cannot load model in agent %s", internalModelName)
			http.NotFound(w, r)
		} else {
			r.URL.Path = rewritePath(r.URL.Path, internalModelName)
			rp.logger.Debugf("Calling %s", r.URL.Path)
			startTime := time.Now()
			proxy.ServeHTTP(w, r)
			elapsedTime := time.Since(startTime).Seconds()
			go rp.metrics.AddInferMetrics(externalModelName, internalModelName, metrics.MethodTypeRest, elapsedTime)
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
	proxy.Transport = &http.Transport{
		MaxIdleConns:        maxIdleConnsHTTP,
		MaxIdleConnsPerHost: maxIdleConnsPerHostHTTP,
		DisableKeepAlives:   disableKeepAlivesHTTP,
		MaxConnsPerHost:     maxConnsPerHostHTTP,
	}
	rp.logger.Infof("Start reverse proxy on port %d for %s", rp.port, backend)
	rp.server = &http.Server{Addr: ":" + strconv.Itoa(int(rp.port)), Handler: rp.addHandlers(proxy)}
	// TODO: check for errors? we rely for now on Ready
	go func() {
		rp.mu.Lock()
		rp.serverReady = true
		rp.mu.Unlock()
		err := rp.server.ListenAndServe()
		rp.logger.WithError(err).Info("HTTP/REST reverse proxy debug service stopped")
		rp.mu.Lock()
		rp.serverReady = false
		rp.mu.Unlock()
	}()
	return nil
}

func (rp *reverseHTTPProxy) Stop() error {
	// Shutdown is graceful
	rp.mu.Lock()
	defer rp.mu.Unlock()
	err := rp.server.Shutdown(context.TODO())
	rp.serverReady = false
	return err
}

func (rp *reverseHTTPProxy) Ready() bool {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.serverReady
}

func (rp *reverseHTTPProxy) SetState(stateManager *LocalStateManager) {
	rp.stateManager = stateManager
}

func (rp *reverseHTTPProxy) Name() string {
	return "Reverse HTTP/REST Proxy"
}

func NewReverseHTTPProxy(
	logger log.FieldLogger,
	port uint,
	metrics metrics.MetricsHandler,
) *reverseHTTPProxy {

	rp := reverseHTTPProxy{
		logger:  logger.WithField("Source", "HTTPProxy"),
		port:    port,
		metrics: metrics,
	}

	return &rp
}

func rewritePath(path string, modelName string) string {
	re := regexp.MustCompile(`(/v2/models/)([\w\-]+)(/versions/\w+)?(.*)$`)
	// ${3}, i.e. versions/<ver_num> is removed
	s := fmt.Sprintf("${1}%s${4}", modelName)
	return re.ReplaceAllString(path, s)
}

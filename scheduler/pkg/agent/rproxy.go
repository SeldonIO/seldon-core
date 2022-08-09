package agent

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strconv"
	"sync"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/pkg/metrics"
	log "github.com/sirupsen/logrus"
)

const (
	DefaultReverseProxyHTTPPort = 9999
	maxIdleConnsHTTP            = 10
	maxIdleConnsPerHostHTTP     = 10
	disableKeepAlivesHTTP       = false
	maxConnsPerHostHTTP         = 20
	defaultTimeoutSeconds       = 5
	idleConnTimeoutSeconds      = 60
)

type reverseHTTPProxy struct {
	stateManager          *LocalStateManager
	logger                log.FieldLogger
	server                *http.Server
	serverReady           bool
	backendHTTPServerHost string
	backendHTTPServerPort uint
	servicePort           uint
	mu                    sync.RWMutex
	metrics               metrics.AgentMetricsHandler
}

// in the case the model is not loaded on server (return 404), we attempt to load it and then retry request
type lazyModelLoadTransport struct {
	loader func(string) *V2Err
	http.RoundTripper
	metrics metrics.AgentMetricsHandler
}

// RoundTrip implements http.RoundTripper for the Transport type.
// It calls its underlying http.RoundTripper to execute the request, and
// adds retry logic if we get 404
func (t *lazyModelLoadTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var originalBody []byte
	var err error

	externalModelName := req.Header.Get(resources.SeldonModelHeader)
	internalModelName := req.Header.Get(resources.SeldonInternalModelHeader)
	startTime := time.Now()
	if req.Body != nil {
		originalBody, err = io.ReadAll(req.Body)
	}
	if err != nil {
		return nil, err
	}

	// reset main request body
	req.Body = io.NopCloser(bytes.NewBuffer(originalBody))
	res, err := t.RoundTripper.RoundTrip(req)
	if err != nil {
		return res, err
	}

	if res.StatusCode == http.StatusNotFound {
		internalModelName := req.Header.Get(resources.SeldonInternalModelHeader)
		t.loader(internalModelName)

		req2 := req.Clone(req.Context())

		req2.Body = io.NopCloser(bytes.NewBuffer(originalBody))
		res, err = t.RoundTripper.RoundTrip(req2)
	}

	elapsedTime := time.Since(startTime).Seconds()
	go t.metrics.AddModelInferMetrics(externalModelName, internalModelName, metrics.MethodTypeRest, elapsedTime, metrics.HttpCodeToString(res.StatusCode))

	return res, err

}

// need to rewrite the host of the outbound request with the host of the incoming request
// (added by ReverseProxy)
func (rp *reverseHTTPProxy) rewriteHostHandler(r *http.Request) {
	r.Host = r.URL.Host
}

func (rp *reverseHTTPProxy) addHandlers(proxy http.Handler) http.Handler {
	return otelhttp.NewHandler(rp.metrics.AddModelHistogramMetricsHandler(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rp.rewriteHostHandler(r)

		externalModelName := r.Header.Get(resources.SeldonModelHeader)
		internalModelName := r.Header.Get(resources.SeldonInternalModelHeader)
		//TODO should we return a 404 if headers not found?
		if externalModelName == "" || internalModelName == "" {
			rp.logger.Warnf("Failed to extract model name %s:[%s] %s:[%s]", resources.SeldonInternalModelHeader, internalModelName, resources.SeldonModelHeader, externalModelName)
			proxy.ServeHTTP(w, r)
			return
		} else {
			rp.logger.Debugf("Extracted model name %s:%s %s:%s", resources.SeldonInternalModelHeader, internalModelName, resources.SeldonModelHeader, externalModelName)
		}

		if err := rp.stateManager.EnsureLoadModel(internalModelName); err != nil {
			rp.logger.Errorf("Cannot load model in agent %s", internalModelName)
			elapsedTime := time.Since(startTime).Seconds()
			go rp.metrics.AddModelInferMetrics(externalModelName, internalModelName, metrics.MethodTypeRest, elapsedTime, metrics.HttpCodeToString(http.StatusNotFound))
			http.NotFound(w, r)
		} else {
			r.URL.Path = rewritePath(r.URL.Path, internalModelName)
			rp.logger.Debugf("Calling %s", r.URL.Path)

			proxy.ServeHTTP(w, r)
		}
	}), "seldon-rproxy")
}

func (rp *reverseHTTPProxy) Start() error {
	if rp.stateManager == nil {
		rp.logger.Error("Set state before starting reverse proxy service")
		return fmt.Errorf("State not set, aborting")
	}

	backend := rp.getBackEndPath()
	proxy := httputil.NewSingleHostReverseProxy(backend)
	t := &http.Transport{
		MaxIdleConns:        maxIdleConnsHTTP,
		MaxIdleConnsPerHost: maxIdleConnsPerHostHTTP,
		DisableKeepAlives:   disableKeepAlivesHTTP,
		MaxConnsPerHost:     maxConnsPerHostHTTP,
		IdleConnTimeout:     idleConnTimeoutSeconds * time.Second,
	}
	proxy.Transport = &lazyModelLoadTransport{rp.stateManager.v2Client.LoadModel, t, rp.metrics}
	rp.logger.Infof("Start reverse proxy on port %d for %s", rp.servicePort, backend)
	rp.server = &http.Server{Addr: ":" + strconv.Itoa(int(rp.servicePort)), Handler: rp.addHandlers(proxy)}
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

func (rp *reverseHTTPProxy) getBackEndPath() *url.URL {
	return &url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(rp.backendHTTPServerHost, strconv.Itoa(int(rp.backendHTTPServerPort))),
		Path:   "/",
	}
}

func (rp *reverseHTTPProxy) Stop() error {
	// Shutdown is graceful
	rp.mu.Lock()
	defer rp.mu.Unlock()
	err := rp.server.Shutdown(context.Background())
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
	backendHTTPServerHost string,
	backendHTTPServerPort uint,
	servicePort uint,
	metrics metrics.AgentMetricsHandler,
) *reverseHTTPProxy {

	rp := reverseHTTPProxy{
		logger:                logger.WithField("Source", "HTTPProxy"),
		backendHTTPServerHost: backendHTTPServerHost,
		backendHTTPServerPort: backendHTTPServerPort,
		servicePort:           servicePort,
		metrics:               metrics,
	}

	return &rp
}

func rewritePath(path string, modelName string) string {
	re := regexp.MustCompile(`(/v2/models/)([\w\-]+)(/versions/\w+)?(.*)$`)
	// ${3}, i.e. versions/<ver_num> is removed
	s := fmt.Sprintf("${1}%s${4}", modelName)
	return re.ReplaceAllString(path, s)
}

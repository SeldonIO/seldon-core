/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package agent

import (
	"bytes"
	"context"
	"crypto/tls"
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

	log "github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/modelscaling"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/envoy/resources"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/metrics"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type reverseHTTPProxy struct {
	stateManager               *LocalStateManager
	logger                     log.FieldLogger
	server                     *http.Server
	serverReady                bool
	backendHTTPServerHost      string
	backendHTTPServerPort      uint
	servicePort                uint
	mu                         sync.RWMutex
	metrics                    metrics.AgentMetricsHandler
	tlsOptions                 util.TLSOptions
	modelScalingStatsCollector *modelscaling.DataPlaneStatsCollector
}

// in the case the model is not loaded on server (return 404), we attempt to load it and then retry request
type lazyModelLoadTransport struct {
	loader func(string) *interfaces.ControlPlaneErr
	http.RoundTripper
	metrics                    metrics.AgentMetricsHandler
	modelScalingStatsCollector *modelscaling.DataPlaneStatsCollector
	logger                     log.FieldLogger
}

func addRequestIdToResponse(req *http.Request, res *http.Response) {
	resRequestIds := res.Header[util.RequestIdHeaderCanonical]
	reqRequestIds := req.Header[util.RequestIdHeaderCanonical]
	if len(resRequestIds) == 0 {
		if len(reqRequestIds) == 0 {
			res.Header[util.RequestIdHeaderCanonical] = []string{util.CreateRequestId()}
		} else {
			res.Header[util.RequestIdHeaderCanonical] = reqRequestIds
		}
	}
}

func getRequestId(req *http.Request) string {
	var requestId string
	requestIds := req.Header[util.RequestIdHeaderCanonical]
	if len(requestIds) == 0 {
		requestId = util.CreateRequestId()
		req.Header[util.RequestIdHeaderCanonical] = []string{requestId}
	} else {
		requestId = requestIds[0]
	}
	return requestId
}

// RoundTrip implements http.RoundTripper for the Transport type.
// It calls its underlying http.RoundTripper to execute the request, and
// adds retry logic if we get 404
func (t *lazyModelLoadTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var originalBody []byte
	var err error

	externalModelName := req.Header.Get(resources.SeldonModelHeader)
	internalModelName := req.Header.Get(resources.SeldonInternalModelHeader)

	// to sync between ModelInferEnter and ModelInferExit calls running in go routines
	var wg sync.WaitGroup
	wg.Add(1)

	requestId := getRequestId(req)
	go func() {
		if err := t.modelScalingStatsCollector.ModelInferEnter(internalModelName, requestId); err != nil {
			t.logger.WithError(err).Warnf("cannot collect scaling stats for model %s", internalModelName)
		}
		wg.Done()
	}()
	defer func() {
		go func() {
			wg.Wait()
			if err := t.modelScalingStatsCollector.ModelInferExit(internalModelName, requestId); err != nil {
				t.logger.WithError(err).Warnf("cannot collect scaling stats for model %s", internalModelName)
			}
		}()
	}()

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

	// in the case of triton, a request to a model that is not found is considered a bad request
	// this is likely to increase latency for genuine bad requests as we will retry twice
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusBadRequest {
		internalModelName := req.Header.Get(resources.SeldonInternalModelHeader)
		if v2Err := t.loader(internalModelName); v2Err != nil {
			t.logger.WithError(v2Err).Warnf("cannot load model %s", internalModelName)
		}

		req2 := req.Clone(req.Context())

		req2.Body = io.NopCloser(bytes.NewBuffer(originalBody))
		res, err = t.RoundTripper.RoundTrip(req2)
	}

	addRequestIdToResponse(req, res)

	elapsedTime := time.Since(startTime).Seconds()
	go t.metrics.AddModelInferMetrics(externalModelName, internalModelName, metrics.MethodTypeRest, elapsedTime, metrics.HttpCodeToString(res.StatusCode))
	return res, err

}

func (rp *reverseHTTPProxy) addHandlers(proxy http.Handler) http.Handler {
	return otelhttp.NewHandler(rp.metrics.AddModelHistogramMetricsHandler(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now()
		rp.logger.Debugf("Received request with host %s and internal header %v", r.Host, r.Header.Values(resources.SeldonInternalModelHeader))
		rewriteHostHandler(r)

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
	var err error
	if rp.stateManager == nil {
		rp.logger.Error("Set state before starting reverse proxy service")
		return fmt.Errorf("State not set, aborting")
	}
	rp.tlsOptions, err = util.CreateUpstreamDataplaneServerTLSOptions()
	if err != nil {
		return err
	}
	backend := rp.getBackEndPath()
	proxy := httputil.NewSingleHostReverseProxy(backend)
	t := &http.Transport{
		MaxIdleConns:        util.MaxIdleConnsHTTP,
		MaxIdleConnsPerHost: util.MaxIdleConnsPerHostHTTP,
		DisableKeepAlives:   util.DisableKeepAlivesHTTP,
		MaxConnsPerHost:     util.MaxConnsPerHostHTTP,
		IdleConnTimeout:     util.IdleConnTimeoutSeconds * time.Second,
	}
	proxy.Transport = &lazyModelLoadTransport{rp.stateManager.v2Client.LoadModel, t, rp.metrics, rp.modelScalingStatsCollector, rp.logger}
	rp.logger.Infof("Start reverse proxy on port %d for %s", rp.servicePort, backend)
	var tlsConfig *tls.Config
	if rp.tlsOptions.TLS {
		tlsConfig = rp.tlsOptions.Cert.CreateServerTLSConfig()
	}
	rp.server = &http.Server{
		Addr:      ":" + strconv.Itoa(int(rp.servicePort)),
		Handler:   rp.addHandlers(proxy),
		TLSConfig: tlsConfig,
	}
	// TODO: check for errors? we rely for now on Ready
	go func() {
		rp.mu.Lock()
		rp.serverReady = true
		rp.mu.Unlock()
		if rp.tlsOptions.TLS {
			err := rp.server.ListenAndServeTLS("", "")
			rp.logger.WithError(err).Info("HTTPS/REST reverse proxy debug service stopped")
		} else {
			err := rp.server.ListenAndServe()
			rp.logger.WithError(err).Info("HTTP/REST reverse proxy debug service stopped")
		}
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
	rp.logger.Info("Start graceful shutdown")
	// Shutdown is graceful
	rp.mu.Lock()
	defer rp.mu.Unlock()
	var err error
	if rp.server != nil {
		err = rp.server.Shutdown(context.Background())
	}
	rp.serverReady = false
	rp.logger.Info("Finished graceful shutdown")
	return err
}

func (rp *reverseHTTPProxy) Ready() bool {
	rp.mu.RLock()
	defer rp.mu.RUnlock()
	return rp.serverReady
}

func (rp *reverseHTTPProxy) SetState(stateManager interface{}) {
	rp.stateManager = stateManager.(*LocalStateManager)
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
	modelScalingStatsCollector *modelscaling.DataPlaneStatsCollector,
) *reverseHTTPProxy {

	rp := reverseHTTPProxy{
		logger:                     logger.WithField("Source", "HTTPProxy"),
		backendHTTPServerHost:      backendHTTPServerHost,
		backendHTTPServerPort:      backendHTTPServerPort,
		servicePort:                servicePort,
		metrics:                    metrics,
		modelScalingStatsCollector: modelScalingStatsCollector,
	}

	return &rp
}

func rewritePath(path string, modelName string) string {
	re := regexp.MustCompile(`(/v2/models/)([\w\-]+)(/versions/\w+)?(.*)$`)
	// ${3}, i.e. versions/<ver_num> is removed
	s := fmt.Sprintf("${1}%s${4}", modelName)
	return re.ReplaceAllString(path, s)
}

// need to rewrite the host of the outbound request with the host of the incoming request
// (added by ReverseProxy)
func rewriteHostHandler(r *http.Request) {
	r.Host = r.URL.Host
}

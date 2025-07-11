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
	"encoding/json"
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

func parserOpenAIAPI(body []byte, logger log.FieldLogger) (string, error) {
	// define final inference request structure
	inferenceRequest := make(map[string]interface{})

	// unmarshal the body to extract OpenAI API request
	var jsonBody map[string]interface{}
	err := json.Unmarshal(body, &jsonBody)
	if err != nil {
		logger.WithError(err).Warn("Failed to parse OpenAI API request body")
		return "", err
	}

	// extract model name
	modelName, _ := jsonBody["model"]
	delete(jsonBody, "model")
	logger.Debug("OpenAI API request model: ", modelName)

	// parse messages
	roles := make([]string, 0)
	contents := make([]string, 0)
	types := make([]string, 0)
	toolCalls := make([]string, 0)
	toolCallIds := make([]string, 0)

	messages := jsonBody["messages"]
	delete(jsonBody, "messages")
	for i, message := range messages.([]interface{}) {
		msgMap, err := message.(map[string]interface{})
		if !err {
			logger.Warnf("Failed to parse message %d in OpenAI API request", i)
			continue
		}

		// append role
		role := msgMap["role"].(string)
		roles = append(roles, role)

		// append content and type
		content := msgMap["content"]
		var contentType string
		var contentMessage string

		switch content.(type) {
		case string:
			contentType = "text"
			contentMessage = content.(string)
		case []interface{}:
			contentType = content.(map[string]interface{})["type"].(string)
			if contentType == "text" {
				contentMessage = content.(map[string]interface{})[contentType].(string)
			} else {
				jsonContentMessage, err := json.Marshal(content.(map[string]interface{})[contentType])
				if err != nil {
					logger.WithError(err).Warnf("Failed to marshal content for message %d in OpenAI API request", i)
					continue
				}
				contentMessage = string(jsonContentMessage)
			}
		}
		contents = append(contents, contentMessage)
		types = append(types, contentType)

		// append tool calls
		msgToolCalls := msgMap["tools"].(string)
		toolCalls = append(toolCalls, msgToolCalls)

		// append tool call ids
		msgToolCallsIds := msgMap["tool_call_ids"].(string)
		toolCallIds = append(toolCallIds, msgToolCallsIds)
	}

	inputs := make([]interface{}, 0)
	if len(roles) == 1 {
		roleTensor := map[string]interface{}{
			"name":     "role",
			"shape":    []int{1},
			"datatype": "BYTES",
			"data":     []string{roles[0]},
		}
		contentTensor := map[string]interface{}{
			"name":     "content",
			"shape":    []int{1},
			"datatype": "BYTES",
			"data":     []string{contents[0]},
		}
		typeTensor := map[string]interface{}{
			"name":     "type",
			"shape":    []int{1},
			"datatype": "BYTES",
			"data":     []string{types[0]},
		}
		toolCallsTensor := map[string]interface{}{
			"name":     "tool_calls",
			"shape":    []int{1},
			"datatype": "BYTES",
			"data":     []string{toolCalls[0]},
		}
		toolCallIdsTensor := map[string]interface{}{
			"name":     "tool_call_ids",
			"shape":    []int{1},
			"datatype": "BYTES",
			"data":     []string{toolCallIds[0]},
		}
		inputs = append(inputs, roleTensor)
		inputs = append(inputs, contentTensor)
		inputs = append(inputs, typeTensor)
		inputs = append(inputs, toolCallsTensor)
		inputs = append(inputs, toolCallIdsTensor)
	} else if len(roles) > 1 {
		roleTensor := map[string]interface{}{
			"name":     "role",
			"shape":    []int{len(roles)},
			"datatype": "BYTES",
			"data":     roles,
		}

		jsonContents := make([]string, len(contents))
		jsonTypes := make([]string, len(types))
		jsonToolCalls := make([]string, len(toolCalls))
		jsonToolCallIds := make([]string, len(toolCallIds))

		for i, _ := range contents {
			dataContent, err := json.Marshal([]string{contents[i]})
			if err == nil {
				jsonContents[i] = string(dataContent)
			}

			dataType, err := json.Marshal([]string{types[i]})
			if err == nil {
				jsonTypes[i] = string(dataType)
			}

			dataToolCalls, err := json.Marshal([]string{toolCalls[i]})
			if err == nil {
				jsonToolCalls[i] = string(dataToolCalls)
			}

			dataToolCallIds, err := json.Marshal([]string{toolCallIds[i]})
			if err == nil {
				jsonToolCallIds[i] = string(dataToolCallIds)
			}
		}

		contentTensor := map[string]interface{}{
			"name":     "content",
			"shape":    []int{len(contents)},
			"datatype": "BYTES",
			"data":     jsonContents,
		}
		typeTensor := map[string]interface{}{
			"name":     "type",
			"shape":    []int{len(types)},
			"datatype": "BYTES",
			"data":     jsonTypes,
		}
		toolCallsTensor := map[string]interface{}{
			"name":     "tool_calls",
			"shape":    []int{len(toolCalls)},
			"datatype": "BYTES",
			"data":     jsonToolCalls,
		}
		toolCallIdsTensor := map[string]interface{}{
			"name":     "tool_call_ids",
			"shape":    []int{len(toolCallIds)},
			"datatype": "BYTES",
			"data":     jsonToolCallIds,
		}

		inputs = append(inputs, roleTensor)
		inputs = append(inputs, contentTensor)
		inputs = append(inputs, typeTensor)
		inputs = append(inputs, toolCallsTensor)
		inputs = append(inputs, toolCallIdsTensor)
	}
	inferenceRequest["inputs"] = inputs

	// tools
	tools, ok := jsonBody["tools"]
	delete(jsonBody, "tools")
	if ok {
		data, err := json.Marshal(tools)
		if err != nil {
			logger.WithError(err).Warn("Failed to marshal OpenAI API request tools")
			return "", err
		}
		inferenceRequest["tools"] = string(data)
		delete(jsonBody, "tools")
	}

	// llm parameters
	llmParams := make(map[string]interface{})
	for key, value := range jsonBody {
		llmParams[key] = value
	}

	// final OIP infer request
	data, err := json.Marshal(map[string]interface{}{
		"inputs": inputs,
		"parameters": map[string]interface{}{
			"llm_parameters": llmParams,
		},
	})
	if err != nil {
		logger.WithError(err).Warn("Failed to marshal OpenAI API request inputs")
		return "", err
	}
	return string(data), nil
}

// RoundTrip implements http.RoundTripper for the Transport type.
// It calls its underlying http.RoundTripper to execute the request, and
// adds retry logic if we get 404
func (t *lazyModelLoadTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var originalBody []byte
	var err error

	internalModelName := req.Header.Get(util.SeldonInternalModelHeader)
	// externalModelName is the name of the model as it is known to the client, we should not use
	// util.SeldonModelHeader though as it can contain the experiment tag (used for routing by envoy)
	// however for the metrics we need the actual model name and this is done by using util.SeldonInternalModelHeader
	externalModelName, _, err := util.GetOrignalModelNameAndVersion(internalModelName)
	if err != nil {
		t.logger.WithError(err).Warnf("cannot extract model name from %s, revert to actual header", internalModelName)
		externalModelName = req.Header.Get(util.SeldonModelHeader)
	}

	// to sync between scalingMetricsSetup and scalingMetricsTearDown calls running in go routines
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		if err := t.modelScalingStatsCollector.ScalingMetricsSetup(&wg, internalModelName); err != nil {
			t.logger.WithError(err).Warnf("cannot collect scaling stats for model %s", internalModelName)
		}
	}()
	defer func() {
		go func() {
			if err := t.modelScalingStatsCollector.ScalingMetricsTearDown(&wg, internalModelName); err != nil {
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

	openAIHeader := req.Header.Get(util.SeldonOpenAIHeader)
	if openAIHeader != "" {
		body, _ := parserOpenAIAPI(originalBody, t.logger)
		res := &http.Response{
			StatusCode: http.StatusBadRequest,
			Header:     make(http.Header),
			Body:       io.NopCloser(bytes.NewBufferString(body)),
		}
		return res, nil
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
		internalModelName := req.Header.Get(util.SeldonInternalModelHeader)
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
		rp.logger.Debugf("Received request with host %s and internal header %v", r.Host, r.Header.Values(util.SeldonInternalModelHeader))
		rewriteHostHandler(r)

		internalModelName := r.Header.Get(util.SeldonInternalModelHeader)
		// externalModelName is the name of the model as it is known to the client, we should not use
		// util.SeldonModelHeader though as it can contain the experiment tag (used for routing by envoy)
		// however for the metrics we need the actual model name and this is done by using util.SeldonInternalModelHeader
		externalModelName, _, err := util.GetOrignalModelNameAndVersion(internalModelName)
		if err != nil {
			rp.logger.WithError(err).Warnf("cannot extract model name from %s, revert to actual header", internalModelName)
			externalModelName = r.Header.Get(util.SeldonModelHeader)
		}

		//TODO should we return a 404 if headers not found?
		if externalModelName == "" || internalModelName == "" {
			rp.logger.Warnf("Failed to extract model name %s:[%s] %s:[%s]", util.SeldonInternalModelHeader, internalModelName, util.SeldonModelHeader, externalModelName)
			proxy.ServeHTTP(w, r)
			return
		} else {
			rp.logger.Debugf("Extracted model name %s:%s %s:%s", util.SeldonInternalModelHeader, internalModelName, util.SeldonModelHeader, externalModelName)
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
	proxy.Transport = &lazyModelLoadTransport{
		rp.stateManager.v2Client.LoadModel,
		t,
		rp.metrics,
		rp.modelScalingStatsCollector,
		rp.logger,
	}
	rp.logger.Infof("Start reverse proxy on port %d for %s", rp.servicePort, backend)
	var tlsConfig *tls.Config
	if rp.tlsOptions.TLS {
		tlsConfig = rp.tlsOptions.Cert.CreateServerTLSConfig()
	}
	rp.server = &http.Server{
		Addr:      ":" + strconv.Itoa(int(rp.servicePort)),
		Handler:   rp.addHandlers(proxy),
		TLSConfig: tlsConfig,
		BaseContext: func(net.Listener) context.Context {
			// BaseContext is called once the server has spun up and is accepting connections
			rp.mu.Lock()
			rp.serverReady = true
			rp.mu.Unlock()
			return context.Background()
		},
	}

	// TODO: check for errors? we rely for now on Ready
	go func() {
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
		ctx, cancel := context.WithTimeout(context.Background(), util.ServerControlPlaneTimeout)
		defer cancel()
		err = rp.server.Shutdown(ctx)
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

func (rp *reverseHTTPProxy) GetType() interfaces.SubServiceType {
	return interfaces.CriticalDataPlaneService
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

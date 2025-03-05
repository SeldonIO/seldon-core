/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package agent

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/mux"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/modelscaling"
	testing_utils2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/internal/testing_utils"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/metrics"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type mockMLServerState struct {
	models         map[string]bool
	mu             *sync.Mutex
	modelsNotFound map[string]bool
}

func (mlserver *mockMLServerState) v2Infer(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	modelName := params["model_name"]
	if _, ok := mlserver.modelsNotFound[modelName]; ok {
		http.NotFound(w, req)
	}
	_, _ = w.Write([]byte("Model inference: " + modelName))
}

func (mlserver *mockMLServerState) v2InferStream(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	modelName := params["model_name"]
	if _, ok := mlserver.modelsNotFound[modelName]; ok {
		http.NotFound(w, req)
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "text/plain")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	chunks := []string{"Model ", "inference: ", modelName}
	for _, chunk := range chunks {
		newLineChunk := chunk + "\n"
		_, _ = w.Write([]byte(newLineChunk)) // Write a chunk
		flusher.Flush()                      // Flush to send immediately
	}
}

func (mlserver *mockMLServerState) v2Load(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	modelName := params["model_name"]
	delete(mlserver.modelsNotFound, modelName)
	mlserver.setModel(modelName, true)
	_, _ = w.Write([]byte("Model load: " + modelName))
}

func (mlserver *mockMLServerState) v2Unload(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	modelName := params["model_name"]
	mlserver.setModel(modelName, false)
	_, _ = w.Write([]byte("Model unload: " + modelName))
}

func (mlserver *mockMLServerState) setModel(modelId string, val bool) {
	mlserver.mu.Lock()
	defer mlserver.mu.Unlock()
	mlserver.models[modelId] = val
}

func (mlserver *mockMLServerState) setModelServerUnloaded(modelId string) {
	mlserver.modelsNotFound[modelId] = true
}

func (mlserver *mockMLServerState) isModelLoaded(modelId string) bool {
	mlserver.mu.Lock()
	defer mlserver.mu.Unlock()
	val, loaded := mlserver.models[modelId]
	if loaded {
		return val
	}
	return false
}

func setupMockMLServer(mockMLServerState *mockMLServerState, serverPort int) *http.Server {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/v2/models/{model_name:\\w+}/infer", mockMLServerState.v2Infer).Methods("POST")
	rtr.HandleFunc("/v2/models/{model_name:\\w+}/infer_stream", mockMLServerState.v2InferStream).Methods("POST")
	rtr.HandleFunc("/v2/repository/models/{model_name:\\w+}/load", mockMLServerState.v2Load).Methods("POST")
	rtr.HandleFunc("/v2/repository/models/{model_name:\\w+}/unload", mockMLServerState.v2Unload).Methods("POST")
	return &http.Server{Addr: ":" + strconv.Itoa(serverPort), Handler: rtr}
}

type loadModelSateValue struct {
	memory uint64
	isLoad bool
	isSoft bool
}

type inferModelSateValue struct {
	internalModelName string
	method            string
	code              string
}

type fakeMetricsHandler struct {
	modelLoadState  map[string]loadModelSateValue
	modelInferState map[string]inferModelSateValue
	mu              *sync.Mutex
}

func (f fakeMetricsHandler) AddModelHistogramMetricsHandler(baseHandler http.HandlerFunc) http.HandlerFunc {
	return baseHandler
}

func (f fakeMetricsHandler) HttpCodeToString(code int) string {
	return fmt.Sprintf("%d", code)
}

func (f fakeMetricsHandler) AddModelInferMetrics(externalModelName string, internalModelName string, method string, elapsedTime float64, code string) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.modelInferState[externalModelName] = inferModelSateValue{
		internalModelName: internalModelName,
		method:            method,
		code:              code,
	}
}

func (f fakeMetricsHandler) AddLoadedModelMetrics(internalModelName string, memory uint64, isLoad, isSoft bool) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.modelLoadState[internalModelName] = loadModelSateValue{
		memory: memory,
		isLoad: isLoad,
		isSoft: isSoft,
	}
}

func (f fakeMetricsHandler) AddServerReplicaMetrics(memory uint64, memoryWithOvercommit float32) {
}

func newFakeMetricsHandler() fakeMetricsHandler {
	return fakeMetricsHandler{
		modelLoadState:  map[string]loadModelSateValue{},
		modelInferState: map[string]inferModelSateValue{},
		mu:              &sync.Mutex{},
	}
}

func (f fakeMetricsHandler) UnaryServerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
}

func setupReverseProxy(logger log.FieldLogger, numModels int, modelPrefix string, rpPort, serverPort int, metricsHandler metrics.AgentMetricsHandler) *reverseHTTPProxy {
	v2Client := testing_utils.NewV2RestClientForTest("localhost", serverPort, logger)
	localCacheManager := setupLocalTestManager(numModels, modelPrefix, v2Client, numModels-2, 1)
	modelScalingStatsCollector := modelscaling.NewDataPlaneStatsCollector(
		modelscaling.NewModelReplicaLagsKeeper(),
		modelscaling.NewModelReplicaLastUsedKeeper(),
	)
	rp := NewReverseHTTPProxy(
		logger,
		"localhost",
		uint(serverPort),
		uint(rpPort),
		metricsHandler,
		modelScalingStatsCollector,
	)
	rp.SetState(localCacheManager)
	return rp
}

func TestReverseProxySmoke(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	type test struct {
		name                     string
		modelToLoad              string
		modelToRequest           string
		modelExternalHeader      string
		expectedModelExternalTag string
		statusCode               int
		isLoadedonServer         bool
	}

	tests := []test{
		{
			name:                     "model exists",
			modelToLoad:              "foo_1",
			modelToRequest:           "foo_1",
			modelExternalHeader:      "foo",
			expectedModelExternalTag: "foo",
			statusCode:               http.StatusOK,
			isLoadedonServer:         true,
		},
		{
			name:                     "model exists, part of experiment",
			modelToLoad:              "foo_1",
			modelToRequest:           "foo_1",
			modelExternalHeader:      "foo-experiment.experiment",
			expectedModelExternalTag: "foo",
			statusCode:               http.StatusOK,
			isLoadedonServer:         true,
		},
		{
			name:                     "model exists on agent but not loaded on server",
			modelToLoad:              "foo_1",
			modelToRequest:           "foo_1",
			modelExternalHeader:      "foo",
			expectedModelExternalTag: "foo",
			statusCode:               http.StatusOK,
			isLoadedonServer:         false,
		},
		{
			name:                     "model does not exists",
			modelToLoad:              "foo_1",
			modelToRequest:           "foo2_1",
			modelExternalHeader:      "foo2",
			expectedModelExternalTag: "foo2",
			statusCode:               http.StatusNotFound,
			isLoadedonServer:         false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockMLServerState := &mockMLServerState{
				models:         make(map[string]bool),
				modelsNotFound: make(map[string]bool),
				mu:             &sync.Mutex{},
			}
			serverPort, err := testing_utils2.GetFreePortForTest()
			if err != nil {
				t.Fatal(err)
			}
			mlserver := setupMockMLServer(mockMLServerState, serverPort)
			go func() {
				_ = mlserver.ListenAndServe()
			}()

			rpPort, err := testing_utils2.GetFreePortForTest()
			if err != nil {
				t.Fatal(err)
			}
			fakeMetricsHandler := newFakeMetricsHandler()
			rpHTTP := setupReverseProxy(logger, 3, test.modelToLoad, rpPort, serverPort, fakeMetricsHandler)
			err = rpHTTP.Start()
			g.Expect(err).To(BeNil())
			time.Sleep(500 * time.Millisecond)

			// load model
			err = rpHTTP.stateManager.LoadModelVersion(getDummyModelDetails(test.modelToLoad, uint64(1), uint32(1)))
			g.Expect(err).To(BeNil())

			if !test.isLoadedonServer {
				// this will set a model to fail infer until load is called
				mockMLServerState.setModelServerUnloaded(test.modelToLoad)
			}

			// make a dummy predict call with any model name, URL does not matter, only headers
			createRequest := func(endpoint string) *http.Request {
				inferV2Path := "/v2/models/RANDOM/" + endpoint
				logger.Debug("inferV2Path:", inferV2Path)

				url := "http://localhost:" + strconv.Itoa(rpPort) + inferV2Path
				req, err := http.NewRequest(http.MethodPost, url, nil)
				g.Expect(err).To(BeNil())
				req.Header.Set("contentType", "application/json")
				req.Header.Set(util.SeldonModelHeader, test.modelExternalHeader)
				req.Header.Set(util.SeldonInternalModelHeader, test.modelToRequest)
				return req
			}

			// infer request
			req := createRequest("infer")
			resp, err := http.DefaultClient.Do(req)
			g.Expect(err).To(BeNil())

			g.Expect(resp.StatusCode).To(Equal(test.statusCode))
			if test.statusCode == http.StatusOK {
				bodyBytes, err := io.ReadAll(resp.Body)
				g.Expect(err).To(BeNil())
				bodyString := string(bodyBytes)
				g.Expect(strings.Contains(bodyString, test.modelToLoad)).To(BeTrue())
			}

			// infer_stream request
			req = createRequest("infer_stream")
			resp, err = http.DefaultClient.Do(req)
			g.Expect(err).To(BeNil())

			g.Expect(resp.StatusCode).To(Equal(test.statusCode))
			if test.statusCode == http.StatusOK {
				scanner := bufio.NewScanner(resp.Body)
				messages := make([]string, 0)
				for scanner.Scan() {
					messages = append(messages, scanner.Text())
				}

				g.Expect(scanner.Err()).To(BeNil())

				messages_concat := strings.Join(messages, "")
				g.Expect(strings.Contains(messages_concat, test.modelToLoad)).To(BeTrue())
			}

			//  test model scaling stats
			if test.statusCode == http.StatusOK {
				g.Expect(rpHTTP.modelScalingStatsCollector.ModelLagStats.Get(test.modelToRequest)).To(Equal(uint32(0)))
				g.Expect(rpHTTP.modelScalingStatsCollector.ModelLastUsedStats.Get(test.modelToRequest)).Should(BeNumerically("<=", time.Now().Unix())) // only triggered when we get results back
			}

			//  test infer metrics
			g.Expect(fakeMetricsHandler.modelInferState[test.expectedModelExternalTag].internalModelName).To(Equal(test.modelToRequest))
			g.Expect(fakeMetricsHandler.modelInferState[test.expectedModelExternalTag].method).To(Equal("rest"))
			if test.statusCode == http.StatusOK {
				g.Expect(fakeMetricsHandler.modelInferState[test.expectedModelExternalTag].code).To(Equal("200"))
			} else {
				g.Expect(fakeMetricsHandler.modelInferState[test.expectedModelExternalTag].code).To(Equal("404"))
			}
			g.Expect(rpHTTP.Ready()).To(BeTrue())
			_ = rpHTTP.Stop()
			g.Expect(rpHTTP.Ready()).To(BeFalse())

			resp.Body.Close()

			_ = mlserver.Shutdown(context.Background())
		})
	}
}

func TestReverseEarlyStop(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	fakeMetricsHandler := newFakeMetricsHandler()
	rpHTTP := setupReverseProxy(logger, 0, "dummy", 1, 1, fakeMetricsHandler)
	err := rpHTTP.Stop()
	g.Expect(err).To(BeNil())
	ready := rpHTTP.Ready()
	g.Expect(ready).To(BeFalse())
}

func TestRewritePath(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name         string
		path         string
		modelName    string
		expectedPath string
	}
	tests := []test{
		{
			name:         "default infer",
			path:         "/v2/models/iris/infer",
			modelName:    "foo",
			expectedPath: "/v2/models/foo/infer",
		},
		{
			name:         "default infer model with dash",
			path:         "/v2/models/iris-1/infer",
			modelName:    "foo",
			expectedPath: "/v2/models/foo/infer",
		},
		{
			name:         "default infer model with underscore",
			path:         "/v2/models/iris_1/infer",
			modelName:    "foo",
			expectedPath: "/v2/models/foo/infer",
		},
		{
			name:         "metadata for model",
			path:         "/v2/models/iris",
			modelName:    "foo",
			expectedPath: "/v2/models/foo",
		},
		{
			name:         "for server calls no change",
			path:         "/v2/health/live",
			modelName:    "foo",
			expectedPath: "/v2/health/live",
		},
		{
			name:         "model ready",
			path:         "/v2/models/iris/ready",
			modelName:    "foo",
			expectedPath: "/v2/models/foo/ready",
		},
		{
			name:         "versioned infer",
			path:         "/v2/models/iris/versions/1/infer",
			modelName:    "foo",
			expectedPath: "/v2/models/foo/infer",
		},
		{
			name:         "versioned metadata",
			path:         "/v2/models/iris/versions/1/infer",
			modelName:    "foo",
			expectedPath: "/v2/models/foo/infer",
		},
		{
			name:         "versioned model ready",
			path:         "/v2/models/iris/versions/1/ready",
			modelName:    "foo",
			expectedPath: "/v2/models/foo/ready",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rewrittenPath := rewritePath(test.path, test.modelName)
			g.Expect(rewrittenPath).To(Equal(test.expectedPath))
		})
	}
}

func TestLazyLoadRoundTripper(t *testing.T) {
	g := NewGomegaWithT(t)
	dummyModel := "foo"

	type test struct {
		name      string
		dummyBody []byte
	}
	tests := []test{
		{
			name:      "non-empty body",
			dummyBody: []byte{97, 98, 99, 100, 101, 102},
		},
		{
			name:      "empty body",
			dummyBody: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockMLServerState := &mockMLServerState{
				models:         make(map[string]bool),
				modelsNotFound: make(map[string]bool),
				mu:             &sync.Mutex{},
			}
			serverPort, err := testing_utils2.GetFreePortForTest()
			if err != nil {
				t.Fatal(err)
			}
			mlserver := setupMockMLServer(mockMLServerState, serverPort)
			go func() {
				_ = mlserver.ListenAndServe()
			}()

			time.Sleep(util.GRPCRetryBackoff)

			defer func() {
				_ = mlserver.Shutdown(context.Background())
			}()

			basePath := "http://localhost:" + strconv.Itoa(serverPort)

			loader := func(model string) *interfaces.ControlPlaneErr {
				loadV2Path := basePath + "/v2/repository/models/" + model + "/load"
				httpClient := http.DefaultClient
				httpClient.Transport = http.DefaultTransport
				req, _ := http.NewRequest(http.MethodPost, loadV2Path, nil)
				_, _ = httpClient.Do(req)
				return nil
			}

			inferV2Path := "/v2/models/" + dummyModel + "/infer"
			inferUrl := basePath + inferV2Path
			req, err := http.NewRequest(http.MethodPost, inferUrl, bytes.NewBuffer(test.dummyBody))
			g.Expect(err).To(BeNil())
			req.Header.Set("contentType", "application/json")
			httpClient := http.DefaultClient
			metricsHandler := newFakeMetricsHandler()
			modelScalingStatsCollector := modelscaling.NewDataPlaneStatsCollector(
				modelscaling.NewModelReplicaLagsKeeper(), modelscaling.NewModelReplicaLastUsedKeeper())
			httpClient.Transport = &lazyModelLoadTransport{
				loader, http.DefaultTransport, metricsHandler, modelScalingStatsCollector, log.New()}
			mockMLServerState.setModelServerUnloaded(dummyModel)
			req.Header.Set(util.SeldonInternalModelHeader, dummyModel)
			resp, err := httpClient.Do(req)
			g.Expect(err).To(BeNil())
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	}
}

func TestAddRequestIdToResponse(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name               string
		req                *http.Request
		res                *http.Response
		expectNewRequestId bool
		expectedRequestId  string
	}
	tests := []test{
		{
			name: "no request id present",
			req: &http.Request{
				Header: map[string][]string{},
			},
			res:                &http.Response{Header: map[string][]string{}},
			expectNewRequestId: true,
		},
		{
			name: "request id in request",
			req: &http.Request{
				Header: map[string][]string{util.RequestIdHeaderCanonical: {"1234"}},
			},
			res:                &http.Response{Header: map[string][]string{}},
			expectNewRequestId: false,
			expectedRequestId:  "1234",
		},
		{
			name: "request id in response",
			req: &http.Request{
				Header: map[string][]string{},
			},
			res:                &http.Response{Header: map[string][]string{util.RequestIdHeaderCanonical: {"1234"}}},
			expectNewRequestId: false,
			expectedRequestId:  "1234",
		},
		{
			name: "request id in request and response",
			req: &http.Request{
				Header: map[string][]string{util.RequestIdHeaderCanonical: {"9999"}},
			},
			res:                &http.Response{Header: map[string][]string{util.RequestIdHeaderCanonical: {"1234"}}},
			expectNewRequestId: false,
			expectedRequestId:  "1234",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			addRequestIdToResponse(test.req, test.res)
			headers := test.res.Header[util.RequestIdHeaderCanonical]
			if test.expectNewRequestId {
				g.Expect(len(headers)).To(Equal(1))
			} else {
				g.Expect(headers[0]).To(Equal(test.expectedRequestId))
			}
		})
	}
}

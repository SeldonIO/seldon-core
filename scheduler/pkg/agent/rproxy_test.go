package agent

import (
	"context"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/seldonio/seldon-core/scheduler/pkg/envoy/resources"
	"google.golang.org/grpc"

	"github.com/gorilla/mux"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

const (
	backEndServerPort = 7777
)

func getFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

type mockMLServerState struct {
	models map[string]bool
	mu     *sync.Mutex
}

func (mlserver *mockMLServerState) v2Infer(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	model_name := params["model_name"]
	_, _ = w.Write([]byte("Model inference: " + model_name))
}

func (mlserver *mockMLServerState) v2Load(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)
	modelName := params["model_name"]
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

func (mlserver *mockMLServerState) isModelLoaded(modelId string) bool {
	mlserver.mu.Lock()
	defer mlserver.mu.Unlock()
	val, loaded := mlserver.models[modelId]
	if loaded {
		return val
	}
	return false
}

func isRegistered(port int) bool {
	timeout := 5 * time.Second
	conn, err := net.DialTimeout("tcp", ":"+strconv.Itoa(port), timeout)
	if err != nil {
		return false
	}

	if conn != nil {
		conn.Close()
		return true
	}

	return false
}
func setupMockMLServer(mockMLServerState *mockMLServerState) {
	if isRegistered(backEndServerPort) {
		log.Warnf("Port %d already running", backEndServerPort)
		return
	}
	rtr := mux.NewRouter()
	rtr.HandleFunc("/v2/models/{model_name:\\w+}/infer", mockMLServerState.v2Infer).Methods("POST")
	rtr.HandleFunc("/v2/repository/models/{model_name:\\w+}/load", mockMLServerState.v2Load).Methods("POST")
	rtr.HandleFunc("/v2/repository/models/{model_name:\\w+}/unload", mockMLServerState.v2Unload).Methods("POST")

	http.Handle("/", rtr)

	if err := http.ListenAndServe(":"+strconv.Itoa(backEndServerPort), nil); err != nil {
		log.Warn(err)
	}
}

type loadModelSateValue struct {
	memory uint64
	isLoad bool
	isSoft bool
}

type fakeMetricsHandler struct {
	modelLoadState map[string]loadModelSateValue
	mu             *sync.Mutex
}

func (f fakeMetricsHandler) AddHistogramMetricsHandler(baseHandler http.HandlerFunc) http.HandlerFunc {
	return baseHandler
}

func (f fakeMetricsHandler) AddInferMetrics(externalModelName string, internalModelName string, method string, elapsedTime float64) {
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
		modelLoadState: map[string]loadModelSateValue{},
		mu:             &sync.Mutex{},
	}
}

func (f fakeMetricsHandler) UnaryServerInterceptor() func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		return handler(ctx, req)
	}
}

func setupReverseProxy(logger log.FieldLogger, numModels int, modelPrefix string, rpPort int) *reverseHTTPProxy {
	v2Client := NewV2Client("localhost", backEndServerPort, logger)
	localCacheManager := setupLocalTestManager(numModels, modelPrefix, v2Client, numModels-2, 1)
	rp := NewReverseHTTPProxy(logger, uint(rpPort), fakeMetricsHandler{})
	rp.SetState(localCacheManager)
	return rp
}

func TestReverseProxySmoke(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := log.New()
	logger.SetLevel(log.DebugLevel)

	type test struct {
		name           string
		modelToLoad    string
		modelToRequest string
		statusCode     int
	}

	tests := []test{
		{
			name:           "model exists",
			modelToLoad:    "foo",
			modelToRequest: "foo",
			statusCode:     200,
		},
		{
			name:           "model does not exists",
			modelToLoad:    "foo",
			modelToRequest: "foo2",
			statusCode:     404,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockMLServerState := &mockMLServerState{
				models: make(map[string]bool),
				mu:     &sync.Mutex{},
			}
			go setupMockMLServer(mockMLServerState)
			rpPort, err := getFreePort()
			if err != nil {
				t.Fatal(err)
			}
			rpHTTP := setupReverseProxy(logger, 3, test.modelToLoad, rpPort)
			err = rpHTTP.Start()
			g.Expect(err).To(BeNil())
			time.Sleep(100 * time.Millisecond)

			// load model
			_, _ = rpHTTP.stateManager.modelVersions.addModelVersion(
				getDummyModelDetails(test.modelToLoad, uint64(1), uint32(1)))

			// make a dummy predict call with any model name
			inferV2Path := "/v2/models/RANDOM/infer"
			url := "http://localhost:" + strconv.Itoa(rpPort) + inferV2Path
			req, err := http.NewRequest(http.MethodPost, url, nil)
			g.Expect(err).To(BeNil())
			req.Header.Set("contentType", "application/json")
			req.Header.Set(resources.SeldonModelHeader, test.modelToRequest)
			req.Header.Set(resources.SeldonInternalModelHeader, test.modelToRequest)
			resp, err := http.DefaultClient.Do(req)
			g.Expect(err).To(BeNil())
			defer resp.Body.Close()

			g.Expect(resp.StatusCode).To(Equal(test.statusCode))
			if test.statusCode == 200 {
				bodyBytes, err := io.ReadAll(resp.Body)
				g.Expect(err).To(BeNil())
				bodyString := string(bodyBytes)
				g.Expect(strings.Contains(bodyString, test.modelToLoad)).To(BeTrue())
			}
			g.Expect(rpHTTP.Ready()).To(BeTrue())
			_ = rpHTTP.Stop()
			g.Expect(rpHTTP.Ready()).To(BeFalse())
		})
	}

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

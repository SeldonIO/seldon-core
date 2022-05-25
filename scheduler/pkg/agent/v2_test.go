package agent

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

type v2State struct {
	models map[string]bool
	mu     sync.Mutex
}

func (s *v2State) loadResponder(model string, status int) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		s.setModel(model, true)
		if status == 200 {
			return httpmock.NewStringResponse(status, ""), nil
		} else {
			return httpmock.NewStringResponse(status, ""), ErrV2BadRequest
		}
	}
}

func (s *v2State) unloadResponder(model string, status int) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		s.setModel(model, false)
		if status == 200 {
			return httpmock.NewStringResponse(status, ""), nil
		} else {
			return httpmock.NewStringResponse(status, ""), ErrV2BadRequest
		}
	}
}

func (s *v2State) setModel(modelId string, val bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.models[modelId] = val
}

func (s *v2State) isModelLoaded(modelId string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	val, loaded := s.models[modelId]
	if loaded {
		return val
	}
	return false
}

func createTestV2ClientMockResponders(host string, port int, modelName string, status int, state *v2State) {

	httpmock.RegisterResponder("POST", fmt.Sprintf("http://%s:%d/v2/repository/models/%s/load", host, port, modelName),
		state.loadResponder(modelName, status))
	httpmock.RegisterResponder("POST", fmt.Sprintf("http://%s:%d/v2/repository/models/%s/unload", host, port, modelName),
		state.unloadResponder(modelName, status))
	// we do not care about ready in tests
	httpmock.RegisterResponder("GET", fmt.Sprintf("http://%s:%d/v2/health/ready", host, port),
		httpmock.NewStringResponder(200, `{}`))
}

func createTestV2ClientwithState(models []string, status int) (*V2Client, *v2State) {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	host := "model-server"
	port := 8080
	v2 := NewV2Client(host, port, logger, false)
	state := &v2State{
		models: make(map[string]bool, len(models)),
	}

	for _, model := range models {
		createTestV2ClientMockResponders(host, port, model, status, state)
	}
	return v2, state
}

func createTestV2Client(models []string, status int) *V2Client {
	v2, _ := createTestV2ClientwithState(models, status)
	return v2
}

func TestCommunicationErrors(t *testing.T) {
	// this should fail because of dns error
	g := NewGomegaWithT(t)
	modelName := "dummy"
	r := createTestV2Client([]string{modelName}, 200)
	err := r.LoadModel(modelName)
	g.Expect(err.errCode).To(Equal(V2CommunicationErrCode))
}

func TestRequestErrors(t *testing.T) {
	// in this test we are not enabling httpmock
	// and therefore all http requests should fail because of dns / client error
	g := NewGomegaWithT(t)
	modelName := "dummy"
	v2 := NewV2Client("httpwrong://server", 0, log.New(), false)
	err := v2.LoadModel(modelName)
	g.Expect(err.errCode).To(Equal(V2RequestErrCode))
}

func TestLoad(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		models []string
		status int
		err    error
	}
	tests := []test{
		{models: []string{"iris"}, status: 200, err: nil},
		{models: []string{"iris"}, status: 400, err: ErrV2BadRequest},
	}
	for _, test := range tests {
		r := createTestV2Client(test.models, test.status)
		httpmock.ActivateNonDefault(r.httpClient)
		for _, model := range test.models {
			err := r.LoadModel(model)
			if test.status == 200 {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
				if test.err != nil {
					g.Expect(errors.Is(err.err, test.err)).To(BeTrue())
				}
			}
		}
		g.Expect(httpmock.GetTotalCallCount()).To(Equal(len(test.models)))
		httpmock.DeactivateAndReset()
	}
}

func TestUnload(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		models []string
		status int
		err    error
	}
	tests := []test{
		{models: []string{"iris"}, status: 200, err: nil},
		{models: []string{"iris"}, status: 400, err: ErrV2BadRequest},
	}
	for _, test := range tests {
		r := createTestV2Client(test.models, test.status)
		httpmock.ActivateNonDefault(r.httpClient)
		for _, model := range test.models {
			err := r.UnloadModel(model)
			if test.status == 200 {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
				if test.err != nil {
					g.Expect(errors.Is(err.err, test.err)).To(BeTrue())
				}
			}
		}
		g.Expect(httpmock.GetTotalCallCount()).To(Equal(len(test.models)))
		httpmock.DeactivateAndReset()
	}
}

func TestGrpcV2(t *testing.T) {
	g := NewGomegaWithT(t)

	mockMLServer := &mockGRPCMLServer{}
	backEndGRPCPort, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}
	_ = mockMLServer.setup(uint(backEndGRPCPort))
	go func() {
		_ = mockMLServer.start()
	}()
	defer mockMLServer.stop()

	time.Sleep(10 * time.Millisecond)

	v2Client := NewV2Client("", backEndGRPCPort, log.New(), true)

	dummModel := "dummy"

	v2Err := v2Client.LoadModel(dummModel)
	g.Expect(v2Err).To(BeNil())

	v2Err = v2Client.UnloadModel(dummModel)
	g.Expect(v2Err).To(BeNil())

	v2Err = v2Client.UnloadModel(modelNameMissing)
	g.Expect(v2Err.IsNotFound()).To(BeTrue())

	err = v2Client.Ready()
	g.Expect(err).To(BeNil())

}

func TestGrpcV2WithError(t *testing.T) {
	g := NewGomegaWithT(t)

	// note no grpc server to respond

	backEndGRPCPort, err := getFreePort()
	if err != nil {
		t.Fatal(err)
	}
	v2Client := NewV2Client("", backEndGRPCPort, log.New(), true)

	dummModel := "dummy"

	v2Err := v2Client.LoadModel(dummModel)
	g.Expect(v2Err).NotTo(BeNil())

	v2Err = v2Client.UnloadModel(dummModel)
	g.Expect(v2Err).NotTo(BeNil())

	err = v2Client.Ready()
	g.Expect(err).NotTo(BeNil())

}

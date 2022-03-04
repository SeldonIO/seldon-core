package agent

import (
	"errors"
	"fmt"
	"net/http"
	"sync"
	"testing"

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
			return httpmock.NewStringResponse(status, ""), V2BadRequestErr
		}
	}
}

func (s *v2State) unloadResponder(model string, status int) func(req *http.Request) (*http.Response, error) {
	return func(req *http.Request) (*http.Response, error) {
		s.setModel(model, false)
		if status == 200 {
			return httpmock.NewStringResponse(status, ""), nil
		} else {
			return httpmock.NewStringResponse(status, ""), V2BadRequestErr
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
	v2 := NewV2Client(host, port, logger)
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

func TestLoad(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		models []string
		status int
		err    error
	}
	tests := []test{
		{models: []string{"iris"}, status: 200, err: nil},
		{models: []string{"iris"}, status: 400, err: V2BadRequestErr},
	}
	for _, test := range tests {
		httpmock.Activate()
		r := createTestV2Client(test.models, test.status)
		for _, model := range test.models {
			err := r.LoadModel(model)
			if test.status == 200 {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
				if test.err != nil {
					g.Expect(errors.Is(err, test.err)).To(BeTrue())
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
		{models: []string{"iris"}, status: 400, err: V2BadRequestErr},
	}
	for _, test := range tests {
		httpmock.Activate()
		r := createTestV2Client(test.models, test.status)
		for _, model := range test.models {
			err := r.UnloadModel(model)
			if test.status == 200 {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
				if test.err != nil {
					g.Expect(errors.Is(err, test.err)).To(BeTrue())
				}
			}
		}
		g.Expect(httpmock.GetTotalCallCount()).To(Equal(len(test.models)))
		httpmock.DeactivateAndReset()
	}
}

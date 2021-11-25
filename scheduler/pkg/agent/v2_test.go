package agent

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func createTestV2ClientMockResponders(host string, port int, modelName string, status int) {
	httpmock.RegisterResponder("POST", fmt.Sprintf("http://%s:%d/v2/repository/models/%s/load", host, port, modelName),
		httpmock.NewStringResponder(status, `{}`))
	httpmock.RegisterResponder("POST", fmt.Sprintf("http://%s:%d/v2/repository/models/%s/unload", host, port, modelName),
		httpmock.NewStringResponder(status, `{}`))
}

func createTestV2Client(models []string, status int) *V2Client {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	host := "model-server"
	port := 8080
	v2 := NewV2Client(host, port, logger)
	for _, model := range models {
		createTestV2ClientMockResponders(host, port, model, status)
	}
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

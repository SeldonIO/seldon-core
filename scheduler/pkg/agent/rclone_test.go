package agent

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func createTestRCloneMockResponders(host string, port int, status int) {
	httpmock.RegisterResponder("POST", fmt.Sprintf("=~http://%s:%d/", host, port),
		httpmock.NewStringResponder(status, `{}`))
}

func createTestRCloneClient(status int) *RCloneClient {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	host := "rclone-server"
	port := 5572
	r := NewRCloneClient(host, port, "/tmp/rclone", logger)
	createTestRCloneMockResponders(host, port, status)
	return r
}

func TestRcloneReady(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	t.Logf("Started")
	g := NewGomegaWithT(t)
	r := createTestRCloneClient(200)
	err := r.Ready()
	g.Expect(err).To(BeNil())
	g.Expect(httpmock.GetTotalCallCount()).To(Equal(1))
}

func TestRcloneCopy(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	type test struct {
		modelName string
		uri       string
		status    int
	}
	tests := []test{
		{modelName: "iris", uri: "gs://seldon-models/sklearn/iris-0.23.2/lr_model", status: 200},
		{modelName: "iris", uri: "gs://seldon-models/sklearn/iris-0.23.2/lr_model", status: 400},
	}
	for _, test := range tests {
		httpmock.Activate()
		r := createTestRCloneClient(test.status)
		err := r.Copy(test.modelName, test.uri)
		if test.status == 200 {
			g.Expect(err).To(BeNil())
		} else {
			g.Expect(err).ToNot(BeNil())
		}
		g.Expect(httpmock.GetTotalCallCount()).To(Equal(1))
		httpmock.DeactivateAndReset()
	}
}
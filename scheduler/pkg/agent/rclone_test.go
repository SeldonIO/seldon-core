package agent

import (
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func createTestRCloneMockResponders(host string, port int, status int, body string) {
	httpmock.RegisterResponder("POST", fmt.Sprintf("=~http://%s:%d/", host, port),
		httpmock.NewStringResponder(status, body))
}

func createTestRCloneClient(status int, body string) *RCloneClient {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	host := "rclone-server"
	port := 5572
	r := NewRCloneClient(host, port, "/tmp/rclone", logger)
	createTestRCloneMockResponders(host, port, status, body)
	return r
}

func TestRcloneReady(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	t.Logf("Started")
	g := NewGomegaWithT(t)
	r := createTestRCloneClient(200, "{}")
	err := r.Ready()
	g.Expect(err).To(BeNil())
	g.Expect(httpmock.GetTotalCallCount()).To(Equal(1))
}

func TestRcloneCopy(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	type test struct {
		modelName    string
		modelVersion string
		uri          string
		status       int
		body         string
	}
	tests := []test{
		{modelName: "iris", modelVersion: "1", uri: "gs://seldon-models/sklearn/iris-0.23.2/lr_model", status: 200, body: "{}"},
		{modelName: "iris", modelVersion: "1", uri: "gs://seldon-models/sklearn/iris-0.23.2/lr_model", status: 400, body: "{}"},
	}
	for _, test := range tests {
		httpmock.Activate()
		r := createTestRCloneClient(test.status, test.body)
		err := r.Copy(test.modelName, test.modelVersion, test.uri)
		if test.status == 200 {
			g.Expect(err).To(BeNil())
		} else {
			g.Expect(err).ToNot(BeNil())
		}
		g.Expect(httpmock.GetTotalCallCount()).To(Equal(1))
		httpmock.DeactivateAndReset()
	}
}

func TestRcloneConfig(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	type test struct {
		name               string
		modelName          string
		modelVersion       string
		config             []byte
		expectedPath       string
		existsBody         []byte
		createUpdateStatus int
		err                bool
	}
	tests := []test{
		{
			name:               "CreateOK",
			modelName:          "iris",
			modelVersion:       "1",
			config:             []byte(`{"name":"mys3","type":"s3","parameters":{"foo":"bar"}}`),
			expectedPath:       "/config/create",
			existsBody:         []byte(`{}`),
			createUpdateStatus: 200,
			err:                false,
		},
		{
			name:               "CreateFail",
			modelName:          "iris",
			modelVersion:       "1",
			config:             []byte(`{"name":"mys3","type":"s3","parameters":{"foo":"bar"}}`),
			expectedPath:       "/config/create",
			existsBody:         []byte(`{}`),
			createUpdateStatus: 400,
			err:                true,
		},
		{
			name:               "CreateBadConfig",
			modelName:          "iris",
			modelVersion:       "1",
			config:             []byte(`{"foo":"mys3","type":"s3","parameters":{"foo":"bar"}}`),
			expectedPath:       "/config/create",
			existsBody:         []byte(`{}`),
			createUpdateStatus: 200,
			err:                true,
		},
		{
			name:               "Update",
			modelName:          "iris",
			modelVersion:       "1",
			config:             []byte(`{"name":"mys3","type":"s3","parameters":{"foo":"bar"}}`),
			expectedPath:       "/config/update",
			existsBody:         []byte(`{"name":"mys3"}`),
			createUpdateStatus: 200,
			err:                false,
		},
		{
			name:               "UpdateBadConfig",
			modelName:          "iris",
			modelVersion:       "1",
			config:             []byte(`{"foo":"mys3","type":"s3","parameters":{"foo":"bar"}}`),
			expectedPath:       "/config/update",
			existsBody:         []byte(`{"name":"mys3"}`),
			createUpdateStatus: 200,
			err:                true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpmock.Activate()
			logger := log.New()
			log.SetLevel(log.DebugLevel)
			host := "rclone-server"
			port := 5572
			r := NewRCloneClient(host, port, "/tmp/rclone", logger)
			httpmock.RegisterResponder("POST", fmt.Sprintf("=~http://%s:%d%s", host, port, test.expectedPath),
				httpmock.NewStringResponder(test.createUpdateStatus, "{}"))
			httpmock.RegisterResponder("POST", fmt.Sprintf("=~http://%s:%d/config/get", host, port),
				httpmock.NewStringResponder(200, string(test.existsBody)))
			err := r.Config(test.modelName, test.modelVersion, test.config)
			if !test.err {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
			}
			httpmock.DeactivateAndReset()
		})
	}
}

func TestGetRemoteName(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	type test struct {
		name     string
		uri      string
		expected string
		err      bool
	}
	tests := []test{
		{
			name:     "simple",
			uri:      "s3://models/iris",
			expected: "s3",
			err:      false,
		},
		{
			name: "Fail",
			uri:  "s3//models/iris",
			err:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := getRemoteName(test.uri)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(res).To(Equal(test.expected))
			}

		})
	}
}
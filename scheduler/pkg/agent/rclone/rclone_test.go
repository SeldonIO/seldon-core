/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rclone

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func createTestRCloneMockRespondersBasic(host string, port int, status int, body string) {
	httpmock.RegisterResponder("POST", fmt.Sprintf("=~http://%s:%d/", host, port),
		httpmock.NewStringResponder(status, body))
}

func createTestRCloneMockResponders(host string, port int, status int, body string, createLocalFolder bool) {
	httpmock.RegisterResponder("POST", fmt.Sprintf("=~http://%s:%d/", host, port),
		func(req *http.Request) (*http.Response, error) {
			if status == http.StatusOK && createLocalFolder {
				b, err := io.ReadAll(req.Body)
				if err != nil {
					return nil, err
				}
				rcloneCopy := RcloneCopy{}
				err = json.Unmarshal(b, &rcloneCopy)
				if err != nil {
					return nil, err
				}
				err = os.MkdirAll(rcloneCopy.DstFs, os.ModePerm)
				if err != nil {
					return nil, err
				}
			}
			return httpmock.NewStringResponse(status, body), nil
		})
}

func createFakeRcloneClient(status int, body string) *RCloneClient {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	host := "rclone-server"
	port := 5572
	r := NewRCloneClient(host, port, "/tmp/rclone", logger, "default")
	createTestRCloneMockRespondersBasic(host, port, status, body)
	return r
}

func createFakeRcloneClientForCopy(t *testing.T, status int, body string, createLocalFolder bool) *RCloneClient {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	host := "rclone-server"
	port := 5572
	r := NewRCloneClient(host, port, t.TempDir(), logger, "default")
	createTestRCloneMockResponders(host, port, status, body, createLocalFolder)
	return r
}

func TestRcloneReady(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()
	t.Logf("Started")
	g := NewGomegaWithT(t)
	r := createFakeRcloneClient(200, "{}")
	err := r.Ready()
	g.Expect(err).To(BeNil())
	g.Expect(httpmock.GetTotalCallCount()).To(Equal(1))
}

func TestRcloneCopy(t *testing.T) {
	type test struct {
		name              string
		modelName         string
		createLocalFolder bool
		uri               string
		status            int
		expectError       bool
		body              string
	}

	tests := []test{
		{
			name:              "ok",
			modelName:         "iris",
			uri:               "gs://seldon-models/sklearn/iris-0.23.2/lr_model",
			status:            200,
			body:              "{}",
			createLocalFolder: true,
		},
		{
			name:              "badResponse",
			modelName:         "iris",
			uri:               "gs://seldon-models/sklearn/iris-0.23.2/lr_model",
			status:            400,
			body:              "{}",
			createLocalFolder: true,
			expectError:       true,
		},
		{
			name:              "noFiles",
			modelName:         "iris",
			uri:               "gs://seldon-models/scv2/xyz",
			status:            200,
			body:              "{}",
			createLocalFolder: false,
			expectError:       true,
		},
	}

	t.Logf("Started")
	g := NewGomegaWithT(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			r := createFakeRcloneClientForCopy(t, test.status, test.body, test.createLocalFolder)
			_, err := r.Copy(test.modelName, test.uri, []byte{})

			if !test.expectError {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
			}
			g.Expect(httpmock.GetTotalCallCount()).To(Equal(1))
		})
	}
}

func TestRcloneConfig(t *testing.T) {
	type test struct {
		name               string
		config             []byte
		expectedPath       string
		existsBody         []byte
		createUpdateStatus int
		expectedName       string
		err                bool
	}

	tests := []test{
		{
			name:               "CreateOK",
			config:             []byte(`{"name":"mys3","type":"s3","parameters":{"foo":"bar"}}`),
			expectedPath:       "/config/create",
			existsBody:         []byte(`{}`),
			createUpdateStatus: 200,
			expectedName:       "mys3",
			err:                false,
		},
		{
			name:               "CreateFail",
			config:             []byte(`{"name":"mys3","type":"s3","parameters":{"foo":"bar"}}`),
			expectedPath:       "/config/create",
			existsBody:         []byte(`{}`),
			createUpdateStatus: 400,
			err:                true,
		},
		{
			name:               "CreateBadConfig",
			config:             []byte(`{"foo":"mys3","type":"s3","parameters":{"foo":"bar"}}`),
			expectedPath:       "/config/create",
			existsBody:         []byte(`{}`),
			createUpdateStatus: 200,
			err:                true,
		},
		{
			name:               "Update",
			config:             []byte(`{"name":"mys3","type":"s3","parameters":{"foo":"bar"}}`),
			expectedPath:       "/config/update",
			existsBody:         []byte(`{"name":"mys3"}`),
			createUpdateStatus: 200,
			expectedName:       "mys3",
			err:                false,
		},
		{
			name:               "UpdateBadConfig",
			config:             []byte(`{"foo":"mys3","type":"s3","parameters":{"foo":"bar"}}`),
			expectedPath:       "/config/update",
			existsBody:         []byte(`{"name":"mys3"}`),
			createUpdateStatus: 200,
			err:                true,
		},
	}

	t.Logf("Started")
	g := NewGomegaWithT(t)

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()

			logger := log.New()
			log.SetLevel(log.DebugLevel)
			host := "rclone-server"
			port := 5572
			r := NewRCloneClient(host, port, "/tmp/rclone", logger, "default")

			httpmock.RegisterResponder(
				"POST",
				fmt.Sprintf("=~http://%s:%d%s", host, port, test.expectedPath),
				httpmock.NewStringResponder(test.createUpdateStatus, "{}"),
			)
			httpmock.RegisterResponder(
				"POST",
				fmt.Sprintf("=~http://%s:%d/config/get", host, port),
				httpmock.NewStringResponder(200, string(test.existsBody)),
			)

			name, err := r.Config(test.config)

			if !test.err {
				g.Expect(err).To(BeNil())
				g.Expect(name).To(Equal(test.expectedName))
			} else {
				g.Expect(err).ToNot(BeNil())
			}
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
			name: "BadURIFail",
			uri:  "s3//models/iris",
			err:  true,
		},
		{
			name: "InlineFail",
			uri:  ":,provider=minio,env_auth=false,access_key_id=minioadmin,secret_access_key=minoadmin,endpoint='http://172.18.255.1:9000'://models/iris",
			err:  true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := getRemoteName(test.uri)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(res).To(Equal(test.expected))
			}

		})
	}
}

func TestCreateUriWithConfig(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	type test struct {
		name        string
		uri         string
		config      []byte
		expectedUri string
		err         bool
	}
	tests := []test{
		{
			name:        "simple",
			uri:         "s3://models/iris",
			config:      []byte(`{"name":"s3","type":"s3","parameters":{"foo":"bar"}}`),
			expectedUri: ":s3,foo=bar://models/iris",
			err:         false,
		},
		{
			name:        "ContainsColon",
			uri:         "s3://models/iris",
			config:      []byte(`{"name":"s3","type":"s3","parameters":{"endpoint":"http://minio:9000"}}`),
			expectedUri: `:s3,endpoint="http://minio:9000"://models/iris`,
			err:         false,
		},
		{
			name:        "ContainsComma",
			uri:         "s3://models/iris",
			config:      []byte(`{"name":"s3","type":"s3","parameters":{"foo":"a,b"}}`),
			expectedUri: `:s3,foo="a,b"://models/iris`,
			err:         false,
		},
		{
			name:        "ContainsDoubleQuoteWithColon",
			uri:         "s3://models/iris",
			config:      []byte(`{"name":"s3","type":"s3","parameters":{"foo":"a:\"b"}}`),
			expectedUri: `:s3,foo="a:""b"://models/iris`,
			err:         false,
		},
		{
			name: "BadUri",
			uri:  "s3//models/iris",
			err:  true,
		},
		{
			name:   "BadConfigMissingFields",
			uri:    "s3://models/iris",
			config: []byte(`{"name":"s3"}`),
			err:    true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			log.SetLevel(log.DebugLevel)
			host := "rclone-server"
			port := 5572
			r := NewRCloneClient(host, port, "/tmp/rclone", logger, "default")
			res, err := r.createUriWithConfig(test.uri, test.config)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(res).To(Equal(test.expectedUri))
			}

		})
	}
}

func TestListRemotes(t *testing.T) {
	t.Logf("Started")
	g := NewGomegaWithT(t)
	type test struct {
		name           string
		rcloneResponse *RcloneListRemotes
		expected       []string
		err            bool
	}
	tests := []test{
		{
			name: "simple",
			rcloneResponse: &RcloneListRemotes{
				Remotes: []string{"a", "b"},
			},
			expected: []string{"a", "b"},
		},
		{
			name: "empty",
			rcloneResponse: &RcloneListRemotes{
				Remotes: []string{},
			},
			expected: []string{},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			logger := log.New()
			log.SetLevel(log.DebugLevel)
			host := "rclone-server"
			port := 5572
			r := NewRCloneClient(host, port, "/tmp/rclone", logger, "default")
			b, err := json.Marshal(test.rcloneResponse)
			g.Expect(err).To(BeNil())
			httpmock.RegisterResponder("POST", fmt.Sprintf("=~http://%s:%d%s", host, port, "/config/listremotes"),
				httpmock.NewBytesResponder(200, b))
			remotes, err := r.ListRemotes()
			if !test.err {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).ToNot(BeNil())
				g.Expect(remotes).To(Equal(test.expected))
			}
		})
	}
}

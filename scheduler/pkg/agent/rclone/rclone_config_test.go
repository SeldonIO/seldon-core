package rclone

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/scheduler/pkg/agent/config"
	"github.com/sirupsen/logrus"
)

func TestLoadRcloneConfig(t *testing.T) {
	t.Logf("Started")
	logrus.SetLevel(logrus.DebugLevel)
	g := gomega.NewGomegaWithT(t)

	type test struct {
		name                string
		agentConfiguration  *config.AgentConfiguration
		rcloneListRemotes   *RcloneListRemotes
		rcloneGetResponse   string
		expectedDeleteCalls int
		expectedUpdateCalls int
		expectedCreateCalls int
		error               bool
	}

	tests := []test{
		{
			name: "config",
			agentConfiguration: &config.AgentConfiguration{
				Rclone: &config.RcloneConfiguration{
					Config: []string{`{"type":"google cloud storage","name":"gs","parameters":{"anonymous":true}}`},
				},
			},
			rcloneListRemotes:   &RcloneListRemotes{Remotes: []string{"gs"}},
			expectedDeleteCalls: 0,
			expectedCreateCalls: 1,
			rcloneGetResponse:   "{}",
		},
		{
			name: "multipleCreate",
			agentConfiguration: &config.AgentConfiguration{
				Rclone: &config.RcloneConfiguration{
					Config: []string{
						`{"type":"google cloud storage","name":"gs","parameters":{"anonymous":true}}`,
						`{"type":"google cloud storage","name":"gs2","parameters":{"anonymous":true}}`,
					},
				},
			},
			rcloneListRemotes:   &RcloneListRemotes{Remotes: []string{"gs", "gs2"}},
			expectedDeleteCalls: 0,
			expectedCreateCalls: 2,
			rcloneGetResponse:   "{}",
		},
		{
			name: "multipleUpdate",
			agentConfiguration: &config.AgentConfiguration{
				Rclone: &config.RcloneConfiguration{
					Config: []string{
						`{"type":"google cloud storage","name":"gs","parameters":{"anonymous":true}}`,
						`{"type":"google cloud storage","name":"gs2","parameters":{"anonymous":true}}`,
					},
				},
			},
			rcloneGetResponse:   `{"type":"google cloud storage"}`,
			expectedUpdateCalls: 2,
			rcloneListRemotes:   &RcloneListRemotes{Remotes: []string{"gs", "gs2"}},
			expectedDeleteCalls: 0,
			expectedCreateCalls: 0,
		},
		{
			name: "configDeleted",
			agentConfiguration: &config.AgentConfiguration{
				Rclone: &config.RcloneConfiguration{
					Config: []string{`{"type":"google cloud storage","name":"gs","parameters":{"anonymous":true}}`},
				},
			},
			rcloneListRemotes:   &RcloneListRemotes{Remotes: []string{"gs", "extra"}},
			expectedDeleteCalls: 1,
			expectedCreateCalls: 1,
			rcloneGetResponse:   "{}",
		},
		{
			name: "configUpdated",
			agentConfiguration: &config.AgentConfiguration{
				Rclone: &config.RcloneConfiguration{
					Config: []string{`{"type":"google cloud storage","name":"gs","parameters":{"anonymous":true}}`},
				},
			},
			rcloneGetResponse:   `{"type":"google cloud storage"}`,
			expectedUpdateCalls: 1,
			expectedCreateCalls: 0,
			rcloneListRemotes:   &RcloneListRemotes{Remotes: []string{"gs"}},
			expectedDeleteCalls: 0,
		},
		{
			name: "badConfig",
			agentConfiguration: &config.AgentConfiguration{
				Rclone: &config.RcloneConfiguration{
					Config: []string{`{"foo":"google cloud storage","bar":"gs","parameters":{"anonymous":true}}`},
				},
			},
			error: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			logger := logrus.New()
			logrus.SetLevel(logrus.DebugLevel)
			host := "rclone-server"
			port := 5572
			rcloneClient := NewRCloneClient(host, port, "/tmp/rclone", logger, "default")

			// Add expected Rclone list remotes response
			b, err := json.Marshal(test.rcloneListRemotes)
			g.Expect(err).To(gomega.BeNil())
			deleteURI := fmt.Sprintf("=~http://%s:%d%s", host, port, "/config/delete")
			updateURI := fmt.Sprintf("=~http://%s:%d%s", host, port, "/config/update")
			createURI := fmt.Sprintf("=~http://%s:%d%s", host, port, "/config/create")
			listURI := fmt.Sprintf("=~http://%s:%d%s", host, port, "/config/listremotes")
			getURI := fmt.Sprintf("=~http://%s:%d%s", host, port, "/config/get")
			httpmock.RegisterResponder("POST", listURI,
				httpmock.NewBytesResponder(200, b))
			httpmock.RegisterResponder("POST", deleteURI,
				httpmock.NewStringResponder(200, "{}"))
			httpmock.RegisterResponder("POST", getURI,
				httpmock.NewStringResponder(200, test.rcloneGetResponse))
			httpmock.RegisterResponder("POST", createURI,
				httpmock.NewStringResponder(200, "{}"))
			httpmock.RegisterResponder("POST", updateURI,
				httpmock.NewStringResponder(200, "{}"))

			g.Expect(err).To(gomega.BeNil())
			err = rcloneClient.loadRcloneConfiguration(test.agentConfiguration)
			if test.error {
				g.Expect(err).ToNot(gomega.BeNil())
			} else {
				g.Expect(err).To(gomega.BeNil())
				// Test the expected calls to each endpoint of rclone
				calls := httpmock.GetCallCountInfo()
				for k, v := range calls {
					switch k {
					case deleteURI:
						g.Expect(v).To(gomega.Equal(test.expectedDeleteCalls))
					case updateURI:
						g.Expect(v).To(gomega.Equal(test.expectedUpdateCalls))
					case createURI:
						g.Expect(v).To(gomega.Equal(test.expectedCreateCalls))
					case listURI, getURI:
						g.Expect(v).To(gomega.Equal(len(test.agentConfiguration.Rclone.Config)))
					}

				}
			}
		})
	}
}

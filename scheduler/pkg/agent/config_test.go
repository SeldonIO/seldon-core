package agent

import (
	"encoding/json"
	"os"
	"path"
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func TestLoadConfigRcloneSecrets(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)
	type test struct {
		name     string
		config   string
		expected []string
		err      bool
	}
	tests := []test{
		{
			name:     "yaml",
			config:   `{"rclone":{"config_secrets":["a","b"]}}`,
			expected: []string{"a", "b"},
		},
		{
			name: "json",
			config: `rclone:
                         config_secrets:
                          - a
                          - b`,
			expected: []string{"a", "b"},
		},
		{
			name:     "badJson",
			config:   `{"rclone":{"config_secrets":["a","b"]}`,
			expected: []string{"a", "b"},
			err:      true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configHandler, err := NewAgentConfigHandler("", "", logger)
			g.Expect(err).To(BeNil())
			err = configHandler.updateConfig([]byte(test.config))
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(configHandler.config.Rclone.ConfigSecrets).To(Equal(test.expected))
			}
		})
	}
}

func TestWatchFile(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)
	type test struct {
		name      string
		contents1 *AgentConfiguration
		contents2 *AgentConfiguration
	}
	tests := []test{
		{
			name: "simple",
			contents1: &AgentConfiguration{
				Rclone: &RcloneConfiguration{
					Config: []string{`{"type":"google cloud storage","name":"gs","parameters":{"anonymous":true}}`},
				},
			},
			contents2: &AgentConfiguration{
				Rclone: &RcloneConfiguration{
					Config: []string{`{"type":"google cloud storage","name":"gs2","parameters":{"anonymous":true}}`},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tdir := t.TempDir()
			configFile := path.Join(tdir, "agent.json")
			b, err := json.Marshal(test.contents1)
			g.Expect(err).To(BeNil())
			err = os.WriteFile(configFile, b, 0644)
			g.Expect(err).To(BeNil())
			configHandler, err := NewAgentConfigHandler(tdir, "", logger)
			defer func() { _ = configHandler.Close() }()
			g.Expect(err).To(BeNil())
			g.Expect(configHandler.config).To(Equal(test.contents1))
			b, err = json.Marshal(test.contents2)
			g.Expect(err).To(BeNil())
			err = os.WriteFile(configFile, b, 0644)
			g.Expect(err).To(BeNil())
			g.Eventually(configHandler.getConfiguration).Should(Equal(test.contents2))
		})
	}
}
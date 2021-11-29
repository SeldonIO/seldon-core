package agent

import (
	log "github.com/sirupsen/logrus"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
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
			err = configHandler.updateConfig(strings.NewReader(test.config))
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(configHandler.config.Rclone.ConfigSecrets).To(Equal(test.expected))
			}
		})
	}
}

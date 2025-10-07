/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package config

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
			name:     "test.json",
			config:   `{"rclone":{"config_secrets":["a","b"]}}`,
			expected: []string{"a", "b"},
		},
		{
			name: "test.yaml",
			config: `rclone:
                         config_secrets:
                          - a
                          - b`,
			expected: []string{"a", "b"},
		},
		{
			name:     "badJson.json",
			config:   `{"rclone":{"config_secrets":["a","b"]}`,
			expected: []string{"a", "b"},
			err:      true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configHandler, err := NewAgentConfigHandler("", "", logger, nil)
			g.Expect(err).To(BeNil())
			err = configHandler.UpdateConfig([]byte(test.config), test.name)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(configHandler.GetConfiguration().Rclone.ConfigSecrets).To(Equal(test.expected))
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
			configFile := path.Join(tdir, AgentConfigYamlFilename)
			b, err := json.Marshal(test.contents1)
			g.Expect(err).To(BeNil())
			err = os.WriteFile(configFile, b, 0644)
			g.Expect(err).To(BeNil())
			configHandler, err := NewAgentConfigHandler(tdir, "", logger, nil)
			defer func() { _ = configHandler.Close() }()
			g.Expect(err).To(BeNil())

			g.Expect(configHandler.GetConfiguration()).To(Equal(*test.contents1))

			b, err = json.Marshal(test.contents2)
			g.Expect(err).To(BeNil())
			err = os.WriteFile(configFile, b, 0644)
			g.Expect(err).To(BeNil())
			g.Eventually(configHandler.GetConfiguration).Should(Equal(*test.contents2))
		})
	}
}

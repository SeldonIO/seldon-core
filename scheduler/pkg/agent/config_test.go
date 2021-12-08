package agent

import (
	"context"
	"encoding/json"
	"os"
	"path"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes/fake"
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
			configHandler, err := NewAgentConfigHandler("", "", logger, nil)
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
			configFile := path.Join(tdir, AgentConfigYamlFilename)
			b, err := json.Marshal(test.contents1)
			g.Expect(err).To(BeNil())
			err = os.WriteFile(configFile, b, 0644)
			g.Expect(err).To(BeNil())
			configHandler, err := NewAgentConfigHandler(tdir, "", logger, nil)
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

func TestWatchConfigMap(t *testing.T) {
	t.Logf("Started")
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	g := NewGomegaWithT(t)
	type test struct {
		name              string
		configMapV1       *v1.ConfigMap
		expectedSecretsV1 []string
		configMapV2       *v1.ConfigMap
		expectedSecretsV2 []string
		errorOnInit       bool
	}
	namespace := "seldon-mesh"
	tests := []test{
		{
			name: "simple",
			configMapV1: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: ConfigMapName, Namespace: namespace},
				Data: map[string]string{
					AgentConfigYamlFilename: "{\"rclone\":{\"config_secrets\":[\"rclone-gs-public\"]}}",
				},
			},
			expectedSecretsV1: []string{"rclone-gs-public"},
			configMapV2: &v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{Name: ConfigMapName, Namespace: namespace},
				Data: map[string]string{
					AgentConfigYamlFilename: "{\"rclone\":{\"config_secrets\":[\"rclone-gs-public2\"]}}",
				},
			},
			expectedSecretsV2: []string{"rclone-gs-public2"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeClientset := fake.NewSimpleClientset(test.configMapV1)
			configHandler, err := NewAgentConfigHandler("", namespace, logger, fakeClientset)
			defer func() { _ = configHandler.Close() }()
			getSecrets := func() []string {
				c := configHandler.getConfiguration()
				if c != nil {
					return c.Rclone.ConfigSecrets
				}
				return []string{}
			}
			g.Expect(err).To(BeNil())
			if !test.errorOnInit {
				g.Expect(configHandler.config).ToNot(BeNil())
				g.Eventually(getSecrets).Should(Equal(test.expectedSecretsV1))
				// update
				_, err = fakeClientset.CoreV1().ConfigMaps(namespace).Update(context.Background(), test.configMapV2, metav1.UpdateOptions{})
				g.Expect(err).To(BeNil())
				g.Eventually(getSecrets).Should(Equal(test.expectedSecretsV2))
			} else {
				g.Expect(configHandler.config).To(BeNil())
			}
		})
	}
}

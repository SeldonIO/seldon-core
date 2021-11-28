package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	yaml "gopkg.in/yaml.v2"
)

const (
	AgentConfigYamlFilename = "agent.yaml"
	AgentConfigJsonFilename = "agent.json"
)

type AgentConfiguration struct {
	Rclone *RcloneConfiguration `json:"rclone,omitempty" yaml:"rclone,omitempty"`
}

type RcloneConfiguration struct {
	ConfigSecrets []string `json:"config_secrets,omitempty" yaml:"config_secrets,omitempty"`
	Config        []string `json:"config,omitempty" yaml:"config,omitempty"`
}

type AgentConfigHandler struct {
	config *AgentConfiguration
	mu     sync.RWMutex
}

func NewAgentConfigHandler(configPath string, namespace string) (*AgentConfigHandler, error) {
	var config *AgentConfiguration
	if configPath != "" {
		configFile, err := loadConfigFile(configPath)
		if err != nil {
			return nil, err
		}
		config, err = loadConfig(configFile)
		if err != nil {
			return nil, err
		}
	}
	return &AgentConfigHandler{config: config}, nil
}

func (a *AgentConfigHandler) getConfiguration() *AgentConfiguration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config
}

func loadConfigFile(configPath string) (io.Reader, error) {
	yamConfigPath := configPath + "/" + AgentConfigYamlFilename
	if _, err := os.Stat(yamConfigPath); errors.Is(err, os.ErrNotExist) {
		jsonConfigPath := configPath + "/" + AgentConfigJsonFilename
		if _, err := os.Stat(jsonConfigPath); errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("Failed to find config file as either %s or %s", yamConfigPath, jsonConfigPath)
		}
		return os.Open(jsonConfigPath)
	} else {
		return os.Open(yamConfigPath)
	}
}

func loadConfig(file io.Reader) (*AgentConfiguration, error) {
	configData, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	config := AgentConfiguration{}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		err = json.Unmarshal(configData, &config)
		if err != nil {
			return nil, err
		}
	}
	return &config, nil
}

func (c *AgentConfigHandler) watchFile() {

}

//TODO finish
func loadFromK8s(namespace string) *AgentConfiguration {
	return nil
}
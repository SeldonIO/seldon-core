package agent

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
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
	config    *AgentConfiguration
	mu        sync.RWMutex
	listeners []chan string
	logger    log.FieldLogger
	watcher   *fsnotify.Watcher
	done      chan struct{}
}

func NewAgentConfigHandler(configPath string, namespace string, logger log.FieldLogger) (*AgentConfigHandler, error) {
	if configPath != "" {
		configFilePath, configFile, err := loadConfigFile(configPath)
		if err != nil {
			return nil, err
		}
		configHandler := &AgentConfigHandler{
			logger: logger,
		}
		err = configHandler.updateConfig(configFile)
		if err != nil {
			return nil, err
		}
		err = configHandler.watchFile(configFilePath)
		if err != nil {
			return nil, err
		}
		return configHandler, nil
	}
	return &AgentConfigHandler{
		logger: logger,
	}, nil
}

func (a *AgentConfigHandler) Close() error {
	if a.done != nil {
		close(a.done)
	}
	if a.watcher != nil {
		return a.watcher.Close()
	}
	return nil
}

func (a *AgentConfigHandler) AddListener(c chan string) *AgentConfiguration {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.listeners = append(a.listeners, c)
	return a.config
}

func (a *AgentConfigHandler) getConfiguration() *AgentConfiguration {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.config
}

func loadConfigFile(configPath string) (string, io.Reader, error) {
	yamConfigPath := configPath + "/" + AgentConfigYamlFilename
	if _, err := os.Stat(yamConfigPath); errors.Is(err, os.ErrNotExist) {
		jsonConfigPath := configPath + "/" + AgentConfigJsonFilename
		if _, err := os.Stat(jsonConfigPath); errors.Is(err, os.ErrNotExist) {
			return "", nil, fmt.Errorf("Failed to find config file as either %s or %s", yamConfigPath, jsonConfigPath)
		}
		reader, err := os.Open(jsonConfigPath)
		return jsonConfigPath, reader, err
	} else {
		reader, err := os.Open(yamConfigPath)
		return yamConfigPath, reader, err
	}
}

func (c *AgentConfigHandler) updateConfig(file io.Reader) error {
	c.logger.Info("Updating config")
	c.mu.Lock()
	defer c.mu.Unlock()
	configData, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	config := AgentConfiguration{}
	err = yaml.Unmarshal(configData, &config)
	if err != nil {
		err = json.Unmarshal(configData, &config)
		if err != nil {
			return err
		}
	}
	c.config = &config
	return nil
}

// Watch the config file passed for changes and reload and signal listerners when it does
// TODO could be extended to watch config directories created by K8S on configmap mounts
func (c *AgentConfigHandler) watchFile(filePath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		c.logger.Error(err, "Failed to create watcher")
		return err
	}
	c.watcher = watcher

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				c.logger.Infof("Processing event %v", event)
				isCreate := event.Op&fsnotify.Create != 0
				isWrite := event.Op&fsnotify.Write != 0
				if isCreate || isWrite {
					reader, err := os.Open(filePath)
					if err != nil {
						c.logger.WithError(err).Errorf("Failed to open %s", filePath)
					} else {
						err := c.updateConfig(reader)
						if err != nil {
							c.logger.WithError(err).Errorf("Failed to update config %s", filePath)
						} else {
							for _, ch := range c.listeners {
								ch <- "updated"
							}
						}
					}
				}
			case err := <-watcher.Errors:
				c.logger.Error(err, "watcher error")
			case <-c.done:
				return
			}
		}
	}()

	if err = watcher.Add(filePath); err != nil {
		c.logger.Errorf("Failed add filePath %s to watcher", filePath)
		return err
	}
	c.logger.Infof("Start to watch config file %s", filePath)

	return nil
}

//TODO finish
func loadFromK8s(namespace string) *AgentConfiguration {
	return nil
}
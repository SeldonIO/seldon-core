package agent

import (
	"encoding/json"
	"os"
	"path"
	"sync"

	"k8s.io/client-go/kubernetes"

	corev1 "k8s.io/api/core/v1"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/configmap/informer"

	log "github.com/sirupsen/logrus"

	"github.com/fsnotify/fsnotify"
	yaml "gopkg.in/yaml.v2"
)

const (
	AgentConfigYamlFilename = "agent.yaml"
	AgentConfigJsonFilename = "agent.json"
	ConfigMapName           = "seldon-agent"
)

type AgentConfiguration struct {
	Rclone *RcloneConfiguration `json:"rclone,omitempty" yaml:"rclone,omitempty"`
}

type RcloneConfiguration struct {
	ConfigSecrets []string `json:"config_secrets,omitempty" yaml:"config_secrets,omitempty"`
	Config        []string `json:"config,omitempty" yaml:"config,omitempty"`
}

type AgentConfigHandler struct {
	config               *AgentConfiguration
	mu                   sync.RWMutex
	listeners            []chan string
	logger               log.FieldLogger
	watcher              *fsnotify.Watcher
	fileWatcherDone      chan struct{}
	namespace            string
	configFilePath       string
	configMapWatcherDone chan struct{}
}

func NewAgentConfigHandler(configPath string, namespace string, logger log.FieldLogger, clientset kubernetes.Interface) (*AgentConfigHandler, error) {
	configHandler := &AgentConfigHandler{
		logger:    logger,
		namespace: namespace,
	}
	if configPath != "" {
		m, err := configmap.Load(configPath)
		if err != nil {
			return nil, err
		}
		if v, ok := m[AgentConfigYamlFilename]; ok {
			err = configHandler.updateConfig([]byte(v))
			if err != nil {
				return nil, err
			}
			configHandler.configFilePath = path.Join(configPath, AgentConfigYamlFilename)
		} else if v, ok := m[AgentConfigJsonFilename]; ok {
			err = configHandler.updateConfig([]byte(v))
			if err != nil {
				return nil, err
			}
			configHandler.configFilePath = path.Join(configPath, AgentConfigJsonFilename)
		}
	}

	if namespace != "" && clientset != nil { // Running in k8s
		err := configHandler.watchConfigMap(clientset)
		if err != nil {
			return nil, err
		}
	} else if configPath != "" { // Watch local file
		err := configHandler.watchFile(configHandler.configFilePath)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Warnf("No config available on initialization")
	}

	return configHandler, nil
}

func (a *AgentConfigHandler) Close() error {
	if a.fileWatcherDone != nil {
		close(a.fileWatcherDone)
	}
	if a.configMapWatcherDone != nil {
		close(a.configMapWatcherDone)
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
	a.logger.Infof("get configuration call")
	return a.config
}

func (c *AgentConfigHandler) updateConfig(configData []byte) error {
	c.logger.Infof("Updating config %s", configData)
	c.mu.Lock()
	defer c.mu.Unlock()
	config := AgentConfiguration{}
	err := yaml.Unmarshal(configData, &config)
	if err != nil {
		err = json.Unmarshal(configData, &config)
		if err != nil {
			return err
		}
	}
	c.config = &config
	return nil
}

// Watch the config file passed for changes and reload and signal listeners when it does
func (c *AgentConfigHandler) watchFile(filePath string) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		c.logger.Error(err, "Failed to create watcher")
		return err
	}
	c.watcher = watcher
	c.fileWatcherDone = make(chan struct{})

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				c.logger.Infof("Processing event %v", event)
				isCreate := event.Op&fsnotify.Create != 0
				isWrite := event.Op&fsnotify.Write != 0
				if isCreate || isWrite {
					b, err := os.ReadFile(filePath)
					if err != nil {
						c.logger.WithError(err).Errorf("Failed to read %s", filePath)
					} else {
						err := c.updateConfig(b)
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
			case <-c.fileWatcherDone:
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

func (a *AgentConfigHandler) watchConfigMap(clientset kubernetes.Interface) error {
	logger := a.logger.WithField("func", "watchConfigMap")

	watcher := informer.NewInformedWatcher(clientset, a.namespace)
	watcher.Watch(ConfigMapName, func(updated *corev1.ConfigMap) {
		if data, ok := updated.Data[AgentConfigYamlFilename]; ok {
			err := a.updateConfig([]byte(data))
			if err != nil {
				logger.Errorf("Failed to update configmap from data in %s", AgentConfigYamlFilename)
			}
		} else if data, ok := updated.Data[AgentConfigJsonFilename]; ok {
			err := a.updateConfig([]byte(data))
			if err != nil {
				logger.Errorf("Failed to update configmap from data in %s", AgentConfigJsonFilename)
			}
		}
	})
	a.configMapWatcherDone = make(chan struct{})
	err := watcher.Start(a.configMapWatcherDone)
	if err != nil {
		return err
	}
	return nil
}

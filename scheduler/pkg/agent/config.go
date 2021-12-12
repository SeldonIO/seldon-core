package agent

import (
	"fmt"
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
	logger               log.FieldLogger
	mu                   sync.RWMutex
	config               *AgentConfiguration
	listeners            []chan<- AgentConfiguration
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
		err := configHandler.initConfigFromPath(configPath)
		if err != nil {
			return nil, err
		}
	}

	err := configHandler.initWatcher(configPath, namespace, clientset)
	if err != nil {
		return nil, err
	}

	return configHandler, nil
}

func (a *AgentConfigHandler) initConfigFromPath(configPath string) error {
	m, err := configmap.Load(configPath)
	if err != nil {
		return err
	}

	if v, ok := m[AgentConfigYamlFilename]; ok {
		err = a.updateConfig([]byte(v))
		if err != nil {
			return err
		}
		a.configFilePath = path.Join(configPath, AgentConfigYamlFilename)
		return nil
	}
	return fmt.Errorf("Failed to find config file %s", AgentConfigYamlFilename)
}

func (a *AgentConfigHandler) initWatcher(configPath string, namespace string, clientset kubernetes.Interface) error {
	logger := a.logger.WithField("func", "initWatcher")
	if namespace != "" { // Running in k8s
		err := a.watchConfigMap(clientset)
		if err != nil {
			return err
		}
	} else if configPath != "" { // Watch local file
		err := a.watchFile(a.configFilePath)
		if err != nil {
			return err
		}
	} else {
		logger.Warnf("No config available on initialization")
	}
	return nil
}

func (a *AgentConfigHandler) Close() error {
	if a == nil {
		return nil
	}
	if a.fileWatcherDone != nil {
		close(a.fileWatcherDone)
	}
	if a.configMapWatcherDone != nil {
		close(a.configMapWatcherDone)
	}
	if a.watcher != nil {
		return a.watcher.Close()
	}
	for _, c := range a.listeners {
		close(c)
	}
	return nil
}

func (a *AgentConfigHandler) AddListener(c chan AgentConfiguration) *AgentConfiguration {
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

func (a *AgentConfigHandler) updateConfig(configData []byte) error {
	logger := a.logger.WithField("func", "updateConfig")
	logger.Infof("Updating config %s", configData)
	a.mu.Lock()
	defer a.mu.Unlock()
	config := AgentConfiguration{}
	err := yaml.Unmarshal(configData, &config)
	if err != nil {
		return err
	}
	if config.Rclone != nil {
		logger.Infof("Rclone Config loaded %v", config.Rclone)
	}
	a.config = &config
	return nil
}

// Watch the config file passed for changes and reload and signal listeners when it does
func (a *AgentConfigHandler) watchFile(filePath string) error {
	logger := a.logger.WithField("func", "watchFile")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error(err, "Failed to create watcher")
		return err
	}
	a.watcher = watcher
	a.fileWatcherDone = make(chan struct{})

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				logger.Infof("Processing event %v", event)
				isCreate := event.Op&fsnotify.Create != 0
				isWrite := event.Op&fsnotify.Write != 0
				if isCreate || isWrite {
					b, err := os.ReadFile(filePath)
					if err != nil {
						logger.WithError(err).Errorf("Failed to read %s", filePath)
					} else {
						err := a.updateConfig(b)
						if err != nil {
							logger.WithError(err).Errorf("Failed to update config %s", filePath)
						} else {
							a.mu.RLock()
							for _, ch := range a.listeners {
								ch <- *a.config
							}
							a.mu.RUnlock()
						}
					}
				}
			case err := <-watcher.Errors:
				logger.Error(err, "watcher error")
			case <-a.fileWatcherDone:
				return
			}
		}
	}()

	if err = watcher.Add(filePath); err != nil {
		a.logger.Errorf("Failed add filePath %s to watcher", filePath)
		return err
	}
	a.logger.Infof("Start to watch config file %s", filePath)

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
			} else {
				a.mu.RLock()
				for _, ch := range a.listeners {
					ch <- *a.config
				}
				a.mu.RUnlock()
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

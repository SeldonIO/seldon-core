/*
Copyright (c) 2025 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"knative.dev/pkg/configmap"
	"knative.dev/pkg/configmap/informer"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"
)

type ConfigUpdateProcessor[T any, PT util.ConfigHandle[T]] func(config PT, logger log.FieldLogger) error

type ConfigWatcher[T any, PT util.ConfigHandle[T]] struct {
	logger                log.FieldLogger
	mu                    sync.RWMutex
	config                PT
	configFilePath        string
	listeners             []chan<- T
	namespace             string
	useConfigMapInformer  bool
	configMapName         string
	configMapFileName     string
	watcher               *fsnotify.Watcher
	fileWatcherDone       chan struct{}
	configMapWatcherDone  chan struct{}
	configUpdateProcessor ConfigUpdateProcessor[T, PT]
}

func NewConfigWatcher[T any, PT util.ConfigHandle[T]](configPath string, configMapFileName string, namespace string, watchK8sConfigMap bool, configMapName string, clientset kubernetes.Interface, configUpdateProcessor ConfigUpdateProcessor[T, PT], logger log.FieldLogger) (*ConfigWatcher[T, PT], error) {
	configHandler := &ConfigWatcher[T, PT]{
		logger:                logger.WithField("source", "ConfigWatcher"),
		namespace:             namespace,
		useConfigMapInformer:  watchK8sConfigMap,
		configMapName:         configMapName,
		configMapFileName:     configMapFileName,
		configUpdateProcessor: configUpdateProcessor,
	}

	if configPath != "" {
		isDir, err := configPathisDir(configPath)
		if err != nil {
			return nil, fmt.Errorf("config path %s: %w", configPath, err)
		}
		if isDir {
			configPath = path.Join(configPath, configMapFileName)
		}
		logger.Infof("Init config from path %s", configPath)
		err = configHandler.initConfigFromPath(configPath)
		if err != nil {
			return nil, err
		}
		_, filename := path.Split(configPath)
		if filename != "" && filename != configMapFileName {
			logger.Warnf("Watched local config file name %s does not match config map file name %s. This means the config may get updates from two different sources", filename, configMapFileName)
		}
	}

	err := configHandler.initWatcher(configPath, clientset)
	if err != nil {
		return nil, err
	}

	return configHandler, nil
}

func configPathisDir(configPath string) (bool, error) {
	fileInfo, err := os.Stat(configPath)
	if err != nil {
		return false, err
	}
	return fileInfo.IsDir(), nil
}

func (cw *ConfigWatcher[T, PT]) initConfigFromPath(configPath string) error {
	m, err := configmap.Load(path.Dir(configPath))
	if err != nil {
		return err
	}

	_, configFileName := path.Split(configPath)
	if v, ok := m[configFileName]; ok {
		err = cw.UpdateConfig([]byte(v), configFileName)
		if err != nil {
			return err
		}
		cw.configFilePath = path.Clean(configPath)
		return nil
	}
	return fmt.Errorf("configuration watcher failed to find file %s", configPath)
}

func (cw *ConfigWatcher[T, PT]) initWatcher(configPath string, clientset kubernetes.Interface) error {
	logger := cw.logger.WithField("func", "initWatcher")
	if cw.useConfigMapInformer && clientset != nil { // Watch k8s config map
		err := cw.watchConfigMap(clientset)
		if err != nil {
			return err
		}
	} else if configPath != "" { // Watch local file
		err := cw.watchFile(cw.configFilePath)
		if err != nil {
			return err
		}
	} else {
		logger.Warnf("No config available on initialization")
	}
	return nil
}

func (cw *ConfigWatcher[T, PT]) Close() error {
	if cw.fileWatcherDone != nil {
		close(cw.fileWatcherDone)
	}
	if cw.configMapWatcherDone != nil {
		close(cw.configMapWatcherDone)
	}
	var err error
	if cw.watcher != nil {
		err = cw.watcher.Close()
	}
	for _, c := range cw.listeners {
		close(c)
	}
	return err
}

func (cw *ConfigWatcher[T, PT]) AddListener(c chan<- T) {
	cw.mu.Lock()
	defer cw.mu.Unlock()
	cw.listeners = append(cw.listeners, c)
}

func (cw *ConfigWatcher[T, PT]) GetConfiguration() T {
	cw.mu.RLock()
	defer cw.mu.RUnlock()
	if cw.config != nil {
		return cw.config.DeepCopy()
	} else {
		return cw.config.Default()
	}
}

func (cw *ConfigWatcher[T, PT]) UpdateConfig(configData []byte, filename string) error {
	logger := cw.logger.WithField("func", "updateConfig")

	cw.mu.Lock()
	defer cw.mu.Unlock()

	config := new(T)
	canonicalExt := strings.Trim(strings.ToLower(path.Ext(filename)), " ")
	if canonicalExt == ".yaml" {
		err := yaml.Unmarshal(configData, &config)
		if err != nil {
			return err
		}
	} else {
		// assume json if not yaml, irrespective of file extension
		err := json.Unmarshal(configData, &config)
		if err != nil {
			return err
		}
	}

	// The config update processor is passed a pointer to the config so that it can validate and
	// modify it as needed based on application logic. Any changes to the config are made while
	// holding the config watcher (write) lock.
	if cw.configUpdateProcessor != nil {
		err := cw.configUpdateProcessor(config, logger)
		if err != nil {
			return err
		}
	}

	cw.config = config
	return nil
}

// Watch the config file passed for changes, reload and signal listeners when it does
func (cw *ConfigWatcher[T, PT]) watchFile(filePath string) error {
	logger := cw.logger.WithField("func", "watchFile")
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logger.Error(err, "Failed to create watcher")
		return err
	}
	cw.watcher = watcher
	cw.fileWatcherDone = make(chan struct{})

	configDir, _ := filepath.Split(filePath)
	knownConfigFile, _ := filepath.EvalSymlinks(filePath)

	go func() {
		for {
			select {
			case event := <-watcher.Events:
				isCreate := event.Op&fsnotify.Create != 0
				isWrite := event.Op&fsnotify.Write != 0
				isRemove := event.Op&fsnotify.Remove != 0

				// when running in k8s, the file is a symlink that gets replaced on update
				currentConfigFile, _ := filepath.EvalSymlinks(filePath)

				existingFileChanged := filepath.Clean(event.Name) == filePath && (isWrite || isCreate)
				configSymlinkChanged := currentConfigFile != "" && currentConfigFile != knownConfigFile

				if existingFileChanged || configSymlinkChanged {
					knownConfigFile = currentConfigFile
					b, err := os.ReadFile(filePath)
					if err != nil {
						logger.WithError(err).Errorf("Failed to read %s", filePath)
					} else {
						err := cw.UpdateConfig(b, filePath)
						if err != nil {
							logger.WithError(err).Errorf("Failed to update config %s", filePath)
						} else {
							cw.mu.RLock()
							for _, ch := range cw.listeners {
								ch <- cw.config.DeepCopy()
							}
							cw.mu.RUnlock()
						}
					}
				} else if filepath.Clean(event.Name) == filePath && isRemove {
					return
				}
			case err := <-watcher.Errors:
				logger.Error(err, "watcher error")
			case <-cw.fileWatcherDone:
				return
			}
		}
	}()

	if err = watcher.Add(configDir); err != nil {
		cw.logger.Errorf("Failed to add file path %s to config watcher", filePath)
		return err
	}
	cw.logger.Infof("Starting to watch config file %s", filePath)

	return nil
}

func (cw *ConfigWatcher[T, PT]) watchConfigMap(clientset kubernetes.Interface) error {
	logger := cw.logger.WithField("func", "watchConfigMap")

	watcher := informer.NewInformedWatcher(clientset, cw.namespace)
	watcher.Watch(cw.configMapName, func(updated *corev1.ConfigMap) {
		filename := cw.configMapFileName
		if data, ok := updated.Data[filename]; ok {
			err := cw.UpdateConfig([]byte(data), cw.configMapName)
			if err != nil {
				logger.Errorf("Failed to update config with data in configmap %s.%s/%s", cw.configMapName, cw.namespace, filename)
			} else {
				cw.mu.RLock()
				for _, ch := range cw.listeners {
					ch <- cw.config.DeepCopy()
				}
				cw.mu.RUnlock()
			}
		}
	})
	cw.configMapWatcherDone = make(chan struct{})
	err := watcher.Start(cw.configMapWatcherDone)
	if err != nil {
		return err
	}
	return nil
}

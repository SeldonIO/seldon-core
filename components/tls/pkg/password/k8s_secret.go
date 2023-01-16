/*
Copyright 2023 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package password

import (
	"context"
	"fmt"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/k8s"
	"k8s.io/client-go/kubernetes"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

type PasswordSecretHandler struct {
	clientset     kubernetes.Interface
	secretName    string
	namespace     string
	stopper       chan struct{}
	logger        log.FieldLogger
	password      string
	folderHandler *PasswordFolderHandler
	mu            sync.RWMutex
}

func NewPasswordSecretHandler(secretName string, clientset kubernetes.Interface, namespace string, prefix string, locationSuffix string, logger log.FieldLogger) (*PasswordSecretHandler, error) {
	if clientset == nil {
		var err error
		clientset, err = k8s.CreateClientset()
		if err != nil {
			logger.WithError(err).Error("Failed to create clientset for password secret handler")
			return nil, err
		}
	}
	folderHandler, err := NewPasswordFolderHandler(prefix, locationSuffix, logger)
	if err != nil {
		return nil, err
	}
	return &PasswordSecretHandler{
		clientset:     clientset,
		secretName:    secretName,
		namespace:     namespace,
		stopper:       make(chan struct{}),
		logger:        logger,
		folderHandler: folderHandler,
	}, nil
}

func (s *PasswordSecretHandler) GetPassword() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.password
}

func (s *PasswordSecretHandler) Stop() {
	close(s.stopper)
}

func (s *PasswordSecretHandler) savePasswordFromSecret(secret *corev1.Secret) error {
	dataKey := filepath.Base(s.folderHandler.passwordFilePath)
	password, ok := secret.Data[dataKey]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", dataKey, secret.Name)
	}
	s.password = string(password)
	return nil
}

func (s *PasswordSecretHandler) onAdd(obj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logger := s.logger.WithField("func", "onAdd")
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("TLS Secret %s added", s.secretName)
		err := s.savePasswordFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract password from secret %s", secret.Name)
		}
	}
}

func (s *PasswordSecretHandler) onUpdate(oldObj, newObj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logger := s.logger.WithField("func", "onUpdate")
	secret := newObj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("TLS Secret %s updated", s.secretName)
		err := s.savePasswordFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract password from secret %s", secret.Name)
		}
	}
}

func (s *PasswordSecretHandler) onDelete(obj interface{}) {
	logger := s.logger.WithField("func", "onDelete")
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Warnf("Secret %s deleted", secret.Name)
	}
}

func (s *PasswordSecretHandler) loadPassword(secretName string) error {
	secret, err := s.clientset.CoreV1().Secrets(s.namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return s.savePasswordFromSecret(secret)
}

func (s *PasswordSecretHandler) GetPasswordAndWatch() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.loadPassword(s.secretName)
	if err != nil {
		return err
	}
	coreInformers := informers.NewSharedInformerFactoryWithOptions(s.clientset, 0, informers.WithNamespace(s.namespace))
	coreInformers.Core().V1().Secrets().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.onAdd,
		UpdateFunc: s.onUpdate,
		DeleteFunc: s.onDelete,
	})
	coreInformers.WaitForCacheSync(s.stopper)
	coreInformers.Start(s.stopper)
	return nil
}

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

package oauth

import (
	"context"
	"fmt"
	"sync"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/seldonio/seldon-core/components/tls/v2/pkg/k8s"
)

type OAUTHSecretHandler struct {
	clientset   kubernetes.Interface
	secretName  string
	namespace   string
	stopper     chan struct{}
	logger      log.FieldLogger
	mu          sync.RWMutex
	oauthConfig OAUTHConfig
}

func NewOAUTHSecretHandler(secretName string, clientset kubernetes.Interface, namespace string, prefix string, locationSuffix string, logger log.FieldLogger) (*OAUTHSecretHandler, error) {
	if clientset == nil {
		var err error
		clientset, err = k8s.CreateClientset()
		if err != nil {
			logger.WithError(err).Error("Failed to create clientset for OAUTH secret handler")
			return nil, err
		}
	}
	return &OAUTHSecretHandler{
		clientset:  clientset,
		secretName: secretName,
		namespace:  namespace,
		stopper:    make(chan struct{}),
		logger:     logger,
	}, nil
}

func (s *OAUTHSecretHandler) GetOAUTHConfig() OAUTHConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.oauthConfig
}

func (s *OAUTHSecretHandler) Stop() {
	close(s.stopper)
}

func (s *OAUTHSecretHandler) saveOAUTHFromSecret(secret *corev1.Secret) error {
	// Read and Save oauthbearer method
	method, ok := secret.Data[methodKey]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", methodKey, secret.Name)
	}
	s.oauthConfig.Method = string(method)

	// Read and Save oauthbearer client id
	clientID, ok := secret.Data[clientIDKey]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", clientIDKey, secret.Name)
	}
	s.oauthConfig.ClientID = string(clientID)

	// Read and Save oauthbearer client secret
	clientSecret, ok := secret.Data[clientSecretKey]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", clientSecretKey, secret.Name)
	}
	s.oauthConfig.ClientSecret = string(clientSecret)

	// Read and Save oauthbearer token endpoint url
	tokenEndpointURL, ok := secret.Data[tokenEndpointURLKey]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", tokenEndpointURLKey, secret.Name)
	}
	s.oauthConfig.TokenEndpointURL = string(tokenEndpointURL)

	// Read and Save oauthbearer extensions
	extensions, ok := secret.Data[extensionsKey]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", extensionsKey, secret.Name)
	}
	s.oauthConfig.Extensions = string(extensions)

	return nil
}

func (s *OAUTHSecretHandler) onAdd(obj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logger := s.logger.WithField("func", "onAdd")
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("OAUTH Secret %s added", s.secretName)
		err := s.saveOAUTHFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract OAUTH from secret %s", secret.Name)
		}
	}
}

func (s *OAUTHSecretHandler) onUpdate(oldObj, newObj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logger := s.logger.WithField("func", "onUpdate")
	secret := newObj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("OAUTH Secret %s updated", s.secretName)
		err := s.saveOAUTHFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract OAUTH from secret %s", secret.Name)
		}
	}
}

func (s *OAUTHSecretHandler) onDelete(obj interface{}) {
	logger := s.logger.WithField("func", "onDelete")
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Warnf("Secret %s deleted", secret.Name)
	}
}

func (s *OAUTHSecretHandler) loadOAUTH(secretName string) error {
	secret, err := s.clientset.CoreV1().Secrets(s.namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return s.saveOAUTHFromSecret(secret)
}

func (s *OAUTHSecretHandler) GetOAUTHAndWatch() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.loadOAUTH(s.secretName)
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

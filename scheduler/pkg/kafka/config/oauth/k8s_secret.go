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

const (
	fieldMethod       = "method"
	fieldClientID     = "client_id"
	fieldClientSecret = "client_secret"
	fieldScope        = "scope"
	fieldTokenURL     = "token_endpoint_url"
	fieldExtensions   = "extensions"
)

type OAuthSecretHandler struct {
	clientset   kubernetes.Interface
	secretName  string
	namespace   string
	stopper     chan struct{}
	logger      log.FieldLogger
	mu          sync.RWMutex
	oauthConfig OAuthConfig
}

func NewOAuthSecretHandler(secretName string, clientset kubernetes.Interface, namespace string, prefix string, logger log.FieldLogger) (*OAuthSecretHandler, error) {
	if clientset == nil {
		var err error
		clientset, err = k8s.CreateClientset()
		if err != nil {
			logger.WithError(err).Error("Failed to create clientset for OAuth secret handler")
			return nil, err
		}
	}
	return &OAuthSecretHandler{
		clientset:  clientset,
		secretName: secretName,
		namespace:  namespace,
		stopper:    make(chan struct{}),
		logger:     logger,
	}, nil
}

func (s *OAuthSecretHandler) GetOAuthConfig() OAuthConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.oauthConfig
}

func (s *OAuthSecretHandler) Stop() {
	close(s.stopper)
}

func (s *OAuthSecretHandler) saveOAuthFromSecret(secret *corev1.Secret) error {
	// Read and Save oauthbearer method
	method, ok := secret.Data[fieldMethod]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", fieldMethod, secret.Name)
	}
	s.oauthConfig.Method = string(method)

	// Read and Save oauthbearer client id
	clientID, ok := secret.Data[fieldClientID]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", fieldClientID, secret.Name)
	}
	s.oauthConfig.ClientID = string(clientID)

	// Read and Save oauthbearer client secret
	clientSecret, ok := secret.Data[fieldClientSecret]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", fieldClientSecret, secret.Name)
	}
	s.oauthConfig.ClientSecret = string(clientSecret)

	// Read and Save oauthbearer scope
	scope, ok := secret.Data[fieldScope]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", fieldScope, secret.Name)
	}
	s.oauthConfig.Scope = string(scope)

	// Read and Save oauthbearer token endpoint url
	tokenEndpointURL, ok := secret.Data[fieldTokenURL]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", fieldTokenURL, secret.Name)
	}
	s.oauthConfig.TokenEndpointURL = string(tokenEndpointURL)

	// Read and Save oauthbearer extensions
	extensions, ok := secret.Data[fieldExtensions]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", fieldExtensions, secret.Name)
	}
	s.oauthConfig.Extensions = string(extensions)

	return nil
}

func (s *OAuthSecretHandler) onAdd(obj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logger := s.logger.WithField("func", "onAdd")
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("OAuth Secret %s added", s.secretName)
		err := s.saveOAuthFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract OAuth from secret %s", secret.Name)
		}
	}
}

func (s *OAuthSecretHandler) onUpdate(oldObj, newObj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logger := s.logger.WithField("func", "onUpdate")
	secret := newObj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("OAuth Secret %s updated", s.secretName)
		err := s.saveOAuthFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract OAuth from secret %s", secret.Name)
		}
	}
}

func (s *OAuthSecretHandler) onDelete(obj interface{}) {
	logger := s.logger.WithField("func", "onDelete")
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Warnf("Secret %s deleted", secret.Name)
	}
}

func (s *OAuthSecretHandler) loadOAuth(secretName string) error {
	secret, err := s.clientset.CoreV1().Secrets(s.namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return s.saveOAuthFromSecret(secret)
}

func (s *OAuthSecretHandler) GetOAuthAndWatch() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.loadOAuth(s.secretName)
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

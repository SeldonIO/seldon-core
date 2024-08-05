/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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

type k8sSecretStore struct {
	clientset   kubernetes.Interface
	secretName  string
	namespace   string
	stopper     chan struct{}
	logger      log.FieldLogger
	mu          sync.RWMutex
	oauthConfig OAuthConfig
}

func newK8sSecretStore(
	secretName string,
	clientset kubernetes.Interface,
	namespace string,
	prefix string,
	logger log.FieldLogger,
) (*k8sSecretStore, error) {
	logger = logger.WithField("source", "SecretOAuthStore")

	if clientset == nil {
		var err error
		clientset, err = k8s.CreateClientset()
		if err != nil {
			logger.WithError(err).Error("Failed to create clientset for OAuth secret handler")
			return nil, err
		}
	}

	return &k8sSecretStore{
		clientset:  clientset,
		secretName: secretName,
		namespace:  namespace,
		stopper:    make(chan struct{}),
		logger:     logger,
	}, nil
}

func (s *k8sSecretStore) GetOAuthConfig() OAuthConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.oauthConfig
}

func (s *k8sSecretStore) Stop() {
	close(s.stopper)
}

func (s *k8sSecretStore) updateFromSecret(secret *corev1.Secret) error {
	newConfig, err := s.getConfigFromSecret(secret)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.oauthConfig = *newConfig

	return nil
}

func (s *k8sSecretStore) getConfigFromSecret(secret *corev1.Secret) (*OAuthConfig, error) {
	config := &OAuthConfig{}
	noSuchFieldError := func(fieldName string) error {
		return fmt.Errorf("Failed to find field %s in secret %s", fieldName, secret.Name)
	}

	method, ok := secret.Data[fieldMethod]
	if !ok {
		return nil, noSuchFieldError(fieldMethod)
	}
	config.Method = string(method)

	clientID, ok := secret.Data[fieldClientID]
	if !ok {
		return nil, noSuchFieldError(fieldClientID)
	}
	config.ClientID = string(clientID)

	clientSecret, ok := secret.Data[fieldClientSecret]
	if !ok {
		return nil, noSuchFieldError(fieldClientSecret)
	}
	config.ClientSecret = string(clientSecret)

	scope, ok := secret.Data[fieldScope]
	if !ok {
		return nil, noSuchFieldError(fieldScope)
	}
	config.Scope = string(scope)

	tokenEndpointURL, ok := secret.Data[fieldTokenURL]
	if !ok {
		return nil, noSuchFieldError(fieldTokenURL)
	}
	config.TokenEndpointURL = string(tokenEndpointURL)

	extensions, ok := secret.Data[fieldExtensions]
	if !ok {
		return nil, noSuchFieldError(fieldExtensions)
	}
	config.Extensions = string(extensions)

	return config, nil
}

func (s *k8sSecretStore) onAdd(obj interface{}) {
	logger := s.logger.WithField("func", "onAdd")
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("OAuth secret %s added", s.secretName)

		err := s.updateFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract OAuth config from secret %s", secret.Name)
		}
	}
}

func (s *k8sSecretStore) onUpdate(oldObj, newObj interface{}) {
	logger := s.logger.WithField("func", "onUpdate")
	secret := newObj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("OAuth secret %s updated", s.secretName)

		err := s.updateFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract OAuth config from secret %s", secret.Name)
		}
	}
}

func (s *k8sSecretStore) onDelete(obj interface{}) {
	logger := s.logger.WithField("func", "onDelete")
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Warnf("OAuth secret %s deleted", secret.Name)
	}
}

func (s *k8sSecretStore) loadOAuthConfig(secretName string) error {
	secret, err := s.clientset.
		CoreV1().
		Secrets(s.namespace).
		Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return s.updateFromSecret(secret)
}

func (s *k8sSecretStore) loadAndWatchConfig() error {
	err := s.loadOAuthConfig(s.secretName)
	if err != nil {
		return err
	}

	coreInformers := informers.NewSharedInformerFactoryWithOptions(
		s.clientset,
		0,
		informers.WithNamespace(s.namespace),
	)
	_, err = coreInformers.Core().V1().Secrets().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    s.onAdd,
			UpdateFunc: s.onUpdate,
			DeleteFunc: s.onDelete,
		},
	)
	if err != nil {
		return err
	}
	coreInformers.WaitForCacheSync(s.stopper)
	coreInformers.Start(s.stopper)

	return nil
}

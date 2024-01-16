/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package password

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"

	"github.com/seldonio/seldon-core/components/tls/v2/pkg/k8s"
)

type k8sSecretStore struct {
	clientset     kubernetes.Interface
	secretName    string
	namespace     string
	stopper       chan struct{}
	logger        log.FieldLogger
	password      string
	folderHandler *fileStore
	mu            sync.RWMutex
}

func newK8sSecretStore(
	secretName string,
	clientset kubernetes.Interface,
	namespace string,
	prefix string,
	locationSuffix string,
	logger log.FieldLogger,
) (*k8sSecretStore, error) {
	logger = logger.WithField("source", "SecretPasswordStore")

	if clientset == nil {
		var err error
		clientset, err = k8s.CreateClientset()
		if err != nil {
			logger.WithError(err).Error("Failed to create clientset for password secret handler")
			return nil, err
		}
	}
	folderHandler, err := newFileStore(prefix, locationSuffix, logger)
	if err != nil {
		return nil, err
	}
	return &k8sSecretStore{
		clientset:     clientset,
		secretName:    secretName,
		namespace:     namespace,
		stopper:       make(chan struct{}),
		logger:        logger,
		folderHandler: folderHandler,
	}, nil
}

func (s *k8sSecretStore) GetPassword() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.password
}

func (s *k8sSecretStore) Stop() {
	close(s.stopper)
}

func (s *k8sSecretStore) updateFromSecret(secret *corev1.Secret) error {
	dataKey := filepath.Base(s.folderHandler.passwordFilePath)
	password, ok := secret.Data[dataKey]
	if !ok {
		return fmt.Errorf("Failed to find %s in secret %s", dataKey, secret.Name)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.password = string(password)

	return nil
}

func (s *k8sSecretStore) onAdd(obj interface{}) {
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger := s.logger.WithField("func", "onAdd")
		logger.Infof("Password secret %s added", s.secretName)

		err := s.updateFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract password from secret %s", secret.Name)
		}
	}
}

func (s *k8sSecretStore) onUpdate(oldObj, newObj interface{}) {
	secret := newObj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger := s.logger.WithField("func", "onUpdate")
		logger.Infof("Password secret %s updated", s.secretName)

		err := s.updateFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract password from secret %s", secret.Name)
		}
	}
}

func (s *k8sSecretStore) onDelete(obj interface{}) {
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger := s.logger.WithField("func", "onDelete")
		logger.Warnf("Password secret %s deleted", secret.Name)
	}
}

func (s *k8sSecretStore) loadPassword(secretName string) error {
	secret, err := s.clientset.
		CoreV1().
		Secrets(s.namespace).
		Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	return s.updateFromSecret(secret)
}

func (s *k8sSecretStore) loadAndWatchPassword() error {
	err := s.loadPassword(s.secretName)
	if err != nil {
		return err
	}

	coreInformers := informers.NewSharedInformerFactoryWithOptions(
		s.clientset,
		0,
		informers.WithNamespace(s.namespace),
	)
	coreInformers.Core().V1().Secrets().Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    s.onAdd,
			UpdateFunc: s.onUpdate,
			DeleteFunc: s.onDelete,
		},
	)
	coreInformers.WaitForCacheSync(s.stopper)
	coreInformers.Start(s.stopper)

	return nil
}

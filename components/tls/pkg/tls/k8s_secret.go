/*
Copyright 2022 Seldon Technologies Ltd.

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

package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
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

type TlsSecretHandler struct {
	clientset     kubernetes.Interface
	secretName    string
	namespace     string
	stopper       chan struct{}
	logger        log.FieldLogger
	folderHandler *TlsFolderHandler
	validation    bool
	cert          *CertificateWrapper
	mu            sync.RWMutex
}

func NewTlsSecretHandler(secretName string, clientset kubernetes.Interface, namespace string, prefix string, validationSecret bool, logger log.FieldLogger) (*TlsSecretHandler, error) {
	if clientset == nil {
		var err error
		clientset, err = k8s.CreateClientset()
		if err != nil {
			logger.WithError(err).Error("Failed to create clientset for TLS secret handler")
			return nil, err
		}
	}
	folderHandler, err := NewTlsFolderHandler(prefix, validationSecret, logger)
	if err != nil {
		return nil, err
	}
	return &TlsSecretHandler{
		clientset:     clientset,
		secretName:    secretName,
		namespace:     namespace,
		stopper:       make(chan struct{}),
		logger:        logger,
		folderHandler: folderHandler,
		validation:    validationSecret,
	}, nil
}

func (t *TlsSecretHandler) GetCertificate() *CertificateWrapper {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cert
}

func (s *TlsSecretHandler) Stop() {
	close(s.stopper)
}

func (s *TlsSecretHandler) getTlsCertificate(secretName string) (*CertificateWrapper, error) {
	logger := s.logger.WithField("func", "getTlsCertificate")
	logger.Infof("Get certificate secret %s from namespace %s", secretName, s.namespace)
	secret, err := s.clientset.CoreV1().Secrets(s.namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		logger.WithError(err).Errorf("Failed to get certificate secret %s from namespace %s", secretName, s.namespace)
		return nil, err
	}
	logger.Infof("Got certificate secret %s from namespace %s", secret.Name, secret.Namespace)
	return s.saveCertificateFromSecret(secret)
}

func saveCert(data []byte, path string) error {
	folder := filepath.Dir(path)
	err := os.MkdirAll(folder, os.ModePerm)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}
	return nil
}

func (s *TlsSecretHandler) saveCertificateFromSecret(secret *corev1.Secret) (*CertificateWrapper, error) {
	c := CertificateWrapper{}
	if !s.validation {
		var err error
		dataKey := filepath.Base(s.folderHandler.keyFilePath)
		key, ok := secret.Data[dataKey]
		if !ok {
			return nil, fmt.Errorf("Failed to find %s in secret %s", dataKey, secret.Name)
		}
		c.KeyPath = s.folderHandler.keyFilePath
		c.KeyRaw = key
		err = saveCert(key, s.folderHandler.keyFilePath)
		if err != nil {
			return nil, err
		}
		dataKey = filepath.Base(s.folderHandler.certFilePath)
		crt, ok := secret.Data[dataKey]
		if !ok {
			return nil, fmt.Errorf("Failed to find %s in secret %s", dataKey, secret.Name)
		}
		c.CrtPath = s.folderHandler.certFilePath
		c.CrtRaw = crt
		err = saveCert(crt, s.folderHandler.certFilePath)
		if err != nil {
			return nil, err
		}
		certificate, err := tls.X509KeyPair(crt, key)
		if err != nil {
			return nil, err
		}
		c.Certificate = &certificate
	}

	dataKey := filepath.Base(s.folderHandler.caFilePath)
	ca, ok := secret.Data[dataKey]
	if !ok {
		return nil, fmt.Errorf("Failed to find %s in secret %s", dataKey, secret.Name)
	}
	c.CaPath = s.folderHandler.caFilePath
	c.CaRaw = ca
	err := saveCert(ca, s.folderHandler.caFilePath)
	if err != nil {
		return nil, err
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("Failed to load ca crt from secret %s", secret.Name)
	}
	c.Ca = capool

	return &c, nil
}

func (s *TlsSecretHandler) onAdd(obj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logger := s.logger.WithField("func", "onAdd")
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("TLS Secret %s added", s.secretName)
		var err error
		s.cert, err = s.saveCertificateFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract TLS certificate from secret %s", secret.Name)
		}
	}
}

func (s *TlsSecretHandler) onUpdate(oldObj, newObj interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	logger := s.logger.WithField("func", "onUpdate")
	secret := newObj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("TLS Secret %s updated", s.secretName)
		var err error
		s.cert, err = s.saveCertificateFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract TLS certificate from secret %s", secret.Name)
		}
	}
}

func (s *TlsSecretHandler) onDelete(obj interface{}) {
	logger := s.logger.WithField("func", "onDelete")
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Warnf("Secret %s deleted", secret.Name)
	}
}

func (s *TlsSecretHandler) GetCertificateAndWatch() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	var err error
	s.cert, err = s.getTlsCertificate(s.secretName)
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

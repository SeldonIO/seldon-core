package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type TlsSecretHandler struct {
	clientset  kubernetes.Interface
	secretName string
	namespace  string
	stopper    chan struct{}
	updater    UpdateCertificateHandler
	logger     log.FieldLogger
}

func NewTlsSecretHandler(secretName string, namespace string, clientset kubernetes.Interface, logger log.FieldLogger) (*TlsSecretHandler, error) {
	return &TlsSecretHandler{
		secretName: secretName,
		clientset:  clientset,
		namespace:  namespace,
		stopper:    make(chan struct{}),
		logger:     logger,
	}, nil
}

func (s *TlsSecretHandler) Stop() {
	close(s.stopper)
}

func (s *TlsSecretHandler) getTlsCertificate(secretName string) (*Certificate, error) {
	secret, err := s.clientset.CoreV1().Secrets(s.namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return s.extractCertificateFromSecret(secret)
}

func (s *TlsSecretHandler) extractCertificateFromSecret(secret *corev1.Secret) (*Certificate, error) {
	key, ok := secret.Data[DefaultKeyName]
	if !ok {
		return nil, fmt.Errorf("Failed to find %s in secret %s", DefaultKeyName, secret.Name)
	}
	crt, ok := secret.Data[DefaultCrtName]
	if !ok {
		return nil, fmt.Errorf("Failed to find %s in secret %s", DefaultCrtName, secret.Name)
	}
	ca, ok := secret.Data[DefaultCaName]
	if !ok {
		return nil, fmt.Errorf("Failed to find %s in secret %s", DefaultCaName, secret.Name)
	}

	certificate, err := tls.X509KeyPair(crt, key)
	if err != nil {
		return nil, err
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("Failed to load ca crt from secret %s", secret.Name)
	}

	return &Certificate{
		Certificate: &certificate,
		Ca:          capool,
	}, nil
}

func (s *TlsSecretHandler) onAdd(obj interface{}) {
	logger := s.logger.WithField("func", "onAdd")
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("TLS Secret %s added", s.secretName)
		cert, err := s.extractCertificateFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract TLS certificate from secret %s", secret.Name)
		}
		if s.updater != nil {
			s.updater.UpdateCertificate(cert)
		}
	}
}

func (s *TlsSecretHandler) onUpdate(oldObj, newObj interface{}) {
	logger := s.logger.WithField("func", "onUpdate")
	secret := newObj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("TLS Secret %s updated", s.secretName)
		cert, err := s.extractCertificateFromSecret(secret)
		if err != nil {
			logger.WithError(err).Errorf("Failed to extract TLS certificate from secret %s", secret.Name)
		}
		if s.updater != nil {
			s.updater.UpdateCertificate(cert)
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

func (s *TlsSecretHandler) GetCertificateAndWatch(updater UpdateCertificateHandler) (*Certificate, error) {
	cert, err := s.getTlsCertificate(s.secretName)
	if err != nil {
		return nil, err
	}
	s.updater = updater
	coreInformers := informers.NewSharedInformerFactoryWithOptions(s.clientset, 0, informers.WithNamespace(s.namespace))
	coreInformers.Core().V1().Secrets().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.onAdd,
		UpdateFunc: s.onUpdate,
		DeleteFunc: s.onDelete,
	})
	coreInformers.WaitForCacheSync(s.stopper)
	coreInformers.Start(s.stopper)
	return cert, nil
}

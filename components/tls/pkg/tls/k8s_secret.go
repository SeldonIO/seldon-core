package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/seldonio/seldon-core-v2/components/tls/pkg/k8s"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

type TlsSecretHandler struct {
	opts          CertificateStoreOptions
	secretName    string
	namespace     string
	stopper       chan struct{}
	updater       UpdateCertificateHandler
	logger        log.FieldLogger
	folderHandler *TlsFolderHandler
}

func NewTlsSecretHandler(secretName string, namespace string, opts CertificateStoreOptions, logger log.FieldLogger) (*TlsSecretHandler, error) {
	if opts.clientset == nil {
		var err error
		opts.clientset, err = k8s.CreateClientset()
		if err != nil {
			logger.WithError(err).Error("Failed to create clientset for TLS secret handler")
			return nil, err
		}
	}
	folderHandler, err := NewTlsFolderHandler(opts, logger)
	if err != nil {
		return nil, err
	}
	return &TlsSecretHandler{
		opts:          opts,
		secretName:    secretName,
		namespace:     namespace,
		stopper:       make(chan struct{}),
		logger:        logger,
		folderHandler: folderHandler,
	}, nil
}

func (s *TlsSecretHandler) Stop() {
	close(s.stopper)
}

func (s *TlsSecretHandler) getTlsCertificate(secretName string) (*CertificateWrapper, error) {
	secret, err := s.opts.clientset.CoreV1().Secrets(s.namespace).Get(context.Background(), secretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
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
	if !s.opts.caOnly {
		var err error
		dataKey := filepath.Base(s.folderHandler.keyFilePath)
		key, ok := secret.Data[dataKey]
		if !ok {
			return nil, fmt.Errorf("Failed to find %s in secret %s", dataKey, secret.Name)
		}
		c.KeyPath = s.folderHandler.keyFilePath
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
	logger := s.logger.WithField("func", "onAdd")
	secret := obj.(*corev1.Secret)
	if secret.Name == s.secretName {
		logger.Infof("TLS Secret %s added", s.secretName)
		cert, err := s.saveCertificateFromSecret(secret)
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
		cert, err := s.saveCertificateFromSecret(secret)
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

func (s *TlsSecretHandler) GetCertificateAndWatch(updater UpdateCertificateHandler) (*CertificateWrapper, error) {
	cert, err := s.getTlsCertificate(s.secretName)
	if err != nil {
		return nil, err
	}
	s.updater = updater
	coreInformers := informers.NewSharedInformerFactoryWithOptions(s.opts.clientset, 0, informers.WithNamespace(s.namespace))
	coreInformers.Core().V1().Secrets().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    s.onAdd,
		UpdateFunc: s.onUpdate,
		DeleteFunc: s.onDelete,
	})
	coreInformers.WaitForCacheSync(s.stopper)
	coreInformers.Start(s.stopper)
	return cert, nil
}

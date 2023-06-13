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
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/credentials"
	"k8s.io/client-go/kubernetes"

	"github.com/seldonio/seldon-core/components/tls/v2/pkg/util"
)

const (
	envSecretSuffix = "_TLS_SECRET_NAME"
	envNamespace    = "POD_NAMESPACE"
)

type funcTLSServerOption struct {
	f func(options *CertificateStoreOptions)
}

func (fdo *funcTLSServerOption) apply(do *CertificateStoreOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(options *CertificateStoreOptions)) *funcTLSServerOption {
	return &funcTLSServerOption{
		f: f,
	}
}

type TLSServerOption interface {
	apply(options *CertificateStoreOptions)
}

type CertificateStore struct {
	opts   CertificateStoreOptions
	logger logrus.FieldLogger

	certificateManager CertificateManager
	validationManager  CertificateManager
}

type CertificateStoreOptions struct {
	prefix           string
	validationPrefix string
	clientset        kubernetes.Interface
	validationOnly   bool
}

func (c CertificateStoreOptions) String() string {
	return fmt.Sprintf("prefix=%s validationPrefix=%s clientset=%v",
		c.prefix, c.validationPrefix, c.clientset)
}

func getDefaultCertificateStoreOptions() CertificateStoreOptions {
	return CertificateStoreOptions{}
}

func Prefix(prefix string) TLSServerOption {
	return newFuncServerOption(func(o *CertificateStoreOptions) {
		o.prefix = prefix
	})
}

func ValidationPrefix(prefix string) TLSServerOption {
	return newFuncServerOption(func(o *CertificateStoreOptions) {
		o.validationPrefix = prefix
	})
}

func ValidationOnly(validationOnly bool) TLSServerOption {
	return newFuncServerOption(func(o *CertificateStoreOptions) {
		o.validationOnly = validationOnly
	})
}

func ClientSet(clientSet kubernetes.Interface) TLSServerOption {
	return newFuncServerOption(func(o *CertificateStoreOptions) {
		o.clientset = clientSet
	})
}

func NewCertificateStore(opt ...TLSServerOption) (*CertificateStore, error) {
	opts := getDefaultCertificateStoreOptions()
	for _, o := range opt {
		o.apply(&opts)
	}
	logger := logrus.New().WithField("source", "CertificateStore")
	logger.Infof("Options:%s", opts.String())
	var err error
	var manager CertificateManager
	var validationManager CertificateManager
	if !opts.validationOnly {
		if secretName, ok := util.GetEnv(opts.prefix, envSecretSuffix); ok {
			logger.Infof("Starting new certificate store for %s from secret %s", opts.prefix, secretName)
			namespace, ok := os.LookupEnv(envNamespace)
			if !ok {
				return nil, fmt.Errorf("Namespace env var %s not found and needed for secret TLS", envNamespace)
			}
			manager, err = NewTlsSecretHandler(secretName, opts.clientset, namespace, opts.prefix, false, logger)
			if err != nil {
				return nil, err
			}

			// optionally add a validation secret ca
			if opts.validationPrefix != "" {
				if secretName, ok := util.GetEnv(opts.validationPrefix, envSecretSuffix); ok {
					logger.Infof("Starting new certificate store for %s from secret %s", opts.validationPrefix, secretName)
					validationManager, err = NewTlsSecretHandler(secretName, opts.clientset, namespace, opts.validationPrefix, true, logger)
					if err != nil {
						return nil, err
					}
				}
			}
		} else {
			manager, err = NewTlsFolderHandler(opts.prefix, false, logger)
			if err != nil {
				return nil, err
			}
			validationFolderHandler, err := NewTlsFolderHandler(opts.validationPrefix, true, logger)
			if validationFolderHandler != nil {
				validationManager = validationFolderHandler
			}
			if err != nil {
				return nil, err
			}
		}
	} else if opts.validationPrefix != "" {
		logger.Info("Just looking for validation cert")
		if secretName, ok := util.GetEnv(opts.validationPrefix, envSecretSuffix); ok {
			namespace, ok := os.LookupEnv(envNamespace)
			if !ok {
				return nil, fmt.Errorf("Namespace env var %s not found and needed for secret TLS", envNamespace)
			}
			logger.Infof("Starting new certificate store for %s from secret %s", opts.validationPrefix, secretName)
			validationManager, err = NewTlsSecretHandler(secretName, opts.clientset, namespace, opts.validationPrefix, true, logger)
			if err != nil {
				return nil, err
			}
		} else {
			validationFolderHandler, err := NewTlsFolderHandler(opts.validationPrefix, true, logger)
			if validationFolderHandler != nil {
				validationManager = validationFolderHandler
			}
			if err != nil {
				return nil, err
			}
		}
	}

	return createCertStoreFromManager(opts, logger, manager, validationManager)
}

func createCertStoreFromManager(opts CertificateStoreOptions, logger logrus.FieldLogger, manager CertificateManager, validationManager CertificateManager) (*CertificateStore, error) {
	certStore := CertificateStore{
		opts:               opts,
		logger:             logger,
		certificateManager: manager,
		validationManager:  validationManager,
	}
	var err error
	// Set cert if available
	if manager != nil {
		logger.Infof("Getting certificate for %s", opts.prefix)
		err = manager.GetCertificateAndWatch()
		if err != nil {
			return nil, err
		}
	}
	// Set validation cert if available
	if validationManager != nil {
		logger.Infof("Getting validation certificate for %s", opts.validationPrefix)
		err = validationManager.GetCertificateAndWatch()
		if err != nil {
			return nil, err
		}
	}
	return &certStore, nil
}

func (s *CertificateStore) GetServerCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	cert := s.certificateManager.GetCertificate()
	if cert != nil {
		return cert.Certificate, nil
	} else {
		return nil, fmt.Errorf("Nil certificate for %s", s.opts.String())
	}
}

func (s *CertificateStore) GetClientCertificate(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
	cert := s.certificateManager.GetCertificate()
	if cert != nil {
		return cert.Certificate, nil
	} else {
		return nil, fmt.Errorf("Nil certificate for %s", s.opts.String())
	}
}

func (s *CertificateStore) getCertificates() (*CertificateWrapper, *CertificateWrapper) {
	var certificate *CertificateWrapper
	var validationCA *CertificateWrapper
	if s.certificateManager != nil {
		certificate = s.certificateManager.GetCertificate()
	}
	if s.validationManager != nil {
		validationCA = s.validationManager.GetCertificate()
	}
	return certificate, validationCA
}

func (s *CertificateStore) CreateClientTLSConfig() *tls.Config {
	logger := s.logger.WithField("func", "CreateClientTransportCredentials")
	var rootCAs *x509.CertPool
	certificate, validationCA := s.getCertificates()
	if certificate != nil {
		rootCAs = certificate.Ca
		logger.Info("Using rootCA from cert resource")
	}
	if validationCA != nil {
		rootCAs = validationCA.Ca
		logger.Info("Using rootCA from validation resource")
	}
	// Create tlsConfig
	tlsConfig := &tls.Config{
		RootCAs: rootCAs,
	}
	// Add updater method
	if certificate != nil {
		tlsConfig.GetClientCertificate = s.GetClientCertificate
	}
	return tlsConfig
}

func (s *CertificateStore) CreateClientTransportCredentials() credentials.TransportCredentials {
	return credentials.NewTLS(s.CreateClientTLSConfig())
}

func (s *CertificateStore) CreateServerTLSConfig() *tls.Config {
	certificate, validationCA := s.getCertificates()
	// Assumes there is always a cert for a server
	clientCAs := certificate.Ca
	if validationCA != nil {
		clientCAs = validationCA.Ca
	}
	tlsConfig := &tls.Config{
		ClientAuth:     tls.RequireAndVerifyClientCert,
		GetCertificate: s.GetServerCertificate,
		ClientCAs:      clientCAs,
	}
	return tlsConfig
}

func (s *CertificateStore) CreateServerTransportCredentials() credentials.TransportCredentials {
	return credentials.NewTLS(s.CreateServerTLSConfig())
}

func (s *CertificateStore) GetCertificate() *CertificateWrapper {
	if s.certificateManager != nil {
		return s.certificateManager.GetCertificate()
	} else {
		return nil
	}
}

func (s *CertificateStore) GetValidationCertificate() *CertificateWrapper {
	if s.validationManager != nil {
		return s.validationManager.GetCertificate()
	} else {
		return nil
	}
}

func (s *CertificateStore) Stop() {
	if s.certificateManager != nil {
		s.certificateManager.Stop()
	}
	if s.validationManager != nil {
		s.validationManager.Stop()
	}
}

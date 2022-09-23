package tls

import (
	"crypto/tls"
	"fmt"
	"k8s.io/client-go/kubernetes"
	"os"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/credentials"
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
	opts               CertificateStoreOptions
	logger             logrus.FieldLogger
	mu                 sync.RWMutex
	certificateManager CertificateManager
	certificate        *CertificateWrapper
}

type CertificateStoreOptions struct {
	caOnly    bool
	prefix    string
	clientset kubernetes.Interface
}

func (c CertificateStoreOptions) String() string {
	return fmt.Sprintf("prefix=%s clientset=%v caOnly=%v",
		c.prefix, c.clientset, c.caOnly)
}

func getDefaultCertificateStoreOptions() CertificateStoreOptions {
	return CertificateStoreOptions{}
}

func getEnvVarKey(prefix string, suffix string) string {
	return fmt.Sprintf("%s%s", prefix, suffix)
}

func getEnv(prefix string, suffix string) (string, bool) {
	secretName, ok := os.LookupEnv(getEnvVarKey(prefix, suffix))
	if ok {
		ok = strings.TrimSpace(secretName) != ""
	}
	return secretName, ok
}

func Prefix(prefix string) TLSServerOption {
	return newFuncServerOption(func(o *CertificateStoreOptions) {
		o.prefix = prefix
	})
}

func ClientSet(clientSet kubernetes.Interface) TLSServerOption {
	return newFuncServerOption(func(o *CertificateStoreOptions) {
		o.clientset = clientSet
	})
}

func CaOnly(caOnly bool) TLSServerOption {
	return newFuncServerOption(func(o *CertificateStoreOptions) {
		o.caOnly = caOnly
	})
}

func NewCertificateStore(opt ...TLSServerOption) (*CertificateStore, error) {
	opts := getDefaultCertificateStoreOptions()
	for _, o := range opt {
		o.apply(&opts)
	}
	logger := logrus.New().WithField("source", "CertificateStore")
	logger.Infof("Options:%s", opts.String())
	if secretName, ok := getEnv(opts.prefix, envSecretSuffix); ok {
		logger.Infof("Starting new certificate store for %s from secret %s", opts.prefix, secretName)
		namespace, ok := os.LookupEnv(envNamespace)
		if !ok {
			return nil, fmt.Errorf("Namespace env var %s not found and needed for secret TLS", envNamespace)
		}
		manager, err := NewTlsSecretHandler(secretName, namespace, opts, logger)
		if err != nil {
			return nil, err
		}
		return createCertStoreFromManager(opts, logger, manager)
	} else {
		manager, err := NewTlsFolderHandler(opts, logger)
		if err != nil {
			return nil, err
		}
		return createCertStoreFromManager(opts, logger, manager)
	}
}

func createCertStoreFromManager(opts CertificateStoreOptions, logger logrus.FieldLogger, manager CertificateManager) (*CertificateStore, error) {
	certStore := CertificateStore{
		opts:               opts,
		logger:             logger,
		certificateManager: manager,
	}
	cert, err := manager.GetCertificateAndWatch(&certStore)
	if err != nil {
		return nil, err
	}
	certStore.certificate = cert
	return &certStore, nil
}

func (m *CertificateStore) UpdateCertificate(certificate *CertificateWrapper) {
	logger := m.logger.WithField("func", "UpdateCertificate")
	logger.Infof("Updating certificate %s", m.opts.prefix)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.certificate = certificate
}

func (m *CertificateStore) GetServerCertificate(*tls.ClientHelloInfo) (*tls.Certificate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.certificate.Certificate, nil
}

func (m *CertificateStore) GetClientCertificate(*tls.CertificateRequestInfo) (*tls.Certificate, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.certificate.Certificate, nil
}

func (s *CertificateStore) CreateClientTransportCredentials() credentials.TransportCredentials {
	tlsConfig := &tls.Config{
		GetClientCertificate: s.GetClientCertificate,
		RootCAs:              s.certificate.Ca,
	}
	return credentials.NewTLS(tlsConfig)
}

func (s *CertificateStore) CreateServerTransportCredentials() credentials.TransportCredentials {
	tlsConfig := &tls.Config{
		ClientAuth:     tls.RequireAndVerifyClientCert,
		GetCertificate: s.GetServerCertificate,
		ClientCAs:      s.certificate.Ca,
	}
	return credentials.NewTLS(tlsConfig)
}

func (s *CertificateStore) GetCertificate() *CertificateWrapper {
	return s.certificate
}

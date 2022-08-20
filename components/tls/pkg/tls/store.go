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
	envFolderSuffix = "_TLS_FOLDER_PATH"
	envNamespace    = "POD_NAMESPACE"
)

type CertificateStore struct {
	prefix             string
	logger             logrus.FieldLogger
	mu                 sync.RWMutex
	certificateManager CertificateManager
	certificate        *Certificate
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

func NewCertificateStore(prefix string, clientset kubernetes.Interface) (*CertificateStore, error) {
	logger := logrus.New()
	if secretName, ok := getEnv(prefix, envSecretSuffix); ok {
		logger.Infof("Starting new certificate store for %s from secret %s", prefix, secretName)
		namespace, ok := os.LookupEnv(envNamespace)
		if !ok {
			return nil, fmt.Errorf("Namespace env var %s not found and needed for secret TLS", envNamespace)
		}
		manager, err := NewTlsSecretHandler(secretName, namespace, clientset, logger)
		if err != nil {
			return nil, err
		}
		return createCertStoreFromManager(prefix, logger, manager)
	}
	if folderPath, ok := getEnv(prefix, envFolderSuffix); ok {
		logger.Infof("Starting new certificate store for %s from path %s", prefix, folderPath)
		manager, err := NewTlsFolderHandler(folderPath, logger)
		if err != nil {
			return nil, err
		}
		return createCertStoreFromManager(prefix, logger, manager)
	}
	return nil, nil
}

func createCertStoreFromManager(prefix string, logger logrus.FieldLogger, manager CertificateManager) (*CertificateStore, error) {
	certStore := CertificateStore{
		prefix:             prefix,
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

func (m *CertificateStore) UpdateCertificate(certificate *Certificate) {
	logger := m.logger.WithField("func", "UpdateCertificate")
	logger.Infof("Updating certificate %s", m.prefix)
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

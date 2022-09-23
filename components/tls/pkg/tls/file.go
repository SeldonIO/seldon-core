package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"istio.io/pkg/filewatcher"
)

const (
	envKeyLocationSuffix = "_TLS_KEY_LOCATION"
	envCrtLocationSuffix = "_TLS_CRT_LOCATION"
	envCaLocationSuffix  = "_TLS_CA_LOCATION"
)

type TlsFolderHandler struct {
	opts         CertificateStoreOptions
	certFilePath string
	keyFilePath  string
	caFilePath   string
	updater      UpdateCertificateHandler
	logger       log.FieldLogger
	watcher      filewatcher.FileWatcher
}

func NewTlsFolderHandler(opts CertificateStoreOptions, logger log.FieldLogger) (*TlsFolderHandler, error) {
	var certFilePath, keyFilePath string
	var ok bool
	if !opts.caOnly {
		certFilePath, ok = getEnv(opts.prefix, envCrtLocationSuffix)
		if !ok {
			return nil, fmt.Errorf("Failed to find %s%s", opts.prefix, envCrtLocationSuffix)
		}
		keyFilePath, ok = getEnv(opts.prefix, envKeyLocationSuffix)
		if !ok {
			return nil, fmt.Errorf("Failed to find %s%s", opts.prefix, envKeyLocationSuffix)
		}
	}
	caFilePath, ok := getEnv(opts.prefix, envCaLocationSuffix)
	if !ok {
		return nil, fmt.Errorf("Failed to find %s%s", opts.prefix, envCaLocationSuffix)
	}

	return &TlsFolderHandler{
		opts:         opts,
		certFilePath: certFilePath,
		keyFilePath:  keyFilePath,
		caFilePath:   caFilePath,
		watcher:      filewatcher.NewWatcher(),
		logger:       logger,
	}, nil
}

func (t *TlsFolderHandler) loadCertificate() (*CertificateWrapper, error) {
	var err error
	c := CertificateWrapper{
		KeyPath: t.keyFilePath,
		CrtPath: t.certFilePath,
		CaPath:  t.caFilePath,
	}
	if !t.opts.caOnly {
		certificate, err := tls.LoadX509KeyPair(t.certFilePath, t.keyFilePath)
		if err != nil {
			return nil, err
		}
		c.Certificate = &certificate
	}

	caRaw, err := os.ReadFile(t.caFilePath)
	if err != nil {
		return nil, err
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(caRaw) {
		return nil, fmt.Errorf("Failed to load ca crt from %s", t.caFilePath)
	}
	c.Ca = capool
	return &c, nil
}

func (t *TlsFolderHandler) reloadCertificate() {
	logger := t.logger.WithField("func", "reloadCertificate")
	cert, err := t.loadCertificate()
	if err != nil {
		logger.WithError(err).Error("Failed to reload certificate")
		return
	}
	t.updater.UpdateCertificate(cert)
}

func (t *TlsFolderHandler) GetCertificateAndWatch(updater UpdateCertificateHandler) (*CertificateWrapper, error) {
	cert, err := t.loadCertificate()
	if err != nil {
		return nil, err
	}
	t.updater = updater
	if !t.opts.caOnly {
		addFileWatcher(t.watcher, t.keyFilePath, t.reloadCertificate)
		addFileWatcher(t.watcher, t.certFilePath, t.reloadCertificate)
	}
	addFileWatcher(t.watcher, t.caFilePath, t.reloadCertificate)
	return cert, nil
}

// Utility function as used in istio to use filewatcher.
// See https://github.com/istio/istio/blob/d1bd0fc297cfb0b1bb349066ffbbfa341220561e/pkg/config/mesh/watcher.go
func addFileWatcher(fileWatcher filewatcher.FileWatcher, file string, callback func()) {
	_ = fileWatcher.Add(file)
	go func() {
		var timerC <-chan time.Time
		for {
			select {
			case <-timerC:
				timerC = nil
				callback()
			case <-fileWatcher.Events(file):
				// Use a timer to debounce configuration updates
				if timerC == nil {
					timerC = time.After(200 * time.Millisecond)
				}
			}
		}
	}()
}

func (t *TlsFolderHandler) Stop() {
	logger := t.logger.WithField("func", "Stop")
	if !t.opts.caOnly {
		err := t.watcher.Remove(t.keyFilePath)
		if err != nil {
			logger.WithError(err).Errorf("Failed to stop watch on %s", t.keyFilePath)
		}
		err = t.watcher.Remove(t.certFilePath)
		if err != nil {
			logger.WithError(err).Errorf("Failed to stop watch on %s", t.certFilePath)
		}
	}
	err := t.watcher.Remove(t.caFilePath)
	if err != nil {
		logger.WithError(err).Errorf("Failed to stop watch on %s", t.caFilePath)
	}
	err = t.watcher.Close()
	if err != nil {
		logger.WithError(err).Errorf("Failed to close watcher")
	}
}

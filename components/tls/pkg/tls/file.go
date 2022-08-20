package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"
	"istio.io/pkg/filewatcher"
)

type TlsFolderHandler struct {
	folderPath   string
	certFilePath string
	keyFilePath  string
	caFilePath   string
	updater      UpdateCertificateHandler
	logger       log.FieldLogger
	watcher      filewatcher.FileWatcher
}

func NewTlsFolderHandler(folderPath string, logger log.FieldLogger) (*TlsFolderHandler, error) {
	return &TlsFolderHandler{
		folderPath:   folderPath,
		certFilePath: filepath.Join(folderPath, DefaultCrtName),
		keyFilePath:  filepath.Join(folderPath, DefaultKeyName),
		caFilePath:   filepath.Join(folderPath, DefaultCaName),
		watcher:      filewatcher.NewWatcher(),
		logger:       logger,
	}, nil
}

func (t *TlsFolderHandler) loadCertificate() (*Certificate, error) {
	certificate, err := tls.LoadX509KeyPair(t.certFilePath, t.keyFilePath)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(t.caFilePath)
	if err != nil {
		return nil, err
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("Failed to load ca crt from %s", t.caFilePath)
	}

	return &Certificate{
		Certificate: &certificate,
		Ca:          capool,
	}, nil
}

func (t *TlsFolderHandler) reloadCertificate() {
	logger := t.logger.WithField("func", "reloadCertificate")
	cert, err := t.loadCertificate()
	if err != nil {
		logger.WithError(err).Errorf("Failed to reload certificate at %s", t.folderPath)
		return
	}
	t.updater.UpdateCertificate(cert)
}

func (t *TlsFolderHandler) GetCertificateAndWatch(updater UpdateCertificateHandler) (*Certificate, error) {
	cert, err := t.loadCertificate()
	if err != nil {
		return nil, err
	}
	t.updater = updater
	addFileWatcher(t.watcher, t.keyFilePath, t.reloadCertificate)
	addFileWatcher(t.watcher, t.certFilePath, t.reloadCertificate)
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
	err := t.watcher.Remove(t.keyFilePath)
	if err != nil {
		logger.WithError(err).Errorf("Failed to stop watch on %s", t.keyFilePath)
	}
	err = t.watcher.Remove(t.certFilePath)
	if err != nil {
		logger.WithError(err).Errorf("Failed to stop watch on %s", t.certFilePath)
	}
	err = t.watcher.Remove(t.caFilePath)
	if err != nil {
		logger.WithError(err).Errorf("Failed to stop watch on %s", t.caFilePath)
	}
	err = t.watcher.Close()
	if err != nil {
		logger.WithError(err).Errorf("Failed to close watcher")
	}
}

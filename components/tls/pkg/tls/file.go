/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"istio.io/pkg/filewatcher"

	"github.com/seldonio/seldon-core/components/tls/v2/pkg/util"
)

const (
	EnvKeyLocationSuffix = "_TLS_KEY_LOCATION"
	EnvCrtLocationSuffix = "_TLS_CRT_LOCATION"
	EnvCaLocationSuffix  = "_TLS_CA_LOCATION"
)

type TlsFolderHandler struct {
	prefix       string
	certFilePath string
	keyFilePath  string
	caFilePath   string
	logger       log.FieldLogger
	watcher      filewatcher.FileWatcher
	validation   bool
	cert         *CertificateWrapper
	mu           sync.RWMutex
}

func NewTlsFolderHandler(prefix string, validation bool, logger log.FieldLogger) (*TlsFolderHandler, error) {
	var certFilePath, keyFilePath string
	var ok bool
	if !validation {
		certFilePath, ok = util.GetNonEmptyEnv(prefix, EnvCrtLocationSuffix)
		if !ok {
			return nil, fmt.Errorf("Failed to find %s%s or empty value", prefix, EnvCrtLocationSuffix)
		}
		keyFilePath, ok = util.GetNonEmptyEnv(prefix, EnvKeyLocationSuffix)
		if !ok {
			return nil, fmt.Errorf("Failed to find %s%s or empty value", prefix, EnvKeyLocationSuffix)
		}
	}
	caFilePath, ok := util.GetNonEmptyEnv(prefix, EnvCaLocationSuffix)
	if !ok {
		if validation {
			return nil, nil // Allow ca only to be optional and return nil
		} else {
			return nil, fmt.Errorf("Failed to find %s%s or empty value", prefix, EnvCaLocationSuffix)
		}
	}

	return &TlsFolderHandler{
		prefix:       prefix,
		certFilePath: certFilePath,
		keyFilePath:  keyFilePath,
		caFilePath:   caFilePath,
		watcher:      filewatcher.NewWatcher(),
		logger:       logger,
		validation:   validation,
	}, nil
}

func (t *TlsFolderHandler) GetCertificate() *CertificateWrapper {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.cert
}

func (t *TlsFolderHandler) loadCertificate() (*CertificateWrapper, error) {
	var err error
	c := CertificateWrapper{
		KeyPath: t.keyFilePath,
		CrtPath: t.certFilePath,
		CaPath:  t.caFilePath,
	}
	if !t.validation {
		certificate, err := tls.LoadX509KeyPair(t.certFilePath, t.keyFilePath)
		if err != nil {
			return nil, err
		}
		c.Certificate = &certificate
		// Load raw versions
		c.CrtRaw, err = os.ReadFile(t.certFilePath)
		if err != nil {
			return nil, err
		}
		c.KeyRaw, err = os.ReadFile(t.keyFilePath)
		if err != nil {
			return nil, err
		}
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
	c.CaRaw = caRaw
	return &c, nil
}

func (t *TlsFolderHandler) reloadCertificate() {
	logger := t.logger.WithField("func", "reloadCertificate")
	var err error
	t.mu.Lock()
	t.cert, err = t.loadCertificate()
	t.mu.Unlock()
	if err != nil {
		logger.WithError(err).Error("Failed to reload certificate")
		return
	}
}

func (t *TlsFolderHandler) GetCertificateAndWatch() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	var err error
	t.cert, err = t.loadCertificate()
	if err != nil {
		return err
	}
	if !t.validation {
		addFileWatcher(t.watcher, t.keyFilePath, t.reloadCertificate)
		addFileWatcher(t.watcher, t.certFilePath, t.reloadCertificate)
	}
	addFileWatcher(t.watcher, t.caFilePath, t.reloadCertificate)
	return nil
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
	if !t.validation {
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

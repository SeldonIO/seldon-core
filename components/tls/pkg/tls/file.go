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
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/util"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"istio.io/pkg/filewatcher"
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

func getDefaultPath(prefix string, suffix string) string {
	var filename string
	switch suffix {
	case EnvKeyLocationSuffix:
		filename = "tls.key"
	case EnvCrtLocationSuffix:
		filename = "tls.crt"
	default:
		filename = "ca.crt"
	}
	return fmt.Sprintf("/tmp/certs/%s%s/%s", prefix, suffix, filename)
}

func NewTlsFolderHandler(prefix string, validation bool, logger log.FieldLogger) (*TlsFolderHandler, error) {
	var certFilePath, keyFilePath string
	var ok bool
	if !validation {
		certFilePath, ok = util.GetEnv(prefix, EnvCrtLocationSuffix)
		if !ok {
			certFilePath = getDefaultPath(prefix, EnvCrtLocationSuffix)
		}
		keyFilePath, ok = util.GetEnv(prefix, EnvKeyLocationSuffix)
		if !ok {
			keyFilePath = getDefaultPath(prefix, EnvKeyLocationSuffix)
		}
	}
	caFilePath, ok := util.GetEnv(prefix, EnvCaLocationSuffix)
	if !ok {
		caFilePath = getDefaultPath(prefix, EnvCaLocationSuffix)
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

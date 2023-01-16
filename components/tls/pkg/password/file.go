/*
Copyright 2023 Seldon Technologies Ltd.

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

package password

import (
	"fmt"
	"github.com/seldonio/seldon-core/components/tls/v2/pkg/util"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"istio.io/pkg/filewatcher"
)

type PasswordFolderHandler struct {
	passwordFilePath string
	logger           log.FieldLogger
	watcher          filewatcher.FileWatcher
	password         string
	mu               sync.RWMutex
}

func NewPasswordFolderHandler(prefix string, suffix string, logger log.FieldLogger) (*PasswordFolderHandler, error) {
	passwordFilePath, ok := util.GetEnv(prefix, suffix)
	if !ok {
		return nil, fmt.Errorf("Failed to find %s%s or empty value", prefix, suffix)
	}

	return &PasswordFolderHandler{
		passwordFilePath: passwordFilePath,
		watcher:          filewatcher.NewWatcher(),
		logger:           logger,
	}, nil
}

func (s *PasswordFolderHandler) GetPassword() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.password
}

func (t *PasswordFolderHandler) loadPassword() error {
	var err error
	passwordRaw, err := os.ReadFile(t.passwordFilePath)
	if err != nil {
		return err
	}
	t.password = string(passwordRaw)
	return nil
}

func (t *PasswordFolderHandler) reloadPassword() {
	logger := t.logger.WithField("func", "reloadPassword")
	t.mu.Lock()
	defer t.mu.Unlock()
	err := t.loadPassword()
	if err != nil {
		logger.WithError(err).Error("Failed to reload password")
	}
}

func (t *PasswordFolderHandler) GetPasswordAndWatch() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	err := t.loadPassword()
	if err != nil {
		return err
	}
	addFileWatcher(t.watcher, t.passwordFilePath, t.reloadPassword)
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

func (t *PasswordFolderHandler) Stop() {
	logger := t.logger.WithField("func", "Stop")
	err := t.watcher.Remove(t.passwordFilePath)
	if err != nil {
		logger.WithError(err).Errorf("Failed to stop watch on %s", t.passwordFilePath)
	}
	err = t.watcher.Close()
	if err != nil {
		logger.WithError(err).Errorf("Failed to close watcher")
	}
}

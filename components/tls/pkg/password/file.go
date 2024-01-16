/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package password

import (
	"fmt"
	"os"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"istio.io/pkg/filewatcher"

	"github.com/seldonio/seldon-core/components/tls/v2/pkg/util"
)

type fileStore struct {
	passwordFilePath string
	logger           log.FieldLogger
	watcher          filewatcher.FileWatcher
	password         string
	mu               sync.RWMutex
}

func newFileStore(prefix string, suffix string, logger log.FieldLogger) (*fileStore, error) {
	logger = logger.WithField("source", "FilePasswordStore")

	passwordFilePath, ok := util.GetNonEmptyEnv(prefix, suffix)
	if !ok {
		return nil, fmt.Errorf("Failed to find %s%s or empty value", prefix, suffix)
	}

	return &fileStore{
		passwordFilePath: passwordFilePath,
		watcher:          filewatcher.NewWatcher(),
		logger:           logger,
	}, nil
}

func (s *fileStore) GetPassword() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.password
}

func (t *fileStore) loadPassword() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	logger := t.logger.WithField("func", "loadPassword")

	passwordRaw, err := os.ReadFile(t.passwordFilePath)
	if err != nil {
		logger.WithError(err).Error("failed to load password")
		return err
	}

	t.password = string(passwordRaw)
	return nil
}

func (t *fileStore) loadAndWatchPassword() error {
	err := t.loadPassword()
	if err != nil {
		return err
	}

	addFileWatcher(
		t.watcher,
		t.passwordFilePath,
		func() {
			t.logger.Info("file has changed; reloading password")
			_ = t.loadPassword()
		},
	)
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

func (t *fileStore) Stop() {
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

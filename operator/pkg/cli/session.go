/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/grpc/metadata"
)

const (
	sessionFilename = "session"
	separator       = ","
)

func saveStickySessionKeyHttp(headers http.Header) (bool, error) {
	sessionKeys := headers.Values(SeldonRouteHeader)
	if sessionKeys != nil {
		err := saveSessionKeysToFile(sessionKeys)
		if err != nil {
			return false, err
		} else {
			return true, err
		}

	}
	return false, nil
}

func saveStickySessionKeyGrpc(headers metadata.MD) (bool, error) {
	sessionKey := headers[SeldonRouteHeader]
	if sessionKey != nil {
		err := saveSessionKeysToFile(sessionKey)
		if err != nil {
			return false, err
		} else {
			return true, err
		}

	}
	return false, nil
}

func getStickySessionKeys() ([]string, error) {
	return loadSessionKeyFromFile()
}

func getSessionFile() string {
	return filepath.Join(getConfigDir(), sessionFilename)
}

func saveSessionKeysToFile(keys []string) error {
	path := getConfigDir()
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return os.WriteFile(getSessionFile(), []byte(strings.Join(keys, separator)), os.ModePerm)
}

func loadSessionKeyFromFile() ([]string, error) {
	data, err := os.ReadFile(getSessionFile())
	if err != nil {
		return nil, err
	}
	keys := string(data)
	return strings.Split(keys, separator), nil
}

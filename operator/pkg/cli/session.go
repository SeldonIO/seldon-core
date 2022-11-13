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

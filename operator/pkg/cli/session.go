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

func getCfgSessionPath() string {
	return filepath.Join(getCfgPath(), sessionFilename)
}

func saveSessionKeysToFile(keys []string) error {
	path := getCfgPath()
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return os.WriteFile(getCfgSessionPath(), []byte(strings.Join(keys, separator)), os.ModePerm)
}

func loadSessionKeyFromFile() ([]string, error) {
	data, err := os.ReadFile(getCfgSessionPath())
	if err != nil {
		return nil, err
	}
	keys := string(data)
	return strings.Split(keys, separator), nil
}

package cli

import (
	"io/ioutil"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

const (
	seldonCfgFilepath = ".config/seldon/cli"
	sessionFilename   = "session"
)

func saveStickySessionKey(headers http.Header) (bool, error) {
	sessionKey := findSeldonRouteHeader(headers)
	if sessionKey != "" {
		err := saveSessionKeyToFile(sessionKey)
		if err != nil {
			return false, err
		} else {
			return true, err
		}

	}
	return false, nil
}

func getStickySessionKey() (string, error) {
	return loadSessionKeyFromFile()
}

func findSeldonRouteHeader(headers http.Header) string {
	for k, v := range headers {
		if strings.ToLower(k) == SeldonRouteHeader {
			return v[0]
		}
	}
	return ""
}

func getCfgPath() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, seldonCfgFilepath)
}

func getCfgSessionPath() string {
	return filepath.Join(getCfgPath(), sessionFilename)
}

func saveSessionKeyToFile(key string) error {
	path := getCfgPath()
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(getCfgSessionPath(), []byte(key), os.ModePerm)
}

func loadSessionKeyFromFile() (string, error) {
	data, err := ioutil.ReadFile(getCfgSessionPath())
	if err != nil {
		return "", err
	}
	return string(data), nil
}

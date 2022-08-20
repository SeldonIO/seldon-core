package cli

import (
	"encoding/json"
	"os"
	"os/user"
	"path/filepath"
)

// start config struct
type SeldonCLIConfig struct {
	SchedulerHost string `json:"schedulerHost,omitempty"`
	TlsKeyPath    string `json:"tlsKeyPath,omitempty"`
	TlsCrtPath    string `json:"tlsCrtPath,omitempty"`
	CaCrtPath     string `json:"caCrtPath,omitempty"`
}

// end config struct

const (
	seldonCfgFilepath = ".config/seldon/cli"
	configFilename    = "config.json"
)

func LoadSeldonCLIConfig() (*SeldonCLIConfig, error) {
	path := getCfgFilePath()
	_, err := os.Stat(path)
	if err != nil {
		return &SeldonCLIConfig{}, nil
	}

	byteValue, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := SeldonCLIConfig{}
	err = json.Unmarshal(byteValue, &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func getCfgFilePath() string {
	return filepath.Join(getCfgPath(), configFilename)
}

func getCfgPath() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, seldonCfgFilepath)
}

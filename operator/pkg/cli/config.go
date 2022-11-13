package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
)

// start config struct
type SeldonCLIConfig struct {
	Dataplane    *Dataplane    `json:"dataplane,omitempty"`
	Controlplane *ControlPlane `json:"controlplane,omitempty"`
	Kafka        *KafkaConfig  `json:"kafka,omitempty"`
}

type Dataplane struct {
	InferHost     string `json:"inferHost,omitempty"`
	Tls           bool   `json:"tls,omitempty"`
	SkipSSLVerify bool   `json:"skipSSLVerify,omitempty"`
	KeyPath       string `json:"keyPath,omitempty"`
	CrtPath       string `json:"crtPath,omitempty"`
	CaPath        string `json:"caPath,omitempty"`
}

type ControlPlane struct {
	SchedulerHost string `json:"schedulerHost,omitempty"`
	Tls           bool   `json:"tls,omitempty"`
	KeyPath       string `json:"keyPath,omitempty"`
	CrtPath       string `json:"crtPath,omitempty"`
	CaPath        string `json:"caPath,omitempty"`
}

type KafkaConfig struct {
	Bootstrap string `json:"bootstrap,omitempty"`
	Tls       bool   `json:"tls,omitempty"`
	KeyPath   string `json:"keyPath,omitempty"`
	CrtPath   string `json:"crtPath,omitempty"`
	CaPath    string `json:"caPath,omitempty"`
}

// end config struct

const (
	seldonCfgFilepath = ".config/seldon/cli"
	configFilename    = "config.json"
)

func LoadSeldonCLIConfig() (*SeldonCLIConfig, error) {
	path := getConfigFile()
	_, err := os.Stat(path)
	if err != nil {
		return &SeldonCLIConfig{}, nil
	}

	byteValue, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to load CLI config from %s: %s", path, err.Error())
	}
	cfg := SeldonCLIConfig{}
	err = json.Unmarshal(byteValue, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to load CLI config from %s: %s", path, err.Error())
	}
	return &cfg, nil
}

func getConfigFile() string {
	return filepath.Join(getConfigDir(), configFilename)
}

func getConfigDir() string {
	usr, _ := user.Current()
	return filepath.Join(usr.HomeDir, seldonCfgFilepath)
}

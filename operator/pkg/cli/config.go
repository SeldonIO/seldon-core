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

const (
	KafkaConfigProtocolSSL          = "ssl"
	KafkaConfigProtocolSASLSSL      = "sasl_ssl"
	KafkaConfigProtocolSASLPlaintxt = "sasl_plaintxt"
)

type KafkaConfig struct {
	Bootstrap    string `json:"bootstrap,omitempty"`
	Namespace    string `json:"namespace,omitempty"`
	Protocol     string `json:"protocol,omitempty"`
	KeyPath      string `json:"keyPath,omitempty"`
	CrtPath      string `json:"crtPath,omitempty"`
	CaPath       string `json:"caPath,omitempty"`
	SaslUsername string `json:"saslUsername,omitempty"`
	SaslPassword string `json:"saslPassword,omitempty"`
}

// end config struct

const (
	seldonCfgFilepath = ".config/seldon/cli"
	configFilename    = "config.json"
)

func LoadSeldonCLIConfig() (*SeldonCLIConfig, error) {
	configs, err := LoadSeldonCLIConfigs()
	if err != nil {
		return nil, err
	}

	if path, ok := configs.getActiveConfigPath(); ok {
		return loadConfig(path)
	}
	return &SeldonCLIConfig{}, nil
}

func loadConfig(path string) (*SeldonCLIConfig, error) {
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

func (sc SeldonCLIConfig) print() {
	b, err := json.MarshalIndent(sc, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
}

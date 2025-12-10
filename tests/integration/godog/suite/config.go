/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package suite

import (
	"encoding/json"
	"fmt"
	"os"
)

type GodogConfig struct {
	Namespace   string    `json:"namespace"`
	LogLevel    string    `json:"log_level"`
	SkipCleanup bool      `json:"skip_cleanup"`
	Inference   Inference `json:"inference"`
}

type Inference struct {
	Host     string `json:"host"`
	HTTPPort uint   `json:"httpPort"`
	GRPCPort uint   `json:"grpcPort"`
	SSL      bool   `json:"ssl"`
}

func LoadConfig() (*GodogConfig, error) {
	configFile := os.Getenv("GODOG_CONFIG_FILE")
	if configFile == "" {
		configFile = "./godog-config.json"
	}

	if _, err := os.Stat(configFile); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", configFile)
		}
		return nil, fmt.Errorf("failed to open config file %s: %w", configFile, err)
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configFile, err)
	}

	var config GodogConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", configFile, err)
	}

	return &config, nil
}

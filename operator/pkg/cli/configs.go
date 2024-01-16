/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"
)

type SeldonCLIConfigs struct {
	Configs map[string]string `json:"configs,omitempty"`
	Active  *string           `json:"active,omitempty"`
}

const (
	configsFilename = "configs.json"
)

func LoadSeldonCLIConfigs() (*SeldonCLIConfigs, error) {
	path := getConfigsFile()
	_, err := os.Stat(path)
	if err != nil {
		return &SeldonCLIConfigs{
			Configs: map[string]string{},
		}, nil
	}

	byteValue, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to load CLI configs from %s: %s", path, err.Error())
	}
	cfg := SeldonCLIConfigs{}
	err = json.Unmarshal(byteValue, &cfg)
	if err != nil {
		return nil, fmt.Errorf("Failed to load CLI configs from %s: %s", path, err.Error())
	}
	if cfg.Configs == nil {
		cfg.Configs = map[string]string{}
	}
	return &cfg, nil
}

func getConfigsFile() string {
	return filepath.Join(getConfigDir(), configsFilename)
}

func (sc *SeldonCLIConfigs) getActiveConfigPath() (string, bool) {
	if sc.Active != nil {
		if path, ok := sc.Configs[*sc.Active]; ok {
			return path, true
		}
	}
	return "", false
}

func (sc *SeldonCLIConfigs) save() error {
	b, err := json.Marshal(sc)
	if err != nil {
		return err
	}
	path := getConfigsFile()
	return os.WriteFile(path, b, os.ModePerm)
}

func (sc *SeldonCLIConfigs) listConfigs() error {
	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', tabwriter.AlignRight)
	_, err := fmt.Fprintln(writer, "config\tpath\tactive")
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(writer, "------\t----\t------")
	if err != nil {
		return err
	}
	for k, v := range sc.Configs {
		if sc.Active != nil && k == *sc.Active {
			_, err = fmt.Fprintf(writer, "%s\t%s\t%s\n", k, v, "*")
			if err != nil {
				return err
			}
		} else {
			_, err = fmt.Fprintf(writer, "%s\t%s\t%s\n", k, v, "")
			if err != nil {
				return err
			}
		}
	}
	err = writer.Flush()
	if err != nil {
		return err
	}
	return nil
}

func (sc *SeldonCLIConfigs) listConfig(key string) error {
	if path, ok := sc.Configs[key]; ok {
		config, err := loadConfig(path)
		if err != nil {
			return err
		}
		config.print()
		return nil
	} else {
		return fmt.Errorf("Config key %s not found", key)
	}
}

func (sc *SeldonCLIConfigs) List(key string) error {
	if key == "" {
		return sc.listConfigs()
	} else {
		return sc.listConfig(key)
	}
}

func (sc *SeldonCLIConfigs) Add(key string, configPath string) error {
	sc.Configs[key] = configPath
	return sc.save()
}

func (sc *SeldonCLIConfigs) Remove(key string) error {
	if sc.Active != nil && *sc.Active == key {
		sc.Active = nil
	}
	delete(sc.Configs, key)
	return sc.save()
}

func (sc *SeldonCLIConfigs) Activate(key string) error {
	if _, ok := sc.Configs[key]; ok {
		sc.Active = &key
		return sc.save()
	} else {
		return fmt.Errorf("Config key %s not found", key)
	}
}

func (sc *SeldonCLIConfigs) Deactivate(key string) error {
	if _, ok := sc.Configs[key]; ok {
		sc.Active = nil
		return sc.save()
	} else {
		return fmt.Errorf("Config key %s not found", key)
	}
}

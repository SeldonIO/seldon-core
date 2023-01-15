/*
Copyright 2023 Seldon Technologies Ltd.

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

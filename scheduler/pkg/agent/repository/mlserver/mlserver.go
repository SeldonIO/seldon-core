/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package mlserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	seldontls "github.com/seldonio/seldon-core/components/tls/v2/pkg/tls"
)

const (
	mlserverConfigFilename = "model-settings.json"
	inferUriKey            = "infer_uri"
	explainerTypeKey       = "explainer_type"
	sslVerifyPath          = "ssl_verify_path"
)

type MLServerRepositoryHandler struct {
	logger log.FieldLogger
	SSL    bool
}

func NewMLServerRepositoryHandler(logger log.FieldLogger) *MLServerRepositoryHandler {
	protocol := seldontls.GetSecurityProtocolFromEnv(seldontls.EnvSecurityPrefixEnvoy)
	return &MLServerRepositoryHandler{
		logger: logger,
		SSL:    protocol == seldontls.SecurityProtocolSSL,
	}
}

type Settings struct {
	Debug                bool    `json:"debug,omitempty"`
	ModelRespositoryRoot *string `json:"model_repository_root,omitempty"`
	LoadModelsAtStartup  bool    `json:"load_models_at_startup,omitempty"`
	ServerName           *string
}

// MLServer model settings. Only a subset of fields included as needed
type ModelSettings struct {
	Name            string                 `json:"name"`
	Inputs          []ModelMetadataTensors `json:"inputs,omitempty"`
	Outputs         []ModelMetadataTensors `json:"outputs,omitempty"`
	Platform        string                 `json:"platform,omitempty"`
	Versions        []string               `json:"versions,omitempty"`
	ParallelWorkers *int                   `json:"parallel_workers,omitempty"`
	Implementation  string                 `json:"implementation,omitempty"`
	Parameters      *ModelParameters       `json:"parameters,omitempty"`
}

type ModelMetadataTensors struct {
	Name     string  `json:"name,omitempty"`
	Shape    []int64 `json:"shape,omitempty"`
	Datatype string  `json:"datatype,omitempty"`
}

// MLServer model parameters.
type ModelParameters struct {
	//URI where the model artifacts can be found.
	//This path must be either absolute or relative to where MLServer is running.
	Uri string `json:"uri,omitempty"`
	//Version of the model
	Version string `json:"version,omitempty"`
	//Format of the model (only available on certain runtimes).
	Format             string                 `json:"format,omitempty"`
	ContentType        string                 `json:"content_type,omitempty"`
	Extra              map[string]interface{} `json:"extra,omitempty"`
	EnvironmentTarball string                 `json:"environment_tarball,omitempty"`
}

// No need to update anything at top level for mlserver
func (m *MLServerRepositoryHandler) UpdateModelRepository(_ string, _ string, _ bool, _ string) error {
	return nil
}

func (m *MLServerRepositoryHandler) UpdateModelVersion(modelName string, version uint32, path string, modelSpec *scheduler.ModelSpec) error {
	versionStr := fmt.Sprintf("%d", version)
	settingsPath := filepath.Join(path, mlserverConfigFilename)
	if _, err := os.Stat(settingsPath); errors.Is(err, os.ErrNotExist) {
		// model-settings does not exist so try to create it from model spec
		err := createModelSettingsFile(path, modelSpec)
		if err != nil {
			return err
		}
	}
	//Modify model-settings
	err := m.updateNameAndVersion(path, modelName, versionStr)
	return err
}

// In order of precedence
// 1. a model repo with a default model at top level - always taken irrespective of whether version specified
// 2. a model repo with a matching version folder to version passed in
// 3. the highest numbered version folder
func (m *MLServerRepositoryHandler) FindModelVersionFolder(modelName string, version *uint32, path string) (string, bool, error) {
	logger := m.logger.WithField("func", "FindModelVersionFolder")
	// If there is just 1 model version we will take that irrespective of foldername or model-settings fields
	mvp, err := m.getDefaultModelSettingsPath(path)
	if err != nil {
		return "", false, err
	}
	if mvp == "" {
		if version != nil {
			mvp, err = m.findModelVersionInPath(path, *version)
			if err != nil {
				return "", false, err
			}
		} else {
			// Find highest numbered version folder
			mvp, err = m.findHighestVersionInPath(path)
			if err != nil {
				return "", false, err
			}
		}
	}
	// Return top level folder if we can't find a model-settings.json and are not looking for a specific version
	if mvp == "" {
		if version == nil {
			return path, false, nil
		}
		return "", false, fmt.Errorf("Failed to find an mlserver model-settings.json file for model in %s for %s for passed in version %v", path, modelName, version)
	} else {
		logger.Debugf("Found model settings for %s at %s for passed in version %v", modelName, mvp, version)
		return mvp, true, nil
	}
}

func (m *MLServerRepositoryHandler) updateNameAndVersion(path string, modelName string, version string) error {
	settingsPath := filepath.Join(path, mlserverConfigFilename)
	ms, err := m.loadModelSettingsFromFile(settingsPath)
	if err != nil {
		return err
	}
	ms.Name = modelName
	if ms.Parameters != nil {
		ms.Parameters.Version = version
	} else {
		ms.Parameters = &ModelParameters{Version: version}
	}
	data, err := json.Marshal(ms)
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, fs.ModePerm)
}

func (m *MLServerRepositoryHandler) SetExplainer(modelRepoPath string, explainerSpec *scheduler.ExplainerSpec, envoyHost string, envoyPort int) error {
	if explainerSpec != nil {
		workers := 0
		settingsPath := filepath.Join(modelRepoPath, mlserverConfigFilename)
		ms, err := m.loadModelSettingsFromFile(settingsPath)
		if err != nil {
			return err
		}
		//TODO: temporary fix for issue in mlserver with explainers
		ms.ParallelWorkers = &workers
		if ms.Parameters == nil {
			ms.Parameters = &ModelParameters{}
		}
		if ms.Parameters.Extra == nil {
			ms.Parameters.Extra = map[string]interface{}{}
		}
		ms.Parameters.Extra[explainerTypeKey] = &explainerSpec.Type
		scheme := "http"
		if m.SSL {
			scheme = "https"
			ms.Parameters.Extra[sslVerifyPath] = "/mnt/certs/ca.crt"
		}
		if explainerSpec.ModelRef != nil {
			inferUri := fmt.Sprintf("%s://%s:%d/v2/models/%s/infer", scheme, envoyHost, envoyPort, *explainerSpec.ModelRef)
			ms.Parameters.Extra[inferUriKey] = &inferUri
		} else if explainerSpec.PipelineRef != nil {
			inferUri := fmt.Sprintf("%s://%s:%d/v2/pipelines/%s/infer", scheme, envoyHost, envoyPort, *explainerSpec.PipelineRef)
			ms.Parameters.Extra[inferUriKey] = &inferUri
		}
		data, err := json.Marshal(ms)
		if err != nil {
			return err
		}
		return os.WriteFile(settingsPath, data, fs.ModePerm)
	}
	return nil
}

func (m *MLServerRepositoryHandler) SetExtraParameters(modelRepoPath string, parameters []*scheduler.ParameterSpec) error {
	settingsPath := filepath.Join(modelRepoPath, mlserverConfigFilename)
	ms, err := m.loadModelSettingsFromFile(settingsPath)
	if err != nil {
		return err
	}
	if ms.Parameters == nil {
		ms.Parameters = &ModelParameters{}
	}
	if ms.Parameters.Extra == nil {
		ms.Parameters.Extra = map[string]interface{}{}
	}
	for _, param := range parameters {
		ms.Parameters.Extra[param.Name] = param.Value
	}
	data, err := json.Marshal(ms)
	if err != nil {
		return err
	}
	return os.WriteFile(settingsPath, data, fs.ModePerm)
}

func (m *MLServerRepositoryHandler) loadModelSettingsFromFile(path string) (*ModelSettings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return m.loadModelSettingsFromBytes(data)
}

func (m *MLServerRepositoryHandler) loadModelSettingsFromBytes(data []byte) (*ModelSettings, error) {
	modelSettings := ModelSettings{}
	err := json.Unmarshal(data, &modelSettings)
	if err != nil {
		return nil, err
	}
	return &modelSettings, nil
}

func (m *MLServerRepositoryHandler) findModelVersionInPath(modelPath string, version uint32) (string, error) {
	var found []string
	versionStr := fmt.Sprintf("%d", version)
	err := filepath.Walk(modelPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Base(path) == mlserverConfigFilename {
			versionFolder := filepath.Base(filepath.Dir(path))
			// Just check folder name matches the desired version
			// We ignore the parameters file vesion settings for now to be consistent
			if versionFolder == versionStr {
				found = append(found, filepath.Dir(path))
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	switch len(found) {
	case 0:
		return "", nil
	case 1:
		return found[0], nil
	default:
		return "", fmt.Errorf("Found multiple folders with version %d %v", version, found)
	}
}

func (m *MLServerRepositoryHandler) getDefaultModelSettingsPath(modelPath string) (string, error) {
	var found []string
	err := filepath.Walk(modelPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && modelPath != path { //Don't descend into directories
			return filepath.SkipDir
		}
		if !info.IsDir() && filepath.Base(path) == mlserverConfigFilename {
			found = append(found, filepath.Dir(path))
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	// Only return result if we found one root model-settings.json
	if len(found) == 1 {
		return found[0], nil
	}
	return "", nil
}

func (m *MLServerRepositoryHandler) findHighestVersionInPath(modelPath string) (string, error) {
	highestVersionFolderNum := -1
	var highestVersionPath string
	err := filepath.Walk(modelPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Base(path) == mlserverConfigFilename {
			dir := filepath.Dir(path)
			if dir != modelPath { // Don't include top level model-settings.json file
				dirName := filepath.Base(dir)
				i, err := strconv.Atoi(dirName)
				if err != nil {
					return nil
				}
				if i > highestVersionFolderNum {
					highestVersionFolderNum = i
					highestVersionPath = dir
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if highestVersionFolderNum > -1 {
		return highestVersionPath, nil
	}
	return "", nil
}

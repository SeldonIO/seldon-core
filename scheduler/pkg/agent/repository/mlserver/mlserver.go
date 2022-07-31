package mlserver

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/seldonio/seldon-core/scheduler/apis/mlops/scheduler"
	log "github.com/sirupsen/logrus"
)

const (
	mlserverConfigFilename = "model-settings.json"
)

type MLServerRepositoryHandler struct {
	logger log.FieldLogger
}

func NewMLServerRepositoryHandler(logger log.FieldLogger) *MLServerRepositoryHandler {
	return &MLServerRepositoryHandler{
		logger: logger,
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
	Name            string           `json:"name"`
	Platform        string           `json:"platform,omitempty"`
	Versions        []string         `json:"versions,omitempty"`
	ParallelWorkers int              `json:"parallel_workers"`
	Implementation  string           `json:"implementation,omitempty"`
	Parameters      *ModelParameters `json:"parameters,omitempty"`
}

// MLServer model parameters.
type ModelParameters struct {
	//URI where the model artifacts can be found.
	//This path must be either absolute or relative to where MLServer is running.
	Uri string `json:"uri,omitempty"`
	//Version of the model
	Version string `json:"version,omitempty"`
	//Format of the model (only available on certain runtimes).
	Format      string          `json:"format,omitempty"`
	ContentType string          `json:"content_type,omitempty"`
	Extra       ExtraParameters `json:"extra,omitempty"`
}

type ExtraParameters struct {
	InferUri       *string                `json:"infer_uri,omitempty"`
	ExplainerType  *string                `json:"explainer_type,omitempty"`
	InitParameters map[string]interface{} `json:"init_parameters,omitempty"`
	PredictFn      *string                `json:"predict_fn,omitempty"`
	//TODO we should add headers here that MLServer can add to request
}

// No need to update anything at top level for mlserver
func (m *MLServerRepositoryHandler) UpdateModelRepository(_ string, _ string, _ string) error {
	return nil
}

func (m *MLServerRepositoryHandler) UpdateModelVersion(modelName string, version uint32, path string) error {
	versionStr := fmt.Sprintf("%d", version)
	//Modify model-settings
	err := m.UpdateNameAndVersion(path, modelName, versionStr)
	return err
}

// In order of precedence
// 1. a model repo with a default model at top level - always taken irrespective of whether version specified
// 2. a model repo with a matching version folder to version passed in
// 3. the highest numbered version folder
func (m *MLServerRepositoryHandler) FindModelVersionFolder(modelName string, version *uint32, path string) (string, error) {
	logger := m.logger.WithField("func", "FindModelVersionFolder")
	// If there is just 1 model version we will take that irrespective of foldername or model-settings fields
	mvp, err := m.getDefaultModelSettingsPath(path)
	if err != nil {
		return "", err
	}
	if mvp == "" {
		if version != nil {
			mvp, err = m.findModelVersionInPath(path, *version)
			if err != nil {
				return "", err
			}
		} else {
			// Find highest numbered version folder
			mvp, err = m.findHighestVersionInPath(path)
			if err != nil {
				return "", err
			}
		}
	}
	if mvp == "" {
		return "", fmt.Errorf("Failed to find an mlserver settings file model in %s for %s for passed in version %v", path, modelName, version)
	}
	logger.Debugf("Found model settings for %s at %s for passed in version %v", modelName, mvp, version)
	return mvp, nil
}

func (m *MLServerRepositoryHandler) UpdateNameAndVersion(path string, modelName string, version string) error {
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
		settingsPath := filepath.Join(modelRepoPath, mlserverConfigFilename)
		ms, err := m.loadModelSettingsFromFile(settingsPath)
		if err != nil {
			return err
		}
		if ms.Parameters == nil {
			ms.Parameters = &ModelParameters{}
		}
		ms.Parameters.Extra.ExplainerType = &explainerSpec.Type
		if explainerSpec.ModelRef != nil {
			inferUri := fmt.Sprintf("http://%s:%d/v2/models/%s/infer", envoyHost, envoyPort, *explainerSpec.ModelRef)
			ms.Parameters.Extra.InferUri = &inferUri
		} else if explainerSpec.PipelineRef != nil {
			inferUri := fmt.Sprintf("http://%s:%d/v2/pipelines/%s/infer", envoyHost, envoyPort, *explainerSpec.PipelineRef)
			ms.Parameters.Extra.InferUri = &inferUri
		}
		data, err := json.Marshal(ms)
		if err != nil {
			return err
		}
		return os.WriteFile(settingsPath, data, fs.ModePerm)
	}
	return nil
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

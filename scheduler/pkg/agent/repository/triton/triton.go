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

package triton

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	copy2 "github.com/otiai10/copy"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
	pb "github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/repository/triton/config"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/prototext"
)

const (
	TritonConfigFile = "config.pbtxt"
)

type TritonRepositoryHandler struct {
	logger log.FieldLogger
}

func NewTritonRepositoryHandler(logger log.FieldLogger) *TritonRepositoryHandler {
	return &TritonRepositoryHandler{logger: logger.WithField("name", "TritonRepositoryHandler")}
}

// Copy config file to dst if it doesn't exist and set modelName
func (t *TritonRepositoryHandler) UpdateModelRepository(modelName string, versionPath, modelRepoPath string) error {
	configFilePathRepo := filepath.Join(modelRepoPath, TritonConfigFile)
	configFilePathFromVersion := filepath.Join(filepath.Dir(versionPath), TritonConfigFile)
	if _, err := os.Stat(configFilePathFromVersion); err != nil {
		return t.createConfigFileWithName(modelName, configFilePathRepo)
	}
	// Always copy config.pbtxt overwriting existing as we may have changes in configuration
	err := copy2.Copy(configFilePathFromVersion, configFilePathRepo)
	if err != nil {
		return err
	}
	return t.updateModelNameInConfig(modelName, configFilePathRepo)
}

func saveConfigFile(path string, config *pb.ModelConfig) error {
	data, err := prototext.Marshal(config)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, fs.ModePerm)
}

func (t *TritonRepositoryHandler) createConfigFileWithName(modelName string, path string) error {
	config := pb.ModelConfig{}
	config.Name = modelName
	return saveConfigFile(path, &config)
}

func (t *TritonRepositoryHandler) updateModelNameInConfig(modelName string, path string) error {
	s, err := t.loadConfigFromFile(path)
	if err != nil {
		return err
	}
	s.Name = modelName
	return saveConfigFile(path, s)
}

func (t *TritonRepositoryHandler) FindModelVersionFolder(_ string, version *uint32, path string) (string, error) {
	if version != nil {
		return t.findModelVersionInPath(path, *version)
	} else {
		return t.findHighestVersionInPath(path)
	}
}

// We don't need to change Triton model folders
func (t *TritonRepositoryHandler) UpdateModelVersion(_ string, _ uint32, _ string, _ *scheduler.ModelSpec) error {
	return nil
}

func (t *TritonRepositoryHandler) findModelVersionInPath(modelPath string, version uint32) (string, error) {
	var found []string
	versionStr := fmt.Sprintf("%d", version)
	err := filepath.Walk(modelPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			// if first level directory with matching version
			if filepath.Base(path) == versionStr &&
				filepath.Dir(path) == modelPath {
				found = append(found, path)
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	switch len(found) {
	case 0:
		return "", fmt.Errorf("Failed to find requested version %d in path %s", version, modelPath)
	case 1:
		return found[0], nil
	default:
		return "", fmt.Errorf("Found multiple folders with version %d %v", version, found)
	}
}

func (m *TritonRepositoryHandler) findHighestVersionInPath(modelPath string) (string, error) {
	highestVersionFolderNum := -1
	var highestVersionPath string
	err := filepath.Walk(modelPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != modelPath {
			dirName := filepath.Base(path)
			i, err := strconv.Atoi(dirName)
			if err != nil {
				return nil
			}
			if i > highestVersionFolderNum {
				highestVersionFolderNum = i
				highestVersionPath = path
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if highestVersionFolderNum > 0 { // Triton versions need to be >0
		return highestVersionPath, nil
	}
	//return "", fmt.Errorf("Failed to find triton model version folder in path %s", modelPath)
	//If we don't find a version assume folder is the default version we want
	return modelPath, nil
}

func (t *TritonRepositoryHandler) loadConfigFromFile(path string) (*pb.ModelConfig, error) {
	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return t.loadConfigFromBytes(dat)
}

func (t *TritonRepositoryHandler) loadConfigFromBytes(dat []byte) (*pb.ModelConfig, error) {
	config := pb.ModelConfig{}
	err := prototext.Unmarshal(dat, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func (t *TritonRepositoryHandler) SetExplainer(modelRepoPath string, explainerSpec *scheduler.ExplainerSpec, envoyHost string, envoyPort int) error {
	return nil
}

func (t *TritonRepositoryHandler) SetExtraParameters(modelRepoPath string, parameters []*scheduler.ParameterSpec) error {
	return nil
}

/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package triton

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	copy2 "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"
	"google.golang.org/protobuf/encoding/prototext"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/agent"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	pb "github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/repository/triton/config"
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

func copyNonConfigFilesToModelRepo(src string, dst string) error {
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && src != path { // Don't descend into directories
			return filepath.SkipDir
		}
		// Copy non- config.pbtxt files to dst folder
		if !info.IsDir() && filepath.Base(path) != TritonConfigFile {
			err := copy2.Copy(path, filepath.Join(dst, filepath.Base(path)))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return err
}

// Copy config file to dst creating if required and setting the correct model name
func (t *TritonRepositoryHandler) UpdateModelRepository(modelName string, rclonePath string, isVersionFolder bool, modelRepoPath string) error {
	configFilePathRepo := filepath.Join(modelRepoPath, TritonConfigFile)
	var configFilePath string
	if isVersionFolder {
		t.logger.Infof("Copy files from versioned folder %s to %s", filepath.Dir(rclonePath), modelRepoPath)
		// copy all non-config.pbtxt files from folder above version to repo folder
		err := copyNonConfigFilesToModelRepo(filepath.Dir(rclonePath), modelRepoPath)
		if err != nil {
			return err
		}
		// look for config.pbtxt in folder above of current folder if this is a version folder
		configFilePath = filepath.Join(filepath.Dir(rclonePath), TritonConfigFile)
		if _, err := os.Stat(configFilePath); err != nil {
			// Create basic config.pbtxt as we didn't find it
			return t.createConfigFileWithName(modelName, configFilePathRepo)
		}
	} else {
		t.logger.Infof("Copy files from non-versioned folder %s to %s", rclonePath, modelRepoPath)
		// copy all non-config.pbtxt files from folder to repo folder
		err := copyNonConfigFilesToModelRepo(rclonePath, modelRepoPath)
		if err != nil {
			return err
		}
		// look for config.pbtxt in same folder as model artifacts
		configFilePath = filepath.Join(rclonePath, TritonConfigFile)
		if _, err := os.Stat(configFilePath); err != nil {
			// Create basic config.pbtxt as we didn't find it
			return t.createConfigFileWithName(modelName, configFilePathRepo)
		}
	}

	// If we are here we found a config.pbtxt
	// Always copy config.pbtxt overwriting existing as we may have changes in configuration
	err := copy2.Copy(configFilePath, configFilePathRepo)
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

func (t *TritonRepositoryHandler) FindModelVersionFolder(_ string, version *uint32, path string) (string, bool, error) {
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

func (t *TritonRepositoryHandler) findModelVersionInPath(modelPath string, version uint32) (string, bool, error) {
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
		return "", false, err
	}
	switch len(found) {
	case 0:
		return "", false, fmt.Errorf("Failed to find requested version %d in path %s", version, modelPath)
	case 1:
		return found[0], true, nil
	default:
		return "", false, fmt.Errorf("Found multiple folders with version %d %v", version, found)
	}
}

func (m *TritonRepositoryHandler) findHighestVersionInPath(modelPath string) (string, bool, error) {
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
		return "", false, err
	}
	if highestVersionFolderNum > 0 { // Triton versions need to be >0
		return highestVersionPath, true, nil
	}
	// return modelPath if no version found
	return modelPath, false, nil
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

func (t *TritonRepositoryHandler) GetModelRuntimeInfo(path string) (*agent.ModelRuntimeInfo, error) {
	configPath := filepath.Join(path, TritonConfigFile)
	tritonConfig, err := t.loadConfigFromFile(configPath)
	tritonRuntimeInfo := &agent.ModelRuntimeInfo_Triton{
		Triton: &agent.TritonModelConfig{
			Cpu: []*agent.TritonCPU{
				{InstanceCount: 1},
			},
		},
	}
	if err == nil {
		instanceGroups := tritonConfig.InstanceGroup
		if len(instanceGroups) > 0 {
			var instanceCount int32 = 0
			backend := tritonConfig.Backend
			for _, instanceGroup := range instanceGroups {
				// only take the value from the first KIND_CPU that's found
				if instanceGroup.Kind == pb.ModelInstanceGroup_KIND_CPU && instanceCount == 0 {
					if instanceGroup.Count < 1 {
						if strings.ToLower(backend) == "tensorflow" || strings.ToLower(backend) == "onnxruntime" {
							instanceCount = 2
						} else {
							instanceCount = 1
						}
					} else {
						instanceCount += instanceGroup.Count
					}
				}
			}
			// Default to 1 if no KIND_CPU is found, as KIND_GPU is currently not supported
			if instanceCount < 1 {
				instanceCount = 1
			}
			tritonRuntimeInfo.Triton.Cpu = []*agent.TritonCPU{{InstanceCount: uint32(instanceCount)}}
		}
	}
	return &agent.ModelRuntimeInfo{ModelRuntimeInfo: tritonRuntimeInfo}, nil
}

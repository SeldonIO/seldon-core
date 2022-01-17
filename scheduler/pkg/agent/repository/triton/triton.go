package triton

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	copy2 "github.com/otiai10/copy"
	pb "github.com/seldonio/seldon-core/scheduler/pkg/agent/repository/triton/config"
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
		return fmt.Errorf("Can't find %s", configFilePathFromVersion)
	}
	// Always copy config.pbtxt overwriting existing as we may have changes in configuration
	err := copy2.Copy(configFilePathFromVersion, configFilePathRepo)
	if err != nil {
		return err
	}
	return t.updateModelNameInConfig(modelName, configFilePathRepo)
}

func (t *TritonRepositoryHandler) updateModelNameInConfig(modelName string, path string) error {
	s, err := t.loadConfigFromFile(path)
	if err != nil {
		return err
	}
	s.Name = modelName
	data, err := prototext.Marshal(s)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, fs.ModePerm)
}

func (t *TritonRepositoryHandler) FindModelVersionFolder(_ string, version *uint32, path string) (string, error) {
	if version != nil {
		return t.findModelVersionInPath(path, *version)
	} else {
		return t.findHighestVersionInPath(path)
	}
}

// We don't need to change Triton model folders
func (t *TritonRepositoryHandler) UpdateModelVersion(modelName string, version uint32, path string) error {
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
		return "", nil
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
	return "", nil
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

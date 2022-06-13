package repository

import (
	"fmt"
	"os"
	"path/filepath"

	copy2 "github.com/otiai10/copy"

	"github.com/seldonio/seldon-core/scheduler/pkg/agent/rclone"
	log "github.com/sirupsen/logrus"
)

type ModelRepositoryHandler interface {
	FindModelVersionFolder(modelName string, version *uint32, path string) (string, error)
	UpdateModelVersion(modelName string, version uint32, path string) error
	UpdateModelRepository(modelName string, versionPath, modelRepoPath string) error
}

type ModelRepository interface {
	DownloadModelVersion(modelName string, version uint32, artifactVersion *uint32, srcUri string, config []byte) (*string, error)
	RemoveModelVersion(modelName string) error
	Ready() error
}

type V2ModelRepository struct {
	logger                 log.FieldLogger
	rcloneClient           *rclone.RCloneClient
	repoPath               string
	modelrepositoryHandler ModelRepositoryHandler
}

func NewModelRepository(logger log.FieldLogger,
	rcloneClient *rclone.RCloneClient,
	repoPath string,
	modelRepositoryHandler ModelRepositoryHandler) *V2ModelRepository {
	return &V2ModelRepository{
		logger:                 logger.WithField("Name", "V2ModelRepository"),
		rcloneClient:           rcloneClient,
		repoPath:               repoPath,
		modelrepositoryHandler: modelRepositoryHandler,
	}
}

func (r *V2ModelRepository) DownloadModelVersion(modelName string, version uint32, artifactVersion *uint32, srcUri string, config []byte) (*string, error) {
	logger := r.logger.WithField("func", "DownloadModelVersion")
	logger.Debugf("running with model %s:%d srcUri %s", modelName, version, srcUri)

	// Run rclone copy sync
	rclonePath, err := r.rcloneClient.Copy(modelName, srcUri, config)
	if err != nil {
		return nil, err
	}

	// Find the version folder we want
	modelVersionFolder, err := r.modelrepositoryHandler.FindModelVersionFolder(modelName, artifactVersion, rclonePath)
	if err != nil {
		return nil, err
	}
	logger.Debugf("Found model %s:%d artifactVersion %d for %s at %s ", modelName, version, artifactVersion, srcUri, modelVersionFolder)

	// Create model directory if needed in model repo
	modelPathInRepo := filepath.Join(r.repoPath, modelName)
	// Ensure path exists
	err = os.MkdirAll(modelPathInRepo, os.ModePerm)
	if err != nil {
		return nil, err
	}

	// Copy version folder to final location in model repo
	// Note there is also fileutil.CopyDirs() if we don't want dependency on github.com/otiai10/copy
	versionStr := fmt.Sprintf("%d", version)
	modelVersionPathInRepo := filepath.Join(modelPathInRepo, versionStr)
	opt := copy2.Options{
		OnDirExists: func(sr, dst string) copy2.DirExistsAction { return copy2.Replace },
		Sync:        true,
	}
	err = copy2.Copy(modelVersionFolder, modelVersionPathInRepo, opt)
	if err != nil {
		return nil, err
	}

	// Update model version in repo
	err = r.modelrepositoryHandler.UpdateModelVersion(modelName, version, modelVersionPathInRepo)
	if err != nil {
		return nil, err
	}

	// Update global model configuration
	err = r.modelrepositoryHandler.UpdateModelRepository(modelName, modelVersionFolder, modelPathInRepo)
	if err != nil {
		return nil, err
	}

	return &modelVersionFolder, nil
}

// Remove version folder and return number of remaining versions calculated as found model-settings files
func (r *V2ModelRepository) RemoveModelVersion(modelName string) error {
	modelPath := filepath.Join(r.repoPath, modelName)
	err := os.RemoveAll(modelPath)
	if err != nil {
		return err
	}
	return nil
}

func (r *V2ModelRepository) Ready() error {
	return r.rcloneClient.Ready()
}

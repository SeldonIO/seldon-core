/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	copy2 "github.com/otiai10/copy"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/rclone"
)

type ModelRepositoryHandler interface {
	FindModelVersionFolder(modelName string, version *uint32, path string) (string, bool, error)
	UpdateModelVersion(modelName string, version uint32, path string, modelSpec *scheduler.ModelSpec) error
	UpdateModelRepository(modelName string, path string, isVersionFolder bool, modelRepoPath string) error
	SetExplainer(modelRepoPath string, explainerSpec *scheduler.ExplainerSpec, envoyHost string, envoyPort int) error
	SetExtraParameters(modelRepoPath string, parameters []*scheduler.ParameterSpec) error
}

type ModelRepository interface {
	DownloadModelVersion(modelName string, version uint32, modelSpec *scheduler.ModelSpec, config []byte) (*string, error)
	RemoveModelVersion(modelName string) error
	Ready() error
}

type V2ModelRepository struct {
	logger                 log.FieldLogger
	rcloneClient           *rclone.RCloneClient
	repoPath               string
	modelrepositoryHandler ModelRepositoryHandler
	envoyHost              string
	envoyPort              int
}

func NewModelRepository(logger log.FieldLogger,
	rcloneClient *rclone.RCloneClient,
	repoPath string,
	modelRepositoryHandler ModelRepositoryHandler,
	envoyHost string,
	envoyPort int) *V2ModelRepository {
	return &V2ModelRepository{
		logger:                 logger.WithField("Name", "V2ModelRepository"),
		rcloneClient:           rcloneClient,
		repoPath:               repoPath,
		modelrepositoryHandler: modelRepositoryHandler,
		envoyHost:              envoyHost,
		envoyPort:              envoyPort,
	}
}

func (r *V2ModelRepository) DownloadModelVersion(
	modelName string,
	version uint32,
	modelSpec *scheduler.ModelSpec,
	config []byte,
) (modelVersionFolderPtr *string, err error) {
	logger := r.logger.WithField("func", "DownloadModelVersion")

	// Setup key vars
	artifactVersion := modelSpec.ArtifactVersion
	srcUri := modelSpec.Uri
	explainerSpec := modelSpec.GetExplainer()
	parameters := modelSpec.GetParameters()

	logger.Debugf("running with model %s:%d srcUri %s", modelName, version, srcUri)

	// Run rclone copy sync
	rclonePath, err := r.rcloneClient.Copy(modelName, srcUri, config)
	if err != nil {
		return nil, err
	}
	defer func() {
		// Once the model artifact has been downloaded via rclone, ensure that we clean it up,
		// even in the presence of errors (e.g. if the PVC doesn't have enough space to copy the
		// artifact into the inference server's model repository path).
		err_purge := r.rcloneClient.PurgeLocal(rclonePath)
		if err_purge != nil {
			err = errors.Join(err, err_purge)
		}
	}()

	// Find the version folder we want
	modelVersionFolder, foundVersionFolder, err := r.modelrepositoryHandler.FindModelVersionFolder(
		modelName,
		artifactVersion,
		rclonePath,
	)
	if err != nil {
		return nil, err
	}

	logger.Debugf(
		"Found model %s:%d artifactVersion %d for %s at %s ",
		modelName,
		version,
		artifactVersion,
		srcUri,
		modelVersionFolder,
	)

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
	err = r.modelrepositoryHandler.UpdateModelVersion(
		modelName,
		version,
		modelVersionPathInRepo,
		modelSpec,
	)
	if err != nil {
		return nil, err
	}

	// Update details for blackbox explainer
	if explainerSpec != nil {
		err = r.modelrepositoryHandler.SetExplainer(
			modelVersionPathInRepo,
			explainerSpec,
			r.envoyHost,
			r.envoyPort,
		)
		if err != nil {
			return nil, err
		}
	}

	// Set init parameters inside model
	err = r.modelrepositoryHandler.SetExtraParameters(modelVersionPathInRepo, parameters)
	if err != nil {
		return nil, err
	}

	// Update global model configuration
	err = r.modelrepositoryHandler.UpdateModelRepository(
		modelName,
		modelVersionFolder,
		foundVersionFolder,
		modelPathInRepo,
	)
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

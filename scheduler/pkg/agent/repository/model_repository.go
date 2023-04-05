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

package repository

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	copy2 "github.com/otiai10/copy"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/rclone"
	log "github.com/sirupsen/logrus"
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
) (*string, error) {
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

	// Purge rclone path now we have loaded successfully
	err = r.rcloneClient.PurgeLocal(rclonePath)
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

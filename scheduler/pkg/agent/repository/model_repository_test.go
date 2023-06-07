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
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/rclone"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/repository/mlserver"
)

func createTestRCloneMockResponders(host string, port int, status int, body string) {
	httpmock.RegisterResponder("POST", fmt.Sprintf("=~http://%s:%d/", host, port),
		httpmock.NewStringResponder(status, body))
}

func createFakeRcloneClient(status int, path string) *rclone.RCloneClient {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	host := "rclone-server"
	port := 5572
	r := rclone.NewRCloneClient(host, port, path, logger, "default")
	createTestRCloneMockResponders(host, port, status, "")
	return r
}

func TestDownloadModelVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		folders      map[string]*mlserver.ModelSettings
		root         *mlserver.ModelSettings
		modelSpec    *scheduler.ModelSpec
		modelName    string
		modelVersion uint32
		chosenFolder string
		error        bool
	}

	getArtifactVersion := func(version uint32) *uint32 {
		return &version
	}

	tests := []test{
		{
			name: "Simple",
			folders: map[string]*mlserver.ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &mlserver.ModelParameters{
						Version: "1",
					},
				},
			},
			modelSpec: &scheduler.ModelSpec{
				Uri:             "gs://model",
				ArtifactVersion: getArtifactVersion(1),
			},
			modelName:    "foo",
			modelVersion: 1,
			chosenFolder: "1",
		},
		{
			name: "ArtifactVersionMissing",
			folders: map[string]*mlserver.ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &mlserver.ModelParameters{
						Version: "1",
					},
				},
			},
			modelSpec: &scheduler.ModelSpec{
				Uri: "gs://model",
			},
			modelName:    "foo",
			modelVersion: 1,
			chosenFolder: "1",
		},
		{
			name: "HighestVersion",
			folders: map[string]*mlserver.ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &mlserver.ModelParameters{
						Version: "1",
					},
				},
				"2": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &mlserver.ModelParameters{
						Version: "2",
					},
				},
			},
			modelSpec: &scheduler.ModelSpec{
				Uri: "gs://model",
			},
			modelName:    "foo",
			modelVersion: 1,
			chosenFolder: "2",
		},
		{
			name: "ArtifactVersionDifferentfromVersion",
			folders: map[string]*mlserver.ModelSettings{
				"2": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &mlserver.ModelParameters{
						Version: "2",
					},
				},
			},
			modelSpec: &scheduler.ModelSpec{
				Uri:             "gs://model",
				ArtifactVersion: getArtifactVersion(2),
			},
			modelName:    "foo",
			modelVersion: 1,
			chosenFolder: "2",
		},
		{
			name: "SettingsContradictsFolder",
			folders: map[string]*mlserver.ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &mlserver.ModelParameters{
						Version: "2",
					},
				},
			},
			modelSpec: &scheduler.ModelSpec{
				Uri:             "gs://model",
				ArtifactVersion: getArtifactVersion(1),
			},
			modelName:    "foo",
			modelVersion: 1,
			chosenFolder: "1",
		},
		{
			name:    "VersionInRootVersionMatches",
			folders: map[string]*mlserver.ModelSettings{},
			root: &mlserver.ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &mlserver.ModelParameters{
					Version: "1",
				},
			},
			modelSpec: &scheduler.ModelSpec{
				Uri:             "gs://model",
				ArtifactVersion: getArtifactVersion(1),
			},
			modelName:    "foo",
			modelVersion: 1,
		},
		{
			name:    "VersionInRootVersionMisMatches",
			folders: map[string]*mlserver.ModelSettings{},
			root: &mlserver.ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &mlserver.ModelParameters{
					Version: "2",
				},
			},
			modelSpec: &scheduler.ModelSpec{
				Uri:             "gs://model",
				ArtifactVersion: getArtifactVersion(1),
			},
			modelName:    "foo",
			modelVersion: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			httpmock.Activate()
			defer httpmock.DeactivateAndReset()
			rclonePath := t.TempDir()
			for folderName, ms := range test.folders {
				hash, err := rclone.CreateRcloneModelHash(test.modelName, test.modelSpec.Uri)
				g.Expect(err).To(BeNil())
				folderPath := filepath.Join(rclonePath, fmt.Sprintf("%d/%s", hash, folderName))
				err = os.MkdirAll(folderPath, fs.ModePerm)
				g.Expect(err).To(BeNil())
				data, err := json.Marshal(ms)
				g.Expect(err).To(BeNil())
				settingsFilePath := filepath.Join(folderPath, "model-settings.json")
				err = os.WriteFile(settingsFilePath, data, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			if test.root != nil {
				data, err := json.Marshal(test.root)
				g.Expect(err).To(BeNil())
				hash, err := rclone.CreateRcloneModelHash(test.modelName, test.modelSpec.Uri)
				g.Expect(err).To(BeNil())
				folderPath := filepath.Join(rclonePath, fmt.Sprintf("%d", hash))
				err = os.MkdirAll(folderPath, fs.ModePerm)
				g.Expect(err).To(BeNil())
				settingsFilePath := filepath.Join(folderPath, "model-settings.json")
				err = os.WriteFile(settingsFilePath, data, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			logger := log.New()
			rcloneClient := createFakeRcloneClient(200, rclonePath)
			modelRepoPath := t.TempDir()
			mr := NewModelRepository(logger, rcloneClient, modelRepoPath, mlserver.NewMLServerRepositoryHandler(logger), "0.0.0.0", 9000)
			chosenFolder, err := mr.DownloadModelVersion(test.modelName, test.modelVersion, test.modelSpec, nil)
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				if test.chosenFolder != "" {
					g.Expect(filepath.Base(*chosenFolder)).To(Equal(test.chosenFolder))
				}
			}
		})
	}
}

func TestRemoveModelVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		folders   map[string]*mlserver.ModelSettings
		modelName string
	}

	tests := []test{
		{
			name: "Simple",
			folders: map[string]*mlserver.ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &mlserver.ModelParameters{
						Version: "1",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir()
			for folderName, ms := range test.folders {
				folderPath := filepath.Join(path, folderName)
				err := os.Mkdir(folderPath, fs.ModePerm)
				g.Expect(err).To(BeNil())
				data, err := json.Marshal(ms)
				g.Expect(err).To(BeNil())
				settingsFilePath := filepath.Join(folderPath, "model-settings.json")
				err = os.WriteFile(settingsFilePath, data, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			logger := log.New()
			logger.SetLevel(log.DebugLevel)
			mr := NewModelRepository(logger, nil, path, nil, "0.0.0.0", 9000)
			err := mr.RemoveModelVersion(test.modelName)
			g.Expect(err).To(BeNil())
			modelPath := filepath.Join(path, test.modelName)
			_, err = os.Stat(modelPath)
			g.Expect(err).ToNot(BeNil())
			g.Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
		})
	}
}

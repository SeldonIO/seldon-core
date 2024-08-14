/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/rclone"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/repository/mlserver"
)

const (
	RcloneHost = "rclone-server"
	RclonePort = 5572
)

func createTestRCloneMockResponders(host string, port int, responder httpmock.Responder) {
	httpmock.RegisterResponder("POST", fmt.Sprintf("=~http://%s:%d/", host, port), responder)
}

func createFakeRcloneClient(path string, responder httpmock.Responder) *rclone.RCloneClient {
	logger := log.New()
	log.SetLevel(log.DebugLevel)
	config, _ := config.NewAgentConfigHandler("", "", logger, nil)
	r := rclone.NewRCloneClient(RcloneHost, RclonePort, path, logger, "default", config)
	createTestRCloneMockResponders(RcloneHost, RclonePort, responder)
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
		downloadFail bool
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
			name: "DownloadFail",
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
			downloadFail: true,
			chosenFolder: "1",
			error:        true,
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
		{
			name: "ArtifactVersionMismatch",
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
				ArtifactVersion: getArtifactVersion(1),
			},
			modelName:    "foo",
			modelVersion: 2,
			error:        true,
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
			responder := func(req *http.Request) (*http.Response, error) {
				if req.URL.Path == rclone.RcloneSyncCopyPath && test.downloadFail {
					return httpmock.NewStringResponse(404, ""), nil
				}
				return httpmock.NewStringResponse(200, ""), nil
			}
			rcloneClient := createFakeRcloneClient(rclonePath, responder)

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

			httpInfo := httpmock.GetCallCountInfo()
			// Check that rclonePath was cleaned-up even in the presence of errors during
			// DownloadModelVersion, with the exception of the download itself failing
			purgeUrl := fmt.Sprintf("POST http://%s:%d%s", RcloneHost, RclonePort, rclone.RcloneOperationsPurgePath)
			expectedPurgeCount := 1
			if test.downloadFail {
				expectedPurgeCount = 0
			}
			g.Expect(httpInfo[purgeUrl]).To(Equal(expectedPurgeCount),
				"Expected %d calls to %s; mock http calls map: %v",
				expectedPurgeCount, purgeUrl, httpInfo)
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

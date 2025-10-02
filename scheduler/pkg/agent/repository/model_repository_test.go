/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/aws/smithy-go/ptr"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
	"go.uber.org/mock/gomock"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/config"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/filemanager/mocks"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/rclone"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/repository/mlserver"
	mocks2 "github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/repository/mocks"
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

func TestDownloadModelBackoffRetry(t *testing.T) {
	g := NewGomegaWithT(t)

	repoPath, err := os.MkdirTemp(os.TempDir(), "")
	g.Expect(err).To(BeNil())

	type expect struct {
		error  bool
		folder string
	}

	type test struct {
		name         string
		modelSpec    *scheduler.ModelSpec
		modelName    string
		modelVersion uint32
		config       []byte
		setupMocks   func(fileManager *mocks.MockFileManager, modelRepo *mocks2.MockModelRepositoryHandler, t *test)
		expect       expect
	}

	tests := []test{
		{
			name: "success - no retry required",
			modelSpec: &scheduler.ModelSpec{
				Uri:             "/some-model-uri",
				ArtifactVersion: ptr.Uint32(1234),
			},
			modelName:    "my-model",
			modelVersion: 1,
			config:       []byte("{}"),
			setupMocks: func(fileManager *mocks.MockFileManager, modelRepo *mocks2.MockModelRepositoryHandler, t *test) {
				tempModelDir, err := os.MkdirTemp(os.TempDir(), "")
				g.Expect(err).To(BeNil())
				rClonePath := path.Join(tempModelDir, "rclone-client")

				fileManager.EXPECT().Copy(gomock.Any(), t.modelName, t.modelSpec.Uri, t.config).
					Return(rClonePath, nil)

				fileManager.EXPECT().PurgeLocal(rClonePath).Return(nil)

				modelVersionFolder, err := os.MkdirTemp(os.TempDir(), "")
				t.expect.folder = modelVersionFolder
				g.Expect(err).To(BeNil())

				modelPath := filepath.Join(repoPath, t.modelName, fmt.Sprintf("%d", t.modelVersion))

				modelRepo.EXPECT().FindModelVersionFolder(t.modelName, t.modelSpec.ArtifactVersion, rClonePath).Return(modelVersionFolder, true, nil)
				modelRepo.EXPECT().UpdateModelVersion(t.modelName, t.modelVersion, modelPath, t.modelSpec).Return(nil)
				modelRepo.EXPECT().SetExtraParameters(modelPath, nil).Return(nil)
				modelRepo.EXPECT().UpdateModelRepository(t.modelName, modelVersionFolder, true, filepath.Join(repoPath, t.modelName)).Return(nil)
			},
			expect: expect{
				error: false,
			},
		},
		{
			name: "success - one Copy retry required",
			modelSpec: &scheduler.ModelSpec{
				Uri:             "/some-model-uri",
				ArtifactVersion: ptr.Uint32(1234),
			},
			modelName:    "my-model",
			modelVersion: 1,
			config:       []byte("{}"),
			setupMocks: func(fileManager *mocks.MockFileManager, modelRepo *mocks2.MockModelRepositoryHandler, t *test) {
				tempModelDir, err := os.MkdirTemp(os.TempDir(), "")
				g.Expect(err).To(BeNil())
				rClonePath := path.Join(tempModelDir, "rclone-client")

				// first attempts errors
				fileManager.EXPECT().Copy(gomock.Any(), t.modelName, t.modelSpec.Uri, t.config).
					Return(rClonePath, &url.Error{})

				// second attempt successful
				fileManager.EXPECT().Copy(gomock.Any(), t.modelName, t.modelSpec.Uri, t.config).
					Return(rClonePath, nil)

				fileManager.EXPECT().PurgeLocal(rClonePath).Return(nil)

				modelVersionFolder, err := os.MkdirTemp(os.TempDir(), "")
				t.expect.folder = modelVersionFolder
				g.Expect(err).To(BeNil())

				modelPath := filepath.Join(repoPath, t.modelName, fmt.Sprintf("%d", t.modelVersion))

				modelRepo.EXPECT().FindModelVersionFolder(t.modelName, t.modelSpec.ArtifactVersion, rClonePath).Return(modelVersionFolder, true, nil)
				modelRepo.EXPECT().UpdateModelVersion(t.modelName, t.modelVersion, modelPath, t.modelSpec).Return(nil)
				modelRepo.EXPECT().SetExtraParameters(modelPath, nil).Return(nil)
				modelRepo.EXPECT().UpdateModelRepository(t.modelName, modelVersionFolder, true, filepath.Join(repoPath, t.modelName)).Return(nil)
			},
			expect: expect{
				error: false,
			},
		},
		{
			name: "failure - max retries exceeded",
			modelSpec: &scheduler.ModelSpec{
				Uri:             "/some-model-uri",
				ArtifactVersion: ptr.Uint32(1234),
			},
			modelName:    "my-model",
			modelVersion: 1,
			config:       []byte("{}"),
			setupMocks: func(fileManager *mocks.MockFileManager, modelRepo *mocks2.MockModelRepositoryHandler, t *test) {
				tempModelDir, err := os.MkdirTemp(os.TempDir(), "")
				g.Expect(err).To(BeNil())
				rClonePath := path.Join(tempModelDir, "rclone-client")

				// 2 or more attempts fail to download
				fileManager.EXPECT().Copy(gomock.Any(), t.modelName, t.modelSpec.Uri, t.config).
					Return(rClonePath, &url.Error{}).MinTimes(2)
			},
			expect: expect{
				error: true,
			},
		},
		{
			name: "failure - don't retry, non url.Error is returned",
			modelSpec: &scheduler.ModelSpec{
				Uri:             "/some-model-uri",
				ArtifactVersion: ptr.Uint32(1234),
			},
			modelName:    "my-model",
			modelVersion: 1,
			config:       []byte("{}"),
			setupMocks: func(fileManager *mocks.MockFileManager, modelRepo *mocks2.MockModelRepositoryHandler, t *test) {
				tempModelDir, err := os.MkdirTemp(os.TempDir(), "")
				g.Expect(err).To(BeNil())
				rClonePath := path.Join(tempModelDir, "rclone-client")

				fileManager.EXPECT().Copy(gomock.Any(), t.modelName, t.modelSpec.Uri, t.config).
					Return(rClonePath, errors.New("some error"))
			},
			expect: expect{
				error: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			logger := log.New()

			tempModelDir, err := os.MkdirTemp(os.TempDir(), "")
			g.Expect(err).To(BeNil())
			test.modelSpec.Uri = tempModelDir + "/" + test.modelName

			rcloneMock := mocks.NewMockFileManager(ctrl)
			modelRepoHandlerMock := mocks2.NewMockModelRepositoryHandler(ctrl)

			test.setupMocks(rcloneMock, modelRepoHandlerMock, &test)

			mr := NewModelRepository(logger, rcloneMock,
				repoPath, modelRepoHandlerMock, "0.0.0.0", 9000, 2*time.Second)

			folder, err := mr.DownloadModelVersion(context.Background(), test.modelName, test.modelVersion, test.modelSpec, test.config)
			if test.expect.error {
				g.Expect(err).To(HaveOccurred())
				return
			}

			g.Expect(*folder).To(Equal(test.expect.folder))
		})
	}
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
			mr := NewModelRepository(logger, rcloneClient, modelRepoPath,
				mlserver.NewMLServerRepositoryHandler(logger), "0.0.0.0", 9000, time.Second)
			chosenFolder, err := mr.DownloadModelVersion(context.Background(), test.modelName, test.modelVersion, test.modelSpec, nil)

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
			mr := NewModelRepository(logger, nil, path, nil, "0.0.0.0", 9000, time.Second)
			err := mr.RemoveModelVersion(test.modelName)
			g.Expect(err).To(BeNil())
			modelPath := filepath.Join(path, test.modelName)
			_, err = os.Stat(modelPath)
			g.Expect(err).ToNot(BeNil())
			g.Expect(errors.Is(err, os.ErrNotExist)).To(BeTrue())
		})
	}
}

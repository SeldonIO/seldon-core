/*
Copyright 2023 Seldon Technologies Ltd.

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

package mlserver

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/otiai10/copy"
	"github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestCreateSettingsFile(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		fileToCreate string
		pathToCreate string
		modelSpec    *scheduler.ModelSpec
		err          bool
		expected     *ModelSettings
	}

	tests := []test{
		{
			name:         "xgboost bst, top folder, ok",
			fileToCreate: "model.bst",
			modelSpec: &scheduler.ModelSpec{
				Requirements: []string{"xgboost"},
			},
			expected: &ModelSettings{
				Implementation: "mlserver_xgboost.XGBoostModel",
				Parameters: &ModelParameters{
					Uri: "./model.bst",
				},
			},
		},
		{
			name:         "lightgbm bst, top folder, ok",
			fileToCreate: "model.bst",
			modelSpec: &scheduler.ModelSpec{
				Requirements: []string{"lightgbm"},
			},
			expected: &ModelSettings{
				Implementation: "mlserver_lightgbm.LightGBMModel",
				Parameters: &ModelParameters{
					Uri: "./model.bst",
				},
			},
		},
		{
			name:         "xgboost json, top folder, ok",
			fileToCreate: "model.json",
			modelSpec: &scheduler.ModelSpec{
				Requirements: []string{"xgboost"},
			},
			expected: &ModelSettings{
				Implementation: "mlserver_xgboost.XGBoostModel",
				Parameters: &ModelParameters{
					Uri: "./model.json",
				},
			},
		},
		{
			name:         "joblib, top folder, ok",
			fileToCreate: "model.joblib",
			modelSpec: &scheduler.ModelSpec{
				Requirements: []string{"sklearn"},
			},
			expected: &ModelSettings{
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Uri: "./model.joblib",
				},
			},
		},
		{
			name:         "joblib, sub folder, not ok",
			fileToCreate: "/tt/model.joblib",
			modelSpec: &scheduler.ModelSpec{
				Requirements: []string{"sklearn"},
			},
			err: true,
		},
		{
			name:         "pickle, top folder, ok",
			fileToCreate: "model.pickle",
			modelSpec: &scheduler.ModelSpec{
				Requirements: []string{"sklearn"},
			},
			expected: &ModelSettings{
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Uri: "./model.pickle",
				},
			},
		},
		{
			name:         "pkl, top folder, ok",
			fileToCreate: "model.pkl",
			modelSpec: &scheduler.ModelSpec{
				Requirements: []string{"sklearn"},
			},
			expected: &ModelSettings{
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Uri: "./model.pkl",
				},
			},
		},
		{
			name: "unknown, top folder, not ok",
			modelSpec: &scheduler.ModelSpec{
				Requirements: []string{"sklearn"},
			},
			fileToCreate: "model.foo",
			err:          true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir()
			folderPath := filepath.Join(path, test.pathToCreate)
			err := os.MkdirAll(folderPath, fs.ModePerm)
			g.Expect(err).To(BeNil())
			if test.fileToCreate != "" {
				artifactFilePath := filepath.Join(folderPath, test.fileToCreate)
				err := os.MkdirAll(filepath.Dir(artifactFilePath), fs.ModePerm)
				g.Expect(err).To(BeNil())
				err = os.WriteFile(artifactFilePath, []byte{}, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			err = createModelSettingsFile(path, test.modelSpec)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				mlserverHandler := NewMLServerRepositoryHandler(logrus.New())
				ms, err := mlserverHandler.loadModelSettingsFromFile(filepath.Join(path, mlserverConfigFilename))
				g.Expect(err).To(BeNil())
				g.Expect(ms).To(Equal(test.expected))
			}
		})
	}
}

func TestCreateSKLearnModelSettings(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		fileToCreate string
		err          bool
		expected     *ModelSettings
	}

	tests := []test{
		{
			name:         "joblib, top folder, ok",
			fileToCreate: "model.joblib",
			expected: &ModelSettings{
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Uri: "./model.joblib",
				},
			},
		},
		{
			name:         "joblib, sub folder, not ok",
			fileToCreate: "/tt/model.joblib",
			err:          true,
		},
		{
			name:         "pickle, top folder, ok",
			fileToCreate: "model.pickle",
			expected: &ModelSettings{
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Uri: "./model.pickle",
				},
			},
		},
		{
			name:         "pkl, top folder, ok",
			fileToCreate: "model.pkl",
			expected: &ModelSettings{
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Uri: "./model.pkl",
				},
			},
		},
		{
			name:         "unknown, top folder, not ok",
			fileToCreate: "model.foo",
			err:          true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir()
			if test.fileToCreate != "" {
				artifactFilePath := filepath.Join(path, test.fileToCreate)
				err := os.MkdirAll(filepath.Dir(artifactFilePath), fs.ModePerm)
				g.Expect(err).To(BeNil())
				err = os.WriteFile(artifactFilePath, []byte{}, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			ms, err := createSKLearnModelSettings(path)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(ms).To(Equal(test.expected))
			}
		})
	}
}

func TestCreateXGBoostModelSettings(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		fileToCreate string
		err          bool
		expected     *ModelSettings
	}

	tests := []test{
		{
			name:         "bst, top folder, ok",
			fileToCreate: "model.bst",
			expected: &ModelSettings{
				Implementation: "mlserver_xgboost.XGBoostModel",
				Parameters: &ModelParameters{
					Uri: "./model.bst",
				},
			},
		},
		{
			name:         "json, top folder, ok",
			fileToCreate: "model.json",
			expected: &ModelSettings{
				Implementation: "mlserver_xgboost.XGBoostModel",
				Parameters: &ModelParameters{
					Uri: "./model.json",
				},
			},
		},
		{
			name:         "unknown, top folder, not ok",
			fileToCreate: "model.foo",
			err:          true,
		},
		{
			name:         "bst, sub folder, not ok",
			fileToCreate: "/tt/model.bst",
			err:          true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir()
			if test.fileToCreate != "" {
				artifactFilePath := filepath.Join(path, test.fileToCreate)
				err := os.MkdirAll(filepath.Dir(artifactFilePath), fs.ModePerm)
				g.Expect(err).To(BeNil())
				err = os.WriteFile(artifactFilePath, []byte{}, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			ms, err := createXGBoostModelSettings(path)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(ms).To(Equal(test.expected))
			}
		})
	}
}

func TestCreateLightgbmModelSettings(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		fileToCreate string
		err          bool
		expected     *ModelSettings
	}

	tests := []test{
		{
			name:         "bst, top folder, ok",
			fileToCreate: "model.bst",
			expected: &ModelSettings{
				Implementation: "mlserver_lightgbm.LightGBMModel",
				Parameters: &ModelParameters{
					Uri: "./model.bst",
				},
			},
		},
		{
			name:         "unknown, top folder, not ok",
			fileToCreate: "model.foo",
			err:          true,
		},
		{
			name:         "bst, sub folder, not ok",
			fileToCreate: "/tt/model.bst",
			err:          true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir()
			if test.fileToCreate != "" {
				artifactFilePath := filepath.Join(path, test.fileToCreate)
				err := os.MkdirAll(filepath.Dir(artifactFilePath), fs.ModePerm)
				g.Expect(err).To(BeNil())
				err = os.WriteFile(artifactFilePath, []byte{}, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			ms, err := createLightGBMModelSettings(path)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(ms).To(Equal(test.expected))
			}
		})
	}
}

func TestCreatePythonModelSettings(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		testFile string
		err      bool
		expected *ModelSettings
	}

	tests := []test{
		{
			name:     "model.py",
			testFile: "model.py",
			expected: &ModelSettings{
				Implementation: "model.PandasQueryRuntime",
				Parameters: &ModelParameters{
					Uri: "./model.py",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir()
			if test.testFile != "" {
				err := copy.Copy(fmt.Sprintf("testdata/%s", test.testFile), fmt.Sprintf("%s/%s", path, test.testFile))
				g.Expect(err).To(BeNil())
			}
			ms, err := createCustomPythonModelSettings(path)
			if test.err {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(ms).To(Equal(test.expected))
			}
		})
	}
}

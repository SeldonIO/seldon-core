/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package mlserver

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestSetExplainer(t *testing.T) {
	g := NewGomegaWithT(t)

	envoyHost := "0.0.0.0"
	envoyPort := 9000
	type test struct {
		name          string
		data          []byte
		explainerSpec *scheduler.ExplainerSpec
		expected      *ModelSettings
	}

	getStrPr := func(str string) *string { return &str }
	tests := []test{
		{
			name: "basic",
			data: []byte(`{"name": "iris","implementation": "mlserver_sklearn.SKLearnModel",
"parameters": {"version": "1", "extra":{}}}`),
			explainerSpec: &scheduler.ExplainerSpec{
				Type:     "anchor_tabular",
				ModelRef: getStrPr("mymodel"),
			},
			expected: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
					Extra: map[string]interface{}{
						explainerTypeKey: "anchor_tabular",
						inferUriKey:      "http://0.0.0.0:9000/v2/models/mymodel/infer",
					},
				},
			},
		},
		{
			name: "explainer parameters",
			data: []byte(`{"name": "iris","implementation": "mlserver_sklearn.SKLearnModel",
"parameters": {"version": "1", "extra":{"init_parameters":{"threshold":0.95}}}}`),
			explainerSpec: &scheduler.ExplainerSpec{
				Type:     "anchor_tabular",
				ModelRef: getStrPr("mymodel"),
			},
			expected: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
					Extra: map[string]interface{}{
						explainerTypeKey: "anchor_tabular",
						inferUriKey:      "http://0.0.0.0:9000/v2/models/mymodel/infer",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			modelRepoPath := t.TempDir()
			settingsFile := filepath.Join(modelRepoPath, mlserverConfigFilename)
			err := os.WriteFile(settingsFile, test.data, os.ModePerm)
			g.Expect(err).To(BeNil())
			m := &MLServerRepositoryHandler{}
			err = m.SetExplainer(modelRepoPath, test.explainerSpec, envoyHost, envoyPort)
			g.Expect(err).To(BeNil())
			modelSettings, err := m.loadModelSettingsFromFile(settingsFile)
			g.Expect(err).To(BeNil())
			g.Expect(modelSettings.Parameters.Extra[explainerTypeKey]).To(Equal(test.expected.Parameters.Extra[explainerTypeKey]))
			g.Expect(modelSettings.Parameters.Extra[inferUriKey]).To(Equal(test.expected.Parameters.Extra[inferUriKey]))
		})
	}
}

func TestLoadFromBytes(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		data     []byte
		expected *ModelSettings
		error    bool
	}

	getIntPtr := func(val int) *int {
		return &val
	}
	tests := []test{
		{
			name: "Sklearn",
			data: []byte(`{"name": "iris","implementation": "mlserver_sklearn.SKLearnModel",
"parameters": {"version": "1"}}`),
			expected: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
				},
			},
		},
		{
			name: "parallel_workers",
			data: []byte(`{"name": "iris","implementation": "mlserver_sklearn.SKLearnModel",
"parameters": {"version": "1"},"parallel_workers":1}`),
			expected: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
				},
				ParallelWorkers: getIntPtr(1),
			},
		},
		{
			name: "ExtraFields",
			data: []byte(`{"foo":"bar","name": "iris","implementation": "mlserver_sklearn.SKLearnModel",
"parameters": {"version": "1"}}`),
			expected: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
				},
			},
		},
		{
			name: "BadVersionField",
			data: []byte(`{"name": "iris","implementation": "mlserver_sklearn.SKLearnModel",
"parameters": {"version": 1}}`),
			expected: &ModelSettings{},
			error:    true,
		},
		{
			name: "environment tarball",
			data: []byte(`{"name": "iris","implementation": "mlserver_sklearn.SKLearnModel",
"parameters": {"version": "1", "environment_tarball":"file.tar.gz"}}`),
			expected: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version:            "1",
					EnvironmentTarball: "file.tar.gz",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := &MLServerRepositoryHandler{}
			ms, err := m.loadModelSettingsFromBytes(test.data)
			if !test.error {
				g.Expect(err).To(BeNil())
				g.Expect(ms).To(Equal(test.expected))
			} else {
				g.Expect(err).ToNot(BeNil())
			}
		})
	}
}

func TestFindModelVersionInPath(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name          string
		folders       map[string]*ModelSettings
		root          *ModelSettings
		version       uint32
		expectedFound bool
	}

	tests := []test{
		{
			name: "Simple",
			folders: map[string]*ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "1",
					},
				},
			},
			version:       1,
			expectedFound: true,
		},
		{
			name: "SettingsContradictsFolder",
			folders: map[string]*ModelSettings{
				"2": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "1",
					},
				},
			},
			version:       1,
			expectedFound: true,
		},
		{
			name: "NotFound",
			folders: map[string]*ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "1",
					},
				},
			},
			version:       2,
			expectedFound: false,
		},
		{
			name:    "VersionInRoot",
			folders: map[string]*ModelSettings{},
			root: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
				},
			},
			version:       1,
			expectedFound: true,
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
			if test.root != nil {
				data, err := json.Marshal(test.root)
				g.Expect(err).To(BeNil())
				settingsFilePath := filepath.Join(path, "model-settings.json")
				err = os.WriteFile(settingsFilePath, data, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			m := &MLServerRepositoryHandler{}
			foundPath, err := m.findModelVersionInPath(path, test.version)
			g.Expect(err).To(BeNil())
			if test.expectedFound {
				g.Expect(foundPath).ToNot(BeNil())
			} else {
				g.Expect(foundPath).To(Equal(""))
			}
		})
	}
}

func TestGetDefaultModelSettingsPath(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name          string
		folders       map[string]*ModelSettings
		root          *ModelSettings
		expectedFound bool
	}

	tests := []test{
		{
			name:          "NoRootOrVersions",
			folders:       map[string]*ModelSettings{},
			expectedFound: false,
		},
		{
			name:    "VersionInRoot",
			folders: map[string]*ModelSettings{},
			root: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
				},
			},
			expectedFound: true,
		},
		{
			name: "OnlyVersion",
			folders: map[string]*ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "1",
					},
				},
			},
			expectedFound: false,
		},
		{
			name: "VersionAndRoot - ignored root chosen",
			folders: map[string]*ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "1",
					},
				},
			},
			root: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
				},
			},
			expectedFound: true,
		},
		{
			name: "MultipleVersions - ignored root chosen",
			folders: map[string]*ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "1",
					},
				},
				"2": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "2",
					},
				},
			},
			root: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
				},
			},
			expectedFound: true,
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
			if test.root != nil {
				data, err := json.Marshal(test.root)
				g.Expect(err).To(BeNil())
				settingsFilePath := filepath.Join(path, "model-settings.json")
				err = os.WriteFile(settingsFilePath, data, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			m := &MLServerRepositoryHandler{}
			foundPath, err := m.getDefaultModelSettingsPath(path)
			g.Expect(err).To(BeNil())
			if test.expectedFound {
				g.Expect(foundPath).ToNot(BeNil())
			} else {
				g.Expect(foundPath).To(Equal(""))
			}
		})
	}
}

func TestUpdateVersion(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name      string
		settings  *ModelSettings
		modelName string
		version   string
		error     bool
	}

	tests := []test{
		{
			name: "Simple",
			settings: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
				},
			},
			modelName: "foo",
			version:   "2",
			error:     false,
		},
		{
			name: "ExtraParameters",
			settings: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version:     "1",
					ContentType: "foo",
					Format:      "bar",
				},
			},
			modelName: "foo",
			version:   "2",
			error:     false,
		},
		{
			name: "NoParameters",
			settings: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
			},
			modelName: "foo",
			version:   "2",
			error:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := t.TempDir()
			data, err := json.Marshal(test.settings)
			g.Expect(err).To(BeNil())
			settingsFilePath := filepath.Join(path, "model-settings.json")
			err = os.WriteFile(settingsFilePath, data, fs.ModePerm)
			g.Expect(err).To(BeNil())
			m := &MLServerRepositoryHandler{}
			err = m.updateNameAndVersion(path, test.modelName, test.version)
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				ms, err := m.loadModelSettingsFromFile(settingsFilePath)
				g.Expect(err).To(BeNil())
				g.Expect(ms.Parameters.Version).To(Equal(test.version))
				g.Expect(ms.Name).To(Equal(test.modelName))
				if test.settings.Parameters != nil {
					g.Expect(ms.Parameters.Uri).To(Equal(test.settings.Parameters.Uri))
					g.Expect(ms.Parameters.ContentType).To(Equal(test.settings.Parameters.ContentType))
					g.Expect(ms.Parameters.Format).To(Equal(test.settings.Parameters.Format))
				}

			}
		})
	}
}

func TestFindModelVersionFolder(t *testing.T) {
	g := NewGomegaWithT(t)

	getUintPtr := func(val uint32) *uint32 {
		return &val
	}

	type test struct {
		name            string
		folders         map[string]*ModelSettings
		root            *ModelSettings
		version         *uint32
		error           bool
		expectedPathDir string
		modelName       string
	}

	tests := []test{
		{
			name: "Simple with version",
			folders: map[string]*ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "",
					},
				},
			},
			version:         getUintPtr(1),
			expectedPathDir: "1",
			modelName:       "iris",
		},
		{
			name: "Root and version so root is chosen",
			folders: map[string]*ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "",
					},
				},
			},
			root: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
				},
			},
			version:         getUintPtr(1),
			expectedPathDir: "iris",
			modelName:       "iris",
		},
		{
			name: "path not matching version so should fail",
			folders: map[string]*ModelSettings{
				"2": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "1",
					},
				},
			},
			version:         getUintPtr(1),
			error:           true,
			expectedPathDir: "2",
			modelName:       "iris",
		},
		{
			name:    "root only",
			folders: map[string]*ModelSettings{},
			root: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
				},
			},
			version:         getUintPtr(1),
			expectedPathDir: "iris",
			modelName:       "iris",
		},
		{
			name:    "Version is in root but model setting version does not match but that's ok",
			folders: map[string]*ModelSettings{},
			root: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "2",
				},
			},
			version:         getUintPtr(1),
			expectedPathDir: "iris",
			modelName:       "iris",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), test.modelName)
			err := os.MkdirAll(path, fs.ModePerm)
			g.Expect(err).To(BeNil())
			for folderName, ms := range test.folders {
				folderPath := filepath.Join(path, folderName)
				err := os.MkdirAll(folderPath, fs.ModePerm)
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
				settingsFilePath := filepath.Join(path, "model-settings.json")
				err = os.WriteFile(settingsFilePath, data, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}
			logger := log.New()
			m := NewMLServerRepositoryHandler(logger)
			foundPath, found, err := m.FindModelVersionFolder("iris", test.version, path)
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				g.Expect(found).To(BeTrue())
				g.Expect(foundPath).ToNot(BeNil())
				pathBase := filepath.Base(foundPath)
				g.Expect(pathBase).To(Equal(test.expectedPathDir))
			}
		})
	}
}

func TestFindHighestVersionInPath(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name            string
		folders         map[string]*ModelSettings
		root            *ModelSettings
		expectedVersion string
		expectedFound   bool
	}

	tests := []test{
		{
			name: "FolderOne",
			folders: map[string]*ModelSettings{
				"1": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "1",
					},
				},
			},
			expectedVersion: "1",
			expectedFound:   true,
		},
		{
			name: "FolderTwo",
			folders: map[string]*ModelSettings{
				"2": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "1",
					},
				},
			},
			expectedVersion: "2",
			expectedFound:   true,
		},
		{
			name: "RootOnly",
			root: &ModelSettings{
				Name:           "iris",
				Implementation: "mlserver_sklearn.SKLearnModel",
				Parameters: &ModelParameters{
					Version: "1",
				},
			},
			expectedFound: false,
		},
		{
			name: "FolderTwo",
			folders: map[string]*ModelSettings{
				"11": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "1",
					},
				},
				"22": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "1",
					},
				},
				"33": {
					Name:           "iris",
					Implementation: "mlserver_sklearn.SKLearnModel",
					Parameters: &ModelParameters{
						Version: "1",
					},
				},
			},
			expectedVersion: "33",
			expectedFound:   true,
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
			if test.root != nil {
				data, err := json.Marshal(test.root)
				g.Expect(err).To(BeNil())
				settingsFilePath := filepath.Join(path, "model-settings.json")
				err = os.WriteFile(settingsFilePath, data, fs.ModePerm)
				g.Expect(err).To(BeNil())
			}

			logger := log.New()
			m := NewMLServerRepositoryHandler(logger)
			foundPath, err := m.findHighestVersionInPath(path)
			g.Expect(err).To(BeNil())
			if test.expectedFound {
				g.Expect(foundPath).ToNot(BeNil())
				g.Expect(filepath.Base(foundPath)).To(Equal(test.expectedVersion))
			} else {
				g.Expect(foundPath).To(Equal(""))
			}
		})
	}
}

func TestDefaultModelSettings(t *testing.T) {
	g := NewGomegaWithT(t)

	getIntPtr := func(val int) *int {
		return &val
	}
	tests := []struct {
		name          string
		modelSettings *ModelSettings
		expected      []byte
	}{
		{
			name:          "omits all empty fields",
			modelSettings: &ModelSettings{Name: "foo"},
			expected:      []byte("{\"name\":\"foo\"}"),
		},
		{
			name:          "add parallel workers",
			modelSettings: &ModelSettings{Name: "foo", ParallelWorkers: getIntPtr(1)},
			expected:      []byte("{\"name\":\"foo\",\"parallel_workers\":1}"),
		},
		{
			name:          "adds empty parameters dict",
			modelSettings: &ModelSettings{Name: "foo", Parameters: &ModelParameters{}},
			expected:      []byte("{\"name\":\"foo\",\"parameters\":{}}"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			data, _ := json.Marshal(test.modelSettings)
			g.Expect(data).To(Equal(test.expected))
		})
	}
}

func TestGetModelConfig(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name     string
		expected uint32
		err      bool
	}{
		{
			name:     "defaults to 1",
			expected: 1,
			err:      false,
		},
		{
			name:     "should pick up env var",
			expected: 10,
			err:      false,
		},
		{
			name:     "returns 1 on err",
			expected: 1,
			err:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := log.New()
			m := NewMLServerRepositoryHandler(logger)

			if test.err {
				os.Setenv(parallelWorkersEnvVar, "uh-oh")
			} else if test.expected > 1 {
				os.Setenv(parallelWorkersEnvVar, strconv.FormatInt(int64(test.expected), 10))
			}

			config, err := m.GetModelConfig("test-model")
			// mlserver should never return an error
			g.Expect(err).To(BeNil())
			g.Expect(config.InstanceCount).To(Equal(test.expected))
		})
	}
}

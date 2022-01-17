package mlserver

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	log "github.com/sirupsen/logrus"
)

func TestLoadFromBytes(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		data     []byte
		expected *ModelSettings
		error    bool
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
					Extra: map[string]string{
						"foo": "bar",
					},
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
			err = m.UpdateNameAndVersion(path, test.modelName, test.version)
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
					for k, v := range test.settings.Parameters.Extra {
						g.Expect(ms.Parameters.Extra[k]).To(Equal(v))
					}
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
			foundPath, err := m.FindModelVersionFolder("iris", test.version, path)
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
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

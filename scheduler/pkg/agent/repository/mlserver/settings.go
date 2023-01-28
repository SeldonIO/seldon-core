package mlserver

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

var (
	sklearnExtensions  = []string{"joblib", "pkl", "pickle"}
	xgboostExtensions  = []string{"json", "bst"}
	lightgbmExtensions = []string{"bst"}
)

// Look at requirements and take first matching one to our known list of ones we can handle
func createSettingsFile(path string, modelSpec *scheduler.ModelSpec) error {
	modelSettings, err := getModelSettings(path, modelSpec)
	if err != nil {
		return err
	}
	return saveModelSettings(path, modelSettings)
}

func getModelSettings(path string, modelSpec *scheduler.ModelSpec) (*ModelSettings, error) {
	for _, requirement := range modelSpec.Requirements {
		switch requirement {
		case "alibi-detect":
			return createAlibiDetectModelSettings()
		case "alibi-explain":
			return createAlibiExplainModelSettings(path)
		case "lightgbm":
			return createLightGBMModelSettings(path)
		case "mlflow":
			return createMLFlowModelSettings()
		case "python":
			return createCustomPythonModelSettings(path)
		case "sklearn":
			return createSKLearnModelSettings(path)
		case "xgboost":
			return createXGBoostModelSettings(path)
		}
	}
	return nil, fmt.Errorf("Can't create model-settings from requirements %v", modelSpec.Requirements)
}

func saveModelSettings(path string, modelSettings *ModelSettings) error {
	data, err := json.Marshal(modelSettings)
	if err != nil {
		return err
	}
	settingsPath := filepath.Join(path, mlserverConfigFilename)
	return os.WriteFile(settingsPath, data, fs.ModePerm)
}

func findFilesMatchingExtension(path string, ext string) ([]string, error) {
	return filepath.Glob(fmt.Sprintf("%s/*.%s", path, ext))
}

func findModelUri(path string, extensions []string) (string, error) {
	modelUri := ""
	for _, ext := range extensions {
		matches, err := findFilesMatchingExtension(path, ext)
		if err != nil {
			return "", err
		}
		if matches != nil {
			modelUri = matches[0]
			break
		}
	}
	modelUri = strings.TrimPrefix(modelUri, path)
	if modelUri == "" {
		return "", fmt.Errorf("Failed to find sklearn artifact in %s", path)
	}
	return modelUri, nil
}

func createModelSettingsFromUri(modelUri string, implementation string) *ModelSettings {
	return &ModelSettings{
		Implementation: implementation,
		Parameters: &ModelParameters{
			Uri: fmt.Sprintf(".%s", modelUri),
		},
	}
}

func createSKLearnModelSettings(path string) (*ModelSettings, error) {
	modelUri, err := findModelUri(path, sklearnExtensions)
	if err != nil {
		return nil, err
	}
	return createModelSettingsFromUri(modelUri, "mlserver_sklearn.SKLearnModel"), nil
}

func createXGBoostModelSettings(path string) (*ModelSettings, error) {
	modelUri, err := findModelUri(path, xgboostExtensions)
	if err != nil {
		return nil, err
	}
	return createModelSettingsFromUri(modelUri, "mlserver_xgboost.XGBoostModel"), nil
}

func createLightGBMModelSettings(path string) (*ModelSettings, error) {
	modelUri, err := findModelUri(path, lightgbmExtensions)
	if err != nil {
		return nil, err
	}
	return createModelSettingsFromUri(modelUri, "mlserver_lightgbm.LightGBMModel"), nil
}

func createAlibiDetectModelSettings() (*ModelSettings, error) {
	return &ModelSettings{
		Implementation: "mlserver_alibi_detect.AlibiDetectRuntime",
		Parameters: &ModelParameters{
			Uri: "./",
		},
	}, nil
}

func createMLFlowModelSettings() (*ModelSettings, error) {
	return &ModelSettings{
		Implementation: "mlserver_mlflow.MLflowRuntime",
		Parameters: &ModelParameters{
			Uri: "./",
		},
	}, nil
}

func createAlibiExplainModelSettings(path string) (*ModelSettings, error) {
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.IsDir() {
			parallelWorkers := 0
			return &ModelSettings{
				Implementation:  "mlserver_alibi_explain.AlibiExplainRuntime",
				ParallelWorkers: &parallelWorkers,
				Parameters: &ModelParameters{
					Uri: fmt.Sprintf("./%s", f.Name()),
				},
			}, nil
		}
	}
	return nil, fmt.Errorf("Failed to find alibi-explain saved folder in %s", path)
}

// This carries out a very simplistic logic:
// Find all python files
// Search for a file that extends MLModel and use that class
func createCustomPythonModelSettings(path string) (*ModelSettings, error) {
	matches, err := findFilesMatchingExtension(path, "py")
	if err != nil {
		return nil, err
	}
	for _, filename := range matches {
		data, err := os.ReadFile(filename)
		if err != nil {
			return nil, err
		}
		re := regexp.MustCompile(`.*class (.*)\(MLModel\):`)
		matches := re.FindStringSubmatch(string(data))
		if matches != nil {
			class := matches[1]
			base := filepath.Base(filename)
			return &ModelSettings{
				Implementation: fmt.Sprintf("%s.%s", strings.TrimSuffix(base, ".py"), class),
				Parameters: &ModelParameters{
					Uri: fmt.Sprintf("./%s", base),
				},
			}, nil
		}
	}
	return nil, fmt.Errorf("Failed to find MLServer custom python class that extends MLModel file in %s", path)
}

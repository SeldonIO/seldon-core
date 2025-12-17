/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"

	"github.com/cucumber/godog"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/generated/clientset/versioned"
	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	"github.com/seldonio/seldon-core/tests/integration/godog/steps/assertions"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Experiment struct {
	namespace      string
	label          map[string]string
	experiment     *mlopsv1alpha1.Experiment
	k8sClient      versioned.Interface
	watcherStorage k8sclient.WatcherStorage
	log            logrus.FieldLogger
}

func newExperiment(label map[string]string, namespace string, k8sClient versioned.Interface, log logrus.FieldLogger, watcherStorage k8sclient.WatcherStorage) *Experiment {
	return &Experiment{label: label, experiment: &mlopsv1alpha1.Experiment{}, log: log, namespace: namespace, k8sClient: k8sClient, watcherStorage: watcherStorage}
}

func LoadExperimentSteps(scenarioCtx *godog.ScenarioContext, w *World) {
	scenarioCtx.Step(`^I deploy experiment spec with timeout "([^"]+)":$`, func(timeout string, docString *godog.DocString) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.currentExperiment.deployExperiment(ctx, docString.Content)
		})
	})
	scenarioCtx.Step(`^the experiment should eventually become Ready with timeout "([^"]+)"$`, func(timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.currentExperiment.waitForExperimentReady(ctx)
		})
	})
	scenarioCtx.Step(`^I send "([^"]+)" HTTP inference requests to the experiment and expect all models in response, with payoad:$`, func(count int, docString *godog.DocString) error {
		return withTimeoutCtx("30s", func(ctx context.Context) error {
			return w.currentExperiment.sendHTTPInferenceRequestsToExperiment(ctx, count, docString.Content, w.infer)
		})
	})
}

func (e *Experiment) deployExperiment(ctx context.Context, yamlSpec string) error {
	var experiment mlopsv1alpha1.Experiment
	if err := yaml.Unmarshal([]byte(yamlSpec), &experiment); err != nil {
		return fmt.Errorf("failed to unmarshal experiment spec: %w", err)
	}

	experiment.Namespace = e.namespace
	e.experiment = &experiment
	e.applyScenarioLabel()

	_, err := e.k8sClient.MlopsV1alpha1().Experiments(e.namespace).Create(
		ctx,
		e.experiment,
		metav1.CreateOptions{},
	)
	if err != nil {
		return fmt.Errorf("failed to create experiment: %w", err)
	}

	return nil
}

func (e *Experiment) applyScenarioLabel() {
	if e.experiment.Labels == nil {
		e.experiment.Labels = make(map[string]string)
	}

	maps.Copy(e.experiment.Labels, e.label)

	// todo: change this approach
	for k, v := range k8sclient.DefaultCRDTestSuiteLabelMap {
		e.experiment.Labels[k] = v
	}
}

func (e *Experiment) waitForExperimentReady(ctx context.Context) error {
	return e.watcherStorage.WaitForExperimentCondition(
		ctx,
		e.experiment.Name,
		assertions.ExperimentReady)
}

func (e *Experiment) sendHTTPInferenceRequestsToExperiment(ctx context.Context, count int, body string, infer inference) error {
	modelRespCount := make(map[string]uint)

	type respJSON struct {
		Model string `json:"model_name"`
	}

	for i := 0; i < count; i++ {
		if err := infer.doHTTPExperimentInferenceRequest(ctx, e.experiment.Name, body); err != nil {
			return fmt.Errorf("request %d failed: %w", i+1, err)
		}

		if infer.lastHTTPResponse.StatusCode != http.StatusOK {
			return fmt.Errorf("request failed with status %d", infer.lastHTTPResponse.StatusCode)
		}

		body, err := io.ReadAll(infer.lastHTTPResponse.Body)
		if err != nil {
			return fmt.Errorf("failed reading resp body: %w", err)
		}

		e.log.Debugf("Got response HTTP body %+v", string(body))

		var resp respJSON
		if err := json.Unmarshal(body, &resp); err != nil {
			return fmt.Errorf("failed unmarshalling response body: %w", err)
		}

		if resp.Model == "" {
			return fmt.Errorf("response has no model name: %s", string(body))
		}

		index := strings.Index(resp.Model, "_")
		if index == -1 {
			return fmt.Errorf("response model name missing _: %s", string(body))
		}
		gotModel := resp.Model[0:index]
		modelRespCount[gotModel]++
	}

	for _, candidate := range e.experiment.Spec.Candidates {
		count, ok := modelRespCount[candidate.Name]
		if !ok {
			return fmt.Errorf("model %s not found in any HTTP response", candidate.Name)
		}
		if count == 0 {
			return fmt.Errorf("model %s response is zero", candidate.Name)
		}
	}

	return nil
}

package steps

import (
	"context"
	"fmt"
	"maps"

	"github.com/cucumber/godog"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/pkg/generated/clientset/versioned"
	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	"github.com/seldonio/seldon-core/tests/integration/godog/steps/assertions"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type Pipeline struct {
	label          map[string]string
	namespace      string
	pipeline       *mlopsv1alpha1.Pipeline
	pipelineName   string
	k8sClient      versioned.Interface
	watcherStorage k8sclient.WatcherStorage
	log            logrus.FieldLogger
}

func NewPipeline(label map[string]string, namespace string, k8sClient versioned.Interface, log logrus.FieldLogger, watcherStorage k8sclient.WatcherStorage) *Pipeline {
	return &Pipeline{label: label, pipeline: &mlopsv1alpha1.Pipeline{}, log: log, namespace: namespace, k8sClient: k8sClient, watcherStorage: watcherStorage}
}

func LoadCustomPipelineSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^I deploy pipeline spec with timeout "([^"]+)":$`, func(timeout string, spec *godog.DocString) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.currentPipeline.deployPipelineSpec(ctx, spec)
		})
	})

	scenario.Step(`^the pipeline "([^"]+)" (?:should )?eventually become Ready with timeout "([^"]+)"$`, func(name, timeout string) error {
		ctx, cancel, err := timeoutToContext(timeout)
		if err != nil {
			return err
		}
		defer cancel()

		return w.currentPipeline.waitForPipelineNameReady(ctx, name)
	})
}

func (p *Pipeline) deployPipelineSpec(ctx context.Context, spec *godog.DocString) error {
	pipelineSpec := &mlopsv1alpha1.Pipeline{}
	if err := yaml.Unmarshal([]byte(spec.Content), &pipelineSpec); err != nil {
		return fmt.Errorf("failed unmarshalling pipeline spec: %w", err)
	}

	p.pipeline = pipelineSpec
	p.pipeline.Namespace = p.namespace
	p.pipelineName = p.pipeline.Name
	p.applyScenarioLabel()

	if _, err := p.k8sClient.MlopsV1alpha1().Pipelines(p.namespace).Create(ctx, p.pipeline, metav1.CreateOptions{}); err != nil {
		return fmt.Errorf("failed creating pipeline: %w", err)
	}
	return nil
}

func (p *Pipeline) applyScenarioLabel() {
	if p.pipeline.Labels == nil {
		p.pipeline.Labels = make(map[string]string)
	}

	maps.Copy(p.pipeline.Labels, p.label)

	// todo: change this approach
	for k, v := range k8sclient.DefaultCRDTestSuiteLabelMap {
		p.pipeline.Labels[k] = v
	}
}

func (p *Pipeline) waitForPipelineNameReady(ctx context.Context, name string) error {

	return p.watcherStorage.WaitForKey(
		ctx,
		name,
		assertions.PipelineReady,
	)
}

func (p *Pipeline) waitForPipelineReady(ctx context.Context) error {

	return p.watcherStorage.WaitForObject(
		ctx,
		p.pipeline,
		assertions.PipelineReady,
	)
}

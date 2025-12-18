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

func newPipeline(label map[string]string, namespace string, k8sClient versioned.Interface, log logrus.FieldLogger, watcherStorage k8sclient.WatcherStorage) *Pipeline {
	return &Pipeline{label: label, pipeline: &mlopsv1alpha1.Pipeline{}, log: log, namespace: namespace, k8sClient: k8sClient, watcherStorage: watcherStorage}
}

func LoadCustomPipelineSteps(scenario *godog.ScenarioContext, w *World) {
	scenario.Step(`^I deploy a pipeline spec with timeout "([^"]+)":$`, func(timeout string, spec *godog.DocString) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.currentPipeline.deployPipelineSpec(ctx, spec)
		})
	})
	scenario.Step(`^the pipeline "([^"]+)" (?:should )?eventually become (Ready|NotReady) with timeout "([^"]+)"$`, func(name, readiness, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			switch readiness {
			case "Ready":
				return w.currentPipeline.waitForPipelineNameReady(ctx, name)
			case "NotReady":
				return w.currentPipeline.waitForPipelineNameNotReady(ctx, name)
			default:
				return fmt.Errorf("unknown readiness type: %s", readiness)
			}
		})
	})
	scenario.Step(`^the pipeline (?:should )?eventually become (Ready|NotReady) with timeout "([^"]+)"$`, func(readiness, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			switch readiness {
			case "Ready":
				return w.currentPipeline.waitForPipelineReady(ctx)
			case "NotReady":
				return w.currentPipeline.waitForPipelineNotReady(ctx)
			default:
				return fmt.Errorf("unknown readiness type: %s", readiness)
			}
		})
	})
	scenario.Step(`^the pipeline status (?:should )?eventually become (PipelineFailed) with timeout "([^"]+)"$`, func(status, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			switch status {
			case "PipelineFailed":
				return w.currentPipeline.waitForPipelineReady(ctx)
			case "NotReady":
				return w.currentPipeline.waitForPipelineNotReady(ctx)
			default:
				return fmt.Errorf("unknown pipeline status type: %s", status)
			}
		})
	})
	scenario.Step(`^I delete pipeline "([^"]+)" with timeout "([^"]+)"$`, func(pipeline, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.currentPipeline.deletePipelineName(ctx, pipeline)
		})
	})
	scenario.Step(`^I delete pipeline (?:the )?with timeout "([^"]+)"$`, func(timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.currentPipeline.deletePipeline(ctx)
		})
	})
	scenario.Step(`^the pipeline "([^"]+)" should eventually not exist with timeout "([^"]+)"$`, func(pipeline, timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.currentPipeline.waitForPipelineNameIsDeleted(ctx, pipeline)
		})
	})
	scenario.Step(`^the pipeline should eventually not exist with timeout "([^"]+)"$`, func(timeout string) error {
		return withTimeoutCtx(timeout, func(ctx context.Context) error {
			return w.currentPipeline.waitForPipelineIsDeleted(ctx)
		})
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

func (p *Pipeline) deletePipeline(ctx context.Context) error {
	return p.k8sClient.MlopsV1alpha1().Servers(p.namespace).Delete(ctx, p.pipelineName, metav1.DeleteOptions{})
}

func (p *Pipeline) deletePipelineName(ctx context.Context, pipeline string) error {
	return p.k8sClient.MlopsV1alpha1().Servers(p.namespace).Delete(ctx, pipeline, metav1.DeleteOptions{})
}

func (p *Pipeline) waitForPipelineNameReady(ctx context.Context, name string) error {
	return p.watcherStorage.WaitForPipelineCondition(
		ctx,
		name,
		assertions.PipelineReady,
	)
}

func (p *Pipeline) waitForPipelineNameNotReady(ctx context.Context, name string) error {
	return p.watcherStorage.WaitForPipelineCondition(
		ctx,
		name,
		assertions.PipelineNotReady,
	)
}

func (p *Pipeline) waitForPipelineReady(ctx context.Context) error {
	return p.watcherStorage.WaitForObjectCondition(
		ctx,
		p.pipeline,
		assertions.PipelineReady,
	)
}

func (p *Pipeline) waitForPipelineNotReady(ctx context.Context) error {
	return p.watcherStorage.WaitForObjectCondition(
		ctx,
		p.pipeline,
		assertions.PipelineNotReady,
	)
}

func (p *Pipeline) waitForPipelineStatus(ctx context.Context, status string) error {
	return p.watcherStorage.WaitForObjectCondition(
		ctx,
		p.pipeline,
		assertions.PipelineNotReady,
	)
}

func (p *Pipeline) waitForPipelineNameIsDeleted(ctx context.Context, pipeline string) error {
	return p.watcherStorage.WaitForPipelineCondition(
		ctx,
		pipeline,
		assertions.PipelineDeleted,
	)
}

func (p *Pipeline) waitForPipelineIsDeleted(ctx context.Context) error {
	return p.watcherStorage.WaitForPipelineCondition(
		ctx,
		p.pipelineName,
		assertions.PipelineDeleted,
	)
}

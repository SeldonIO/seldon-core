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
	"errors"
	"fmt"

	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// deleteModel we have to wait for model to be deleted, as there is a finalizer attached so the scheduler can confirm
// when model has been unloaded from inference pod, model-gw, dataflow-engine, pipeline-gw and controller will remove
// finalizer so deletion can complete.
func (m *Model) deleteModel(ctx context.Context, model string) error {
	modelCR, err := m.k8sClient.MlopsV1alpha1().Models(m.namespace).Get(ctx, model, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return fmt.Errorf("model %s can't be deleted, does not exist", model)
		}
		return fmt.Errorf("failed to get model %s", model)
	}

	if err := m.k8sClient.MlopsV1alpha1().Models(m.namespace).Delete(ctx, model, metav1.DeleteOptions{}); err != nil {
		return fmt.Errorf("failed deleting model: %w", err)
	}

	m.log.Debugf("Delete request for model %s sent", model)

	watcher, err := m.k8sClient.MlopsV1alpha1().Models(m.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector:   fmt.Sprintf("metadata.name=%s", model),
		ResourceVersion: modelCR.ResourceVersion,
	})
	if err != nil {
		return fmt.Errorf("failed watching model: %w", err)
	}
	defer watcher.Stop()

	m.log.Debugf("Waiting for %s model deletion confirmation", model)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return errors.New("watcher channel closed")
			}
			if event.Type == watch.Error {
				return fmt.Errorf("watch error: %v", event.Object)
			}
			if event.Type == watch.Deleted {
				return nil
			}
		}
	}
}

func (m *Model) waitForModelReady(ctx context.Context, model string) error {
	foundModel, err := m.k8sClient.MlopsV1alpha1().Models(m.namespace).Get(ctx, model, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed getting model: %w", err)
	}

	if foundModel.Status.IsReady() {
		return nil
	}

	watcher, err := m.k8sClient.MlopsV1alpha1().Models(m.namespace).Watch(ctx, metav1.ListOptions{
		FieldSelector:   fmt.Sprintf("metadata.name=%s", model),
		ResourceVersion: foundModel.ResourceVersion,
		Watch:           true,
	})
	if err != nil {
		return fmt.Errorf("failed subscribed to watch model: %w", err)
	}
	defer watcher.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event, ok := <-watcher.ResultChan():
			if !ok {
				return fmt.Errorf("watch channel closed")
			}

			if event.Type == watch.Error {
				return fmt.Errorf("watch error: %v", event.Object)
			}

			if event.Type == watch.Added || event.Type == watch.Modified {
				model := event.Object.(*v1alpha1.Model)
				if model.Status.IsReady() {
					return nil
				}
				m.log.Debugf("got watch event: model %s is not ready, still waiting", model)
				continue
			}

			if event.Type == watch.Deleted {
				return fmt.Errorf("resource was deleted")
			}
		}
	}
}

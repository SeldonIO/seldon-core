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

	"github.com/cucumber/godog"
	"github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"go.yaml.in/yaml/v3"
)

func (w *World) deployModelSpec(spec *godog.DocString) error {
	modelSpec := &v1alpha1.Model{}
	if err := yaml.Unmarshal([]byte(spec.Content), &modelSpec); err != nil {
		return fmt.Errorf("failed unmarshalling model spec: %w", err)
	}
	modelSpec.Namespace = w.namespace
	// TODO: uncomment when auto-gen k8s client merged
	//if _, err := w.k8sClient.MlopsV1alpha1().Models(w.namespace).Create(context.TODO(), modelSpec, metav1.CreateOptions{}); err != nil {
	//	return fmt.Errorf("failed creating model: %w", err)
	//}
	return nil
}

func (w *World) waitForModelReady(ctx context.Context, model string) error {
	// TODO: uncomment when auto-gen k8s client merged

	//foundModel, err := w.k8sClient.MlopsV1alpha1().Models(w.namespace).Get(ctx, model, metav1.GetOptions{})
	//if err != nil {
	//	return fmt.Errorf("failed getting model: %w", err)
	//}
	//
	//if foundModel.Status.IsReady() {
	//	return nil
	//}
	//
	//watcher, err := w.k8sClient.MlopsV1alpha1().Models(w.namespace).Watch(ctx, metav1.ListOptions{
	//	FieldSelector:   fmt.Sprintf("metadata.name=%s", model),
	//	ResourceVersion: foundModel.ResourceVersion,
	//	Watch:           true,
	//})
	//if err != nil {
	//	return fmt.Errorf("failed subscribed to watch model: %w", err)
	//}
	//defer watcher.Stop()
	//
	//for {
	//	select {
	//	case <-ctx.Done():
	//		return ctx.Err()
	//	case event, ok := <-watcher.ResultChan():
	//		if !ok {
	//			return fmt.Errorf("watch channel closed")
	//		}
	//
	//		if event.Type == watch.Error {
	//			return fmt.Errorf("watch error: %v", event.Object)
	//		}
	//
	//		if event.Type == watch.Added || event.Type == watch.Modified {
	//			model := event.Object.(*v1alpha1.Model)
	//			if model.Status.IsReady() {
	//				return nil
	//			}
	//		}
	//
	//		if event.Type == watch.Deleted {
	//			return fmt.Errorf("resource was deleted")
	//		}
	//	}
	//}

	return nil
}

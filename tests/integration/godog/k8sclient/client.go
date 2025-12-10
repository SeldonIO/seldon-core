/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package k8sclient

import (
	"context"
	"fmt"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	log "github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type K8sClient struct {
	logger     log.FieldLogger
	namespace  string
	KubeClient client.WithWatch
}

var DefaultCRDLabelMap = map[string]string{
	"test-suite": "godog",
}

const (
	DefaultCRDLabel = "test-suite=godog"
)

// New todo: separate k8s client init and pass to new
func New(namespace string, logger *log.Logger) (*K8sClient, error) {
	if logger == nil {
		logger = log.New()
	}

	k8sScheme := runtime.NewScheme()

	if err := scheme.AddToScheme(k8sScheme); err != nil {
		return nil, err
	}

	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed init k8s client: %v", err)
	}

	if err = mlopsv1alpha1.AddToScheme(k8sScheme); err != nil {
		return nil, fmt.Errorf("failed init k8s client: %v", err)
	}

	cl, err := client.NewWithWatch(cfg, client.Options{Scheme: k8sScheme})
	if err != nil {
		return nil, fmt.Errorf("failed init k8s client: %v", err)
	}

	return &K8sClient{
		logger:     logger.WithField("client", "k8sClient"),
		namespace:  namespace,
		KubeClient: cl,
	}, nil
}

func (k8s *K8sClient) ApplyModel(model *mlopsv1alpha1.Model) error {
	ctx := context.Background()

	// Ensure namespace is set
	if model.Namespace == "" {
		model.Namespace = k8s.namespace
	}

	// Ensure labels exist
	if model.Labels == nil {
		model.Labels = map[string]string{}
	}

	// add labels
	for k, v := range DefaultCRDLabelMap {
		model.Labels[k] = v
	}

	existing := &mlopsv1alpha1.Model{}
	key := client.ObjectKey{
		Namespace: model.Namespace,
		Name:      model.Name,
	}

	err := k8s.KubeClient.Get(ctx, key, existing)
	if apierrors.IsNotFound(err) {
		// Doesn't exist → create
		return k8s.KubeClient.Create(ctx, model)
	}
	if err != nil {
		// Some other error
		return err
	}

	// Exists → preserve ResourceVersion and update
	model.ResourceVersion = existing.ResourceVersion
	return k8s.KubeClient.Update(ctx, model)
}

func (k8s *K8sClient) DeleteScenarioResources(ctx context.Context, labels client.MatchingLabels) error {
	// 1) Delete Models
	if err := k8s.KubeClient.DeleteAllOf(
		ctx,
		&mlopsv1alpha1.Model{},
		client.InNamespace(k8s.namespace),
		labels,
	); err != nil {
		return err
	}

	return nil
}

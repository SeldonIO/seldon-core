package k8sclient

import (
	"context"
	"fmt"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Client interface {
	ApplyModel(model *mlopsv1alpha1.Model) error
	GetModel(model string) (*mlopsv1alpha1.Model, error)
	ApplyPipeline(pipeline *mlopsv1alpha1.Pipeline) error
	GetPipeline(pipeline string) (*mlopsv1alpha1.Pipeline, error)
}

type K8sClient struct {
	namespace  string
	kubeClient client.Client
}

func New(namespace string) (*K8sClient, error) {
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

	cl, err := client.New(cfg, client.Options{Scheme: k8sScheme})
	if err != nil {
		return nil, fmt.Errorf("failed init k8s client: %v", err)
	}

	return &K8sClient{
		namespace:  namespace,
		kubeClient: cl,
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

	// Add your label
	model.Labels["test-suite"] = "godog"

	existing := &mlopsv1alpha1.Model{}
	key := client.ObjectKey{
		Namespace: model.Namespace,
		Name:      model.Name,
	}

	err := k8s.kubeClient.Get(ctx, key, existing)
	if apierrors.IsNotFound(err) {
		// Doesn't exist → create
		return k8s.kubeClient.Create(ctx, model)
	}
	if err != nil {
		// Some other error
		return err
	}

	// Exists → preserve ResourceVersion and update
	model.ResourceVersion = existing.ResourceVersion
	return k8s.kubeClient.Update(ctx, model)
}

func (k8s *K8sClient) DeleteGodogTestModels(ctx context.Context) error {

	list := &mlopsv1alpha1.ModelList{}
	err := k8s.kubeClient.List(ctx, list,
		client.InNamespace(k8s.namespace),
		client.MatchingLabels{"test-suite": "godog"},
	)
	if err != nil {
		return err
	}

	for _, m := range list.Items {
		// Copy because Delete expects a pointer
		model := m
		if err := k8s.kubeClient.Delete(ctx, &model); err != nil {
			return err
		}
	}

	return nil
}

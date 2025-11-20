package godogtests

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cucumber/godog"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/yaml"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
)

const (
	defaultNamespace = "seldon-mesh"
	pollInterval     = 5 * time.Second
	waitTimeout      = 5 * time.Minute
)

type scalingContext struct {
	namespace      string
	kubeClient     client.Client
	serverManifest string
	modelManifest  string
	applied        []client.Object
}

func newScalingContext() (*scalingContext, error) {
	scheme := runtime.NewScheme()
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := mlopsv1alpha1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig: %w", err)
	}
	cl, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create controller-runtime client: %w", err)
	}
	return &scalingContext{
		namespace:  defaultNamespace,
		kubeClient: cl,
		applied:    make([]client.Object, 0),
	}, nil
}

func InitializeScenario(sc *godog.ScenarioContext) {
	var suite *scalingContext

	sc.Before(func(ctx context.Context, scenario *godog.Scenario) (context.Context, error) {
		var err error
		suite, err = newScalingContext()
		if err != nil {
			return ctx, err
		}
		return ctx, nil
	})

	sc.After(func(ctx context.Context, scenario *godog.Scenario, scnErr error) (context.Context, error) {
		if suite != nil {
			suite.cleanup()
		}
		return ctx, nil
	})

	sc.Step(`^the namespace "([^"]+)" is used$`, func(ns string) error {
		if suite == nil {
			return fmt.Errorf("scenario context not initialised")
		}
		suite.namespace = ns
		return nil
	})

	sc.Step(`^the following server resource:$`, func(body *godog.DocString) error {
		if suite == nil {
			return fmt.Errorf("scenario context not initialised")
		}
		suite.serverManifest = strings.TrimSpace(body.Content)
		if suite.serverManifest == "" {
			return fmt.Errorf("server manifest is empty")
		}
		return nil
	})

	sc.Step(`^the following model resource:$`, func(body *godog.DocString) error {
		if suite == nil {
			return fmt.Errorf("scenario context not initialised")
		}
		suite.modelManifest = strings.TrimSpace(body.Content)
		if suite.modelManifest == "" {
			return fmt.Errorf("model manifest is empty")
		}
		return nil
	})

	sc.Step(`^I apply the server resource$`, func() error {
		if suite == nil {
			return fmt.Errorf("scenario context not initialised")
		}
		if suite.serverManifest == "" {
			return fmt.Errorf("server manifest has not been defined")
		}
		return suite.applyServerManifest(suite.serverManifest)
	})

	sc.Step(`^I apply the model resource$`, func() error {
		if suite == nil {
			return fmt.Errorf("scenario context not initialised")
		}
		if suite.modelManifest == "" {
			return fmt.Errorf("model manifest has not been defined")
		}
		return suite.applyModelManifest(suite.modelManifest)
	})

	sc.Step(`^the server "([^"]+)" eventually reports (\d+) replicas$`, func(name string, replicas int) error {
		if suite == nil {
			return fmt.Errorf("scenario context not initialised")
		}
		return suite.waitForServerReplicas(name, suite.namespace, replicas)
	})

	sc.Step(`^the server "([^"]+)" in namespace "([^"]+)" eventually reports (\d+) replicas$`, func(name, namespace string, replicas int) error {
		if suite == nil {
			return fmt.Errorf("scenario context not initialised")
		}
		return suite.waitForServerReplicas(name, namespace, replicas)
	})

	sc.Step(`^the model "([^"]+)" eventually reports (\d+) replicas$`, func(name string, replicas int) error {
		if suite == nil {
			return fmt.Errorf("scenario context not initialised")
		}
		return suite.waitForModelReplicas(name, suite.namespace, replicas)
	})

	sc.Step(`^the model "([^"]+)" in namespace "([^"]+)" eventually reports (\d+) replicas$`, func(name, namespace string, replicas int) error {
		if suite == nil {
			return fmt.Errorf("scenario context not initialised")
		}
		return suite.waitForModelReplicas(name, namespace, replicas)
	})
}

func (s *scalingContext) applyServerManifest(manifest string) error {
	server := &mlopsv1alpha1.Server{}
	if err := yaml.Unmarshal([]byte(manifest), server); err != nil {
		return fmt.Errorf("failed to unmarshal server manifest: %w", err)
	}
	if server.Name == "" {
		return fmt.Errorf("server manifest missing metadata.name")
	}
	if server.Namespace == "" {
		server.Namespace = s.namespace
	}
	if server.TypeMeta.APIVersion == "" {
		server.APIVersion = "mlops.seldon.io/v1alpha1"
	}
	if server.TypeMeta.Kind == "" {
		server.Kind = "Server"
	}
	return s.applyObject(server)
}

func (s *scalingContext) applyModelManifest(manifest string) error {
	model := &mlopsv1alpha1.Model{}
	if err := yaml.Unmarshal([]byte(manifest), model); err != nil {
		return fmt.Errorf("failed to unmarshal model manifest: %w", err)
	}
	if model.Name == "" {
		return fmt.Errorf("model manifest missing metadata.name")
	}
	if model.Namespace == "" {
		model.Namespace = s.namespace
	}
	if model.TypeMeta.APIVersion == "" {
		model.APIVersion = "mlops.seldon.io/v1alpha1"
	}
	if model.TypeMeta.Kind == "" {
		model.Kind = "Model"
	}
	return s.applyObject(model)
}

func (s *scalingContext) applyObject(obj client.Object) error {
	ctx := context.Background()
	if obj.GetNamespace() == "" {
		obj.SetNamespace(s.namespace)
	}
	err := s.kubeClient.Create(ctx, obj)
	switch {
	case apierrors.IsAlreadyExists(err):
		existing := obj.DeepCopyObject().(client.Object)
		if getErr := s.kubeClient.Get(ctx, client.ObjectKeyFromObject(obj), existing); getErr != nil {
			return fmt.Errorf("failed to lookup existing %s/%s: %w", obj.GetNamespace(), obj.GetName(), getErr)
		}
		obj.SetResourceVersion(existing.GetResourceVersion())
		if updateErr := s.kubeClient.Update(ctx, obj); updateErr != nil {
			return fmt.Errorf("failed to update %s/%s: %w", obj.GetNamespace(), obj.GetName(), updateErr)
		}
	case err != nil:
		return fmt.Errorf("failed to create %s/%s: %w", obj.GetNamespace(), obj.GetName(), err)
	}
	s.trackApplied(obj)
	return nil
}

func (s *scalingContext) trackApplied(obj client.Object) {
	gvk := obj.GetObjectKind().GroupVersionKind()
	key := client.ObjectKeyFromObject(obj)
	for _, existing := range s.applied {
		if existing.GetObjectKind().GroupVersionKind() == gvk {
			existingKey := client.ObjectKeyFromObject(existing)
			if existingKey == key {
				return
			}
		}
	}
	s.applied = append(s.applied, obj.DeepCopyObject().(client.Object))
}

func (s *scalingContext) waitForServerReplicas(name, namespace string, replicas int) error {
	var lastStatus string
	err := s.waitForResource(fmt.Sprintf("server %s/%s replicas", namespace, name), func(ctx context.Context) (bool, error) {
		server := &mlopsv1alpha1.Server{}
		if err := s.kubeClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, server); err != nil {
			if apierrors.IsNotFound(err) {
				lastStatus = "server resource not found"
				return false, nil
			}
			return false, err
		}
		lastStatus = formatServerStatus(server)
		if int(server.Status.Replicas) != replicas {
			return false, nil
		}
		if !server.Status.IsReady() {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		if lastStatus != "" {
			return fmt.Errorf("%s (last observed: %s)", err.Error(), lastStatus)
		}
		return err
	}
	return nil
}

func (s *scalingContext) waitForModelReplicas(name, namespace string, replicas int) error {
	var lastStatus string
	err := s.waitForResource(fmt.Sprintf("model %s/%s replicas", namespace, name), func(ctx context.Context) (bool, error) {
		model := &mlopsv1alpha1.Model{}
		if err := s.kubeClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, model); err != nil {
			if apierrors.IsNotFound(err) {
				lastStatus = "model resource not found"
				return false, nil
			}
			return false, err
		}
		lastStatus = formatModelStatus(model)
		if int(model.Status.Replicas) != replicas {
			return false, nil
		}
		if !model.Status.IsReady() {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		if lastStatus != "" {
			return fmt.Errorf("%s (last observed: %s)", err.Error(), lastStatus)
		}
		return err
	}
	return nil
}

func (s *scalingContext) waitForResource(description string, condition wait.ConditionWithContextFunc) error {
	ctx, cancel := context.WithTimeout(context.Background(), waitTimeout)
	defer cancel()
	if err := wait.PollUntilContextTimeout(ctx, pollInterval, waitTimeout, true, func(ctx context.Context) (bool, error) {
		done, pollErr := condition(ctx)
		if pollErr != nil {
			return false, pollErr
		}
		return done, nil
	}); err != nil {
		return fmt.Errorf("timed out waiting for %s: %w", description, err)
	}
	return nil
}

func (s *scalingContext) cleanup() {
	if len(s.applied) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	for i := len(s.applied) - 1; i >= 0; i-- {
		obj := s.applied[i]
		_ = s.kubeClient.Delete(ctx, obj)
	}
	s.applied = nil
}

func formatServerStatus(server *mlopsv1alpha1.Server) string {
	desired := toIntPointer(server.Spec.Replicas)
	min := toIntPointer(server.Spec.MinReplicas)
	max := toIntPointer(server.Spec.MaxReplicas)
	ready := server.Status.IsReady()
	return fmt.Sprintf("spec.replicas=%d min=%d max=%d status.replicas=%d loadedModels=%d ready=%t", desired, min, max, server.Status.Replicas, server.Status.LoadedModelReplicas, ready)
}

func formatModelStatus(model *mlopsv1alpha1.Model) string {
	desired := toIntPointer(model.Spec.Replicas)
	min := toIntPointer(model.Spec.MinReplicas)
	max := toIntPointer(model.Spec.MaxReplicas)
	ready := model.Status.IsReady()
	return fmt.Sprintf("spec.replicas=%d min=%d max=%d status.replicas=%d available=%d ready=%t", desired, min, max, model.Status.Replicas, model.Status.AvailableReplicas, ready)
}

func toIntPointer(v *int32) int32 {
	if v == nil {
		return 0
	}
	return *v
}

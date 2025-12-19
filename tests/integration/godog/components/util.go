package components

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func rolloutRestartDeployment(ctx context.Context, c client.Client, nn types.NamespacedName) error {
	var dep appsv1.Deployment
	if err := c.Get(ctx, nn, &dep); err != nil {
		return err
	}

	if dep.Spec.Template.Annotations == nil {
		dep.Spec.Template.Annotations = map[string]string{}
	}
	dep.Spec.Template.Annotations["tests.seldon.io/restartedAt"] = time.Now().UTC().Format(time.RFC3339Nano)

	return c.Update(ctx, &dep)
}

func rolloutRestartStatefulSet(ctx context.Context, c client.Client, nn types.NamespacedName) error {
	var sts appsv1.StatefulSet
	if err := c.Get(ctx, nn, &sts); err != nil {
		return err
	}

	if sts.Spec.Template.Annotations == nil {
		sts.Spec.Template.Annotations = map[string]string{}
	}
	sts.Spec.Template.Annotations["tests.seldon.io/restartedAt"] = time.Now().UTC().Format(time.RFC3339Nano)

	return c.Update(ctx, &sts)
}

// wait helpers

func waitDeploymentReady(ctx context.Context, c client.Client, nn types.NamespacedName) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for deployment %s/%s ready: %w", nn.Namespace, nn.Name, ctx.Err())
		case <-ticker.C:
			var dep appsv1.Deployment
			if err := c.Get(ctx, nn, &dep); err != nil {
				continue
			}
			want := int32(1)
			if dep.Spec.Replicas != nil {
				want = *dep.Spec.Replicas
			}
			if dep.Status.AvailableReplicas == want {
				return nil
			}
		}
	}
}

func waitStatefulSetReady(ctx context.Context, c client.Client, nn types.NamespacedName) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for statefulset %s/%s ready: %w", nn.Namespace, nn.Name, ctx.Err())
		case <-ticker.C:
			var sts appsv1.StatefulSet
			if err := c.Get(ctx, nn, &sts); err != nil {
				continue
			}
			want := int32(1)
			if sts.Spec.Replicas != nil {
				want = *sts.Spec.Replicas
			}
			if sts.Status.ReadyReplicas == want {
				return nil
			}
		}
	}
}

func (c *SeldonRuntimeComponent) waitWorkloadAtReplicas(ctx context.Context, svc SeldonRuntimeService, want int32) error {
	ref, err := c.discoverWorkload(ctx, svc)
	if err != nil {
		// If scaling to 0, some operators might delete the workload; your runtime seems to keep it
		// (it scales Deployments to 0). If that ever changes, you could treat NotFound as success when want==0.
		return err
	}

	switch ref.kind {
	case "Deployment":
		return waitDeploymentReplicas(ctx, c.k8s.KubeClient, ref.nn, want)
	case "StatefulSet":
		return waitStatefulSetReplicas(ctx, c.k8s.KubeClient, ref.nn, want)
	default:
		return fmt.Errorf("wait replicas: unknown workload kind %q for %s", ref.kind, svc)
	}
}

func waitDeploymentReplicas(ctx context.Context, c client.Client, nn types.NamespacedName, want int32) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for deployment %s/%s replicas=%d: %w", nn.Namespace, nn.Name, want, ctx.Err())
		case <-ticker.C:
			var dep appsv1.Deployment
			if err := c.Get(ctx, nn, &dep); err != nil {
				continue
			}
			cur := int32(1)
			if dep.Spec.Replicas != nil {
				cur = *dep.Spec.Replicas
			}
			if cur != want {
				continue
			}
			// If scaling down, consider it done when desired replicas is 0.
			if want == 0 {
				return nil
			}
			// If scaling up, ensure it’s available too.
			if dep.Status.AvailableReplicas == want {
				return nil
			}
		}
	}
}

func waitStatefulSetReplicas(ctx context.Context, c client.Client, nn types.NamespacedName, want int32) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for statefulset %s/%s replicas=%d: %w", nn.Namespace, nn.Name, want, ctx.Err())
		case <-ticker.C:
			var sts appsv1.StatefulSet
			if err := c.Get(ctx, nn, &sts); err != nil {
				continue
			}
			cur := int32(1)
			if sts.Spec.Replicas != nil {
				cur = *sts.Spec.Replicas
			}
			if cur != want {
				continue
			}
			if want == 0 {
				return nil
			}
			if sts.Status.ReadyReplicas == want {
				return nil
			}
		}
	}
}

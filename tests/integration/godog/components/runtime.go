/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package components

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type SeldonRuntimeService string

const (
	ServiceScheduler       SeldonRuntimeService = "seldon-scheduler"
	ServicePipelineGateway SeldonRuntimeService = "seldon-pipelinegateway"
	ServiceModelGateway    SeldonRuntimeService = "seldon-modelgateway"
	ServiceDataflowEngine  SeldonRuntimeService = "seldon-dataflow-engine"
	ServiceEnvoy           SeldonRuntimeService = "seldon-envoy"
	ServiceHodometer       SeldonRuntimeService = "hodometer"
)

type workloadRef struct {
	kind string // "Deployment" or "StatefulSet"
	nn   types.NamespacedName
}

type SeldonRuntimeComponent struct {
	name ComponentName
	k8s  *k8sclient.K8sClient

	gvk schema.GroupVersionKind
	nn  types.NamespacedName

	mu sync.Mutex

	// snapshot of whole runtime + baseline per-service replicas
	snapExists   bool
	runtimeSnap  *unstructured.Unstructured
	baseReplicas map[SeldonRuntimeService]*int32

	dirty bool
}

// todo: this would be global components for the runtime which are not done or complete

func (c *SeldonRuntimeComponent) MakeUnavailable(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (c *SeldonRuntimeComponent) MakeAvailable(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (c *SeldonRuntimeComponent) Scale(ctx context.Context, replicas int32) error {
	//TODO implement me
	panic("implement me")
}

func NewSeldonRuntimeComponent(k8s *k8sclient.K8sClient, namespace, runtimeName string) *SeldonRuntimeComponent {
	return &SeldonRuntimeComponent{
		name:         "SeldonRuntime",
		k8s:          k8s,
		gvk:          schema.GroupVersionKind{Group: "mlops.seldon.io", Version: "v1alpha1", Kind: "SeldonRuntime"},
		nn:           types.NamespacedName{Namespace: namespace, Name: runtimeName},
		baseReplicas: map[SeldonRuntimeService]*int32{},
	}
}

func (c *SeldonRuntimeComponent) Name() ComponentName { return c.name }

func (c *SeldonRuntimeComponent) Snapshot(ctx context.Context) error {
	if c.runtimeSnap != nil || c.snapExists {
		return nil
	}

	rt := &unstructured.Unstructured{}
	rt.SetGroupVersionKind(c.gvk)

	if err := c.k8s.KubeClient.Get(ctx, c.nn, rt); err != nil {
		if apierrors.IsNotFound(err) {
			c.snapExists = false
			return nil
		}
		return fmt.Errorf("runtime snapshot: get %s/%s: %w", c.nn.Namespace, c.nn.Name, err)
	}

	c.snapExists = true
	c.runtimeSnap = rt.DeepCopy()

	// read baseline overrides replicas for all overrides
	overrides, found, err := unstructured.NestedSlice(rt.Object, "spec", "overrides")
	if err != nil {
		return fmt.Errorf("runtime snapshot: read spec.overrides: %w", err)
	}
	if found {
		for _, v := range overrides {
			m, ok := v.(map[string]any)
			if !ok {
				continue
			}
			name, _ := m["name"].(string)
			if name == "" {
				continue
			}
			// replicas might be missing (scheduler in your sample)
			if r64, ok := m["replicas"].(int64); ok {
				r := int32(r64)
				svc := SeldonRuntimeService(name)
				c.baseReplicas[svc] = &r
			}
		}
	}

	return nil
}

func (c *SeldonRuntimeComponent) Restore(ctx context.Context) error {
	if !c.dirty {
		return nil
	}

	// baseline absent => ensure absent (optional)
	if !c.snapExists {
		_ = c.deleteRuntimeIfExists(ctx)
		c.dirty = false
		return nil
	}

	if c.runtimeSnap == nil {
		return fmt.Errorf("runtime restore: missing snapshot")
	}

	// 1) Restore the runtime object spec (create/update)
	if err := c.applyRuntimeBaseline(ctx); err != nil {
		return err
	}

	// 2) Restore baseline replicas for services that had replicas in baseline
	// (If baseline did not specify replicas for a service, we don't force it.)
	for svc, r := range c.baseReplicas {
		if r == nil {
			continue
		}
		if err := c.SetReplicas(ctx, svc, *r); err != nil {
			return err
		}
	}

	c.dirty = false
	return nil
}

func (c *SeldonRuntimeComponent) mutateRuntime(ctx context.Context, fn func(rt *unstructured.Unstructured) error) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	return retry.OnError(retry.DefaultBackoff, apierrors.IsConflict, func() error {
		rt := &unstructured.Unstructured{}
		rt.SetGroupVersionKind(c.gvk)

		if err := c.k8s.KubeClient.Get(ctx, c.nn, rt); err != nil {
			return err
		}
		if err := fn(rt); err != nil {
			return err
		}
		return c.k8s.KubeClient.Update(ctx, rt)
	})
}

func (c *SeldonRuntimeComponent) SetReplicas(ctx context.Context, svc SeldonRuntimeService, replicas int32) error {
	if err := c.Snapshot(ctx); err != nil {
		return err
	}

	c.dirty = true

	if err := c.mutateRuntime(ctx, func(rt *unstructured.Unstructured) error {
		return setOverrideReplicas(rt, string(svc), replicas)
	}); err != nil {
		return err
	}

	// wait for the managed workload to reflect it
	return c.waitWorkloadAtReplicas(ctx, svc, replicas)
}

func (c *SeldonRuntimeComponent) ScaleDown(ctx context.Context, svc SeldonRuntimeService) error {
	return c.SetReplicas(ctx, svc, 0)
}

func (c *SeldonRuntimeComponent) ScaleUpToBaseline(ctx context.Context, svc SeldonRuntimeService) error {
	if err := c.Snapshot(ctx); err != nil {
		return err
	}
	base := c.baseReplicas[svc]
	if base == nil {
		return fmt.Errorf("no baseline replicas recorded for %s", svc)
	}
	return c.SetReplicas(ctx, svc, *base)
}

func (c *SeldonRuntimeComponent) RestartService(ctx context.Context, svc SeldonRuntimeService) error {
	if err := c.Snapshot(ctx); err != nil {
		return err
	}

	// optional policy: pipeline-gw restart should ensure scheduler ready
	if svc == ServicePipelineGateway {
		if err := c.WaitServiceReady(ctx, ServiceScheduler); err != nil {
			return fmt.Errorf("cannot restart %s while scheduler not ready: %w", svc, err)
		}
	}

	ref, err := c.discoverWorkload(ctx, svc)
	if err != nil {
		return err
	}

	switch ref.kind {
	case "Deployment":
		if err := rolloutRestartDeployment(ctx, c.k8s.KubeClient, ref.nn); err != nil {
			return err
		}
		return waitDeploymentReady(ctx, c.k8s.KubeClient, ref.nn)

	case "StatefulSet":
		if err := rolloutRestartStatefulSet(ctx, c.k8s.KubeClient, ref.nn); err != nil {
			return err
		}
		return waitStatefulSetReady(ctx, c.k8s.KubeClient, ref.nn)

	default:
		return fmt.Errorf("unknown workload kind %q for %s", ref.kind, svc)
	}
}

func (c *SeldonRuntimeComponent) discoverWorkload(ctx context.Context, svc SeldonRuntimeService) (*workloadRef, error) {
	appName := string(svc) // matches app.kubernetes.io/name in your manifests

	// Deployment first
	var deps appsv1.DeploymentList
	if err := c.k8s.KubeClient.List(ctx, &deps,
		client.InNamespace(c.nn.Namespace),
		client.MatchingLabels{"app.kubernetes.io/name": appName},
	); err != nil {
		return nil, err
	}
	if len(deps.Items) == 1 {
		d := deps.Items[0]
		return &workloadRef{kind: "Deployment", nn: types.NamespacedName{Namespace: d.Namespace, Name: d.Name}}, nil
	}
	if len(deps.Items) > 1 {
		return nil, fmt.Errorf("multiple deployments matched app.kubernetes.io/name=%s", appName)
	}

	// StatefulSet
	var stss appsv1.StatefulSetList
	if err := c.k8s.KubeClient.List(ctx, &stss,
		client.InNamespace(c.nn.Namespace),
		client.MatchingLabels{"app.kubernetes.io/name": appName},
	); err != nil {
		return nil, err
	}
	if len(stss.Items) == 1 {
		s := stss.Items[0]
		return &workloadRef{kind: "StatefulSet", nn: types.NamespacedName{Namespace: s.Namespace, Name: s.Name}}, nil
	}
	if len(stss.Items) > 1 {
		return nil, fmt.Errorf("multiple statefulsets matched app.kubernetes.io/name=%s", appName)
	}

	return nil, fmt.Errorf("no deployment/statefulset matched app.kubernetes.io/name=%s", appName)
}

func (c *SeldonRuntimeComponent) WaitServiceReady(ctx context.Context, svc SeldonRuntimeService) error {
	condType := runtimeReadyConditionType(svc) // e.g. "PipelineGatewayReady"
	return c.waitRuntimeCondition(ctx, condType, "True")
}

func runtimeReadyConditionType(svc SeldonRuntimeService) string {
	switch svc {
	case ServicePipelineGateway:
		return "PipelineGatewayReady"
	case ServiceScheduler:
		return "SchedulerReady"
	case ServiceModelGateway:
		return "ModelGatewayReady"
	case ServiceDataflowEngine:
		return "DataflowEngineReady"
	case ServiceEnvoy:
		return "EnvoyReady"
	case ServiceHodometer:
		return "HodometerReady"
	default:
		return ""
	}
}

func findOverrideIndex(overrides []any, name string) int {
	for i, v := range overrides {
		m, ok := v.(map[string]any)
		if !ok {
			continue
		}
		if n, _ := m["name"].(string); n == name {
			return i
		}
	}
	return -1
}

func setOverrideReplicas(rt *unstructured.Unstructured, serviceName string, replicas int32) error {
	overrides, found, err := unstructured.NestedSlice(rt.Object, "spec", "overrides")
	if err != nil {
		return fmt.Errorf("read spec.overrides: %w", err)
	}
	if !found {
		overrides = []any{}
	}

	idx := findOverrideIndex(overrides, serviceName)
	if idx == -1 {
		// Create override if missing
		overrides = append(overrides, map[string]any{
			"name":     serviceName,
			"replicas": int64(replicas),
		})
	} else {
		m := overrides[idx].(map[string]any)
		m["replicas"] = int64(replicas)
		overrides[idx] = m
	}

	if err := unstructured.SetNestedSlice(rt.Object, overrides, "spec", "overrides"); err != nil {
		return fmt.Errorf("write spec.overrides: %w", err)
	}
	return nil
}

func getOverrideReplicas(rt *unstructured.Unstructured, serviceName string) (*int32, error) {
	overrides, found, err := unstructured.NestedSlice(rt.Object, "spec", "overrides")
	if err != nil {
		return nil, fmt.Errorf("read spec.overrides: %w", err)
	}
	if !found {
		return nil, nil
	}

	idx := findOverrideIndex(overrides, serviceName)
	if idx == -1 {
		return nil, nil
	}

	m := overrides[idx].(map[string]any)
	// replicas might be int64 (usual), but be tolerant
	switch v := m["replicas"].(type) {
	case int64:
		r := int32(v)
		return &r, nil
	case float64:
		r := int32(v)
		return &r, nil
	default:
		return nil, nil
	}
}

func (c *SeldonRuntimeComponent) waitRuntimeCondition(ctx context.Context, condType string, wantStatus string) error {
	if condType == "" {
		return fmt.Errorf("runtime condition type is empty")
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for runtime condition %q=%q: %w", condType, wantStatus, ctx.Err())
		case <-ticker.C:
			rt := &unstructured.Unstructured{}
			rt.SetGroupVersionKind(c.gvk)

			if err := c.k8s.KubeClient.Get(ctx, c.nn, rt); err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return err
			}

			conds, found, err := unstructured.NestedSlice(rt.Object, "status", "conditions")
			if err != nil {
				return fmt.Errorf("read status.conditions: %w", err)
			}
			if !found {
				continue
			}

			for _, v := range conds {
				m, ok := v.(map[string]any)
				if !ok {
					continue
				}
				t, _ := m["type"].(string)
				if t != condType {
					continue
				}
				s, _ := m["status"].(string)
				if s == wantStatus {
					return nil
				}
				// Found condition but not in desired state yet.
				break
			}
		}
	}
}

func (c *SeldonRuntimeComponent) deleteRuntimeIfExists(ctx context.Context) error {
	rt := &unstructured.Unstructured{}
	rt.SetGroupVersionKind(c.gvk)

	if err := c.k8s.KubeClient.Get(ctx, c.nn, rt); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return c.k8s.KubeClient.Delete(ctx, rt)
}

func (c *SeldonRuntimeComponent) applyRuntimeBaseline(ctx context.Context) error {
	// Copy snapshot (don’t mutate stored snapshot)
	base := c.runtimeSnap.DeepCopy()

	// Always ensure correct GVK/namespace/name
	base.SetGroupVersionKind(c.gvk)
	base.SetNamespace(c.nn.Namespace)
	base.SetName(c.nn.Name)

	// Strip status to avoid update conflicts (unless you explicitly manage it)
	unstructured.RemoveNestedField(base.Object, "status")

	// Strip server-managed metadata fields that can cause issues
	base.SetResourceVersion("") // set later if updating existing
	base.SetUID("")
	base.SetGeneration(0)
	base.SetManagedFields(nil)

	// Try create; if exists, update
	current := &unstructured.Unstructured{}
	current.SetGroupVersionKind(c.gvk)

	if err := c.k8s.KubeClient.Get(ctx, c.nn, current); err != nil {
		if apierrors.IsNotFound(err) {
			// Create fresh
			return c.k8s.KubeClient.Create(ctx, base)
		}
		return err
	}

	// Update existing: must carry resourceVersion
	base.SetResourceVersion(current.GetResourceVersion())
	return c.k8s.KubeClient.Update(ctx, base)
}

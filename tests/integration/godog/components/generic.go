package components

import (
	"context"
	"fmt"
	"time"

	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

type UnavailableMode string

const (
	UnavailableByScaling  UnavailableMode = "scale"
	UnavailableByDeleting UnavailableMode = "delete"
)

// K8sObjectComponent is a generic component manager for any K8s object (including CRDs)
// referenced by GVK + NamespacedName.
type K8sObjectComponent struct {
	name     ComponentName
	gvk      schema.GroupVersionKind
	resource types.NamespacedName
	k8s      *k8sclient.K8sClient

	unavailableMode UnavailableMode

	// Snapshot
	snapExists   bool
	objectSnap   *unstructured.Unstructured // deep copy of baseline object
	baseReplicas *int32                     // if replicas exist

	dirty bool
}

func NewK8sObjectComponent(
	k8s *k8sclient.K8sClient,
	name ComponentName,
	gvk schema.GroupVersionKind,
	resource types.NamespacedName,
	unavailableMode UnavailableMode,
) *K8sObjectComponent {
	return &K8sObjectComponent{
		name:            name,
		gvk:             gvk,
		resource:        resource,
		k8s:             k8s,
		unavailableMode: unavailableMode,
	}
}

func (c *K8sObjectComponent) Name() ComponentName { return c.name }

func (c *K8sObjectComponent) Snapshot(ctx context.Context) error {
	if c.objectSnap != nil || c.snapExists {
		return nil
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(c.gvk)

	err := c.k8s.KubeClient.Get(ctx, c.resource, obj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Baseline is "absent"
			c.snapExists = false
			return nil
		}
		return fmt.Errorf("%s snapshot: get %s/%s: %w", c.name, c.resource.Namespace, c.resource.Name, err)
	}

	// Baseline is "present": deep copy
	c.snapExists = true
	c.objectSnap = obj.DeepCopy()

	// Try to snapshot replicas if present
	if replicas64, found, _ := unstructured.NestedInt64(obj.Object, "spec", "replicas"); found {
		r := int32(replicas64)
		c.baseReplicas = &r
	}

	return nil
}

func (c *K8sObjectComponent) Restore(ctx context.Context) error {
	if !c.dirty {
		return nil
	}

	// Baseline absent => ensure deleted (deleteIfExists now waits for NotFound)
	if !c.snapExists {
		if err := c.deleteIfExists(ctx); err != nil {
			return err
		}
		c.dirty = false
		return nil
	}

	if c.objectSnap == nil {
		return fmt.Errorf("%s restore: baseline snapshot missing", c.name)
	}

	// 1) Ensure baseline object exists (create/update)
	if err := c.applyBaseline(ctx); err != nil {
		return err
	}

	// 2) Wait until it exists (avoid races with subsequent steps)
	if err := c.waitForExists(ctx); err != nil {
		return err
	}

	// 3) Restore replicas if we captured them, and wait until applied
	if c.baseReplicas != nil {
		if err := c.Scale(ctx, *c.baseReplicas); err != nil {
			return err
		}
		if err := c.waitForReplicas(ctx, *c.baseReplicas); err != nil {
			return err
		}
	}

	c.dirty = false
	return nil
}

func (c *K8sObjectComponent) waitForExists(ctx context.Context) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf(
				"timeout waiting for %s %s/%s to exist: %w",
				c.gvk.Kind, c.resource.Namespace, c.resource.Name, ctx.Err(),
			)
		case <-ticker.C:
			obj := &unstructured.Unstructured{}
			obj.SetGroupVersionKind(c.gvk)

			err := c.k8s.KubeClient.Get(ctx, c.resource, obj)
			if err == nil {
				return nil
			}
			if apierrors.IsNotFound(err) {
				continue
			}
			return fmt.Errorf(
				"error checking existence of %s %s/%s: %w",
				c.gvk.Kind, c.resource.Namespace, c.resource.Name, err,
			)
		}
	}
}

func (c *K8sObjectComponent) waitForReplicas(ctx context.Context, want int32) error {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf(
				"timeout waiting for %s %s/%s replicas=%d: %w",
				c.gvk.Kind, c.resource.Namespace, c.resource.Name, want, ctx.Err(),
			)
		case <-ticker.C:
			obj := &unstructured.Unstructured{}
			obj.SetGroupVersionKind(c.gvk)

			err := c.k8s.KubeClient.Get(ctx, c.resource, obj)
			if err != nil {
				if apierrors.IsNotFound(err) {
					continue
				}
				return fmt.Errorf(
					"error checking replicas for %s %s/%s: %w",
					c.gvk.Kind, c.resource.Namespace, c.resource.Name, err,
				)
			}

			r64, found, err := unstructured.NestedInt64(obj.Object, "spec", "replicas")
			if err != nil {
				return fmt.Errorf(
					"error reading spec.replicas for %s %s/%s: %w",
					c.gvk.Kind, c.resource.Namespace, c.resource.Name, err,
				)
			}
			if !found {
				return fmt.Errorf(
					"%s %s/%s does not have spec.replicas (cannot wait for replicas)",
					c.gvk.Kind, c.resource.Namespace, c.resource.Name,
				)
			}

			if int32(r64) == want {
				return nil
			}
		}
	}
}

func (c *K8sObjectComponent) MakeUnavailable(ctx context.Context) error {
	if err := c.Snapshot(ctx); err != nil {
		return err
	}

	switch c.unavailableMode {
	case UnavailableByDeleting:
		if err := c.deleteIfExists(ctx); err != nil {
			return err
		}
	case UnavailableByScaling:
		if err := c.Scale(ctx, 0); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%s: unknown unavailable mode %q", c.name, c.unavailableMode)
	}

	c.dirty = true
	return nil
}

func (c *K8sObjectComponent) MakeAvailable(ctx context.Context) error {
	if err := c.Snapshot(ctx); err != nil {
		return err
	}
	// MakeAvailable means "restore to baseline"
	c.dirty = true
	return c.Restore(ctx)
}

// Scale attempts to update spec.replicas. If the object doesn't have that field, return a clear error.
func (c *K8sObjectComponent) Scale(ctx context.Context, replicas int32) error {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(c.gvk)

	if err := c.k8s.KubeClient.Get(ctx, c.resource, obj); err != nil {
		return fmt.Errorf("%s scale: get %s/%s: %w", c.name, c.resource.Namespace, c.resource.Name, err)
	}

	// Check the field exists or can be set
	if _, found, _ := unstructured.NestedFieldNoCopy(obj.Object, "spec"); !found {
		return fmt.Errorf("%s scale: object has no spec", c.name)
	}

	if err := unstructured.SetNestedField(obj.Object, int64(replicas), "spec", "replicas"); err != nil {
		return fmt.Errorf("%s scale: set spec.replicas: %w", c.name, err)
	}

	if err := c.k8s.KubeClient.Update(ctx, obj); err != nil {
		return fmt.Errorf("%s scale: update %s/%s: %w", c.name, c.resource.Namespace, c.resource.Name, err)
	}

	c.dirty = true
	return nil
}

// ---- internals ----

func (c *K8sObjectComponent) deleteIfExists(ctx context.Context) error {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(c.gvk)

	// First: try to delete if it exists
	if err := c.k8s.KubeClient.Get(ctx, c.resource, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}

	if err := c.k8s.KubeClient.Delete(ctx, obj); err != nil {
		return err
	}

	// Then: wait until it's actually gone
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf(
				"timeout waiting for %s %s/%s to be deleted: %w",
				c.gvk.Kind, c.resource.Namespace, c.resource.Name, ctx.Err(),
			)
		case <-ticker.C:
			check := &unstructured.Unstructured{}
			check.SetGroupVersionKind(c.gvk)

			err := c.k8s.KubeClient.Get(ctx, c.resource, check)
			if apierrors.IsNotFound(err) {
				// ✅ deletion fully completed
				return nil
			}
			if err != nil {
				return fmt.Errorf(
					"error checking deletion of %s %s/%s: %w",
					c.gvk.Kind, c.resource.Namespace, c.resource.Name, err,
				)
			}
		}
	}
}

func (c *K8sObjectComponent) applyBaseline(ctx context.Context) error {
	desired := c.objectSnap.DeepCopy()

	// Remove fields that should not be set on create/update.
	// (Important when restoring full objects.)
	unstructured.RemoveNestedField(desired.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(desired.Object, "metadata", "uid")
	unstructured.RemoveNestedField(desired.Object, "metadata", "generation")
	unstructured.RemoveNestedField(desired.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(desired.Object, "status")

	// Try update if exists, else create
	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(c.gvk)

	err := c.k8s.KubeClient.Get(ctx, c.resource, existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return c.k8s.KubeClient.Create(ctx, desired)
		}
		return fmt.Errorf("%s restore: get existing %s/%s: %w", c.name, c.resource.Namespace, c.resource.Name, err)
	}

	// Preserve current resourceVersion for Update
	desired.SetResourceVersion(existing.GetResourceVersion())
	return c.k8s.KubeClient.Update(ctx, desired)
}

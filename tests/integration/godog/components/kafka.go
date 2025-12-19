package components

import (
	"context"
	"fmt"
	"time"

	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

// This is an example of how to do a custom Component

type KafkaComponent struct {
	name     ComponentName
	k8s      *k8sclient.K8sClient
	resource types.NamespacedName

	// baseline snapshot
	snapExists bool
	snap       *unstructured.Unstructured

	dirty bool
}

const Kafka ComponentName = "Kafka"

var kafkaNodePoolGVK = schema.GroupVersionKind{
	Group:   "kafka.strimzi.io",
	Version: "v1beta2",
	Kind:    "KafkaNodePool",
}

func NewKafkaComponent(k8s *k8sclient.K8sClient, namespace string) *KafkaComponent {
	return &KafkaComponent{
		name:     Kafka,
		k8s:      k8s,
		resource: types.NamespacedName{Namespace: namespace, Name: "kafka"},
	}
}

func (k *KafkaComponent) Name() ComponentName { return k.name }

func (k *KafkaComponent) Snapshot(ctx context.Context) error {
	if k.snap != nil || k.snapExists {
		return nil
	}

	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(kafkaNodePoolGVK)

	err := k.k8s.KubeClient.Get(ctx, k.resource, obj)
	if err != nil {
		if apierrors.IsNotFound(err) {
			k.snapExists = false
			return nil
		}
		return fmt.Errorf("kafka snapshot: get %s/%s: %w", k.resource.Namespace, k.resource.Name, err)
	}

	k.snapExists = true
	k.snap = obj.DeepCopy()
	return nil
}

func (k *KafkaComponent) Restore(ctx context.Context) error {
	if !k.dirty {
		return nil
	}

	// baseline absent => ensure absent
	if !k.snapExists {
		_ = k.deleteIfExists(ctx)
		k.dirty = false
		return nil
	}
	if k.snap == nil {
		return fmt.Errorf("kafka restore: missing baseline snapshot")
	}

	if err := k.applyBaseline(ctx); err != nil {
		return err
	}

	k.dirty = false
	return nil
}

// --- Availability actions (domain choice) ---

// MakeUnavailable: delete the KafkaNodePool CR (your new requirement)
func (k *KafkaComponent) MakeUnavailable(ctx context.Context) error {
	if err := k.Snapshot(ctx); err != nil {
		return err
	}
	if err := k.deleteIfExists(ctx); err != nil {
		return err
	}
	k.dirty = true
	return nil
}

// MakeAvailable: recreate to baseline
func (k *KafkaComponent) MakeAvailable(ctx context.Context) error {
	if err := k.Snapshot(ctx); err != nil {
		return err
	}
	k.dirty = true
	return k.Restore(ctx)
}

// --- Extra Kafka-specific methods ---

func (k *KafkaComponent) Scale(ctx context.Context, replicas int32) error {
	// optional: only if the resource exists
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(kafkaNodePoolGVK)

	if err := k.k8s.KubeClient.Get(ctx, k.resource, obj); err != nil {
		return fmt.Errorf("kafka scale: get %s/%s: %w", k.resource.Namespace, k.resource.Name, err)
	}

	if err := unstructured.SetNestedField(obj.Object, int64(replicas), "spec", "replicas"); err != nil {
		return fmt.Errorf("kafka scale: set spec.replicas: %w", err)
	}

	if err := k.k8s.KubeClient.Update(ctx, obj); err != nil {
		return fmt.Errorf("kafka scale: update: %w", err)
	}

	k.dirty = true
	return nil
}

func (k *KafkaComponent) WaitForNodePoolReplicas(ctx context.Context, want int32) error {
	deadline, ok := ctx.Deadline()
	for {
		obj := &unstructured.Unstructured{}
		obj.SetGroupVersionKind(kafkaNodePoolGVK)
		err := k.k8s.KubeClient.Get(ctx, k.resource, obj)
		if err == nil {
			if r64, found, _ := unstructured.NestedInt64(obj.Object, "spec", "replicas"); found && int32(r64) == want {
				return nil
			}
		}
		if ok && time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for kafka nodepool replicas=%d", want)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

// --- internals ---

func (k *KafkaComponent) deleteIfExists(ctx context.Context) error {
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(kafkaNodePoolGVK)

	if err := k.k8s.KubeClient.Get(ctx, k.resource, obj); err != nil {
		if apierrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	return k.k8s.KubeClient.Delete(ctx, obj)
}

func (k *KafkaComponent) applyBaseline(ctx context.Context) error {
	desired := k.snap.DeepCopy()

	// strip fields that break create/update
	unstructured.RemoveNestedField(desired.Object, "metadata", "resourceVersion")
	unstructured.RemoveNestedField(desired.Object, "metadata", "uid")
	unstructured.RemoveNestedField(desired.Object, "metadata", "generation")
	unstructured.RemoveNestedField(desired.Object, "metadata", "managedFields")
	unstructured.RemoveNestedField(desired.Object, "status")

	// ensure namespace/name set
	desired.SetNamespace(k.resource.Namespace)
	desired.SetName(k.resource.Name)
	desired.SetGroupVersionKind(kafkaNodePoolGVK)

	existing := &unstructured.Unstructured{}
	existing.SetGroupVersionKind(kafkaNodePoolGVK)

	err := k.k8s.KubeClient.Get(ctx, k.resource, existing)
	if err != nil {
		if apierrors.IsNotFound(err) {
			desired.SetCreationTimestamp(metav1.Time{}) // safe no-op
			return k.k8s.KubeClient.Create(ctx, desired)
		}
		return err
	}

	desired.SetResourceVersion(existing.GetResourceVersion())
	return k.k8s.KubeClient.Update(ctx, desired)
}

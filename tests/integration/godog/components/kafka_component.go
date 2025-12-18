package components

import (
	"context"
	"fmt"

	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/types"
)

type KafkaComponent struct {
	name      ComponentName
	namespace string
	resource  types.NamespacedName

	k8s   *k8sclient.K8sClient
	base  *int32 // baseline replicas
	dirty bool
}

const kafkaComponentName ComponentName = "kafka"

func NewKafkaComponent(k8s *k8sclient.K8sClient, namespace, statefulSetName string) *KafkaComponent {
	return &KafkaComponent{
		name:      kafkaComponentName,
		namespace: namespace,
		resource:  types.NamespacedName{Namespace: namespace, Name: statefulSetName},
		k8s:       k8s,
	}
}

func (k *KafkaComponent) Name() ComponentName { return k.name }

// Snapshot reads the current replicas once and stores them as baseline.
func (k *KafkaComponent) Snapshot(ctx context.Context) error {
	if k.base != nil {
		// already snapshotted
		return nil
	}

	var sts appsv1.StatefulSet
	if err := k.k8s.KubeClient.Get(ctx, k.resource, &sts); err != nil {
		return fmt.Errorf("kafka snapshot: get %s/%s: %w", k.resource.Namespace, k.resource.Name, err)
	}

	if sts.Spec.Replicas == nil {
		return fmt.Errorf("kafka snapshot: statefulset %s/%s has nil .spec.replicas", k.resource.Namespace, k.resource.Name)
	}

	replicas := *sts.Spec.Replicas
	k.base = &replicas
	return nil
}

// Restore only acts if we've modified Kafka in this scenario (dirty=true).
func (k *KafkaComponent) Restore(ctx context.Context) error {
	if !k.dirty || k.base == nil {
		return nil
	}
	if err := k.scale(ctx, *k.base); err != nil {
		return err
	}
	k.dirty = false
	return nil
}

// Public helpers used from steps:

func (k *KafkaComponent) MakeUnavailable(ctx context.Context) error {
	if err := k.Snapshot(ctx); err != nil {
		return err
	}
	if err := k.scale(ctx, 0); err != nil {
		return err
	}
	k.dirty = true
	return nil
}

func (k *KafkaComponent) MakeAvailable(ctx context.Context) error {
	if k.base == nil {
		// nothing to restore to; safest is to no-op or return error
		return nil
	}
	if err := k.scale(ctx, *k.base); err != nil {
		return err
	}
	k.dirty = false
	return nil
}

func (k *KafkaComponent) scale(ctx context.Context, replicas int32) error {
	var sts appsv1.StatefulSet
	if err := k.k8s.KubeClient.Get(ctx, k.resource, &sts); err != nil {
		return fmt.Errorf("kafka scale: get %s/%s: %w", k.resource.Namespace, k.resource.Name, err)
	}
	sts.Spec.Replicas = &replicas
	if err := k.k8s.KubeClient.Update(ctx, &sts); err != nil {
		return fmt.Errorf("kafka scale: update %s/%s: %w", k.resource.Namespace, k.resource.Name, err)
	}
	return nil
}

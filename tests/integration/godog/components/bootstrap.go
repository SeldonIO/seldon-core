/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package components

import (
	"fmt"

	"github.com/seldonio/seldon-core/tests/integration/godog/k8sclient"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

const (
	KafkaNodePool ComponentName = "KafkaNodePool"
)

func StartComponents(k8sClient *k8sclient.K8sClient, namespace string) (*EnvManager, error) {
	var kafkaNodePoolGVK = schema.GroupVersionKind{
		Group:   "kafka.strimzi.io",
		Version: "v1beta2",
		Kind:    "KafkaNodePool",
	}

	kafkaNodePool := NewK8sObjectComponent(
		k8sClient,
		KafkaNodePool,
		kafkaNodePoolGVK,
		types.NamespacedName{Namespace: namespace, Name: "kafka"},
		UnavailableByDeleting,
	)

	runtime := NewSeldonRuntimeComponent(
		k8sClient,
		namespace,
		"seldon",
	)

	env, err := NewEnvManager(kafkaNodePool, runtime)
	if err != nil {
		return nil, fmt.Errorf("failed to bootstrap components: %v", err)
	}

	return env, nil
}

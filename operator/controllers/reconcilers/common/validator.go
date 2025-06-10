/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package common

import (
	"context"
	"fmt"
	"strconv"

	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
)

func ParseInt32(s string) (int32, error) {
	i64, err := strconv.ParseInt(s, 10, 32)
	return int32(i64), err
}

func ValidateDataflowScaleSpec(
	ctx context.Context,
	clt client.Client,
	component *mlopsv1alpha1.ComponentDefn,
	kafkaConfig *mlopsv1alpha1.KafkaConfig,
	namespace *string,
	logger logr.Logger,
) error {
	logger.Info("kafkaConfig.Topics", "Topics", kafkaConfig.Topics)
	numPartitions, err := ParseInt32(kafkaConfig.Topics["numPartitions"].StrVal)
	if err != nil {
		return fmt.Errorf("failed to parse numPartitions from KafkaConfig: %w", err)
	}

	logger.Info("Using numPartitions from KafkaConfig", "numPartitions", numPartitions)

	var pipelineCount int32 = 0
	if namespace != nil {
		// Get the number of Pipeline resources in the namespace
		var pipelineList mlopsv1alpha1.PipelineList
		if err := clt.List(ctx, &pipelineList, client.InNamespace(*namespace)); err != nil {
			return fmt.Errorf("failed to list Pipeline resources in namespace %s: %w", *namespace, err)
		}

		pipelineCount = int32(len(pipelineList.Items))
		logger.Info("Number of Pipeline resources", "namespace", *namespace, "count", pipelineCount)
	}

	maxReplicas := numPartitions
	if pipelineCount != 0 {
		maxReplicas = numPartitions * pipelineCount
	}

	logger.Info("Maximum replicas for dataflow engine", "max_replicas", maxReplicas)

	if component.Replicas != nil && *component.Replicas > maxReplicas {
		component.Replicas = &maxReplicas
		logger.Info("Adjusted dataflow engine replicas to max", "replicas", maxReplicas)
	}
	return nil
}

/*
Copyright 2023 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/ghodss/yaml"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
)

const (
	agentConfigMapName = "seldon-agent"
	kafkaConfigMapName = "seldon-kafka"
	traceConfigMapName = "seldon-tracing"
)

type ConfigMapReconciler struct {
	common.ReconcilerConfig
	configMaps []*v1.ConfigMap
}

func NewConfigMapReconciler(
	common common.ReconcilerConfig,
	config *mlopsv1alpha1.SeldonConfiguration,
	meta metav1.ObjectMeta) (*ConfigMapReconciler, error) {

	configMaps, err := toConfigMaps(config, meta)
	if err != nil {
		return nil, err
	}
	return &ConfigMapReconciler{
		ReconcilerConfig: common,
		configMaps:       configMaps,
	}, nil
}

func (s *ConfigMapReconciler) GetResources() []client.Object {
	var objs []client.Object
	for _, svc := range s.configMaps {
		objs = append(objs, svc)
	}
	return objs
}

func toConfigMaps(config *mlopsv1alpha1.SeldonConfiguration, meta metav1.ObjectMeta) ([]*v1.ConfigMap, error) {
	agentConfigMap, err := getAgentConfigMap(config.AgentConfig, meta.Namespace)
	if err != nil {
		return nil, err
	}
	kafkaConfigMap, err := getKafkaConfigMap(config.KafkaConfig, meta.Namespace)
	if err != nil {
		return nil, err
	}
	tracingConfigMap, err := getTracingConfigMap(config.TracingConfig, meta.Namespace)
	if err != nil {
		return nil, err
	}
	return []*v1.ConfigMap{
		agentConfigMap,
		kafkaConfigMap,
		tracingConfigMap,
	}, nil
}

func getAgentConfigMap(agentConfig mlopsv1alpha1.AgentConfiguration, namespace string) (*v1.ConfigMap, error) {
	agentJson, err := yaml.Marshal(agentConfig)
	if err != nil {
		return nil, err
	}
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      agentConfigMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			"agent.yaml": string(agentJson),
		},
	}, nil
}

func getKafkaConfigMap(kafkaConfig mlopsv1alpha1.KafkaConfig, namespace string) (*v1.ConfigMap, error) {
	kafkaJson, err := json.Marshal(kafkaConfig)
	if err != nil {
		return nil, err
	}
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      kafkaConfigMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			"kafka.json": string(kafkaJson),
		},
	}, nil
}

func getTracingConfigMap(tracingConfig mlopsv1alpha1.TracingConfig, namespace string) (*v1.ConfigMap, error) {
	tracingJson, err := json.Marshal(tracingConfig)
	if err != nil {
		return nil, err
	}
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      traceConfigMapName,
			Namespace: namespace,
		},
		Data: map[string]string{
			"tracing.json":                string(tracingJson),
			"OTEL_JAVAAGENT_ENABLED":      strconv.FormatBool(!tracingConfig.Disable),
			"OTEL_EXPORTER_OTLP_ENDPOINT": fmt.Sprintf("http://%s", tracingConfig.OtelExporterEndpoint),
		},
	}, nil
}

func (s *ConfigMapReconciler) getReconcileOperation(idx int, configMap *v1.ConfigMap) (constants.ReconcileOperation, error) {
	found := &v1.ConfigMap{}
	err := s.Client.Get(context.TODO(), types.NamespacedName{
		Name:      configMap.GetName(),
		Namespace: configMap.GetNamespace(),
	}, found)
	if err != nil {
		if errors.IsNotFound(err) {
			return constants.ReconcileCreateNeeded, nil
		}
		return constants.ReconcileUnknown, err
	}
	if equality.Semantic.DeepEqual(configMap.Data, found.Data) {
		// Update our version so we have Status if needed
		s.configMaps[idx] = found
		return constants.ReconcileNoChange, nil
	}
	// Update resource vesion so the client Update will succeed
	s.configMaps[idx].SetResourceVersion(found.ResourceVersion)
	return constants.ReconcileUpdateNeeded, nil
}

func (s *ConfigMapReconciler) Reconcile() error {
	logger := s.Logger.WithName("ConfigMapReconcile")
	for idx, configMap := range s.configMaps {
		op, err := s.getReconcileOperation(idx, configMap)
		switch op {
		case constants.ReconcileCreateNeeded:
			logger.V(1).Info("ConfigMap Create", "Name", configMap.GetName(), "Namespace", configMap.GetNamespace())
			err = s.Client.Create(s.Ctx, configMap)
			if err != nil {
				logger.Error(err, "Failed to create configmap", "Name", configMap.GetName(), "Namespace", configMap.GetNamespace())
				return err
			}
		case constants.ReconcileUpdateNeeded:
			logger.V(1).Info("ConfigMap Update", "Name", configMap.GetName(), "Namespace", configMap.GetNamespace())
			err = s.Client.Update(s.Ctx, configMap)
			if err != nil {
				logger.Error(err, "Failed to update configmap", "Name", configMap.GetName(), "Namespace", configMap.GetNamespace())
				return err
			}
		case constants.ReconcileNoChange:
			logger.V(1).Info("ConfigMap No Change", "Name", configMap.GetName(), "Namespace", configMap.GetNamespace())
		case constants.ReconcileUnknown:
			logger.Error(err, "Failed to get reconcile operation for configmap", "Name", configMap.GetName(), "Namespace", configMap.GetNamespace())
			return err
		}
	}
	return nil
}

func (s *ConfigMapReconciler) GetConditions() []*apis.Condition {
	return nil
}

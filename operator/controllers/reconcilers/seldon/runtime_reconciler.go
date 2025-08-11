/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package server

import (
	"fmt"
	"strconv"

	"github.com/banzaicloud/k8s-objectmatcher/patch"
	v1 "k8s.io/api/core/v1"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mlopsv1alpha1 "github.com/seldonio/seldon-core/operator/v2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operator/v2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
)

const (
	DEFAULT_NUM_PARTITIONS                    = 1
	DEFAULT_MODELGATEWAY_MAX_NUM_CONSUMERS    = 100
	DEFAULT_PIPELINEGATEWAY_MAX_NUM_CONSUMERS = 100

	MODELGATEWAY_MAX_NUM_CONSUMERS    = "MODELGATEWAY_MAX_NUM_CONSUMERS"
	PIPELINEGATEWAY_MAX_NUM_CONSUMERS = "PIPELINEGATEWAY_MAX_NUM_CONSUMERS"
)

type SeldonRuntimeReconciler struct {
	common.ReconcilerConfig
	componentReconcilers []common.Reconciler
	rbacReconciler       common.Reconciler
	serviceReconciler    common.Reconciler
	configMapReconciler  common.Reconciler
}

func ParseInt32(s string, defaultVal int32) (int32, error) {
	if s == "" {
		return defaultVal, nil
	}

	i64, err := strconv.ParseInt(s, 10, 32)
	return int32(i64), err
}

func getEnvVarValue(podSpec *v1.PodSpec, name string, defaultValue string) string {
	if podSpec != nil && len(podSpec.Containers) > 0 {
		for _, env := range podSpec.Containers[0].Env {
			if env.Name == name {
				return env.Value
			}
		}
	}

	return defaultValue
}

func replicaCalc(resourceCount, maxConsumers, partitions int32) int32 {
	if resourceCount == 0 {
		return 1
	}
	return partitions * min(resourceCount, maxConsumers)
}

func validateScaleSpec(
	component *mlopsv1alpha1.ComponentDefn,
	runtime *mlopsv1alpha1.SeldonRuntime,
	commonConfig common.ReconcilerConfig,
	namespace *string,
	maxConsumersEnvName string,
	defaultMaxConsumers int32,
	eventReason string,
	resourceListObj client.ObjectList,
	countResources func(client.ObjectList) int,
) error {
	ctx, clt, recorder := commonConfig.Ctx, commonConfig.Client, commonConfig.Recorder

	numPartitions, err := ParseInt32(
		runtime.Spec.Config.KafkaConfig.Topics["numPartitions"].StrVal,
		DEFAULT_NUM_PARTITIONS,
	)
	if err != nil {
		return fmt.Errorf("failed to parse numPartitions from KafkaConfig: %w", err)
	}

	var resourceCount int32 = 0
	if namespace != nil {
		if err := clt.List(ctx, resourceListObj, client.InNamespace(*namespace)); err != nil {
			return fmt.Errorf("failed to list resources in namespace %s: %w", *namespace, err)
		}
		resourceCount = int32(countResources(resourceListObj))
	}

	var maxConsumers int32 = defaultMaxConsumers
	if maxConsumersEnvName != "" {
		maxConsumersEnv := getEnvVarValue(component.PodSpec, maxConsumersEnvName, "")
		maxConsumers, err = ParseInt32(maxConsumersEnv, defaultMaxConsumers)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", maxConsumersEnvName, err)
		}
		if maxConsumers == 0 {
			return fmt.Errorf("invalid %s value: %s", maxConsumersEnvName, maxConsumersEnv)
		}
	}

	maxReplicas := replicaCalc(resourceCount, maxConsumers, numPartitions)
	if component.Replicas != nil && *component.Replicas > maxReplicas {
		component.Replicas = &maxReplicas
		recorder.Eventf(
			runtime,
			v1.EventTypeWarning,
			eventReason,
			fmt.Sprintf(
				"%s requested replicas exceeded maximum of %d based on KafkaConfig and resource count, adjusted to %d",
				component.Name, maxReplicas, maxReplicas,
			),
		)
	}

	return nil
}

func ValidateScaleSpecPipelines(
	component *mlopsv1alpha1.ComponentDefn,
	runtime *mlopsv1alpha1.SeldonRuntime,
	commonConfig common.ReconcilerConfig,
	namespace *string,
	envVar string,
	reason string,
) error {
	return validateScaleSpec(
		component,
		runtime,
		commonConfig,
		namespace,
		envVar,
		DEFAULT_PIPELINEGATEWAY_MAX_NUM_CONSUMERS,
		reason,
		&mlopsv1alpha1.PipelineList{},
		func(obj client.ObjectList) int {
			return len(obj.(*mlopsv1alpha1.PipelineList).Items)
		},
	)
}

func ValidateDataflowScaleSpec(
	component *mlopsv1alpha1.ComponentDefn,
	runtime *mlopsv1alpha1.SeldonRuntime,
	commonConfig common.ReconcilerConfig,
	namespace *string,
) error {
	return ValidateScaleSpecPipelines(
		component,
		runtime,
		commonConfig,
		namespace,
		"", // No env var
		"DataflowEngineReplicasAdjusted",
	)
}

func ValidatePipelineGatewaySpec(
	component *mlopsv1alpha1.ComponentDefn,
	runtime *mlopsv1alpha1.SeldonRuntime,
	commonConfig common.ReconcilerConfig,
	namespace *string,
) error {
	return ValidateScaleSpecPipelines(
		component,
		runtime,
		commonConfig,
		namespace,
		PIPELINEGATEWAY_MAX_NUM_CONSUMERS,
		"PipelineGatewayReplicasAdjusted",
	)
}

func ValidateModelGatewaySpec(
	component *mlopsv1alpha1.ComponentDefn,
	runtime *mlopsv1alpha1.SeldonRuntime,
	commonConfig common.ReconcilerConfig,
	namespace *string,
) error {
	return validateScaleSpec(
		component,
		runtime,
		commonConfig,
		namespace,
		MODELGATEWAY_MAX_NUM_CONSUMERS,
		DEFAULT_MODELGATEWAY_MAX_NUM_CONSUMERS,
		"ModelGatewayReplicasAdjusted",
		&mlopsv1alpha1.ModelList{},
		func(obj client.ObjectList) int {
			return len(obj.(*mlopsv1alpha1.ModelList).Items)
		},
	)
}

func ValidateComponent(
	component *mlopsv1alpha1.ComponentDefn,
	runtime *mlopsv1alpha1.SeldonRuntime,
	commonConfig common.ReconcilerConfig,
	namespace *string,
) error {
	if component.Name == mlopsv1alpha1.DataflowEngineName {
		return ValidateDataflowScaleSpec(
			component,
			runtime,
			commonConfig,
			namespace,
		)
	}
	if component.Name == mlopsv1alpha1.PipelineGatewayName {
		return ValidatePipelineGatewaySpec(
			component,
			runtime,
			commonConfig,
			namespace,
		)
	}
	if component.Name == mlopsv1alpha1.ModelGatewayName {
		return ValidateModelGatewaySpec(
			component,
			runtime,
			commonConfig,
			namespace,
		)
	}
	return nil
}

func ComponentOverride(component *mlopsv1alpha1.ComponentDefn, override *mlopsv1alpha1.OverrideSpec) (*mlopsv1alpha1.ComponentDefn, error) {
	if override != nil && override.Replicas != nil {
		component.Replicas = override.Replicas
	} else {
		replicas := int32(1)
		component.Replicas = &replicas
	}
	if override != nil && override.PodSpec != nil {
		var err error
		component.PodSpec, err = common.MergePodSpecs(component.PodSpec, override.PodSpec)
		if err != nil {
			return nil, err
		}
	}
	return component, nil
}

func NewSeldonRuntimeReconciler(
	runtime *mlopsv1alpha1.SeldonRuntime,
	commonConfig common.ReconcilerConfig,
	namespace string,
) (common.Reconciler, error) {
	var err error

	seldonConfig, err := mlopsv1alpha1.GetSeldonConfigForSeldonRuntime(runtime.Spec.SeldonConfig, commonConfig.Client)
	if err != nil {
		return nil, err
	}

	runtime.Spec.Config.AddDefaults(seldonConfig.Spec.Config)

	overrides := make(map[string]*mlopsv1alpha1.OverrideSpec)
	for _, o := range runtime.Spec.Overrides {
		overrides[o.Name] = o
	}

	annotator := patch.NewAnnotator(constants.LastAppliedConfig)

	var componentReconcilers []common.Reconciler
	for _, c := range seldonConfig.Spec.Components {
		override := overrides[c.Name]
		commonConfig.Logger.Info("Creating reconciler", "name", c.Name, "has override", override != nil)
		if override == nil || !override.Disable {
			commonConfig.Logger.Info("Creating component", "name", c.Name)
			c, _ = ComponentOverride(c, override)
			err = ValidateComponent(
				c,
				runtime,
				commonConfig,
				&namespace,
			)
			if err != nil {
				return nil, err
			}

			if len(c.VolumeClaimTemplates) > 0 {
				statefulSetReconcilor, err := NewComponentStatefulSetReconciler(
					c.Name,
					commonConfig,
					runtime.ObjectMeta,
					*c.Replicas,
					c.PodSpec,
					c.VolumeClaimTemplates,
					c.Labels,
					c.Annotations,
					seldonConfig.ObjectMeta,
					annotator,
				)
				if err != nil {
					return nil, err
				}
				componentReconcilers = append(componentReconcilers, statefulSetReconcilor)
			} else {
				deploymentReconcilor, err := NewComponentDeploymentReconciler(
					c.Name,
					commonConfig,
					runtime.ObjectMeta,
					*c.Replicas,
					c.PodSpec,
					c.Labels,
					c.Annotations,
					seldonConfig.ObjectMeta,
					annotator,
				)
				if err != nil {
					return nil, err
				}
				componentReconcilers = append(componentReconcilers, deploymentReconcilor)
			}
		} else {
			commonConfig.Logger.Info("Disabling component", "name", c.Name)
		}
	}
	// Set last applied annotation for update
	for _, cr := range componentReconcilers {
		for _, res := range cr.GetResources() {
			if err := annotator.SetLastAppliedAnnotation(res); err != nil {
				return nil, err
			}
		}
	}

	configMapReconciler, err := NewConfigMapReconciler(commonConfig, &runtime.Spec.Config, runtime.ObjectMeta)
	if err != nil {
		return nil, err
	}

	svcReconciler := NewComponentServiceReconciler(commonConfig, runtime.ObjectMeta, runtime.Spec.Config.ServiceConfig, overrides, annotator)
	for _, res := range svcReconciler.GetResources() {
		if err := annotator.SetLastAppliedAnnotation(res); err != nil {
			return nil, err
		}
	}

	return &SeldonRuntimeReconciler{
		ReconcilerConfig:     commonConfig,
		componentReconcilers: componentReconcilers,
		rbacReconciler:       NewComponentRBACReconciler(commonConfig, runtime.ObjectMeta),
		serviceReconciler:    svcReconciler,
		configMapReconciler:  configMapReconciler,
	}, nil
}

func (s *SeldonRuntimeReconciler) GetResources() []client.Object {
	var objs []client.Object
	for _, c := range s.componentReconcilers {
		objs = append(objs, c.GetResources()...)
	}
	objs = append(objs, s.rbacReconciler.GetResources()...)
	objs = append(objs, s.serviceReconciler.GetResources()...)
	objs = append(objs, s.configMapReconciler.GetResources()...)
	return objs
}

func (s *SeldonRuntimeReconciler) GetConditions() []*apis.Condition {
	var conditions []*apis.Condition
	for _, c := range s.componentReconcilers {
		conditions = append(conditions, c.GetConditions()...)
	}
	conditions = append(conditions, s.rbacReconciler.GetConditions()...)
	conditions = append(conditions, s.serviceReconciler.GetConditions()...)
	conditions = append(conditions, s.configMapReconciler.GetConditions()...)
	return conditions
}

func (s *SeldonRuntimeReconciler) Reconcile() error {
	err := s.rbacReconciler.Reconcile()
	if err != nil {
		return err
	}
	err = s.configMapReconciler.Reconcile()
	if err != nil {
		return err
	}
	err = s.serviceReconciler.Reconcile()
	if err != nil {
		return err
	}
	for _, c := range s.componentReconcilers {
		err := c.Reconcile()
		if err != nil {
			return err
		}
	}
	return nil
}

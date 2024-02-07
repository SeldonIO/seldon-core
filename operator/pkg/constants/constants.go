/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package constants

import "os"

const (
	ModelFinalizerName      = "seldon.model.finalizer"
	PipelineFinalizerName   = "seldon.pipeline.finalizer"
	ExperimentFinalizerName = "seldon.experiment.finalizer"
	RuntimeFinalizerName    = "seldon.runtime.finalizer"
)

func getEnvOrDefault(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

var (
	SeldonNamespace = getEnvOrDefault("POD_NAMESPACE", "seldon-mesh")
)

// Label selector
const (
	KubernetesNameLabelKey    = "app.kubernetes.io/name"
	ServerLabelValue          = "seldon-server"
	ServerLabelNameKey        = "seldon-server-name"
	ServerReplicaLabelKey     = "seldon-server-replica"
	ServerReplicaNameLabelKey = "seldon-server-replica-name"
	ControlPlaneLabelKey      = "control-plane"
	LastAppliedConfig         = "seldon.io/last-applied"
)

// Reconcilliation operations
type ReconcileOperation uint32

const (
	ReconcileUnknown ReconcileOperation = iota
	ReconcileNoChange
	ReconcileUpdateNeeded
	ReconcileCreateNeeded
)

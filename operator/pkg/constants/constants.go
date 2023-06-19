/*
Copyright 2022 Seldon Technologies Ltd.

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

package constants

import "os"

const (
	ModelFinalizerName      = "seldon.model.finalizer"
	ServerFinalizerName     = "seldon.server.finalizer"
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
	AppKey                    = "app"
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

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

package controlplane_factory

import (
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/interfaces"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/agent/modelserver_controlplane/oip"
)

func CreateModelServerControlPlane(
	modelServerType string,
	config interfaces.ModelServerConfig,
) (interfaces.ModelServerControlPlaneClient, error) {
	// we only support v2 for now
	return oip.NewV2Client(config.Host, config.Port, config.Logger), nil
}

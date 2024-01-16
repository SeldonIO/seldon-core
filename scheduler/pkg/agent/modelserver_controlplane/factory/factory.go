/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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

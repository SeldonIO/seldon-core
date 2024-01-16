/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package filters

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type SharingServerFilter struct{}

func (e SharingServerFilter) Name() string {
	return "SharingServerFilter"
}

func (e SharingServerFilter) Filter(model *store.ModelVersion, server *store.ServerSnapshot) bool {
	requestedServer := model.GetRequestedServer()
	return (requestedServer == nil && server.Shared) || (requestedServer != nil && *requestedServer == server.Name)
}

func (e SharingServerFilter) Description(model *store.ModelVersion, server *store.ServerSnapshot) string {
	requestedServer := model.GetRequestedServer()
	if requestedServer != nil {
		return fmt.Sprintf("requested server %s == %s", *requestedServer, server.Name)
	} else {
		return fmt.Sprintf("sharing %v", server.Shared)
	}
}

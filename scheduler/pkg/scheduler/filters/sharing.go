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

package filters

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/pkg/store"
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

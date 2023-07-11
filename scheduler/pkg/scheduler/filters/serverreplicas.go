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

package filters

import (
	"fmt"

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type ServerReplicaFilter struct{}

func (r ServerReplicaFilter) Name() string {
	return "ServerReplicaFilter"
}

func (r ServerReplicaFilter) Filter(model *store.ModelVersion, server *store.ServerSnapshot) bool {
	return len(server.Replicas) > 0
}

func (r ServerReplicaFilter) Description(model *store.ModelVersion, server *store.ServerSnapshot) string {
	return fmt.Sprintf("%d server replicas (waiting for server replicas to connect)", len(server.Replicas))
}

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

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

type DeletedServerFilter struct{}

func (e DeletedServerFilter) Name() string {
	return "DeletedServerFilter"
}

func (e DeletedServerFilter) Filter(model *store.ModelVersion, server *store.ServerSnapshot) bool {
	return server.ExpectedReplicas != 0
}

func (e DeletedServerFilter) Description(model *store.ModelVersion, server *store.ServerSnapshot) string {
	return fmt.Sprintf("expected replicas %d != 0", server.ExpectedReplicas)
}

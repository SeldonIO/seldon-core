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

	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
)

type DeletedServerFilter struct{}

func (e DeletedServerFilter) Name() string {
	return "DeletedServerFilter"
}

func (e DeletedServerFilter) Filter(model *db.ModelVersion, server *db.Server) bool {
	return server.ExpectedReplicas != 0
}

func (e DeletedServerFilter) Description(model *db.ModelVersion, server *db.Server) string {
	return fmt.Sprintf("expected replicas %d != 0", server.ExpectedReplicas)
}

/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package filters

import (
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
)

type ReplicaFilter interface {
	Name() string
	Filter(model *db.ModelVersion, replica *db.ServerReplica) bool
	Description(model *db.ModelVersion, replica *db.ServerReplica) string
}

type ServerFilter interface {
	Name() string
	Filter(model *db.ModelVersion, server *db.Server) bool
	Description(model *db.ModelVersion, server *db.Server) string
}

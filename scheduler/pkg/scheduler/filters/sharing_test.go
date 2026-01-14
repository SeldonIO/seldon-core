/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package filters

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler/db"
	"github.com/seldonio/seldon-core/scheduler/v2/pkg/util"

	pb "github.com/seldonio/seldon-core/apis/go/v2/mlops/scheduler"
)

func TestSharingFilter(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		model    *db.ModelVersion
		server   *db.Server
		expected bool
	}
	serverName := "server1"
	modelExplicitServer := util.NewTestModelVersion(
		&pb.Model{ModelSpec: &pb.ModelSpec{Server: &serverName}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
		1,
		serverName,
		map[int32]*db.ReplicaStatus{3: {State: db.ModelReplicaState_Loading}},
		db.ModelState_ModelProgressing)
	modelSharedServer := util.NewTestModelVersion(
		&pb.Model{ModelSpec: &pb.ModelSpec{}, DeploymentSpec: &pb.DeploymentSpec{Replicas: 1}},
		1,
		serverName,
		map[int32]*db.ReplicaStatus{3: {State: db.ModelReplicaState_Loading}},
		db.ModelState_ModelProgressing)
	tests := []test{
		{name: "ModelAndServerMatchNotShared", model: modelExplicitServer, server: &db.Server{Name: serverName, Shared: false}, expected: true},
		{name: "ModelAndServerMatchShared", model: modelExplicitServer, server: &db.Server{Name: serverName, Shared: true}, expected: true},
		{name: "ModelAndServerDontMatch", model: modelExplicitServer, server: &db.Server{Name: "foo", Shared: true}, expected: false},
		{name: "SharedModelAnyServer", model: modelSharedServer, server: &db.Server{Name: "foo", Shared: true}, expected: true},
		{name: "SharedModelNotSharedServer", model: modelSharedServer, server: &db.Server{Name: "foo", Shared: false}, expected: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := SharingServerFilter{}
			ok := filter.Filter(test.model, test.server)
			g.Expect(ok).To(Equal(test.expected))
		})
	}
}

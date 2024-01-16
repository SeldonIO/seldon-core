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

	"github.com/seldonio/seldon-core/scheduler/v2/pkg/store"
)

func TestReplicaDrainingFilter(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name       string
		isDraining bool
		expected   bool
	}

	tests := []test{
		{name: "WithDraining", isDraining: true, expected: false},
		{name: "NoDraining", isDraining: false, expected: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filter := ReplicaDrainingFilter{}
			replica := store.ServerReplica{}
			if test.isDraining {
				replica.SetIsDraining()
			}
			ok := filter.Filter(nil, &replica)
			g.Expect(ok).To(Equal(test.expected))
		})
	}
}

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

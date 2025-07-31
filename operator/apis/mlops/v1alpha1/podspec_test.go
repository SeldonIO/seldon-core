/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func TestToV1PodSped(t *testing.T) {
	t.Parallel()

	g := NewGomegaWithT(t)

	type test struct {
		name         string
		mlopsPodSpec *PodSpec
		podSpec      *v1.PodSpec
	}

	tests := []test{
		{
			name: "Simple",
			mlopsPodSpec: &PodSpec{
				InitContainers: []v1.Container{
					{
						Name:  "c1",
						Image: "myimagec1:1",
					},
				},
				NodeName: "node",
			},
			podSpec: &v1.PodSpec{
				InitContainers: []v1.Container{
					{
						Name:  "c1",
						Image: "myimagec1:1",
					},
				},
				NodeName: "node",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			podNew, err := test.mlopsPodSpec.ToV1PodSpec()
			g.Expect(err).To(BeNil())
			g.Expect(equality.Semantic.DeepEqual(podNew, test.podSpec)).To(BeTrue())
		})
	}
}

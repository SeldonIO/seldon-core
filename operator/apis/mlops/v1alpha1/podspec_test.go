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

package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
)

func TestToV1PodSped(t *testing.T) {
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
			podNew, err := test.mlopsPodSpec.ToV1PodSpec()
			g.Expect(err).To(BeNil())
			g.Expect(equality.Semantic.DeepEqual(podNew, test.podSpec)).To(BeTrue())
		})
	}
}

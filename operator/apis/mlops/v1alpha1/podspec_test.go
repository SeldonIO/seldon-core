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

package utils

import (
	"encoding/json"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"testing"
)

func getTestDeployment(replicas *int32) *v1.Deployment {
	d := &v1.Deployment{
		Spec: v1.DeploymentSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name: "mycontainer",
						},
					},
				},
			},
		},
	}
	if replicas != nil {
		d.Spec.Replicas = replicas
	}
	return d
}

func TestReplica(t *testing.T) {
	g := NewGomegaWithT(t)
	type test struct {
		name         string
		dep1         *v1.Deployment
		dep2         *v1.Deployment
		expectedDep1 *v1.Deployment
		expectedDep2 *v1.Deployment
	}
	getInt32Ptr := func(v int32) *int32 { return &v }
	tests := []test{
		{
			name:         "no replicas",
			dep1:         getTestDeployment(nil),
			dep2:         getTestDeployment(nil),
			expectedDep1: getTestDeployment(nil),
			expectedDep2: getTestDeployment(nil),
		},
		{
			name:         "first dep with replicas",
			dep1:         getTestDeployment(getInt32Ptr(1)),
			dep2:         getTestDeployment(nil),
			expectedDep1: getTestDeployment(nil),
			expectedDep2: getTestDeployment(nil),
		},
		{
			name:         "second dep with replicas",
			dep1:         getTestDeployment(nil),
			dep2:         getTestDeployment(getInt32Ptr(1)),
			expectedDep1: getTestDeployment(nil),
			expectedDep2: getTestDeployment(nil),
		},
		{
			name:         "both deps with replicas",
			dep1:         getTestDeployment(getInt32Ptr(2)),
			dep2:         getTestDeployment(getInt32Ptr(4)),
			expectedDep1: getTestDeployment(nil),
			expectedDep2: getTestDeployment(nil),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			f := IgnoreReplicas()
			b1, err := json.Marshal(test.dep1)
			g.Expect(err).To(BeNil())
			b2, err := json.Marshal(test.dep2)
			g.Expect(err).To(BeNil())

			b3, b4, err := f(b1, b2)
			g.Expect(err).To(BeNil())

			d1 := &v1.Deployment{}
			err = json.Unmarshal(b3, d1)
			g.Expect(err).To(BeNil())

			d2 := &v1.Deployment{}
			err = json.Unmarshal(b4, d2)
			g.Expect(err).To(BeNil())

			g.Expect(d1).To(Equal(test.expectedDep1))
			g.Expect(d2).To(Equal(test.expectedDep2))
		})
	}
}

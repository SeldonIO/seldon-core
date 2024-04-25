package utils

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
)

const HOST_ANNOTATION = "host-annotation"

var _ = Describe("Controller utils", func() {
	DescribeTable(
		"isEmptyExplainer",
		func(explainer *machinelearningv1.Explainer, expected bool) {
			empty := IsEmptyExplainer(explainer)
			Expect(empty).To(Equal(expected))
		},
		Entry("empty if nil", nil, true),
		Entry("empty if unset type", &machinelearningv1.Explainer{}, true),
		Entry(
			"not empty otherwise",
			&machinelearningv1.Explainer{Type: machinelearningv1.AlibiAnchorsImageExplainer},
			false,
		),
	)
})

func TestHostsFromAnnotation(t *testing.T) {
	var key = HOST_ANNOTATION
	var fallback = "*"

	tests := []struct {
		name  string
		mldep *machinelearningv1.SeldonDeployment
		want  []string
	}{
		{
			name: "No host",
			mldep: &machinelearningv1.SeldonDeployment{
				Spec: machinelearningv1.SeldonDeploymentSpec{
					Annotations: map[string]string{},
				},
			},
			want: []string{fallback},
		},
		{
			name: "Single host",
			mldep: &machinelearningv1.SeldonDeployment{
				Spec: machinelearningv1.SeldonDeploymentSpec{
					Annotations: map[string]string{
						key: "prod.svc.cluster.local",
					},
				},
			},
			want: []string{"prod.svc.cluster.local"},
		},
		{
			name: "Multiple hosts",
			mldep: &machinelearningv1.SeldonDeployment{
				Spec: machinelearningv1.SeldonDeploymentSpec{
					Annotations: map[string]string{
						key: "prod.svc.cluster.local,dev.svc.cluster.local",
					},
				},
			},
			want: []string{"prod.svc.cluster.local", "dev.svc.cluster.local"},
		},
		{
			name: "Multiple hosts with space",
			mldep: &machinelearningv1.SeldonDeployment{
				Spec: machinelearningv1.SeldonDeploymentSpec{
					Annotations: map[string]string{
						key: "prod.svc.cluster.local, dev.svc.cluster.local",
					},
				},
			},
			want: []string{"prod.svc.cluster.local", "dev.svc.cluster.local"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hosts := HostsFromAnnotation(tt.mldep, key, fallback)
			for i, host := range hosts {
				if host != tt.want[i] {
					t.Errorf("got: %v; want %v", host, tt.want[i])
				}
			}
		})
	}
}

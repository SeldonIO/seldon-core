package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/seldonio/seldon-core/operatorv2/pkg/constants"
	testing2 "github.com/seldonio/seldon-core/operatorv2/pkg/utils/testing"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestGetServerConfigForServer(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		serverConfig *ServerConfig
		serverType   ServerType
		err          bool
	}

	tests := []test{
		{
			name: "Mlserver",
			serverConfig: &ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mlserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: ServerConfigSpec{
					PodSpec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "mlserver",
								Image: "seldonio/mlserver:0.5",
							},
						},
					},
				},
			},
			serverType: MLServerServerType,
		},
		{
			name: "Triton",
			serverConfig: &ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "triton",
					Namespace: constants.SeldonNamespace,
				},
				Spec: ServerConfigSpec{
					PodSpec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "triton",
								Image: "nvcr.io/nvidia/tritonserver:21.12-py3",
							},
						},
					},
				},
			},
			serverType: TritonServerType,
		},
		{
			name: "MlserverNotFound",
			serverConfig: &ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "triton",
					Namespace: constants.SeldonNamespace,
				},
				Spec: ServerConfigSpec{
					PodSpec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "triton",
								Image: "nvcr.io/nvidia/tritonserver:21.12-py3",
							},
						},
					},
				},
			},
			serverType: MLServerServerType,
			err:        true,
		},
		{
			name: "No Server type",
			serverConfig: &ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "triton",
					Namespace: constants.SeldonNamespace,
				},
				Spec: ServerConfigSpec{
					PodSpec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "triton",
								Image: "nvcr.io/nvidia/tritonserver:21.12-py3",
							},
						},
					},
				},
			},
			serverType: "",
			err:        true,
		},
		{
			name: "Unknown server type",
			serverConfig: &ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "triton",
					Namespace: constants.SeldonNamespace,
				},
				Spec: ServerConfigSpec{
					PodSpec: v1.PodSpec{
						Containers: []v1.Container{
							{
								Name:  "triton",
								Image: "nvcr.io/nvidia/tritonserver:21.12-py3",
							},
						},
					},
				},
			},
			serverType: "foo",
			err:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			err := AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			client := testing2.NewFakeClient(scheme, test.serverConfig)
			sc, err := GetServerConfigForServer(test.serverType, client)
			if !test.err {
				g.Expect(err).To(BeNil())
				g.Expect(equality.Semantic.DeepEqual(sc.Spec, test.serverConfig.Spec)).To(BeTrue())
			} else {
				g.Expect(err).ToNot(BeNil())
			}
		})
	}

}

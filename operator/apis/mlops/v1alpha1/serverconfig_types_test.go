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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/seldonio/seldon-core/operator/v2/pkg/constants"
	testing2 "github.com/seldonio/seldon-core/operator/v2/pkg/utils/testing"
)

func TestGetServerConfigForServer(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		serverConfig *ServerConfig
		configName   string
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
			configName: "mlserver",
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
			configName: "triton",
		},
		{
			name: "MlserverNotFound",
			serverConfig: &ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mlserver",
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
			configName: "foo",
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
			configName: "",
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
			configName: "foo",
			err:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			scheme := runtime.NewScheme()
			err := AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			client := testing2.NewFakeClient(scheme, test.serverConfig)
			sc, err := GetServerConfigForServer(test.configName, client)
			if !test.err {
				g.Expect(err).To(BeNil())
				g.Expect(equality.Semantic.DeepEqual(sc.Spec, test.serverConfig.Spec)).To(BeTrue())
			} else {
				g.Expect(err).ToNot(BeNil())
			}
		})
	}

}

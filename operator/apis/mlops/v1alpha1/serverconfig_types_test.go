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

package reconcilers

import (
	"context"
	"testing"

	logrtest "github.com/go-logr/logr/testr"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operatorv2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operatorv2/pkg/constants"
	testing2 "github.com/seldonio/seldon-core/operatorv2/pkg/utils/testing"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/gomega"
)

func TestServerReconcile(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                    string
		serverConfig            *mlopsv1alpha1.ServerConfig
		server                  *mlopsv1alpha1.Server
		error                   bool
		expectedSvcNames        []string
		expectedStatefulSetName string
	}
	mlserverConfigName := "mlserver-config"
	tests := []test{
		{
			name: "MLServer",
			serverConfig: &mlopsv1alpha1.ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      mlserverConfigName,
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerConfigSpec{
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
			server: &mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerSpec{
					ServerConfig: mlserverConfigName,
				},
			},
			expectedSvcNames:        []string{"myserver-0"},
			expectedStatefulSetName: "myserver",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrtest.New(t)
			var client client2.Client
			scheme := runtime.NewScheme()
			err := mlopsv1alpha1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			err = v1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			err = appsv1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			if test.serverConfig != nil {
				client = testing2.NewFakeClient(scheme, test.serverConfig)
			} else {
				client = testing2.NewFakeClient(scheme)
			}
			g.Expect(err).To(BeNil())
			sr, err := NewServerReconciler(test.server, common.ReconcilerConfig{Ctx: context.TODO(), Logger: logger, Client: client})
			g.Expect(err).To(BeNil())
			err = sr.Reconcile()
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				for _, svcName := range test.expectedSvcNames {
					svc := &v1.Service{}
					err := client.Get(context.TODO(), types.NamespacedName{
						Name:      svcName,
						Namespace: test.server.GetNamespace(),
					}, svc)
					g.Expect(err).To(BeNil())
				}
				ss := &appsv1.StatefulSet{}
				err := client.Get(context.TODO(), types.NamespacedName{
					Name:      test.expectedStatefulSetName,
					Namespace: test.server.GetNamespace(),
				}, ss)
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func TestNewServerReconciler(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name         string
		serverConfig *mlopsv1alpha1.ServerConfig
		server       *mlopsv1alpha1.Server
		error        bool
	}
	tests := []test{
		{
			name: "MLServer",
			serverConfig: &mlopsv1alpha1.ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mlserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerConfigSpec{
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
			server: &mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerSpec{
					ServerConfig: "mlserver",
				},
			},
		},
		{
			name: "MissingServer",
			server: &mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerSpec{
					ServerConfig: "mlserver",
				},
			},
			error: true,
		},
		{
			name: "CustomPodSpec",
			serverConfig: &mlopsv1alpha1.ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "mlserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerConfigSpec{
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
			server: &mlopsv1alpha1.Server{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "myserver",
					Namespace: constants.SeldonNamespace,
				},
				Spec: mlopsv1alpha1.ServerSpec{
					ServerConfig: "mlserver",
					PodSpec: &mlopsv1alpha1.PodSpec{
						NodeName: "node",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrtest.New(t)
			var client client2.Client
			scheme := runtime.NewScheme()
			err := mlopsv1alpha1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			if test.serverConfig != nil {
				client = testing2.NewFakeClient(scheme, test.serverConfig)
			} else {
				client = testing2.NewFakeClient(scheme)
			}
			g.Expect(err).To(BeNil())
			_, err = NewServerReconciler(test.server, common.ReconcilerConfig{Ctx: context.TODO(), Logger: logger, Client: client})
			if test.error {
				g.Expect(err).ToNot(BeNil())
			} else {
				g.Expect(err).To(BeNil())
			}
		})
	}
}

func TestMergePodSpecs(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name          string
		serverPodSpec *v1.PodSpec
		override      *mlopsv1alpha1.PodSpec
		expected      *v1.PodSpec
	}

	tests := []test{
		{
			name: "NoOverride",
			serverPodSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c1",
						Image:   "myimagec1:1",
						Command: []string{"cmd"},
					},
				},
				NodeName: "node",
			},
			expected: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c1",
						Image:   "myimagec1:1",
						Command: []string{"cmd"},
					},
				},
				NodeName: "node",
			},
		},
		{
			name: "Override with new container",
			serverPodSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c1",
						Image:   "myimagec1:1",
						Command: []string{"cmd"},
					},
				},
				NodeName: "node",
			},
			override: &mlopsv1alpha1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c2",
						Image:   "myimagec2:2",
						Command: []string{"cmd2"},
					},
				},
				NodeName: "node2",
			},
			expected: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c2",
						Image:   "myimagec2:2",
						Command: []string{"cmd2"},
					},
					{
						Name:    "c1",
						Image:   "myimagec1:1",
						Command: []string{"cmd"},
					},
				},
				NodeName: "node2",
			},
		},
		{
			name: "Override with existing container",
			serverPodSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c1",
						Image:   "myimagec1:1",
						Command: []string{"cmd"},
					},
				},
				NodeName: "node",
			},
			override: &mlopsv1alpha1.PodSpec{
				Containers: []v1.Container{
					{
						Name:  "c1",
						Image: "myimagec2:2",
					},
				},
				NodeName: "node2",
			},
			expected: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c1",
						Image:   "myimagec2:2",
						Command: []string{"cmd"},
					},
				},
				NodeName: "node2",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			podSpec, err := mergePodSpecs(test.serverPodSpec, test.override)
			g.Expect(err).To(BeNil())
			g.Expect(equality.Semantic.DeepEqual(podSpec, test.expected)).To(BeTrue())
		})
	}
}

func TestUpdateServerCapabilities(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                 string
		extraCapabilities    []string
		podSpec              *v1.PodSpec
		expectedCapabilities string
	}
	tests := []test{
		{
			name: "add capability",
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c1",
						Image:   "myimagec1:1",
						Command: []string{"cmd"},
						Env: []v1.EnvVar{
							{
								Name:  EnvVarNameCapabilities,
								Value: "foo",
							},
						},
					},
				},
				NodeName: "node",
			},
			extraCapabilities:    []string{"bar"},
			expectedCapabilities: "foo,bar",
		},
		{
			name: "no new capability",
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c1",
						Image:   "myimagec1:1",
						Command: []string{"cmd"},
						Env: []v1.EnvVar{
							{
								Name:  EnvVarNameCapabilities,
								Value: "foo",
							},
						},
					},
				},
				NodeName: "node",
			},
			extraCapabilities:    []string{},
			expectedCapabilities: "foo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updateCapabilities(test.extraCapabilities, test.podSpec)
			for _, container := range test.podSpec.Containers {
				for _, envVar := range container.Env {
					if envVar.Name == EnvVarNameCapabilities {
						g.Expect(envVar.Value).To(Equal(test.expectedCapabilities))
					}
				}
			}
		})
	}
}

func TestMergeContainers(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name     string
		existing []v1.Container
		override []v1.Container
		expected []v1.Container
	}

	tests := []test{
		{
			name: "different containers",
			existing: []v1.Container{
				{
					Name:  "c1",
					Image: "imagec1",
				},
			},
			override: []v1.Container{
				{
					Name:  "c2",
					Image: "imagec2",
				},
			},
			expected: []v1.Container{
				{
					Name:  "c2",
					Image: "imagec2",
				},
				{
					Name:  "c1",
					Image: "imagec1",
				},
			},
		},
		{
			name: "same container",
			existing: []v1.Container{
				{
					Name:    "c1",
					Image:   "imagec1",
					Command: []string{"cmd"},
				},
			},
			override: []v1.Container{
				{
					Name:  "c1",
					Image: "imagec2",
					Args:  []string{"arg"},
				},
			},
			expected: []v1.Container{
				{
					Name:    "c1",
					Image:   "imagec2",
					Command: []string{"cmd"},
					Args:    []string{"arg"},
				},
			},
		},
		{
			name: "mix of containers",
			existing: []v1.Container{
				{
					Name:  "c1",
					Image: "imagec1",
				},
				{
					Name:  "c2",
					Image: "imagec2",
				},
			},
			override: []v1.Container{
				{
					Name:  "c1",
					Image: "imagec2",
				},
				{
					Name:  "c3",
					Image: "imagec3",
				},
			},
			expected: []v1.Container{
				{
					Name:  "c1",
					Image: "imagec2",
				},
				{
					Name:  "c3",
					Image: "imagec3",
				},
				{
					Name:  "c2",
					Image: "imagec2",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			containers, err := mergeContainers(test.existing, test.override)
			g.Expect(err).To(BeNil())
			g.Expect(equality.Semantic.DeepEqual(containers, test.expected)).To(BeTrue())
		})
	}
}

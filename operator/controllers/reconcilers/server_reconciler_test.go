package reconcilers

import (
	"context"
	"testing"

	logrtest "github.com/go-logr/logr/testing"
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
	mlserverServer := mlopsv1alpha1.MLServerServerType
	tests := []test{
		{
			name: "MLServer",
			serverConfig: &mlopsv1alpha1.ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      string(mlserverServer),
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
					Server: mlopsv1alpha1.ServerDefn{
						Type: mlserverServer,
					},
				},
			},
			expectedSvcNames:        []string{"myserver-0"},
			expectedStatefulSetName: "myserver",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrtest.TestLogger{T: t}
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
					Name:      string(mlopsv1alpha1.MLServerServerType),
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
					Server: mlopsv1alpha1.ServerDefn{
						Type: mlopsv1alpha1.MLServerServerType,
					},
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
					Server: mlopsv1alpha1.ServerDefn{
						Type: mlopsv1alpha1.MLServerServerType,
					},
				},
			},
			error: true,
		},
		{
			name: "CustomPodSpec",
			serverConfig: &mlopsv1alpha1.ServerConfig{
				ObjectMeta: metav1.ObjectMeta{
					Name:      string(mlopsv1alpha1.MLServerServerType),
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
					Server: mlopsv1alpha1.ServerDefn{
						Type: mlopsv1alpha1.MLServerServerType,
					},
					PodSpec: &mlopsv1alpha1.PodSpec{
						NodeName: "node",
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrtest.TestLogger{T: t}
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
			name: "Override",
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
						Name:    "c1",
						Image:   "myimagec1:1",
						Command: []string{"cmd"},
					},
					{
						Name:    "c2",
						Image:   "myimagec2:2",
						Command: []string{"cmd2"},
					},
				},
				NodeName: "node2",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			podSepc, err := mergePodSpecs(test.serverPodSpec, test.override)
			g.Expect(err).To(BeNil())
			g.Expect(equality.Semantic.DeepEqual(podSepc, test.expected)).To(BeTrue())
		})
	}
}

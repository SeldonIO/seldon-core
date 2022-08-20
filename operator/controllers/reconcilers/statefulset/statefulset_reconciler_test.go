package statefulset

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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	client2 "sigs.k8s.io/controller-runtime/pkg/client"

	. "github.com/onsi/gomega"
)

func TestReconcile(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                 string
		meta                 metav1.ObjectMeta
		podSpec              *v1.PodSpec
		volumeClaimTemplates []mlopsv1alpha1.PersistentVolumeClaim
		scaling              *mlopsv1alpha1.ScalingSpec
		existing             *appsv1.StatefulSet
		expectedReconcileOp  constants.ReconcileOperation
	}

	getIntPtr := func(i int32) *int32 {
		return &i
	}
	oneG, err := resource.ParseQuantity("1G")
	g.Expect(err).To(BeNil())
	tests := []test{
		{
			name: "Create",
			meta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "default",
			},
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c1",
						Image:   "myimagec1:1",
						Command: []string{"cmd"},
					},
				},
				NodeName: "node",
			},
			volumeClaimTemplates: []mlopsv1alpha1.PersistentVolumeClaim{
				{
					Name: "model-repository",
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: oneG,
							},
						},
					},
				},
			},
			scaling: &mlopsv1alpha1.ScalingSpec{
				Replicas: getIntPtr(2),
			},
			expectedReconcileOp: constants.ReconcileCreateNeeded,
		},
		{
			name: "Update",
			meta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "default",
			},
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c1",
						Image:   "myimagec1:1",
						Command: []string{"cmd"},
					},
				},
				NodeName: "node",
			},
			volumeClaimTemplates: []mlopsv1alpha1.PersistentVolumeClaim{
				{
					Name: "model-repository",
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: oneG,
							},
						},
					},
				},
			},
			scaling: &mlopsv1alpha1.ScalingSpec{
				Replicas: getIntPtr(2),
			},
			existing: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
			},
			expectedReconcileOp: constants.ReconcileUpdateNeeded,
		},
		{
			name: "Same",
			meta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "default",
			},
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c1",
						Image:   "myimagec1:1",
						Command: []string{"cmd"},
					},
				},
				NodeName: "node",
			},
			volumeClaimTemplates: []mlopsv1alpha1.PersistentVolumeClaim{
				{
					Name: "model-repository",
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: oneG,
							},
						},
					},
				},
			},
			scaling: &mlopsv1alpha1.ScalingSpec{
				Replicas: getIntPtr(2),
			},
			existing: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
					Labels:    map[string]string{constants.AppKey: constants.ServerLabelValue},
				},
				Spec: appsv1.StatefulSetSpec{
					ServiceName: "foo",
					Replicas:    getIntPtr(2),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{constants.ServerLabelNameKey: "foo"},
					},
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels:    map[string]string{constants.ServerLabelNameKey: "foo", constants.AppKey: constants.ServerLabelValue},
							Name:      "foo",
							Namespace: "default",
						},
						Spec: v1.PodSpec{
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
					PodManagementPolicy: appsv1.ParallelPodManagement,
					VolumeClaimTemplates: []v1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "model-repository",
							},
							Spec: v1.PersistentVolumeClaimSpec{
								AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceStorage: oneG,
									},
								},
							},
						},
					},
				},
			},
			expectedReconcileOp: constants.ReconcileNoChange,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			logger := logrtest.New(t)
			var client client2.Client
			scheme := runtime.NewScheme()
			err := appsv1.AddToScheme(scheme)
			g.Expect(err).To(BeNil())
			if test.existing != nil {
				client = testing2.NewFakeClient(scheme, test.existing)
			} else {
				client = testing2.NewFakeClient(scheme)
			}
			g.Expect(err).To(BeNil())
			r := NewStatefulSetReconciler(common.ReconcilerConfig{Ctx: context.TODO(), Logger: logger, Client: client}, test.meta, test.podSpec, test.volumeClaimTemplates, test.scaling)
			rop, err := r.getReconcileOperation()
			g.Expect(rop).To(Equal(test.expectedReconcileOp))
			g.Expect(err).To(BeNil())
			err = r.Reconcile()
			g.Expect(err).To(BeNil())
			found := &appsv1.StatefulSet{}
			err = client.Get(context.TODO(), types.NamespacedName{
				Name:      r.StatefulSet.GetName(),
				Namespace: r.StatefulSet.GetNamespace()}, found)
			g.Expect(err).To(BeNil())
			g.Expect(equality.Semantic.DeepEqual(found.Spec, r.StatefulSet.Spec))
		})
	}
}

func TestToStatefulSet(t *testing.T) {
	g := NewGomegaWithT(t)

	type test struct {
		name                 string
		meta                 metav1.ObjectMeta
		podSpec              *v1.PodSpec
		volumeClaimTemplates []mlopsv1alpha1.PersistentVolumeClaim
		scaling              *mlopsv1alpha1.ScalingSpec
		statefulSet          *appsv1.StatefulSet
	}

	getIntPtr := func(i int32) *int32 {
		return &i
	}
	oneG, err := resource.ParseQuantity("1G")
	g.Expect(err).To(BeNil())
	tests := []test{
		{
			name: "Basic",
			meta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "default",
			},
			podSpec: &v1.PodSpec{
				Containers: []v1.Container{
					{
						Name:    "c1",
						Image:   "myimagec1:1",
						Command: []string{"cmd"},
					},
				},
				NodeName: "node",
			},
			volumeClaimTemplates: []mlopsv1alpha1.PersistentVolumeClaim{
				{
					Name: "model-repository",
					Spec: v1.PersistentVolumeClaimSpec{
						AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
						Resources: v1.ResourceRequirements{
							Requests: v1.ResourceList{
								v1.ResourceStorage: oneG,
							},
						},
					},
				},
			},
			scaling: &mlopsv1alpha1.ScalingSpec{
				Replicas: getIntPtr(2),
			},
			statefulSet: &appsv1.StatefulSet{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "foo",
					Namespace: "default",
					Labels:    map[string]string{constants.AppKey: constants.ServerLabelValue},
				},
				Spec: appsv1.StatefulSetSpec{
					ServiceName: "foo",
					Replicas:    getIntPtr(2),
					Selector: &metav1.LabelSelector{
						MatchLabels: map[string]string{constants.ServerLabelNameKey: "foo"},
					},
					Template: v1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels:    map[string]string{constants.ServerLabelNameKey: "foo", constants.AppKey: constants.ServerLabelValue},
							Name:      "foo",
							Namespace: "default",
						},
						Spec: v1.PodSpec{
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
					PodManagementPolicy: appsv1.ParallelPodManagement,
					VolumeClaimTemplates: []v1.PersistentVolumeClaim{
						{
							ObjectMeta: metav1.ObjectMeta{
								Name: "model-repository",
							},
							Spec: v1.PersistentVolumeClaimSpec{
								AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
								Resources: v1.ResourceRequirements{
									Requests: v1.ResourceList{
										v1.ResourceStorage: oneG,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			statefulSet := toStatefulSet(test.meta, test.podSpec, test.volumeClaimTemplates, test.scaling)
			g.Expect(equality.Semantic.DeepEqual(statefulSet, test.statefulSet)).To(BeTrue())
		})
	}
}

package reconcilers

import (
	"github.com/imdario/mergo"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operatorv2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operatorv2/controllers/reconcilers/service"
	"github.com/seldonio/seldon-core/operatorv2/controllers/reconcilers/statefulset"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

type ServerReconciler struct {
	common.ReconcilerConfig
	StatefulSetReconciler common.Reconciler
	ServiceReconciler     common.Reconciler
}

func NewServerReconciler(server *mlopsv1alpha1.Server,
	common common.ReconcilerConfig) (common.Reconciler, error) {
	// Ensure defaults added to server
	server.Default()

	var err error
	sr := &ServerReconciler{
		ReconcilerConfig: common,
	}
	sr.StatefulSetReconciler, err = sr.createStatefulSetReconciler(server)
	if err != nil {
		return nil, err
	}
	sr.ServiceReconciler = service.NewServiceReconciler(common, server.ObjectMeta, &server.Spec.ScalingSpec)
	return sr, nil
}

func (s *ServerReconciler) GetResources() []metav1.Object {
	objs := s.StatefulSetReconciler.GetResources()
	objs = append(objs, s.ServiceReconciler.GetResources()...)
	return objs
}

func (s *ServerReconciler) GetConditions() []*apis.Condition {
	conditions := s.StatefulSetReconciler.GetConditions()
	conditions = append(conditions, s.ServiceReconciler.GetConditions()...)
	return conditions
}

func (s *ServerReconciler) Reconcile() error {
	// Reconcile Services
	err := s.ServiceReconciler.Reconcile()
	if err != nil {
		return err
	}
	// Reconcile StatefulSet
	err = s.StatefulSetReconciler.Reconcile()
	if err != nil {
		return err
	}

	return nil
}

func (s *ServerReconciler) createStatefulSetReconciler(server *mlopsv1alpha1.Server) (*statefulset.StatefulSetReconciler, error) {
	//Get ServerConfig
	serverConfig, err := mlopsv1alpha1.GetServerConfigForServer(server.Spec.Server.Type, s.Client)
	if err != nil {
		return nil, err
	}

	//Merge specs
	podSpec, err := mergePodSpecs(&serverConfig.Spec.PodSpec, server.Spec.PodSpec)
	if err != nil {
		return nil, err
	}

	// Reconcile ReplicaSet
	statefulSetReconciler := statefulset.NewStatefulSetReconciler(s.ReconcilerConfig, server.ObjectMeta, podSpec, serverConfig.Spec.VolumeClaimTemplates, &server.Spec.ScalingSpec)
	return statefulSetReconciler, nil
}

func mergePodSpecs(serverConfigPodSpec *v1.PodSpec, override *mlopsv1alpha1.PodSpec) (*v1.PodSpec, error) {
	dst := serverConfigPodSpec.DeepCopy()
	if override != nil {
		v1PodSpecOverride, err := override.ToV1PodSpec()
		if err != nil {
			return nil, err
		}
		err = mergo.Merge(dst, v1PodSpecOverride, mergo.WithOverride, mergo.WithAppendSlice)
		if err != nil {
			return nil, err
		}
		return dst, nil
	} else {
		return dst, nil
	}
}

package reconcilers

import (
	"fmt"
	"strings"

	"github.com/imdario/mergo"
	mlopsv1alpha1 "github.com/seldonio/seldon-core/operatorv2/apis/mlops/v1alpha1"
	"github.com/seldonio/seldon-core/operatorv2/controllers/reconcilers/common"
	"github.com/seldonio/seldon-core/operatorv2/controllers/reconcilers/service"
	"github.com/seldonio/seldon-core/operatorv2/controllers/reconcilers/statefulset"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"knative.dev/pkg/apis"
)

const (
	EnvVarNameCapabilities = "SELDON_SERVER_CAPABILITIES"
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

func updateCapabilities(extraCapabilities []string, podSpec *v1.PodSpec) {
	if len(extraCapabilities) > 0 {
		for _, container := range podSpec.Containers {
			for idx, envVar := range container.Env {
				if envVar.Name == EnvVarNameCapabilities {
					capabilitiesStr := strings.Join(extraCapabilities, ",")
					val := fmt.Sprintf("%s,%s", envVar.Value, capabilitiesStr)
					container.Env[idx] = v1.EnvVar{Name: EnvVarNameCapabilities, Value: val}
				}
			}
		}
	}
}

func (s *ServerReconciler) createStatefulSetReconciler(server *mlopsv1alpha1.Server) (*statefulset.StatefulSetReconciler, error) {
	//Get ServerConfig
	serverConfig, err := mlopsv1alpha1.GetServerConfigForServer(server.Spec.ServerConfig, s.Client)
	if err != nil {
		return nil, err
	}

	//Merge specs
	podSpec, err := mergePodSpecs(&serverConfig.Spec.PodSpec, server.Spec.PodSpec)
	if err != nil {
		return nil, err
	}

	// Update capabilities
	updateCapabilities(server.Spec.ExtraCapabilities, podSpec)

	// Reconcile ReplicaSet
	statefulSetReconciler := statefulset.NewStatefulSetReconciler(s.ReconcilerConfig, server.ObjectMeta, podSpec, serverConfig.Spec.VolumeClaimTemplates, &server.Spec.ScalingSpec)
	return statefulSetReconciler, nil
}

// TODO only containers are handled correctly for merging via the name of the container. Need to hande other slices
func mergePodSpecs(serverConfigPodSpec *v1.PodSpec, override *mlopsv1alpha1.PodSpec) (*v1.PodSpec, error) {
	dst := serverConfigPodSpec.DeepCopy()
	if override != nil {
		v1PodSpecOverride, err := override.ToV1PodSpec()
		if err != nil {
			return nil, err
		}

		// remove and copy existing containers
		existingContainers := serverConfigPodSpec.Containers
		err = mergo.Merge(dst, v1PodSpecOverride, mergo.WithOverride, mergo.WithAppendSlice)
		if err != nil {
			return nil, err
		}

		// merge containers
		updatedConatiners, err := mergeContainers(existingContainers, override.Containers)
		if err != nil {
			return nil, err
		}
		dst.Containers = updatedConatiners

		return dst, nil
	} else {
		return dst, nil
	}
}

// Allow containers to be overridden. As containers are keys by name we need to merge by this key.
func mergeContainers(existing []v1.Container, overrides []v1.Container) ([]v1.Container, error) {
	var containersNew []v1.Container
	for _, containerOverride := range overrides {
		found := false
		for _, containerExisting := range existing {
			if containerOverride.Name == containerExisting.Name {
				found = true
				err := mergo.Merge(&containerExisting, containerOverride, mergo.WithOverride, mergo.WithAppendSlice)
				if err != nil {
					return nil, err
				}
				containersNew = append(containersNew, containerExisting)
			}
		}
		if !found {
			containersNew = append(containersNew, containerOverride)
		}
	}
	for _, containerExisting := range existing {
		found := false
		for _, containerOverride := range overrides {
			if containerExisting.Name == containerOverride.Name {
				found = true
			}
		}
		if !found {
			containersNew = append(containersNew, containerExisting)
		}
	}
	return containersNew, nil
}

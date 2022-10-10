package controllers

import (
	"context"
	v2 "github.com/emissary-ingress/emissary/v3/pkg/api/getambassador.io/v2"

	"github.com/go-logr/logr"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type IstioResourceCleaner struct {
	instance        *machinelearningv1.SeldonDeployment
	client          client.Client
	virtualServices []*istio.VirtualService
	logger          logr.Logger
}

func (r *IstioResourceCleaner) cleanUnusedVirtualServices() ([]*istio.VirtualService, error) {
	deleted := []*istio.VirtualService{}
	vlist := &istio.VirtualServiceList{}
	err := r.client.List(context.Background(), vlist, &client.ListOptions{Namespace: r.instance.Namespace})
	for _, vsvc := range vlist.Items {
		for _, ownerRef := range vsvc.OwnerReferences {
			if ownerRef.UID == r.instance.GetUID() {
				found := false
				for _, expectedVsvc := range r.virtualServices {
					if expectedVsvc.Name == vsvc.Name {
						found = true
						break
					}
				}
				if !found {
					r.logger.Info("Will delete VirtualService", "name", vsvc.Name, "namespace", vsvc.Namespace)
					r.client.Delete(context.Background(), &vsvc, client.PropagationPolicy(metav1.DeletePropagationForeground))
					deleted = append(deleted, vsvc.DeepCopy())
				}
			}
		}
	}
	return deleted, err
}

type AmbassadoroResourceCleaner struct {
	instance *machinelearningv1.SeldonDeployment
	client   client.Client
	mappings []*v2.Mapping
	logger   logr.Logger
}

func (a *AmbassadoroResourceCleaner) cleanUnusedAmbassadorMappings() ([]*v2.Mapping, error) {
	deleted := []*v2.Mapping{}
	mlist := &v2.MappingList{}
	err := a.client.List(context.Background(), mlist, &client.ListOptions{Namespace: a.instance.Namespace})
	for _, mapping := range mlist.Items {
		for _, ownerRef := range mapping.OwnerReferences {
			if ownerRef.UID == a.instance.GetUID() {
				found := false
				for _, expectedMapping := range a.mappings {
					if expectedMapping.Name == mapping.Name {
						found = true
						break
					}
				}
				if !found {
					a.logger.Info("Will delete Ambassador Maping", "name", mapping.Name, "namespace", mapping.Namespace)
					a.client.Delete(context.Background(), &mapping, client.PropagationPolicy(metav1.DeletePropagationForeground))
					deleted = append(deleted, mapping.DeepCopy())
				}
			}
		}
	}
	return deleted, err
}

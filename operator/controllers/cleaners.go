package controllers

import (
	"context"

	"github.com/go-logr/logr"
	machinelearningv1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	istio "istio.io/client-go/pkg/apis/networking/v1alpha3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceCleaner struct {
	instance        *machinelearningv1.SeldonDeployment
	client          client.Client
	virtualServices []*istio.VirtualService
	logger          logr.Logger
}

func (r *ResourceCleaner) cleanUnusedVirtualServices() ([]*istio.VirtualService, error) {
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

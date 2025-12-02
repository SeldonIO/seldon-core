package steps

import (
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type World struct {
	namespace    string
	kubeClient   client.Client
	Kube         kubernetes.Interface
	CurrentModel *Model
}

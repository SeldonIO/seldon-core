package steps

import (
	"github.com/seldonio/seldon-core/godog/k8sclient"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type World struct {
	namespace            string
	kubeClient           k8sclient.Client
	StartingClusterState string
	CurrentModel         *Model
	Models               map[string]*Model
}

func NewWorld(namespace string, kubeClient client.Client) *World {

}

package steps

import (
	"github.com/seldonio/seldon-core/godog/k8sclient"
)

type World struct {
	namespace            string
	KubeClient           *k8sclient.K8sClient
	StartingClusterState string //todo: this will be a combination of starting state awareness of core 2 such as the
	//todo:  server config,seldon config and seldon runtime to be able to reconcile to starting state should we change
	//todo: the state such as reducing replicas to 0 of scheduler to test unavailability
	CurrentModel *Model
}

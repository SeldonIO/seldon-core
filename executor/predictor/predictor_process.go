package predictor

import (
	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api/client"
	api "github.com/seldonio/seldon-core/executor/api/grpc"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type PredictorProcess struct {
	predictor *v1alpha2.PredictorSpec
	client *client.SeldonMessageClient
	Log logr.Logger
}

func NewPredictorProcess(predictor *v1alpha2.PredictorSpec) *PredictorProcess {
	return &PredictorProcess{
		predictor,
		client.NewSeldonMessageClient(),
		logf.Log.WithName("SeldonMessageClient"),
	}
}


func (p *PredictorProcess) Execute(node *v1alpha2.PredictiveUnit, msg *api.SeldonMessage) (*api.SeldonMessage, *int, error) {

	var resp *api.SeldonMessage
	var respCode *int
	var err error
	if *node.Type == v1alpha2.MODEL {
		resp, respCode, err = p.client.Predict(node.Endpoint.ServiceHost,node.Endpoint.ServicePort,msg)
		if err != nil {
			return resp,respCode, err
		}
	}
	return resp, respCode, nil
}
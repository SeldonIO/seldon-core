package predictor

import (
	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
)

type PredictorProcess struct {
	Predictor *v1alpha2.PredictorSpec
	Client    client.SeldonApiClient
	Log       logr.Logger
}


func (p *PredictorProcess) transformInput(node *v1alpha2.PredictiveUnit, msg client.SeldonPayload) (client.SeldonPayload, error) {
	var resp client.SeldonPayload
	var err error
	switch *node.Type {
	case v1alpha2.MODEL:
		resp, err = p.Client.Predict(node.Endpoint.ServiceHost,node.Endpoint.ServicePort,msg)
	case v1alpha2.TRANSFORMER:
		resp, err = p.Client.Transform(node.Endpoint.ServiceHost,node.Endpoint.ServicePort,msg)
	default:
		return msg, nil
	}
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (p *PredictorProcess) Execute(node *v1alpha2.PredictiveUnit, msg client.SeldonPayload) (client.SeldonPayload, error) {
	return p.transformInput(node,msg)
}
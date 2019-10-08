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

func hasMethod(method v1alpha2.PredictiveUnitMethod,methods *[]v1alpha2.PredictiveUnitMethod) bool {
	if methods != nil {
		for _, m := range *methods {
			if m == method {
				return true
			}
		}
	}
	return false
}

func (p *PredictorProcess) transformInput(node *v1alpha2.PredictiveUnit, msg client.SeldonPayload) (client.SeldonPayload, error) {
	if (*node).Type != nil {
		switch *node.Type {
		case v1alpha2.MODEL:
			resp, err := p.Client.Predict(node.Endpoint.ServiceHost,node.Endpoint.ServicePort,msg)
			if err != nil {
				return resp, err
			} else {
				return resp, nil
			}
		case v1alpha2.TRANSFORMER:
			resp, err := p.Client.TransformInput(node.Endpoint.ServiceHost,node.Endpoint.ServicePort,msg)
			if err != nil {
				return resp, err
			} else {
				return resp, nil
			}
		}
	} else if hasMethod(v1alpha2.TRANSFORM_INPUT, node.Methods) {
		resp, err := p.Client.TransformInput(node.Endpoint.ServiceHost,node.Endpoint.ServicePort,msg)
		if err != nil {
			return resp, err
		} else {
			return resp, nil
		}
	}
	return msg, nil
}

func (p *PredictorProcess) Execute(node *v1alpha2.PredictiveUnit, msg client.SeldonPayload) (client.SeldonPayload, error) {
	return p.transformInput(node,msg)
}
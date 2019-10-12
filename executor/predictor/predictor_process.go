package predictor

import (
	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/machinelearning/v1alpha2"
	"sync"
)

type PredictorProcess struct {
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
			return p.Client.Predict(node.Endpoint.ServiceHost,node.Endpoint.ServicePort,msg)
		case v1alpha2.TRANSFORMER:
			return p.Client.TransformInput(node.Endpoint.ServiceHost,node.Endpoint.ServicePort,msg)
		}
	}
	if hasMethod(v1alpha2.TRANSFORM_INPUT, node.Methods) {
		return p.Client.TransformInput(node.Endpoint.ServiceHost,node.Endpoint.ServicePort,msg)
	}
	return msg, nil
}

func (p *PredictorProcess) transformOutput(node *v1alpha2.PredictiveUnit, msg client.SeldonPayload) (client.SeldonPayload, error) {
	if (*node).Type != nil {
		switch *node.Type {
		case v1alpha2.OUTPUT_TRANSFORMER:
			return p.Client.TransformOutput(node.Endpoint.ServiceHost,node.Endpoint.ServicePort,msg)
		}
	}
	if hasMethod(v1alpha2.TRANSFORM_OUTPUT, node.Methods) {
		return p.Client.TransformOutput(node.Endpoint.ServiceHost,node.Endpoint.ServicePort,msg)
	}
	return msg, nil
}

func (p *PredictorProcess) route(node *v1alpha2.PredictiveUnit, msg client.SeldonPayload) (int, error) {
	if (*node).Type != nil {
		switch *node.Type {
		case v1alpha2.ROUTER:
			return p.Client.Route(node.Endpoint.ServiceHost, node.Endpoint.ServicePort, msg)
		}
	}
	if hasMethod(v1alpha2.ROUTE, node.Methods) {
		return p.Client.Route(node.Endpoint.ServiceHost, node.Endpoint.ServicePort, msg)
	}
	return -1, nil
}


func (p *PredictorProcess) aggregate(node *v1alpha2.PredictiveUnit, msg []client.SeldonPayload) (client.SeldonPayload, error) {
	if (*node).Type != nil {
		switch *node.Type {
		case v1alpha2.COMBINER:
			return p.Client.Combine(node.Endpoint.ServiceHost, node.Endpoint.ServicePort, msg)
		}
	}
	if hasMethod(v1alpha2.AGGREGATE, node.Methods) {
		return p.Client.Combine(node.Endpoint.ServiceHost, node.Endpoint.ServicePort, msg)
	}
	return msg[0], nil
}

func (p *PredictorProcess) routeChildren(node *v1alpha2.PredictiveUnit, msg client.SeldonPayload) (client.SeldonPayload, error) {
	if node.Children != nil && len(node.Children) > 0 {
		route, err := p.route(node, msg)
		if err != nil {
			return nil, err
		}
		var cmsgs []client.SeldonPayload
		if route == -1 {
			cmsgs = make([]client.SeldonPayload,len(node.Children))
			var errs =  make([]error,len(node.Children))
			wg := sync.WaitGroup{}
			for i, nodeChild := range(node.Children) {
				wg.Add(1)
				go func(i int, nodeChild v1alpha2.PredictiveUnit, msg client.SeldonPayload) {
					cmsgs[i], errs[i] = p.Execute(&nodeChild, msg)
					wg.Done()
				} (i, nodeChild, msg)
			}
			wg.Wait()
			for i, err := range(errs) {
				if err != nil {
					return cmsgs[i], err
				}
			}
		} else {
			cmsgs = make([]client.SeldonPayload,1)
			cmsgs[0], err = p.Execute(&node.Children[route], msg)
			if err != nil {
				return cmsgs[0], err
			}
		}
		return  p.aggregate(node, cmsgs)
	} else {
		return msg, nil
	}
}

func (p *PredictorProcess) Execute(node *v1alpha2.PredictiveUnit, msg client.SeldonPayload) (client.SeldonPayload, error) {
	tmsg, err := p.transformInput(node,msg)
	if err != nil {
		return tmsg, err
	}
	cmsg, err := p.routeChildren(node, tmsg)
	if err != nil {
		return tmsg, err
	}
	return p.transformOutput(node, cmsg)
}
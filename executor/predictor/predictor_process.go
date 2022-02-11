package predictor

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"sync"

	"github.com/go-logr/logr"
	guuid "github.com/google/uuid"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/util"

	payloadLogger "github.com/seldonio/seldon-core/executor/logger"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
)

const (
	NilPUIDError                        = "context value for Seldon PUID Header is nil"
	ENV_REQUEST_LOGGER_DEFAULT_ENDPOINT = "REQUEST_LOGGER_DEFAULT_ENDPOINT"
	ENV_ENABLE_ROUTING_INJECTION        = "SELDON_ENABLE_ROUTING_INJECTION"
)

var (
	envRequestLoggerDefaultEndpoint = os.Getenv(ENV_REQUEST_LOGGER_DEFAULT_ENDPOINT)
	envEnableRoutingInjection       = len(os.Getenv(ENV_ENABLE_ROUTING_INJECTION)) != 0
)

type PredictorProcess struct {
	Ctx               context.Context
	Client            client.SeldonApiClient
	Log               logr.Logger
	ServerUrl         *url.URL
	Namespace         string
	Meta              *payload.MetaData
	Routing           map[string]int32
	RoutingMutex      *sync.RWMutex
	ModelNameOverride string
}

func NewPredictorProcess(context context.Context, client client.SeldonApiClient, log logr.Logger, serverUrl *url.URL, namespace string, meta map[string][]string, modelNameOverride string) PredictorProcess {
	return PredictorProcess{
		Ctx:               context,
		Client:            client,
		Log:               log,
		ServerUrl:         serverUrl,
		Namespace:         namespace,
		Meta:              payload.NewFromMap(meta),
		Routing:           make(map[string]int32),
		RoutingMutex:      &sync.RWMutex{},
		ModelNameOverride: modelNameOverride,
	}
}

func hasMethod(method v1.PredictiveUnitMethod, methods *[]v1.PredictiveUnitMethod) bool {
	if methods != nil {
		for _, m := range *methods {
			if m == method {
				return true
			}
		}
	}
	return false
}

func (p *PredictorProcess) getPort(node *v1.PredictiveUnit) int32 {
	if p.Client.IsGrpc() {
		return node.Endpoint.GrpcPort
	} else {
		return node.Endpoint.HttpPort
	}
}

func (p *PredictorProcess) getModelName(node *v1.PredictiveUnit) string {
	modelName := node.Name
	if p.ModelNameOverride != "" {
		modelName = p.ModelNameOverride
	}
	return modelName
}

func (p *PredictorProcess) transformInput(node *v1.PredictiveUnit, msg payload.SeldonPayload, puid string) (tmsg payload.SeldonPayload, err error) {
	callModel := false
	callTransformInput := false
	if (*node).Type != nil {
		switch *node.Type {
		case v1.MODEL:
			callModel = true
		case v1.TRANSFORMER:
			callTransformInput = true
		}
	}
	if hasMethod(v1.TRANSFORM_INPUT, node.Methods) {
		callTransformInput = true
	}

	modelName := p.getModelName(node)

	if callModel || callTransformInput {
		//Log Request
		if node.Logger != nil && (node.Logger.Mode == v1.LogRequest || node.Logger.Mode == v1.LogAll) {
			err := p.logPayload(node.Name, node.Logger, payloadLogger.InferenceRequest, msg, puid)
			if err != nil {
				return nil, err
			}
		}

		msg, err := p.Client.Chain(p.Ctx, modelName, msg)
		if err != nil {
			return nil, err
		}
		p.RoutingMutex.Lock()
		p.Routing[node.Name] = -1
		p.RoutingMutex.Unlock()

		if callTransformInput {
			tmsg, err = p.Client.TransformInput(p.Ctx, modelName, node.Endpoint.ServiceHost, p.getPort(node), msg, p.Meta.Meta)
		} else {
			tmsg, err = p.Client.Predict(p.Ctx, modelName, node.Endpoint.ServiceHost, p.getPort(node), msg, p.Meta.Meta)
		}
		if tmsg != nil && err == nil {
			// Log Response
			if node.Logger != nil && (node.Logger.Mode == v1.LogResponse || node.Logger.Mode == v1.LogAll) {
				err := p.logPayload(node.Name, node.Logger, payloadLogger.InferenceResponse, tmsg, puid)
				if err != nil {
					return nil, err
				}
			}
		}
		return tmsg, err
	} else {
		return msg, nil
	}
}

func (p *PredictorProcess) transformOutput(node *v1.PredictiveUnit, msg payload.SeldonPayload, puid string) (payload.SeldonPayload, error) {
	callClient := false
	if (*node).Type != nil {
		switch *node.Type {
		case v1.OUTPUT_TRANSFORMER:
			callClient = true
		}
	}
	if hasMethod(v1.TRANSFORM_OUTPUT, node.Methods) {
		callClient = true
	}

	modelName := p.getModelName(node)

	if callClient {
		//Log Request
		if node.Logger != nil && (node.Logger.Mode == v1.LogRequest || node.Logger.Mode == v1.LogAll) {
			err := p.logPayload(node.Name, node.Logger, payloadLogger.InferenceRequest, msg, puid)
			if err != nil {
				return nil, err
			}
		}

		msg, err := p.Client.Chain(p.Ctx, modelName, msg)
		if err != nil {
			return nil, err
		}
		tmsg, err := p.Client.TransformOutput(p.Ctx, modelName, node.Endpoint.ServiceHost, p.getPort(node), msg, p.Meta.Meta)
		if tmsg != nil && err == nil {
			// Log Response
			if node.Logger != nil && (node.Logger.Mode == v1.LogResponse || node.Logger.Mode == v1.LogAll) {
				err := p.logPayload(node.Name, node.Logger, payloadLogger.InferenceResponse, tmsg, puid)
				if err != nil {
					return nil, err
				}
			}
		}
		return tmsg, err
	} else {
		return msg, nil
	}

}

func (p *PredictorProcess) feedback(node *v1.PredictiveUnit, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	callClient := false
	if (*node).Type != nil {
		switch *node.Type {
		case v1.MODEL, v1.ROUTER:
			callClient = true
		}
	}
	if hasMethod(v1.SEND_FEEDBACK, node.Methods) {
		callClient = true
	}

	modelName := p.getModelName(node)

	if callClient {
		return p.Client.Feedback(p.Ctx, modelName, node.Endpoint.ServiceHost, p.getPort(node), msg, p.Meta.Meta)
	} else {
		return msg, nil
	}

}

func (p *PredictorProcess) routeFeedback(node *v1.PredictiveUnit, msg payload.SeldonPayload) (int, error) {
	if msg.GetContentType() == payload.APPLICATION_TYPE_PROTOBUF {
		return util.RouteFromFeedbackMessageMeta(msg.GetPayload().(*proto.Feedback), node.Name), nil
	} else {
		return util.RouteFromFeedbackJsonMeta(msg, node.Name), nil
	}
}

func (p *PredictorProcess) route(node *v1.PredictiveUnit, msg payload.SeldonPayload) (int, error) {
	callClient := false
	if (*node).Type != nil {
		switch *node.Type {
		case v1.ROUTER:
			callClient = true
		}
	}
	if hasMethod(v1.ROUTE, node.Methods) {
		callClient = true
	}

	modelName := p.getModelName(node)

	if callClient {
		return p.Client.Route(p.Ctx, modelName, node.Endpoint.ServiceHost, p.getPort(node), msg, p.Meta.Meta)
	} else if node.Implementation != nil && *node.Implementation == v1.RANDOM_ABTEST {
		return p.abTestRouter(node)
	} else {
		return -1, nil
	}
}

func (p *PredictorProcess) aggregate(node *v1.PredictiveUnit, cmsg []payload.SeldonPayload, msg payload.SeldonPayload, puid string) (payload.SeldonPayload, error) {
	callClient := false
	if (*node).Type != nil {
		switch *node.Type {
		case v1.COMBINER:
			callClient = true
		}
	}
	if hasMethod(v1.AGGREGATE, node.Methods) {
		callClient = true
	}

	modelName := p.getModelName(node)

	if callClient {
		//Log Request
		if node.Logger != nil && (node.Logger.Mode == v1.LogRequest || node.Logger.Mode == v1.LogAll) {
			err := p.logPayload(node.Name, node.Logger, payloadLogger.InferenceRequest, msg, puid)
			if err != nil {
				return nil, err
			}
		}
		p.RoutingMutex.Lock()
		p.Routing[node.Name] = -1
		p.RoutingMutex.Unlock()
		tmsg, err := p.Client.Combine(p.Ctx, modelName, node.Endpoint.ServiceHost, p.getPort(node), cmsg, p.Meta.Meta)
		if tmsg != nil && err == nil {
			// Log Response
			if node.Logger != nil && (node.Logger.Mode == v1.LogResponse || node.Logger.Mode == v1.LogAll) {
				err := p.logPayload(node.Name, node.Logger, payloadLogger.InferenceResponse, tmsg, puid)
				if err != nil {
					return nil, err
				}
			}
		}
		return tmsg, err
	} else {
		return cmsg[0], nil
	}
}

func (p *PredictorProcess) predictChildren(node *v1.PredictiveUnit, msg payload.SeldonPayload, puid string) (payload.SeldonPayload, error) {
	if node.Children != nil && len(node.Children) > 0 {
		route, err := p.route(node, msg)
		if err != nil {
			return nil, err
		}
		var cmsgs []payload.SeldonPayload
		if route == -1 {

			cmsgs = make([]payload.SeldonPayload, len(node.Children))
			var errs = make([]error, len(node.Children))
			wg := sync.WaitGroup{}
			for i, nodeChild := range node.Children {
				wg.Add(1)
				go func(i int, nodeChild v1.PredictiveUnit, msg payload.SeldonPayload) {
					cmsgs[i], errs[i] = p.Predict(&nodeChild, msg)
					wg.Done()
				}(i, nodeChild, msg)
			}
			wg.Wait()
			p.RoutingMutex.Lock()
			p.Routing[node.Name] = -1
			p.RoutingMutex.Unlock()
			for i, err := range errs {
				if err != nil {
					return cmsgs[i], err
				}
			}
		} else if route == -2 {
			//Abort and return request
			p.RoutingMutex.Lock()
			p.Routing[node.Name] = -2
			p.RoutingMutex.Unlock()
			return msg, nil
		} else {
			cmsgs = make([]payload.SeldonPayload, 1)
			cmsgs[0], err = p.Predict(&node.Children[route], msg)
			p.RoutingMutex.Lock()
			p.Routing[node.Name] = int32(route)
			p.RoutingMutex.Unlock()
			if err != nil {
				return cmsgs[0], err
			}
		}
		return p.aggregate(node, cmsgs, msg, puid)
	} else {
		// Don't add routing for leaf nodes
		return msg, nil
	}
}

func (p *PredictorProcess) feedbackChildren(node *v1.PredictiveUnit, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	if node.Children != nil && len(node.Children) > 0 {

		route, err := p.routeFeedback(node, msg)
		if err != nil {
			return nil, err
		}
		var cmsgs []payload.SeldonPayload
		if route == -1 {
			cmsgs = make([]payload.SeldonPayload, len(node.Children))
			var errs = make([]error, len(node.Children))
			wg := sync.WaitGroup{}
			for i, nodeChild := range node.Children {
				wg.Add(1)
				go func(i int, nodeChild v1.PredictiveUnit, msg payload.SeldonPayload) {
					cmsgs[i], errs[i] = p.Feedback(&nodeChild, msg)
					wg.Done()
				}(i, nodeChild, msg)
			}
			wg.Wait()
			for i, err := range errs {
				if err != nil {
					return cmsgs[i], err
				}
			}
		} else {
			cmsgs = make([]payload.SeldonPayload, 1)
			cmsgs[0], err = p.Feedback(&node.Children[route], msg)
			if err != nil {
				return cmsgs[0], err
			}
		}
		// Arbitrary return of first feedback
		return cmsgs[0], nil
	} else {
		return msg, nil
	}
}

func (p *PredictorProcess) getLogUrl(logger *v1.Logger) (*url.URL, error) {
	if logger.Url != nil {
		return url.Parse(*logger.Url)
	} else {
		return url.Parse(envRequestLoggerDefaultEndpoint)
	}
}

func (p *PredictorProcess) logPayload(nodeName string, logger *v1.Logger, reqType payloadLogger.LogRequestType, msg payload.SeldonPayload, puid string) error {
	skipLogging := p.Meta.GetAsBoolean(payload.SeldonSkipLoggingHeader, false)
	if skipLogging {
		p.Log.Info("Skipped logging request with", "PUID", puid)
		return nil
	}

	data, err := msg.GetBytes()
	if err != nil {
		return err
	}
	logUrl, err := p.getLogUrl(logger)
	if err != nil {
		return err
	}
	go func() {
		err := payloadLogger.QueueLogRequest(payloadLogger.LogRequest{
			Url:             logUrl,
			Bytes:           &data,
			ContentType:     msg.GetContentType(),
			ContentEncoding: msg.GetContentEncoding(),
			ReqType:         reqType,
			Id:              guuid.New().String(),
			SourceUri:       p.ServerUrl,
			ModelId:         nodeName,
			RequestId:       puid,
		})
		if err != nil {
			p.Log.Error(err, "failed to log request")
		}
	}()
	return nil
}

func (p *PredictorProcess) getPUIDHeader() (string, error) {
	// Check request ID is not nil
	if puid, ok := p.Ctx.Value(payload.SeldonPUIDHeader).(string); ok {
		return puid, nil
	}
	return "", fmt.Errorf(NilPUIDError)
}

func (p *PredictorProcess) Predict(node *v1.PredictiveUnit, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	puid, err := p.getPUIDHeader()
	if err != nil {
		return nil, err
	}

	tmsg, err := p.transformInput(node, msg, puid)
	if err != nil {
		return tmsg, err
	}
	cmsg, err := p.predictChildren(node, tmsg, puid)
	if err != nil {
		return cmsg, err
	}

	response, err := p.transformOutput(node, cmsg, puid)

	if envEnableRoutingInjection {
		if routeResponse, err := util.InsertRouteToSeldonPredictPayload(response, &p.Routing); err == nil {
			return routeResponse, err
		}
	}
	return response, err
}

func (p *PredictorProcess) Status(node *v1.PredictiveUnit, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	if nodeModel := v1.GetPredictiveUnit(node, modelName); nodeModel == nil {
		return nil, fmt.Errorf("Failed to find model %s", modelName)
	} else {
		return p.Client.Status(p.Ctx, modelName, nodeModel.Endpoint.ServiceHost, p.getPort(node), msg, p.Meta.Meta)
	}
}

func (p *PredictorProcess) Metadata(node *v1.PredictiveUnit, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	if nodeModel := v1.GetPredictiveUnit(node, modelName); nodeModel == nil {
		return nil, fmt.Errorf("Failed to find model %s", modelName)
	} else {
		return p.Client.Metadata(p.Ctx, modelName, nodeModel.Endpoint.ServiceHost, p.getPort(node), msg, p.Meta.Meta)
	}
}

func (p *PredictorProcess) GraphMetadata(spec *v1.PredictorSpec) (*GraphMetadata, error) {
	metadataMap, err := p.ModelMetadataMap(&spec.Graph)
	if err != nil {
		return nil, err
	}

	output := &GraphMetadata{
		Name:   spec.Name,
		Models: metadataMap,
	}

	inputNodeMeta, outputNodeMeta := output.getEdgeNodes(&spec.Graph)
	output.GraphInputs = inputNodeMeta.Inputs
	output.GraphOutputs = outputNodeMeta.Outputs

	return output, nil
}

func (p *PredictorProcess) Feedback(node *v1.PredictiveUnit, msg payload.SeldonPayload) (payload.SeldonPayload, error) {

	if node.Logger != nil && (node.Logger.Mode == v1.LogResponse || node.Logger.Mode == v1.LogAll) {
		puid, puiderr := p.getPUIDHeader()
		if puiderr != nil {
			p.Log.Error(puiderr, "Error retrieving uuid for feedback and could not send feedback")
		} else {
			err := p.logPayload(node.Name, node.Logger, payloadLogger.InferenceFeedback, msg, puid)
			if err != nil {
				return nil, err
			}
		}
	}

	tmsg, err := p.feedbackChildren(node, msg)
	if err != nil {
		return tmsg, err
	}
	return p.feedback(node, msg)
}

func (p *PredictorProcess) ModelMetadataMap(node *v1.PredictiveUnit) (map[string]payload.ModelMetadata, error) {
	resPayload, err := p.Client.ModelMetadata(p.Ctx, node.Name, node.Endpoint.ServiceHost, p.getPort(node), nil, p.Meta.Meta)
	if err != nil {
		return nil, err
	}

	var output = map[string]payload.ModelMetadata{
		node.Name: resPayload,
	}
	for _, child := range node.Children {
		childMeta, err := p.ModelMetadataMap(&child)
		if err != nil {
			return nil, err
		}
		for k, v := range childMeta {
			output[k] = v
		}
	}
	return output, nil
}

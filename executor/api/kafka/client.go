package kafka

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/rest"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"io"
)

type KafkaClient struct {
	Hostname       string
	DeploymentName string
	Namespace      string
	Protocol       string
	Transport      string
	predictor      *v1.PredictorSpec
	Broker         string
	Log            logr.Logger
	topicHandlers  map[string]*KafkaRPC
}

func (kc *KafkaClient) IsGrpc() bool {
	return false
}

func NewKafkaClient(hostname, deploymentName, namespace, protocol, transport string, predictor *v1.PredictorSpec, broker string, log logr.Logger) client.SeldonApiClient {
	skc := &KafkaClient{
		Hostname:       hostname,
		DeploymentName: deploymentName,
		Namespace:      namespace,
		Protocol:       protocol,
		Transport:      transport,
		predictor:      predictor,
		Broker:         broker,
		Log:            log.WithName("KafkaClient"),
		topicHandlers:  make(map[string]*KafkaRPC),
	}
	skc.createTopicHandlers(&predictor.Graph)
	return skc
}

func (kc *KafkaClient) createTopicHandlers(node *v1.PredictiveUnit) error {
	th, err := NewKafkaRPC(kc, node.Name)
	if err != nil {
		return err
	}
	th.start()
	kc.topicHandlers[node.Name] = th
	for _, child := range node.Children {
		err = kc.createTopicHandlers(&child)
		if err != nil {
			return err
		}
	}
	return nil
}

func getPuidFromMeta(meta map[string][]string) (string, error) {
	if arr, ok := meta[payload.SeldonPUIDHeader]; ok {
		if len(arr) == 1 {
			return arr[0], nil
		} else {
			return "", fmt.Errorf("Can't findPUID: Invalid number of items in puid meta key: %d", len(arr))
		}
	} else {
		return "", fmt.Errorf("Failed to find header key in meta %s", payload.SeldonPUIDHeader)
	}
}

func (kc *KafkaClient) kafkaRPC(msg payload.SeldonPayload, meta map[string][]string, modelName string, method string) (payload.SeldonPayload, error) {
	bytes, err := msg.GetBytes()
	if err != nil {
		kc.Log.Error(err, "Failed to get bytes from request")
		return nil, err
	}
	puid, err := getPuidFromMeta(meta)
	if err != nil {
		return nil, err
	}
	if kafkaRPC, ok := kc.topicHandlers[modelName]; ok {
		return kafkaRPC.call(bytes, puid, method)
	} else {
		return nil, fmt.Errorf("Failed to find topic handler for model name %s", modelName)
	}
}

func (kc *KafkaClient) Predict(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return kc.kafkaRPC(msg, meta, modelName, client.SeldonPredictPath)
}

func (kc *KafkaClient) TransformInput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return kc.kafkaRPC(msg, meta, modelName, client.SeldonTransformInputPath)
}

func (kc *KafkaClient) Route(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (int, error) {
	res, err := kc.kafkaRPC(msg, meta, modelName, client.SeldonRoutePath)
	if err != nil {
		return 0, err
	} else {
		return rest.ExtractRouteFromJson(res)
	}
}

func (kc *KafkaClient) Combine(ctx context.Context, modelName string, host string, port int32, msgs []payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	req, err := rest.CombineSeldonMessagesToJson(msgs)
	if err != nil {
		return nil, err
	}
	return kc.kafkaRPC(req, meta, modelName, client.SeldonCombinePath)
}

func (kc *KafkaClient) TransformOutput(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return kc.kafkaRPC(msg, meta, modelName, client.SeldonTransformOutputPath)
}

func (kc *KafkaClient) Feedback(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	return kc.kafkaRPC(msg, meta, modelName, client.SeldonFeedbackPath)
}

func (kc *KafkaClient) Chain(ctx context.Context, modelName string, msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	switch kc.Protocol {
	case api.ProtocolSeldon: // Seldon Messages can always be chained together
		return msg, nil
	case api.ProtocolTensorflow: // Attempt to chain tensorflow Payload
		return rest.ChainTensorflow(msg)
	}
	return nil, errors.Errorf("Unknown protocol %s", kc.Protocol)
}

func (kc *KafkaClient) Status(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	panic("Not implemented")
}

func (kc *KafkaClient) Metadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.SeldonPayload, error) {
	panic("Not implemented")
}

func (kc *KafkaClient) CreateErrorPayload(err error) payload.SeldonPayload {
	panic("Not implemented")
}

func (kc *KafkaClient) Marshall(w io.Writer, msg payload.SeldonPayload) error {
	_, err := w.Write(msg.GetPayload().([]byte))
	return err
}

func (kc *KafkaClient) Unmarshall(msg []byte, contentType string) (payload.SeldonPayload, error) {
	reqPayload := payload.BytesPayload{Msg: msg, ContentType: contentType}
	return &reqPayload, nil
}

func (kc *KafkaClient) ModelMetadata(ctx context.Context, modelName string, host string, port int32, msg payload.SeldonPayload, meta map[string][]string) (payload.ModelMetadata, error) {
	panic("implement me")
}

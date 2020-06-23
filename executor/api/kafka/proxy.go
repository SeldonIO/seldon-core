package kafka

import (
	"context"
	"github.com/cloudevents/sdk-go/pkg/bindings/http"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"os"
	"os/signal"
	"syscall"
)

type KafkaProxy struct {
	Client         client.SeldonApiClient
	ModelName      string
	PredictorName  string
	DeploymentName string
	Namespace      string
	Broker         string
	Hostname       string
	Port           int32
	Log            logr.Logger
}

func NewKafkaProxy(client client.SeldonApiClient, modelName, predictorName, deploymentName, namespace, broker, hostname string, port int32, log logr.Logger) *KafkaProxy {
	return &KafkaProxy{
		Client:         client,
		ModelName:      modelName,
		PredictorName:  predictorName,
		DeploymentName: deploymentName,
		Namespace:      namespace,
		Broker:         broker,
		Hostname:       hostname,
		Port:           port,
		Log:            log,
	}
}

func (kp *KafkaProxy) getGroupName() string {
	return kp.PredictorName + "." + kp.DeploymentName + "." + kp.Namespace
}

func (kp *KafkaProxy) getTopicIn() string {
	return kp.ModelName + "." + kp.PredictorName + "." + kp.DeploymentName + "." + kp.Namespace
}

func (kp *KafkaProxy) getDefaultTopicResponse() string {
	return kp.Hostname + "." + kp.ModelName + "." + kp.PredictorName + "." + kp.DeploymentName + "." + kp.Namespace
}

func (kp *KafkaProxy) Consume() error {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":     kp.Broker,
		"broker.address.family": "v4",
		"group.id":              kp.getGroupName(),
		"session.timeout.ms":    6000,
		"auto.offset.reset":     "earliest"})
	if err != nil {
		return err
	}
	kp.Log.Info("Created", "consumer", c.String())

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": kp.Broker})
	if err != nil {
		return err
	}
	kp.Log.Info("Created", "producer", p.String())

	err = c.SubscribeTopics([]string{kp.getTopicIn()}, nil)
	if err != nil {
		return err
	}
	kp.Log.Info("Subscribed", "topic", kp.getTopicIn())

	run := true
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for run == true {
		select {
		case sig := <-sigchan:
			kp.Log.Info("Terminating", "signal", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				kp.Log.Info("Message", "Partition", e.TopicPartition)
				if e.Headers != nil {
					kp.Log.Info("Received", "headers", e.Headers)
				}

				puid := ""
				responseTopic := ""
				method := ""
				for _, header := range e.Headers {
					switch header.Key {
					case payload.SeldonPUIDHeader:
						puid = string(header.Value)
					case KeyTopicResponse:
						responseTopic = string(header.Value)
					case KeyMethod:
						method = string(header.Value)
					default:
						kp.Log.Info("Skipping", "header", string(header.Value))
					}
				}
				kp.Log.Info("Extracted headers", payload.SeldonPUIDHeader, puid, KeyTopicResponse, responseTopic, KeyMethod, method)
				if responseTopic == "" {
					responseTopic = kp.getDefaultTopicResponse()
				}
				if puid == "" {
					kp.Log.Info("No puid found")
					puid = "0"
				}
				if method == "" {
					kp.Log.Info("No method found will use default")
					method = client.SeldonPredictPath
				}
				kp.Log.Info("Extracted headers with defaults", payload.SeldonPUIDHeader, puid, KeyTopicResponse, responseTopic, KeyMethod, method)

				headers := collectHeaders(e.Headers)
				ctx := context.Background()
				// Add Seldon Puid to Context
				ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, puid)

				// Assume JSON if no content type - should maybe be application/octet-stream?
				contentType := rest.ContentTypeJSON
				if ct, ok := headers[http.ContentType]; ok {
					if len(ct) == 1 {
						contentType = ct[0]
					}
				}
				reqPayload, err := kp.Client.Unmarshall(e.Value, contentType)
				if err != nil {
					kp.Log.Error(err, "Failed to unmarshall Payload")
					continue
				}

				var resPayload payload.SeldonPayload

				switch method {
				case client.SeldonPredictPath:
					resPayload, err = kp.Client.Predict(ctx, kp.ModelName, kp.Hostname, kp.Port, reqPayload, headers)
				case client.SeldonCombinePath:
					msgs, err := rest.ExtractSeldonMessagesFromJson(reqPayload)
					if err != nil {
						kp.Log.Error(err, "Failed to extract Payload")
						continue
					}
					resPayload, err = kp.Client.Combine(ctx, kp.ModelName, kp.Hostname, kp.Port, msgs, headers)
				}

				if err != nil {
					kp.Log.Error(err, "Failed prediction")
					continue
				}
				resBytes, err := resPayload.GetBytes()
				if err != nil {
					kp.Log.Error(err, "Failed to get bytes from prediction response")
				}

				err = p.Produce(&kafka.Message{
					TopicPartition: kafka.TopicPartition{Topic: &responseTopic, Partition: kafka.PartitionAny},
					Value:          resBytes,
					Headers:        []kafka.Header{{Key: payload.SeldonPUIDHeader, Value: []byte(puid)}},
				}, nil)
				if err != nil {
					kp.Log.Error(err, "Failed to produce response")
				}
				kp.Log.Info("Produced message", "topic", responseTopic)

			case kafka.Error:
				// Errors should generally be considered
				// informational, the client will try to
				// automatically recover.
				// But in this example we choose to terminate
				// the application if all brokers are down.
				kp.Log.Error(e, "Received kafka error")
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				kp.Log.Info("Ignored", "msg", e)
			}
		}
	}

	kp.Log.Info("Closing consumer")
	c.Close()
	return nil
}

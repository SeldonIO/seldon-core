package kafka

import (
	"context"
	"fmt"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-logr/logr"
	proto2 "github.com/golang/protobuf/proto"
	guuid "github.com/google/uuid"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon"
	"github.com/seldonio/seldon-core/executor/api/grpc/tensorflow"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"github.com/seldonio/seldon-core/executor/predictor"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"syscall"
)

const (
	kafkaPayloadJson  = "json"
	kafkaPayloadProto = "proto"
)

type SeldonKafkaServer struct {
	Client         client.SeldonApiClient
	DeploymentName string
	Namespace      string
	Transport      string
	Predictor      *v1.PredictorSpec
	Broker         string
	TopicIn        string
	TopicOut       string
	ServerUrl      *url.URL
	Log            logr.Logger
}

func NewKafkaServer(graphInternal bool, deploymentName, namespace, protocol, transport string, annotations map[string]string, serverUrl *url.URL, predictor *v1.PredictorSpec, broker, topicIn, topicOut string, log logr.Logger) (*SeldonKafkaServer, error) {
	var apiClient client.SeldonApiClient
	var err error
	if graphInternal {
		apiClient = NewKafkaClient(serverUrl.Hostname(), deploymentName, namespace, protocol, transport, predictor, broker, log)
	} else {
		switch transport {
		case api.TransportRest:
			apiClient, err = rest.NewJSONRestClient(protocol, deploymentName, predictor, annotations)
			if err != nil {
				return nil, err
			}
		case api.TransportGrpc:
			if protocol == "seldon" {
				apiClient = seldon.NewSeldonGrpcClient(predictor, deploymentName, annotations)
			} else {
				apiClient = tensorflow.NewTensorflowGrpcClient(predictor, deploymentName, annotations)
			}
		default:
			return nil, fmt.Errorf("Unknown transport %s", transport)
		}
	}
	return &SeldonKafkaServer{
		Client:         apiClient,
		DeploymentName: deploymentName,
		Namespace:      namespace,
		Transport:      transport,
		Predictor:      predictor,
		Broker:         broker,
		TopicIn:        topicIn,
		TopicOut:       topicOut,
		ServerUrl:      serverUrl,
		Log:            log.WithName("KafkaServer"),
	}, nil
}

func (ks *SeldonKafkaServer) getGroupName() string {
	return ks.Predictor.Name + "." + ks.DeploymentName + "." + ks.Namespace
}

func collectHeaders(headers []kafka.Header) map[string][]string {
	sheaders := make(map[string][]string)
	foundPuid := false
	if headers != nil {
		for _, header := range headers {
			if header.Key == payload.SeldonPUIDHeader {
				foundPuid = true
			}
			if _, ok := sheaders[header.Key]; ok {
				sheaders[header.Key] = append(sheaders[header.Key], string(header.Value))
			} else {
				sheaders[header.Key] = []string{string(header.Value)}
			}
		}
	}
	// PUID if not found
	if !foundPuid {
		sheaders[payload.SeldonPUIDHeader] = []string{guuid.New().String()}
	}
	return sheaders
}

func getProto(messageType string, messageBytes []byte) (proto2.Message, error) {
	pbtype := proto2.MessageType(messageType)
	msg := reflect.New(pbtype.Elem()).Interface().(proto2.Message)
	err := proto2.Unmarshal(messageBytes, msg)
	return msg, err
}

func (ks *SeldonKafkaServer) Serve() error {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers":     ks.Broker,
		"broker.address.family": "v4",
		"group.id":              ks.getGroupName(),
		"session.timeout.ms":    6000,
		"auto.offset.reset":     "earliest"})
	if err != nil {
		return err
	}
	ks.Log.Info("Created", "consumer", c.String())

	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": ks.Broker})
	if err != nil {
		return err
	}
	ks.Log.Info("Created", "producer", p.String())

	err = c.SubscribeTopics([]string{ks.TopicIn}, nil)
	if err != nil {
		return err
	}

	run := true
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	for run == true {
		select {
		case sig := <-sigchan:
			ks.Log.Info("Terminating", "signal", sig)
			run = false
		default:
			ev := c.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				ks.Log.Info("Message", "Partition", e.TopicPartition)
				if e.Headers != nil {
					ks.Log.Info("Received", "headers", e.Headers)
				}
				headers := collectHeaders(e.Headers)

				var reqPayload payload.SeldonPayload
				var err error
				switch ks.Transport {
				case api.TransportRest:
					reqPayload, err = ks.Client.Unmarshall(e.Value)
					if err != nil {
						ks.Log.Error(err, "Failed to unmarshall Payload")
						continue
					}
				case api.TransportGrpc:
					if val, ok := headers[KeyProtoName]; ok && len(val) == 1 {
						protoName := val[0]
						proto, err := getProto(protoName, e.Value)
						if err != nil {
							ks.Log.Error(err, "Failed to get proto from bytes")
							continue
						}
						reqPayload = &payload.ProtoPayload{Msg: proto}

					} else {
						ks.Log.Info("Failed to find proto name in headers")
						continue
					}

				}

				go func() {
					ctx := context.Background()
					// Add Seldon Puid to Context
					ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, headers[payload.SeldonPUIDHeader][0])

					seldonPredictorProcess := predictor.NewPredictorProcess(ctx, ks.Client, logf.Log.WithName("KafkaClient"), ks.ServerUrl, ks.Namespace, headers)

					resPayload, err := seldonPredictorProcess.Predict(ks.Predictor.Graph, reqPayload)
					if err != nil {
						ks.Log.Error(err, "Failed prediction")
						return
					}
					resBytes, err := resPayload.GetBytes()
					if err != nil {
						ks.Log.Error(err, "Failed to get bytes from prediction response")
						return
					}

					kafkaHeaders := make([]kafka.Header, 0)
					if ks.Transport == api.TransportGrpc {
						kafkaHeaders = []kafka.Header{{Key: KeyProtoName, Value: []byte(proto2.MessageName(resPayload.GetPayload().(*payload.ProtoPayload).Msg))}}
					}

					err = p.Produce(&kafka.Message{
						TopicPartition: kafka.TopicPartition{Topic: &ks.TopicOut, Partition: kafka.PartitionAny},
						Value:          resBytes,
						Headers:        kafkaHeaders,
					}, nil)
					if err != nil {
						ks.Log.Error(err, "Failed to produce response")
					}
				}()

			case kafka.Error:
				// Errors should generally be considered
				// informational, the client will try to
				// automatically recover.
				// But in this example we choose to terminate
				// the application if all brokers are down.
				ks.Log.Error(e, "Received kafka error")
				if e.Code() == kafka.ErrAllBrokersDown {
					run = false
				}
			default:
				ks.Log.Info("Ignored", "msg", e)
			}
		}
	}

	ks.Log.Info("Closing consumer")
	c.Close()
	return nil
}

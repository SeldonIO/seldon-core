package kafka

import (
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"time"

	"github.com/cloudevents/sdk-go/pkg/bindings/http"
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
	"github.com/seldonio/seldon-core/executor/api/util"
	"github.com/seldonio/seldon-core/executor/predictor"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
)

const (
	kafkaPayloadJson  = "json"
	kafkaPayloadProto = "proto"
)

const (
	ENV_KAFKA_BROKER       = "KAFKA_BROKER"
	ENV_KAFKA_INPUT_TOPIC  = "KAFKA_INPUT_TOPIC"
	ENV_KAFKA_OUTPUT_TOPIC = "KAFKA_OUTPUT_TOPIC"
	ENV_KAFKA_FULL_GRAPH   = "KAFKA_FULL_GRAPH"
	ENV_KAFKA_WORKERS      = "KAFKA_WORKERS"
)

type SeldonKafkaServer struct {
	Client         client.SeldonApiClient
	Producer       *kafka.Producer
	DeploymentName string
	Namespace      string
	Transport      string
	Predictor      *v1.PredictorSpec
	Broker         string
	TopicIn        string
	TopicOut       string
	ServerUrl      *url.URL
	Workers        int
	Log            logr.Logger
}

func NewKafkaServer(fullGraph bool, workers int, deploymentName, namespace, protocol, transport string, annotations map[string]string, serverUrl *url.URL, predictor *v1.PredictorSpec, broker, topicIn, topicOut string, log logr.Logger) (*SeldonKafkaServer, error) {
	var apiClient client.SeldonApiClient
	var err error
	if fullGraph {
		log.Info("Starting full graph kafka server")
		apiClient = NewKafkaClient(serverUrl.Hostname(), deploymentName, namespace, protocol, transport, predictor, broker, log)
	} else {
		switch transport {
		case api.TransportRest:
			log.Info("Start http kafka graph")
			apiClient, err = rest.NewJSONRestClient(protocol, deploymentName, predictor, annotations)
			if err != nil {
				return nil, err
			}
		case api.TransportGrpc:
			log.Info("Start grpc kafka graph")
			if protocol == "seldon" {
				apiClient = seldon.NewSeldonGrpcClient(predictor, deploymentName, annotations)
			} else {
				apiClient = tensorflow.NewTensorflowGrpcClient(predictor, deploymentName, annotations)
			}
		default:
			return nil, fmt.Errorf("Unknown transport %s", transport)
		}
	}
	sslKakfaServer := util.GetSslElements()
	var producerConfigMap = kafka.ConfigMap{"bootstrap.servers": broker,
		"go.delivery.reports": false, // Need this othewise will get memory leak
	}
	if broker != "" {
		if util.GetKafkaSecurityProtocol() == "SSL" {
			producerConfigMap["security.protocol"] = util.GetKafkaSecurityProtocol()
			producerConfigMap["ssl.ca.location"] = sslKakfaServer.CACertFile
			producerConfigMap["ssl.key.location"] = sslKakfaServer.ClientKeyFile
			producerConfigMap["ssl.certificate.location"] = sslKakfaServer.ClientCertFile
			producerConfigMap["ssl.key.password"] = sslKakfaServer.ClientKeyPass // Key password, if any

		}
	}
	// Create Producer
	log.Info("Creating producer", "broker", broker)
	p, err := kafka.NewProducer(&producerConfigMap)
	if err != nil {
		return nil, err
	}
	log.Info("Created", "producer", p.String())

	return &SeldonKafkaServer{
		Client:         apiClient,
		Producer:       p,
		DeploymentName: deploymentName,
		Namespace:      namespace,
		Transport:      transport,
		Predictor:      predictor,
		Broker:         broker,
		TopicIn:        topicIn,
		TopicOut:       topicOut,
		ServerUrl:      serverUrl,
		Workers:        workers,
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

	err = c.SubscribeTopics([]string{ks.TopicIn}, nil)
	if err != nil {
		return err
	}

	run := true
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// create a cancel channel
	cancelChan := make(chan struct{})
	// make a channel with a capacity of the number of workers
	jobChan := make(chan *KafkaJob, ks.Workers)
	for i := 0; i < ks.Workers; i++ {
		go ks.worker(jobChan, cancelChan)
	}

	//wait for graph to be ready
	ready := false
	for ready == false {
		err := predictor.Ready(&ks.Predictor.Graph)
		ready = err == nil
		if !ready {
			ks.Log.Info("Waiting for graph to be ready")
			time.Sleep(2 * time.Second)
		}
	}

	cnt := 0
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
				cnt += 1
				if cnt%1000 == 0 {
					ks.Log.Info("Processed", "messages", cnt)
				}
				headers := collectHeaders(e.Headers)

				var reqPayload payload.SeldonPayload
				var err error
				switch ks.Transport {
				case api.TransportRest:
					// Assume JSON if no content type - should maybe be application/octet-stream?
					contentType := rest.ContentTypeJSON
					if ct, ok := headers[http.ContentType]; ok {
						if len(ct) == 1 {
							contentType = ct[0]
						}
					}
					reqPayload, err = ks.Client.Unmarshall(e.Value, contentType)
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

				job := KafkaJob{
					headers:    headers,
					reqKey:     e.Key,
					reqPayload: reqPayload,
				}
				// enqueue a job
				jobChan <- &job

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

	ks.Log.Info("Final Processed", "messages", cnt)
	ks.Log.Info("Closing consumer")
	close(cancelChan)
	c.Close()
	return nil
}

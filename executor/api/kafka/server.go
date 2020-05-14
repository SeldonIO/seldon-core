package kafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-logr/logr"
	guuid "github.com/google/uuid"
	"github.com/seldonio/seldon-core/executor/api/client"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/predictor"
	v1 "github.com/seldonio/seldon-core/operator/apis/machinelearning.seldon.io/v1"
	"net/url"
	"os"
	"os/signal"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"syscall"
)

type SeldonKafkaServer struct {
	Client         client.SeldonApiClient
	DeploymentName string
	Namespace      string
	predictor      *v1.PredictorSpec
	Broker         string
	TopicIn        string
	TopicOut       string
	ServerUrl      *url.URL
	Log            logr.Logger
}

func NewKafkaServer(deploymentName, namespace, protocol string, annotations map[string]string, serverUrl *url.URL, predictor *v1.PredictorSpec, broker, topicIn, topicOut string, log logr.Logger) *SeldonKafkaServer {
	client := NewKafkaClient(serverUrl.Hostname(), deploymentName, namespace, protocol, predictor, broker, log)
	return &SeldonKafkaServer{
		Client:         client,
		DeploymentName: deploymentName,
		Namespace:      namespace,
		predictor:      predictor,
		Broker:         broker,
		TopicIn:        topicIn,
		TopicOut:       topicOut,
		ServerUrl:      serverUrl,
		Log:            log.WithName("KafkaServer"),
	}
}

func (ks *SeldonKafkaServer) getGroupName() string {
	return ks.predictor.Name + "." + ks.DeploymentName + "." + ks.Namespace
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
				reqPayload, err := ks.Client.Unmarshall(e.Value)
				if err != nil {
					ks.Log.Error(err, "Failed to unmarshall payload")
					continue
				}

				go func() {
					ctx := context.Background()
					// Add Seldon Puid to Context
					ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, headers[payload.SeldonPUIDHeader][0])

					seldonPredictorProcess := predictor.NewPredictorProcess(ctx, ks.Client, logf.Log.WithName("KafkaClient"), ks.ServerUrl, ks.Namespace, headers)

					resPayload, err := seldonPredictorProcess.Predict(ks.predictor.Graph, reqPayload)
					if err != nil {
						ks.Log.Error(err, "Failed prediction")
						return
					}
					resBytes, err := resPayload.GetBytes()
					if err != nil {
						ks.Log.Error(err, "Failed to get bytes from prediction response")
						return
					}

					err = p.Produce(&kafka.Message{
						TopicPartition: kafka.TopicPartition{Topic: &ks.TopicOut, Partition: kafka.PartitionAny},
						Value:          resBytes,
						Headers:        []kafka.Header{{Key: "myTestHeader", Value: []byte("header values are binary")}},
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

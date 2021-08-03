package kafka

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cloudevents/sdk-go/pkg/bindings/http"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-logr/logr"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/rest"
)

const (
	KeyTopicResponse = "topic-response"
	KeyMethod        = "seldon-method"
	KeyProtoName     = "proto-name"
)

type KafkaRPC struct {
	Client       *KafkaClient
	Producer     *kafka.Producer
	Broker       string
	GroupId      string
	TopicReceive string
	TopicSend    string
	Receivers    map[string]chan<- payload.SeldonPayload
	Lock         sync.RWMutex
	Log          logr.Logger
}

func getTopicReceiveForModel(modelName string, kc *KafkaClient) string {
	return kc.Hostname + "." + modelName + "." + kc.predictor.Name + "." + kc.DeploymentName + "." + kc.Namespace
}

func getTopicSendForModel(modelName string, kc *KafkaClient) string {
	return modelName + "." + kc.predictor.Name + "." + kc.DeploymentName + "." + kc.Namespace
}

func NewKafkaRPC(client *KafkaClient, modelName string) (*KafkaRPC, error) {
	topicReceive := getTopicReceiveForModel(modelName, client)
	// Set group name to be same as receive topic
	groupId := topicReceive

	// Create producer
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": client.Broker})
	if err != nil {
		return nil, err
	}
	client.Log.Info("Created", "producer", p.String())

	return &KafkaRPC{
		Client:       client,
		Producer:     p,
		Broker:       client.Broker,
		GroupId:      groupId,
		TopicSend:    getTopicSendForModel(modelName, client),
		TopicReceive: topicReceive,
		Receivers:    make(map[string]chan<- payload.SeldonPayload),
		Lock:         sync.RWMutex{},
		Log:          client.Log.WithName("KafkaRPC"),
	}, nil
}

func getPuidFromHeaders(headers []kafka.Header) string {
	for _, header := range headers {
		if header.Key == payload.SeldonPUIDHeader {
			return string(header.Value)
		}
	}
	return ""
}

func (tp *KafkaRPC) start() {
	go func() {
		c, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":     tp.Broker,
			"broker.address.family": "v4",
			"group.id":              tp.GroupId,
			"session.timeout.ms":    6000,
			"auto.offset.reset":     "earliest"})
		if err != nil {
			tp.Log.Error(err, "Failed to create consumer", "groupId", tp.GroupId)
		}

		err = c.SubscribeTopics([]string{tp.TopicReceive}, nil)
		if err != nil {
			tp.Log.Error(err, "Failed to subscribe to topic", "topic", tp.TopicReceive)
			return
		}

		tp.Log.Info("Created", "consumer", c.String(), "topic", tp.TopicReceive)
		run := true
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

		for run {
			select {
			case sig := <-sigchan:
				tp.Log.Info("Terminating", "signal", sig)
				run = false
			default:
				ev := c.Poll(100)
				if ev == nil {
					continue
				}

				switch e := ev.(type) {
				case *kafka.Message:
					tp.Log.Info("Message", "Partition", e.TopicPartition)

					headers := collectHeaders(e.Headers)

					// Assume JSON if no content type - should maybe be application/octet-stream?
					contentType := rest.ContentTypeJSON
					if ct, ok := headers[http.ContentType]; ok {
						if len(ct) == 1 {
							contentType = ct[0]
						}
					}
					msg, err := tp.Client.Unmarshall(e.Value, contentType)
					if err != nil {
						tp.Log.Error(err, "Failed to unmarshal consume", "topic")
					} else {
						puid := getPuidFromHeaders(e.Headers)
						if puid == "" {
							tp.Log.Info("Failed to find puid in message", "topic", tp.TopicReceive)
						} else {
							tp.Lock.Lock()
							if c, ok := tp.Receivers[puid]; ok {
								c <- msg
							} else {
								tp.Log.Info("Failed to find receiver key for", "puid", puid)
							}
							tp.Lock.Unlock()
						}
					}

				case kafka.Error:
					// Errors should generally be considered
					// informational, the client will try to
					// automatically recover.
					// But in this example we choose to terminate
					// the application if all brokers are down.
					tp.Log.Error(e, "Received kafka error")
					if e.Code() == kafka.ErrAllBrokersDown {
						run = false
					}
				default:
					tp.Log.Info("Ignored", "msg", e)
				}
			}
		}

		tp.Log.Info("Closing consumer")
		c.Close()
	}()
}

func (tp *KafkaRPC) call(msg []byte, puid string, method string) (payload.SeldonPayload, error) {
	//add to receivers
	c := make(chan payload.SeldonPayload)
	tp.Lock.Lock()
	tp.Receivers[puid] = c
	tp.Lock.Unlock()
	//produce msg with topic for reply in headers
	err := tp.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &tp.TopicSend, Partition: kafka.PartitionAny},
		Value:          msg,
		Headers: []kafka.Header{
			{Key: payload.SeldonPUIDHeader, Value: []byte(puid)},
			{Key: KeyTopicResponse, Value: []byte(tp.TopicReceive)},
			{Key: KeyMethod, Value: []byte(method)},
		}}, nil)
	if err != nil {
		tp.Log.Error(err, "Failed to produce request", "topic", tp.TopicSend)
		return nil, err
	}
	//wait for response

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigchan:
		tp.Log.Info("Terminating", "signal", sig)
		return nil, fmt.Errorf("Terminated")
	case res := <-c:
		return res, nil
	}
}

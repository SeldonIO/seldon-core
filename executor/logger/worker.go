package logger

import (
	"context"
	"errors"
	"fmt"
	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/go-logr/logr"
	"net/http"
	"time"
)

const (
	CEInferenceRequest  = "io.seldon.serving.inference.request"
	CEInferenceResponse = "io.seldon.serving.inference.response"
	CEFeedback          = "io.seldon.serving.feedback"
	// cloud events extension attributes have to be lowercase alphanumeric
	RequestIdAttr            = "requestid"
	ModelIdAttr              = "modelid"
	InferenceServiceNameAttr = "inferenceservicename"
	NamespaceAttr            = "namespace"
	EndpointAttr             = "endpoint"
	KafkaTypeHeader          = "type"
	KafkaContentTypeHeader   = "content-type"
)

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func NewWorker(id int, workerQueue chan chan LogRequest, log logr.Logger, sdepName string, namespace string, predictorName string, kafkaBroker string, kafkaTopic string) (*Worker, error) {

	var producer *kafka.Producer
	var err error
	if kafkaBroker != "" {
		log.Info("Creating producer", "broker", kafkaBroker, "topic", kafkaTopic)
		producer, err = kafka.NewProducer(&kafka.ConfigMap{
			"bootstrap.servers":   kafkaBroker,
			"go.delivery.reports": false, // Need this othewise will get memory leak
		})
		if err != nil {
			return nil, err
		}
		log.Info("Created Logger Kafka Producer", "producer", producer.String())
	}

	// Create, and return the worker.
	return &Worker{
		Log:         log,
		ID:          id,
		Work:        make(chan LogRequest),
		WorkerQueue: workerQueue,
		QuitChan:    make(chan bool),
		Client: http.Client{
			Timeout: 60 * time.Second,
		},
		CeCtx:         cloudevents.ContextWithEncoding(context.Background(), cloudevents.Binary),
		SdepName:      sdepName,
		Namespace:     namespace,
		PredictorName: predictorName,
		KafkaTopic:    kafkaTopic,
		Producer:      producer,
	}, nil
}

type Worker struct {
	Log           logr.Logger
	ID            int
	Work          chan LogRequest
	WorkerQueue   chan chan LogRequest
	QuitChan      chan bool
	Client        http.Client
	CeCtx         context.Context
	CeTransport   transport.Transport
	SdepName      string
	Namespace     string
	PredictorName string
	KafkaTopic    string
	Producer      *kafka.Producer
}

func getCEType(logReq LogRequest) (string, error) {
	switch logReq.ReqType {
	case InferenceRequest:
		return CEInferenceRequest, nil
	case InferenceResponse:
		return CEInferenceResponse, nil
	case InferenceFeedback:
		return CEFeedback, nil
	default:
		return "", fmt.Errorf("Incorrect log request type: %s", errors.New("Incorrect log request type"))
	}
}

func (w *Worker) sendKafkaEvent(logReq LogRequest) error {

	reqType, err := getCEType(logReq)
	if err != nil {
		return err
	}

	kafkaHeaders := []kafka.Header{
		{Key: KafkaTypeHeader, Value: []byte(reqType)},
		{Key: KafkaContentTypeHeader, Value: []byte(logReq.ContentType)},
		{Key: ModelIdAttr, Value: []byte(logReq.ModelId)},
		{Key: RequestIdAttr, Value: []byte(logReq.RequestId)},
		{Key: InferenceServiceNameAttr, Value: []byte(w.SdepName)},
		{Key: NamespaceAttr, Value: []byte(w.Namespace)},
		{Key: EndpointAttr, Value: []byte(w.PredictorName)},
	}

	err = w.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &w.KafkaTopic, Partition: kafka.PartitionAny},
		Value:          *logReq.Bytes,
		Headers:        kafkaHeaders,
	}, nil)
	if err != nil {
		w.Log.Error(err, "Failed to produce response")
		return err
	}

	return nil
}

func (w *Worker) sendCloudEvent(logReq LogRequest) error {

	t, err := cloudevents.NewHTTPTransport(
		cloudevents.WithTarget(logReq.Url.String()),
		cloudevents.WithEncoding(cloudevents.HTTPBinaryV1),
	)

	if err != nil {
		return fmt.Errorf("while creating http transport: %s", err)
	}
	c, err := cloudevents.NewClient(t,
		cloudevents.WithTimeNow(),
	)
	if err != nil {
		return fmt.Errorf("while creating new cloudevents client: %s", err)
	}
	event := cloudevents.NewEvent(cloudevents.VersionV1)
	event.SetID(logReq.Id)
	if refType, err := getCEType(logReq); err == nil {
		event.SetType(refType)
	} else {
		return err
	}

	event.SetExtension(ModelIdAttr, logReq.ModelId)
	event.SetExtension(RequestIdAttr, logReq.RequestId)
	event.SetExtension(InferenceServiceNameAttr, w.SdepName)
	event.SetExtension(NamespaceAttr, w.Namespace)
	//use 'endpoint' for the header to align with kfserving - https://github.com/kubeflow/kfserving/pull/699/files#r385360114
	event.SetExtension(EndpointAttr, w.PredictorName)

	event.SetSource(logReq.SourceUri.String())
	event.SetDataContentType(logReq.ContentType)
	if err := event.SetData(*logReq.Bytes); err != nil {
		return fmt.Errorf("while setting cloudevents data: %s", err)
	}

	//fmt.Printf("%+v\n", event)

	if _, _, err := c.Send(w.CeCtx, event); err != nil {
		return fmt.Errorf("while sending event: %s", err)
	}
	return nil
}

// This function "starts" the worker by starting a goroutine, that is
// an infinite "for-select" loop.
func (w *Worker) Start() {
	go func() {
		for {
			// Add ourselves into the worker queue.
			w.WorkerQueue <- w.Work

			select {
			case work := <-w.Work:
				// Receive a work request.

				if w.KafkaTopic != "" {
					if err := w.sendKafkaEvent(work); err != nil {
						w.Log.Error(err, "Failed to send kafka log", "Topic", w.KafkaTopic)
					}
				} else {
					if err := w.sendCloudEvent(work); err != nil {
						w.Log.Error(err, "Failed to send cloudevent log", "URL", work.Url.String())
					}
				}

			case <-w.QuitChan:
				// We have been asked to stop.
				fmt.Printf("worker %d stopping\n", w.ID)
				return
			}
		}
	}()
}

// Stop tells the worker to stop listening for work requests.
//
// Note that the worker will only stop *after* it has finished its work.
func (w *Worker) Stop() {
	go func() {
		w.QuitChan <- true
	}()
}

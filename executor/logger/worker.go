package logger

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	cloudevents "github.com/cloudevents/sdk-go"
	"github.com/cloudevents/sdk-go/pkg/cloudevents/transport"
	"github.com/go-logr/logr"
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
)

// NewWorker creates, and returns a new Worker object. Its only argument
// is a channel that the worker can add itself to whenever it is done its
// work.
func NewWorker(id int, workerQueue chan chan LogRequest, log logr.Logger, sdepName string, namespace string, predictorName string) Worker {
	// Create, and return the worker.
	return Worker{
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
	}
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
}

func (W *Worker) sendCloudEvent(logReq LogRequest) error {

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
	if logReq.ReqType == InferenceRequest {
		event.SetType(CEInferenceRequest)
	} else if logReq.ReqType == InferenceResponse {
		event.SetType(CEInferenceResponse)
	} else if logReq.ReqType == InferenceFeedback {
		event.SetType(CEFeedback)
	} else {
		return fmt.Errorf("Incorrect log request type: %s", errors.New("Incorrect log request type"))
	}

	event.SetExtension(ModelIdAttr, logReq.ModelId)
	event.SetExtension(RequestIdAttr, logReq.RequestId)
	event.SetExtension(InferenceServiceNameAttr, W.SdepName)
	event.SetExtension(NamespaceAttr, W.Namespace)
	//use 'endpoint' for the header to align with kfserving - https://github.com/kubeflow/kfserving/pull/699/files#r385360114
	event.SetExtension(EndpointAttr, W.PredictorName)

	event.SetSource(logReq.SourceUri.String())
	event.SetDataContentType(logReq.ContentType)
	if err := event.SetData(*logReq.Bytes); err != nil {
		return fmt.Errorf("while setting cloudevents data: %s", err)
	}

	//fmt.Printf("%+v\n", event)

	if _, _, err := c.Send(W.CeCtx, event); err != nil {
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
				fmt.Printf("worker%d: Received work request for %s\n", w.ID, work.Url.String())

				if err := w.sendCloudEvent(work); err != nil {
					w.Log.Error(err, "Failed to send log", "URL", work.Url.String())
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

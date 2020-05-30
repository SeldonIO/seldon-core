package kafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	proto2 "github.com/golang/protobuf/proto"
	"github.com/seldonio/seldon-core/executor/api"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/predictor"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

type KafkaJob struct {
	headers    map[string][]string
	reqPayload payload.SeldonPayload
}

func (ks *SeldonKafkaServer) worker(jobChan <-chan *KafkaJob, cancelChan <-chan struct{}) {
	for {
		select {
		case <-cancelChan:
			return

		case job := <-jobChan:
			ks.processKafkaRequest(job)
		}
	}
}

func (ks *SeldonKafkaServer) processKafkaRequest(job *KafkaJob) {
	ctx := context.Background()
	// Add Seldon Puid to Context
	ctx = context.WithValue(ctx, payload.SeldonPUIDHeader, job.headers[payload.SeldonPUIDHeader][0])

	seldonPredictorProcess := predictor.NewPredictorProcess(ctx, ks.Client, logf.Log.WithName("KafkaClient"), ks.ServerUrl, ks.Namespace, job.headers)

	resPayload, err := seldonPredictorProcess.Predict(ks.Predictor.Graph, job.reqPayload)
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

	err = ks.Producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &ks.TopicOut, Partition: kafka.PartitionAny},
		Value:          resBytes,
		Headers:        kafkaHeaders,
	}, nil)
	if err != nil {
		ks.Log.Error(err, "Failed to produce response")
	}
}

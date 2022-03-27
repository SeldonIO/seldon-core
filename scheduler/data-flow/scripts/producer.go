package main

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
)

//nolint:typecheck
func main() {
	logger := log.New()

	args := parseArguments(logger)

	p, err := newProducer(logger, args)
	if err != nil {
		logger.WithError(err).Fatal("unable to create Kafka producer")
	}
	p.produce(numMessages)
	p.close()
}

type producer struct {
	client         *kafka.Producer
	logger         log.FieldLogger
	topics         []string
	pipelineHeader string
}

func newProducer(logger log.FieldLogger, args *Args) (*producer, error) {
	kConf := &kafka.ConfigMap{
		"bootstrap.servers": args.bootstrapServers,
		"security.protocol": args.securityProtocol,
	}
	kp, err := kafka.NewProducer(kConf)
	if err != nil {
		return nil, err
	}

	return &producer{
		client:         kp,
		logger:         logger,
		topics:         args.inputTopics,
		pipelineHeader: args.pipelineHeader,
	}, nil
}

func (p *producer) close() {
	p.client.Flush(5_000)
	p.client.Close()
}

func (p *producer) produce(n uint) {
	var i uint

	// Should not match pipeline header
	for i = 0; i < n; i++ {
		msg := makeV2Response(int32(i))
		for _, t := range p.topics {
			p.publish(t, false, msg)
		}
	}

	// Should match pipeline header
	for i = 0; i < n; i++ {
		msg := makeV2Response(int32(i))
		for _, t := range p.topics {
			p.publish(t, true, msg)
		}
	}
}

func (p *producer) publish(topic string, useHeader bool, msg *[]byte) {
	h := p.pipelineHeader
	if !useHeader {
		h += "-other"
	}

	err := p.client.Produce(
		&kafka.Message{
			TopicPartition: kafka.TopicPartition{
				Topic:     &topic,
				Partition: kafka.PartitionAny,
			},
			Value: *msg,
			Headers: []kafka.Header{
				{Key: pipelineHeader, Value: []byte(h)},
			},
		},
		nil,
	)
	if err != nil {
		p.logger.WithError(err).Error("unable to publish message")
	}
}

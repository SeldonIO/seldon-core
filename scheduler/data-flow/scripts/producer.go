/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

package main

import (
	"github.com/confluentinc/confluent-kafka-go/kafka"
	log "github.com/sirupsen/logrus"
)

//nolint:typecheck
func main() {
	logger := log.New()

	args := parseArguments(logger)

	c, err := newConsumer(logger, args)
	if err != nil {
		logger.WithError(err).Fatal("unable to create Kafka consumer")
	}
	c.consume(numMessages)
	c.close()
}

type consumer struct {
	client *kafka.Consumer
	logger log.FieldLogger
	topics []string
}

func newConsumer(logger log.FieldLogger, args *Args) (*consumer, error) {
	kConf := &kafka.ConfigMap{
		"bootstrap.servers": args.bootstrapServers,
		"security.protocol": args.securityProtocol,
		"group.id":          "data-flow-test",
		"auto.offset.reset": "earliest",
	}
	kc, err := kafka.NewConsumer(kConf)
	if err != nil {
		return nil, err
	}

	return &consumer{
		client: kc,
		logger: logger,
		topics: args.outputTopics,
	}, nil
}

func (c *consumer) consume(n uint) {
	err := c.client.SubscribeTopics(c.topics, nil)
	if err != nil {
		c.logger.WithError(err).Fatal("unable to subscribe to topics")
	}

	var i uint
	for i < n*uint(len(c.topics)) {
		event := c.client.Poll(50)
		if event == nil {
			continue
		}

		switch et := event.(type) {
		case *kafka.Message:
			i++

			msg, err := parseV2Request(et.Value)
			if err != nil {
				c.logger.WithError(err).Error("failed to parse message as v2 inference response")
			}

			c.logger.Infof(
				"received inference response on topic %s: %v",
				*et.TopicPartition.Topic,
				msg,
			)
		case kafka.Error:
			c.logger.WithError(et).Warn("failed to read message")
		default:
			c.logger.Infof("ignoring event %v", event)
		}
	}
}

func (c *consumer) close() {
	c.client.Close()
}

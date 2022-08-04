package rabbitmq

import (
	"fmt"
	"github.com/go-logr/logr"
	amqp "github.com/rabbitmq/amqp091-go"
	"os"
	"os/signal"
	"syscall"
)

/*
 * based on patterns from https://github.com/dominodatalab/forge/blob/master/internal/message/amqp/publisher.go
 */

type consumer struct {
	connection
	queueName   string
	consumerTag string
}

func NewConsumer(uri, queueName, consumerTag string, logger logr.Logger) (*consumer, error) {
	c, err := NewConnection(uri, logger)
	if err != nil {
		c.log.Error(err, "error creating connection for consumer", "uri", c.uri)
		return nil, fmt.Errorf("error '%w' creating connection to '%v' for consumer", err, uri)
	}
	return &consumer{
		*c,
		queueName,
		consumerTag,
	}, nil
}

// In the event that the underlying connection was closed after connection creation, this function will attempt to
// reconnection to the AMQP broker before performing these operations.
// this is a blocking function while the consumer is running, run it in a goroutine if needed
func (c *consumer) Consume(
	payloadHandler func(*SeldonPayloadWithHeaders) error,
	errorHandler func(args ConsumerError) error,
) error {
	select {
	case <-c.err:
		c.log.Info("attempting to reconnect to rabbitmq", "uri", c.uri)

		if err := c.connect(); err != nil {
			c.log.Error(err, "error reconnecting to rabbitmq")
			return fmt.Errorf("error '%w' reconnecting to rabbitmq", err)
		}
	default:
	}

	_, err := c.DeclareQueue(c.queueName)
	if err != nil {
		c.log.Error(err, "error declaring rabbitmq queue", "uri", c.uri)
		return fmt.Errorf("error '%w' declaring rabbitmq queue", err)
	}

	// default exchange with queue name as name is the same as a direct exchange routed to the queue
	deliveries, err := c.channel.Consume(
		c.queueName,   // name
		c.consumerTag, // consumerTag,
		false,         // autoAck
		false,         // exclusive
		false,         // noLocal
		false,         // noWait
		amqp.Table{},  // arguments TODO should something go here?
	)
	if err != nil {
		c.log.Error(err, "error consuming from rabbitmq queue")
		return fmt.Errorf("error '%w' consuming from rabbitmq queue", err)
	}

	// TODO does this need more error handling?  What about if the connection or channel fail while
	// the handler is running?
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	for delivery := range deliveries {
		select {
		case sig := <-sigChan:
			c.log.Info("terminating due to signal", "signal", sig)
			break
		default:
			pl, err := DeliveryToPayload(delivery)
			if err != nil {
				errorHandlerErr := c.handleConsumerError(ConsumerError{err, delivery, pl}, errorHandler)
				if errorHandlerErr != nil {
					// fatal error, abort the consumer
					return errorHandlerErr
				}
				continue
			}
			err = payloadHandler(pl)
			if err != nil {
				errorHandlerErr := c.handleConsumerError(ConsumerError{err, delivery, pl}, errorHandler)
				if errorHandlerErr != nil {
					// fatal error, abort the consumer
					return errorHandlerErr
				}
				continue
			}
			ackErr := delivery.Ack(false)
			if ackErr != nil {
				// fatal error, abort the consumer
				return ackErr
			}
		}
	}
	close(sigChan)

	return nil
}

type ConsumerError struct {
	err      error
	delivery amqp.Delivery
	pl       *SeldonPayloadWithHeaders // might be nil
}

func (c *consumer) handleConsumerError(
	args ConsumerError,
	errorHandler func(args ConsumerError) error,
) error {
	delivery := args.delivery

	// retry once
	if delivery.Redelivered {
		err := errorHandler(args)
		if err != nil {
			c.log.Error(err, "error handler encountered an error", "original error", args.err)
			return fmt.Errorf("error handler encountered an error '%w' when handling original error '%v'", err, args.err)
		}

		rejectErr := delivery.Reject(false)
		if rejectErr != nil {
			c.log.Error(rejectErr, "error rejecting")
			return fmt.Errorf("error '%w' rejecting", rejectErr)
		}
	} else {
		nackErr := delivery.Nack(false, true)
		if nackErr != nil {
			c.log.Error(nackErr, "error nack-ing")
			return fmt.Errorf("error '%w' nack-ing", nackErr)
		}
	}

	return nil
}

package rabbitmq

import (
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

/*
 * mostly taken from https://github.com/dominodatalab/forge/blob/master/internal/message/amqp/publisher.go
 */

const (
	publishMandatory = true
	publishImmediate = false
)

type publisher struct {
	connection
	queueName string
}

func NewPublisher(uri, queueName string, logger logr.Logger) (*publisher, error) {
	c, err := NewConnection(uri, logger)
	if err != nil {
		return nil, err
	}
	return &publisher{
		*c,
		queueName,
	}, nil
}

// In the event that the underlying connection was closed after connection creation, this function will attempt to
// reconnection to the AMQP broker before performing these operations.
func (p *publisher) Publish(payload SeldonPayloadWithHeaders) error {
	select {
	case <-p.err:
		p.log.Info("attempting to reconnect to rabbitmq", "uri", p.uri)

		if err := p.connect(); err != nil {
			return errors.Wrap(err, "could not reconnect to rabbitmq")
		}
	default:
	}

	_, err := p.DeclareQueue(p.queueName)
	if err != nil {
		return errors.Wrap(err, "failed to declare rabbitmq queue")
	}

	body, err := payload.GetBytes()
	if err != nil {
		return errors.Wrap(err, "could not get payload bytes")
	}
	message := amqp.Publishing{
		Headers:         StringMapToTable(payload.Headers),
		ContentType:     payload.GetContentType(),
		ContentEncoding: payload.GetContentEncoding(),
		Body:            body,
	}
	err = p.channel.Publish(amqpExchange, p.queueName, publishMandatory, publishImmediate, message)
	return errors.Wrap(err, "failed to publish rabbitmq message")
}

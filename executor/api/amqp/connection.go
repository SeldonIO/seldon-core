package amqp

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	seldon "github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"time"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	connectionRetryLimit = 5

	queueDurable    = true
	queueAutoDelete = false
	queueExclusive  = false
	queueNoWait     = false

	amqpExchange     = ""
	publishMandatory = true
	publishImmediate = false
)

var (
	connectionRetryDelay = 5 * time.Second
	queueArgs            = amqp.Table{
		"x-single-active-consumer": true,
	}
)

type connection struct {
	log logr.Logger
	uri string

	conn    Connection
	channel Channel
	err     chan error
}

type publisher struct {
	connection
	queueName string
}

type consumer struct {
	connection
	queueName   string
	consumerTag string
	stop        chan bool
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

func NewConsumer(uri, queueName, consumerTag string, logger logr.Logger) (*consumer, error) {
	c, err := NewConnection(uri, logger)
	if err != nil {
		return nil, err
	}
	return &consumer{
		*c,
		queueName,
		consumerTag,
		make(chan bool),
	}, nil
}

// NewConnection creates a new AMQP connection that targets a specific broker uri and queue.
func NewConnection(uri string, logger logr.Logger) (*connection, error) {
	p := &connection{
		uri: uri,
		err: make(chan error),
		log: logger.WithName("MessagePublisher"),
	}

	if err := p.connect(); err != nil {
		return nil, err
	}
	return p, nil
}

// In the event that the underlying connection was closed after connection creation, this function will attempt to
// reconnection to the AMQP broker before performing these operations.
func (c *consumer) Consume(handler func(<-chan payload.SeldonPayload, <-chan error) <-chan error) (error, <-chan error) {
	select {
	case <-c.err:
		c.log.Info("attempting to reconnect to rabbitmq", "uri", c.uri)

		if err := c.connect(); err != nil {
			return errors.Wrap(err, "could not reconnect to rabbitmq"), nil
		}
	default:
	}

	_, err := c.channel.QueueDeclare(
		c.queueName,
		queueDurable,
		queueAutoDelete,
		queueExclusive,
		queueNoWait,
		queueArgs,
	)
	if err != nil {
		return errors.Wrap(err, "failed to declare rabbitmq queue"), nil
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
		return errors.Wrap(err, "Queue Consume error"), nil
	}

	payloads, errs := MapToPayload(deliveries)
	// TODO does this need more error handling?  What about if the connection or channel fail while
	// the handler is running?
	errs = handler(payloads, errs)

	return nil, errs
}

func MapToPayload(deliveries <-chan amqp.Delivery) (<-chan payload.SeldonPayload, <-chan error) {
	payloads := make(chan payload.SeldonPayload)
	errs := make(chan error)

	go func() {
		for delivery := range deliveries {
			var pl payload.SeldonPayload
			var err error = nil

			switch delivery.ContentType {
			case payload.APPLICATION_TYPE_PROTOBUF:
				var message = &seldon.SeldonMessage{}
				err = proto.Unmarshal(delivery.Body, message)
				if err == nil {
					pl = &payload.ProtoPayload{message}
				}
			case rest.ContentTypeJSON:
				pl = &payload.BytesPayload{
					delivery.Body,
					delivery.ContentType,
					delivery.ContentEncoding,
				}
			default:
				err = fmt.Errorf("unknown payload type: %s", delivery.ContentType)
			}

			if err != nil {
				errs <- err
			} else {
				payloads <- pl
			}
		}
		close(payloads)
		close(errs)
	}()

	return payloads, errs
}

// In the event that the underlying connection was closed after connection creation, this function will attempt to
// reconnection to the AMQP broker before performing these operations.
func (p *publisher) Publish(payload payload.SeldonPayload) error {
	select {
	case <-p.err:
		p.log.Info("attempting to reconnect to rabbitmq", "uri", p.uri)

		if err := p.connect(); err != nil {
			return errors.Wrap(err, "could not reconnect to rabbitmq")
		}
	default:
	}

	q, err := p.channel.QueueDeclare(
		p.queueName,
		queueDurable,
		queueAutoDelete,
		queueExclusive,
		queueNoWait,
		queueArgs,
	)
	if err != nil {
		return errors.Wrap(err, "failed to declare rabbitmq queue")
	}

	body, err := payload.GetBytes()
	if err != nil {
		return errors.Wrap(err, "could not get payload bytes")
	}
	message := amqp.Publishing{
		ContentType:     payload.GetContentType(),
		ContentEncoding: payload.GetContentEncoding(),
		Body:            body,
	}
	err = p.channel.Publish(amqpExchange, q.Name, publishMandatory, publishImmediate, message)
	return errors.Wrap(err, "failed to publish rabbitmq message")
}

// Close will close the underlying AMQP connection if one has been set, and this operation will cascade down to any
// channels created under this connection.
func (p *connection) Close() error {
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

// implements retry logic with delays for establishing AMQP connections.
func (p *connection) connect() error {
	ticker := time.NewTicker(connectionRetryDelay)
	defer ticker.Stop()

	for counter := 0; counter < connectionRetryLimit; <-ticker.C {
		var err error

		p.conn, err = defaultDialerAdapter(p.uri)
		if err != nil {
			p.log.Error(err, "cannot dial rabbitmq", "uri", p.uri, "attempt", counter+1)

			counter++
			continue
		}

		go func() {
			closed := make(chan *amqp.Error, 1)
			p.conn.NotifyClose(closed)

			reason, ok := <-closed
			if ok {
				p.log.Error(reason, "rabbitmq connection closed, registering err signal")
				p.err <- reason
			}
		}()

		p.channel, err = p.conn.Channel()
		return errors.Wrapf(err, "failed to create rabbitmq channel to %q", p.uri)
	}

	return fmt.Errorf("rabbitmq connection retry limit reached: %d", connectionRetryLimit)
}

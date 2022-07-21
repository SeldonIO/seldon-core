package rabbitmq

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	seldon "github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"os"
	"os/signal"
	"syscall"
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
	queueArgs            = amqp.Table{}
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

func (c *connection) DeclareQueue(queueName string) (amqp.Queue, error) {
	return c.channel.QueueDeclare(
		queueName,
		queueDurable,
		queueAutoDelete,
		queueExclusive,
		queueNoWait,
		queueArgs,
	)
}

// In the event that the underlying connection was closed after connection creation, this function will attempt to
// reconnection to the AMQP broker before performing these operations.
// this is a blocking function while the consumer is running, run it in a goroutine if needed
func (c *consumer) Consume(payloadHandler func(SeldonPayloadWithHeaders) error, errorHandler func(error)) error {
	select {
	case <-c.err:
		c.log.Info("attempting to reconnect to rabbitmq", "uri", c.uri)

		if err := c.connect(); err != nil {
			return errors.Wrap(err, "could not reconnect to rabbitmq")
		}
	default:
	}

	_, err := c.DeclareQueue(c.queueName)
	if err != nil {
		return errors.Wrap(err, "failed to declare rabbitmq queue")
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
		return errors.Wrap(err, "Queue Consume error")
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
				HandleConsumerError(err, errorHandler, delivery, c.log)
				continue
			}
			err = payloadHandler(pl)
			if err != nil {
				HandleConsumerError(err, errorHandler, delivery, c.log)
				continue
			}
			ackErr := delivery.Ack(false)
			if ackErr != nil {
				c.log.Error(ackErr, "error ack-ing", "delivery", delivery)
			}
		}
	}
	close(sigChan)

	return nil
}

func HandleConsumerError(err error, errorHandler func(error), delivery amqp.Delivery, log logr.Logger) {
	errorHandler(err)

	// retry once
	if delivery.Redelivered {
		rejectErr := delivery.Reject(false)
		if rejectErr != nil {
			// perhaps we should do more in this case, but I'm not sure what, fail the entire app?
			log.Error(rejectErr, "error rejecting", "delivery", delivery)
		}
	} else {
		nackErr := delivery.Nack(false, true)
		if nackErr != nil {
			// perhaps we should do more in this case, but I'm not sure what, fail the entire app?
			log.Error(nackErr, "error nack-ing", "delivery", delivery)
		}
	}
}

func TableToStringMap(t amqp.Table) map[string][]string {
	stringMap := make(map[string][]string)
	for key, value := range t {
		stringMap[key] = []string{fmt.Sprintf("%v", value)}
	}
	return stringMap
}

func StringMapToTable(m map[string][]string) amqp.Table {
	table := make(map[string]interface{})
	for key, values := range m {
		// just take the first value, at least for now
		table[key] = values[0]
	}
	return table
}

func DeliveryToPayload(delivery amqp.Delivery) (SeldonPayloadWithHeaders, error) {
	var pl SeldonPayloadWithHeaders
	var err error = nil

	headers := TableToStringMap(delivery.Headers)

	switch delivery.ContentType {
	case payload.APPLICATION_TYPE_PROTOBUF:
		var message = &seldon.SeldonMessage{}
		err = proto.Unmarshal(delivery.Body, message)
		if err == nil {
			pl = SeldonPayloadWithHeaders{
				&payload.ProtoPayload{Msg: message},
				headers,
			}
		}
	case rest.ContentTypeJSON:
		pl = SeldonPayloadWithHeaders{
			&payload.BytesPayload{
				Msg:             delivery.Body,
				ContentType:     delivery.ContentType,
				ContentEncoding: delivery.ContentEncoding,
			},
			headers,
		}
	default:
		err = fmt.Errorf("unknown payload type: %s", delivery.ContentType)
	}

	return pl, err

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

	q, err := p.DeclareQueue(p.queueName)
	if err != nil {
		return errors.Wrap(err, "failed to declare rabbitmq queue")
	}
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
	err = p.channel.Publish(amqpExchange, q.Name, publishMandatory, publishImmediate, message)
	return errors.Wrap(err, "failed to publish rabbitmq message")
}

// Close will close the underlying AMQP connection if one has been set, and this operation will cascade down to any
// channels created under this connection.
func (c *connection) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// implements retry logic with delays for establishing AMQP connections.
func (c *connection) connect() error {
	ticker := time.NewTicker(connectionRetryDelay)
	defer ticker.Stop()

	for counter := 0; counter < connectionRetryLimit; <-ticker.C {
		var err error

		c.conn, err = defaultDialerAdapter(c.uri)
		if err != nil {
			c.log.Error(err, "cannot dial rabbitmq", "uri", c.uri, "attempt", counter+1)

			counter++
			continue
		}

		go func() {
			closed := make(chan *amqp.Error, 1)
			c.conn.NotifyClose(closed)

			reason, ok := <-closed
			if ok {
				c.log.Error(reason, "rabbitmq connection closed, registering err signal")
				c.err <- reason
			}
		}()

		c.channel, err = c.conn.Channel()
		return errors.Wrapf(err, "failed to create rabbitmq channel to %q", c.uri)
	}

	return fmt.Errorf("rabbitmq connection retry limit reached: %d", connectionRetryLimit)
}

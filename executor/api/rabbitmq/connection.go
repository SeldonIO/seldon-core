package rabbitmq

import (
	"fmt"
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

	amqpExchange = ""
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

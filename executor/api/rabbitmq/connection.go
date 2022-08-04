package rabbitmq

import (
	"fmt"
	"time"

	"github.com/go-logr/logr"
	amqp "github.com/rabbitmq/amqp091-go"
)

/*
 * mostly taken from https://github.com/dominodatalab/forge/blob/master/internal/message/amqp/publisher.go
 */

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
	conn := &connection{
		uri: uri,
		err: make(chan error),
		log: logger.WithName("Connection"),
	}

	if err := conn.connect(); err != nil {
		conn.log.Error(err, "error connecting", "uri", uri)
		return nil, fmt.Errorf("error '%w' connecting to '%v'", err, uri)
	}
	return conn, nil
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
		err := c.conn.Close()
		if err != nil {
			return fmt.Errorf("error '%w' closing connection '%v'", err, c)
		}
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
		if err != nil {
			c.log.Error(err, "error creating rabbitmq channel", "uri", c.uri)
			return fmt.Errorf("error '%w' creating rabbitmq channel to %q", err, c.uri)
		}
		return nil
	}

	return fmt.Errorf("rabbitmq connection retry limit reached: %d", connectionRetryLimit)
}

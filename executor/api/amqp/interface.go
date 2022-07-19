package amqp

import (
	"github.com/seldonio/seldon-core/executor/api/payload"
	"io"

	amqp "github.com/rabbitmq/amqp091-go"
)

// default implementation leverages the real "streadway/amqp" dialer
var defaultDialerAdapter DialerAdapter = func(url string) (Connection, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}

	return ConnectionAdapter{conn}, nil
}

// DialerAdapter is a function that returns a handle to a Connection type.
type DialerAdapter func(url string) (Connection, error)

// ConnectionAdapter adapts the amqp.Connection type so that it adheres to our libraries interfaces.
type ConnectionAdapter struct {
	*amqp.Connection
}

// Channel adapts an amqp.Channel to our Channel interface.
func (c ConnectionAdapter) Channel() (Channel, error) {
	return c.Connection.Channel()
}

// Connection defines the AMQP connections operations required by this library.
type Connection interface {
	io.Closer

	Channel() (Channel, error)
	NotifyClose(receiver chan *amqp.Error) chan *amqp.Error
}

// Channel defines the AMQP channel operations required by this library.
type Channel interface {
	QueueDeclare(name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp.Table) (amqp.Queue, error)
	Publish(exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error
	Consume(name string, consumerTag string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error)
}

type SeldonPayloadWithHeaders struct {
	payload.SeldonPayload
	Headers map[string][]string
}

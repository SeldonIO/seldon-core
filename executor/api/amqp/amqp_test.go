package amqp

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"
)

type mockDialerAdapter struct {
	mock.Mock
}

func (m *mockDialerAdapter) Dial(url string) (Connection, error) {
	args := m.Called(url)
	amqpConn, _ := args.Get(0).(Connection)

	return amqpConn, args.Error(1)
}

type mockConnection struct {
	mock.Mock
}

func (m *mockConnection) Channel() (Channel, error) {
	args := m.Called()
	amqpCh, _ := args.Get(0).(Channel)

	return amqpCh, args.Error(1)
}

func (m *mockConnection) NotifyClose(receiver chan *amqp.Error) chan *amqp.Error {
	return receiver
}

func (m *mockConnection) Close() error {
	args := m.Called()
	return args.Error(0)
}

type mockChannel struct {
	mock.Mock
}

func (m *mockChannel) QueueDeclare(name string, durable bool, autoDelete bool, exclusive bool, noWait bool, args amqp.Table) (amqp.Queue, error) {
	mArgs := m.Called(name, durable, autoDelete, exclusive, noWait, args)
	return mArgs.Get(0).(amqp.Queue), mArgs.Error(1)
}

func (m *mockChannel) Publish(exchange string, key string, mandatory bool, immediate bool, msg amqp.Publishing) error {
	args := m.Called(exchange, key, mandatory, immediate, msg)
	return args.Error(0)
}

func (m *mockChannel) Consume(name string, consumerTag string, autoAck bool, exclusive bool, noLocal bool, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	mArgs := m.Called(name, consumerTag, autoAck, exclusive, noLocal, noWait, args)
	return mArgs.Get(0).(chan amqp.Delivery), mArgs.Error(1)
}

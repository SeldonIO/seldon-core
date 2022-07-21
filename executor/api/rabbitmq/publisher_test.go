package rabbitmq

import (
	"errors"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPublisher(t *testing.T) {
	testMessage := SeldonPayloadWithHeaders{
		&TestPayload{Msg: `"hello"`},
		make(map[string][]string),
	}

	t.Run("success", func(t *testing.T) {
		mockChan := &mockChannel{}

		mockChan.On("QueueDeclare", queueName, true, false, false, false,
			queueArgs).Return(amqp091.Queue{Name: queueName},
			nil)

		mockChan.On("Publish", "", queueName, true, false, amqp091.Publishing{
			Headers:     make(map[string]interface{}),
			ContentType: "application/json",
			Body:        []byte(`"hello"`),
		}).Return(nil)

		pub := &publisher{
			connection: connection{
				channel: mockChan,
			},
			queueName: queueName,
		}

		assert.NoError(t, pub.Publish(testMessage))

		mockChan.AssertExpectations(t)
	})

	t.Run("publish_failure", func(t *testing.T) {
		mockChan := &mockChannel{}

		mockChan.On("QueueDeclare", queueName, true, false, false, false,
			queueArgs).Return(amqp091.Queue{Name: queueName},
			nil)

		mockChan.On("Publish", "", queueName, true, false, amqp091.Publishing{
			Headers:     make(map[string]interface{}),
			ContentType: "application/json",
			Body:        []byte(`"hello"`),
		}).Return(errors.New("test error"))

		pub := &publisher{
			connection: connection{
				channel: mockChan,
			},
			queueName: queueName,
		}

		assert.EqualError(t, pub.Publish(testMessage), "failed to publish rabbitmq message: test error")

		mockChan.AssertExpectations(t)
	})

	t.Run("queue_declare_failure", func(t *testing.T) {
		mockChan := &mockChannel{}

		mockChan.On("QueueDeclare", queueName, true, false, false, false, queueArgs).Return(amqp091.Queue{},
			errors.New("test error"))

		pub := &publisher{
			connection: connection{
				channel: mockChan,
			},
			queueName: queueName,
		}

		assert.EqualError(t, pub.Publish(testMessage), "failed to declare rabbitmq queue: test error")

		mockChan.AssertExpectations(t)
	})

	//t.Run("bad_input", func(t *testing.T) {
	//	badType := make(chan int)
	//	defer close(badType)
	//	pub := &publisher{}
	//
	//	assert.EqualError(t, pub.Publish(&badType), "cannot marshal rabbitmq event: json: unsupported type: chan int")
	//})

	t.Run("connection_closed", func(t *testing.T) {
		f, reset := setupConnect(func(adapter *mockDialerAdapter, conn *mockConnection, channel *mockChannel) {
			channel.On("QueueDeclare", queueName, true, false, false, false,
				queueArgs).Return(amqp091.Queue{Name: queueName},
				nil)

			channel.On("Publish", "", queueName, true, false, amqp091.Publishing{
				Headers:     make(map[string]interface{}),
				ContentType: "application/json",
				Body:        []byte(`"hello"`),
			}).Return(nil)

			conn.On("Channel").Return(channel, nil)
			adapter.On("Dial", uri).Return(conn, nil)
		})
		defer reset()

		pub := &publisher{
			connection: connection{
				uri: uri,
				log: logger,
				err: make(chan error, 1),
			},
			queueName: queueName,
		}

		pub.err <- errors.New("dang, conn be broke")

		assert.NoError(t, pub.Publish(testMessage))

		f.adapter.AssertExpectations(t)
		f.connection.AssertExpectations(t)
		f.channel.AssertExpectations(t)
	})

	t.Run("connection_closed_retry", func(t *testing.T) {
		f, reset := setupConnect(func(adapter *mockDialerAdapter, conn *mockConnection, channel *mockChannel) {
			channel.On("QueueDeclare", queueName, true, false, false, false,
				queueArgs).Return(amqp091.Queue{Name: queueName},
				nil)

			channel.On("Publish", "", queueName, true, false, amqp091.Publishing{
				Headers:     make(map[string]interface{}),
				ContentType: "application/json",
				Body:        []byte(`"hello"`),
			}).Return(nil)

			conn.On("Channel").Return(channel, nil)

			adapter.On("Dial", uri).Return(nil, errors.New("test dial error")).Once()
			adapter.On("Dial", uri).Return(conn, nil).Once()
		})
		defer reset()

		pub := &publisher{
			connection: connection{
				uri: uri,
				log: logger,
				err: make(chan error, 1),
			},
			queueName: queueName,
		}

		pub.err <- errors.New("dang, conn be broke")

		assert.NoError(t, pub.Publish(testMessage))

		f.adapter.AssertExpectations(t)
		f.connection.AssertExpectations(t)
		f.channel.AssertExpectations(t)
	})

	t.Run("connection_closed_retry_failure", func(t *testing.T) {
		f, reset := setupConnect(func(adapter *mockDialerAdapter, conn *mockConnection, channel *mockChannel) {
			adapter.On("Dial", uri).Return(nil, errors.New("test dial error"))
		})
		defer reset()

		pub := &publisher{
			connection: connection{
				uri: uri,
				log: logger,
				err: make(chan error, 1),
			},
			queueName: queueName,
		}

		pub.err <- errors.New("dang, conn be broke")

		assert.Error(t, pub.Publish(testMessage))

		f.adapter.AssertExpectations(t)
	})
}

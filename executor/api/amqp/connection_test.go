package amqp

import (
	"errors"
	"github.com/golang/protobuf/proto"
	messages "github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"testing"
	"time"

	"github.com/go-logr/zapr"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	uri         = "amqp://test-rabbitmq:5672/"
	logger      = zapr.NewLogger(zap.L())
	queueName   = "test-queue"
	consumerTag = "tag"
)

type connectFixture struct {
	adapter    *mockDialerAdapter
	connection *mockConnection
	channel    *mockChannel
}

func setupConnect(fn func(adapter *mockDialerAdapter, conn *mockConnection, channel *mockChannel)) (*connectFixture, func()) {
	mockChan := &mockChannel{}
	mockConn := &mockConnection{}
	mockAdapter := &mockDialerAdapter{}

	fn(mockAdapter, mockConn, mockChan)

	origAdapter := defaultDialerAdapter
	origRetryDelay := connectionRetryDelay

	defaultDialerAdapter = mockAdapter.Dial
	connectionRetryDelay = 1 * time.Nanosecond

	fixture := &connectFixture{
		adapter:    mockAdapter,
		connection: mockConn,
		channel:    mockChan,
	}
	reset := func() {
		defaultDialerAdapter = origAdapter
		connectionRetryDelay = origRetryDelay
	}
	return fixture, reset
}

func TestNewConnection(t *testing.T) {
	t.Run("connect", func(t *testing.T) {
		f, reset := setupConnect(func(adapter *mockDialerAdapter, conn *mockConnection, channel *mockChannel) {
			conn.On("Channel").Return(channel, nil)
			adapter.On("Dial", uri).Return(conn, nil)
		})
		defer reset()

		actual, err := NewConnection(uri, logger)
		require.NoError(t, err)
		assert.NotNil(t, actual.conn)
		assert.NotNil(t, actual.channel)
		assert.Equal(t, uri, actual.uri)

		f.adapter.AssertExpectations(t)
		f.connection.AssertExpectations(t)
	})

	t.Run("reconnect", func(t *testing.T) {
		f, reset := setupConnect(func(adapter *mockDialerAdapter, conn *mockConnection, channel *mockChannel) {
			conn.On("Channel").Return(channel, nil)
			adapter.On("Dial", uri).Return(nil, errors.New("test dial error")).Once()
			adapter.On("Dial", uri).Return(conn, nil).Once()
		})
		defer reset()

		actual, err := NewConnection(uri, logger)
		require.NoError(t, err)
		assert.NotNil(t, actual.conn)
		assert.NotNil(t, actual.channel)
		assert.Equal(t, uri, actual.uri)

		f.adapter.AssertExpectations(t)
		f.adapter.AssertNumberOfCalls(t, "Dial", 2)
		f.connection.AssertExpectations(t)
	})

	t.Run("channel_failure", func(t *testing.T) {
		f, reset := setupConnect(func(adapter *mockDialerAdapter, conn *mockConnection, channel *mockChannel) {
			conn.On("Channel").Return(nil, errors.New("test channel failure"))
			adapter.On("Dial", uri).Return(conn, nil)
		})
		defer reset()

		_, err := NewConnection(uri, logger)
		assert.Error(t, err)

		f.adapter.AssertExpectations(t)
		f.connection.AssertExpectations(t)
	})

	t.Run("retry_limit_failure", func(t *testing.T) {
		f, reset := setupConnect(func(adapter *mockDialerAdapter, conn *mockConnection, channel *mockChannel) {
			adapter.On("Dial", uri).Return(nil, errors.New("test dial error"))
		})
		defer reset()

		_, err := NewConnection(uri, logger)
		assert.Error(t, err)

		f.adapter.AssertExpectations(t)
		f.adapter.AssertNumberOfCalls(t, "Dial", connectionRetryLimit)
	})
}

type TestPayload struct {
	Msg string
}

func (s *TestPayload) GetPayload() interface{} {
	return s.Msg
}

func (s *TestPayload) GetContentType() string {
	return rest.ContentTypeJSON
}

func (s *TestPayload) GetContentEncoding() string {
	return ""
}

func (s *TestPayload) GetBytes() ([]byte, error) {
	return []byte(s.Msg), nil
}

func TestConnection_Consume(t *testing.T) {
	testDelivery := amqp.Delivery{
		Body:            []byte(`"hello"`),
		ContentType:     rest.ContentTypeJSON,
		ContentEncoding: "",
	}
	seldonMessage := messages.SeldonMessage{
		Status: &messages.Status{
			Status: messages.Status_SUCCESS,
		},
		Meta: nil,
		DataOneof: &messages.SeldonMessage_StrData{
			StrData: `"hello"`,
		},
	}
	seldonMessageEnc, _ := proto.Marshal(&seldonMessage)
	seldonMessage.XXX_sizecache = 0 // to make test cases match
	testDelivery2 := amqp.Delivery{
		Body:            seldonMessageEnc,
		ContentType:     payload.APPLICATION_TYPE_PROTOBUF,
		ContentEncoding: "",
	}

	t.Run("success", func(t *testing.T) {
		mockChan := &mockChannel{}

		mockChan.On("QueueDeclare", queueName, true, false, false, false, amqp.Table{
			"x-single-active-consumer": true,
		}).Return(amqp.Queue{Name: queueName}, nil)

		mockDeliveries := make(chan amqp.Delivery, 1) // buffer 1 so that we send returns before starting consumer
		mockDeliveries <- testDelivery
		close(mockDeliveries)

		mockChan.On("Consume", queueName, consumerTag, false, false, false, false, amqp.Table{}).Return(mockDeliveries, nil)

		cons := &consumer{
			connection: connection{
				log:     logger,
				channel: mockChan,
			},
			queueName:   queueName,
			consumerTag: consumerTag,
		}

		payloadHandler := func(pl SeldonPayloadWithHeaders) error {
			assert.Equal(
				t,
				SeldonPayloadWithHeaders{
					&payload.BytesPayload{
						Msg:             []byte(`"hello"`),
						ContentType:     rest.ContentTypeJSON,
						ContentEncoding: "",
					},
					make(map[string][]string),
				},
				pl,
				"payloads not equal",
			)
			return nil
		}

		errorHandler := func(err error) {
			assert.NoError(t, err, "unexpected error")
		}

		err := cons.Consume(payloadHandler, errorHandler)

		assert.NoError(t, err)
	})

	t.Run("encoded seldon msg", func(t *testing.T) {
		mockChan := &mockChannel{}

		mockChan.On("QueueDeclare", queueName, true, false, false, false, amqp.Table{
			"x-single-active-consumer": true,
		}).Return(amqp.Queue{Name: queueName}, nil)

		mockDeliveries := make(chan amqp.Delivery, 1) // buffer 1 so that we send returns before starting consumer
		mockDeliveries <- testDelivery2
		close(mockDeliveries)

		mockChan.On("Consume", queueName, consumerTag, false, false, false, false, amqp.Table{}).Return(mockDeliveries, nil)

		cons := &consumer{
			connection: connection{
				log:     logger,
				channel: mockChan,
			},
			queueName:   queueName,
			consumerTag: consumerTag,
		}

		payloadHandler := func(pl SeldonPayloadWithHeaders) error {
			assert.Equal(
				t,
				SeldonPayloadWithHeaders{
					&payload.ProtoPayload{
						Msg: &seldonMessage,
					},
					make(map[string][]string),
				},
				pl,
				"payloads not equal",
			)
			return nil
		}

		errorHandler := func(err error) {
			assert.NoError(t, err, "unexpected error")
		}

		err := cons.Consume(payloadHandler, errorHandler)

		assert.NoError(t, err)
	})

}

func TestConnection_Publish(t *testing.T) {
	testMessage := SeldonPayloadWithHeaders{
		&TestPayload{Msg: `"hello"`},
		make(map[string][]string),
	}

	t.Run("success", func(t *testing.T) {
		mockChan := &mockChannel{}

		mockChan.On("QueueDeclare", queueName, true, false, false, false, amqp.Table{
			"x-single-active-consumer": true,
		}).Return(amqp.Queue{Name: queueName}, nil)

		mockChan.On("Publish", "", queueName, true, false, amqp.Publishing{
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

		mockChan.On("QueueDeclare", queueName, true, false, false, false, amqp.Table{
			"x-single-active-consumer": true,
		}).Return(amqp.Queue{Name: queueName}, nil)

		mockChan.On("Publish", "", queueName, true, false, amqp.Publishing{
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

		mockChan.On("QueueDeclare", queueName, true, false, false, false, amqp.Table{
			"x-single-active-consumer": true,
		}).Return(amqp.Queue{}, errors.New("test error"))

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
			channel.On("QueueDeclare", queueName, true, false, false, false, amqp.Table{
				"x-single-active-consumer": true,
			}).Return(amqp.Queue{Name: queueName}, nil)

			channel.On("Publish", "", queueName, true, false, amqp.Publishing{
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
			channel.On("QueueDeclare", queueName, true, false, false, false, amqp.Table{
				"x-single-active-consumer": true,
			}).Return(amqp.Queue{Name: queueName}, nil)

			channel.On("Publish", "", queueName, true, false, amqp.Publishing{
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

func TestConnection_Close(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockConn := &mockConnection{}
		mockConn.On("Close").Return(nil)
		con := &connection{
			conn: mockConn,
		}

		assert.NoError(t, con.Close())
		mockConn.AssertExpectations(t)
	})

	t.Run("failure", func(t *testing.T) {
		mockConn := &mockConnection{}
		mockConn.On("Close").Return(errors.New("test failed to close connection"))
		con := &connection{
			conn: mockConn,
		}

		assert.EqualError(t, con.Close(), "test failed to close connection")
		mockConn.AssertExpectations(t)
	})

	t.Run("no_connection", func(t *testing.T) {
		assert.NoError(t, (&connection{}).Close())
	})
}

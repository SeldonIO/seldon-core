package rabbitmq

import (
	"errors"
	"testing"
	"time"

	"github.com/go-logr/zapr"
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

func setupConnect(fn func(adapter *mockDialerAdapter, conn *mockConnection, channel *mockChannel)) (
	*connectFixture, func(),
) {
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

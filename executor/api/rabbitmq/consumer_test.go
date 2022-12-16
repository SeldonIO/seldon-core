package rabbitmq

import (
	"errors"
	"github.com/go-logr/logr/testr"
	proto2 "github.com/golang/protobuf/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/rest"
	"github.com/stretchr/testify/assert"
	"testing"
)

/*
 * based on patterns from https://github.com/dominodatalab/forge/blob/master/internal/message/amqp/publisher_test.go
 */

func TestConsume(t *testing.T) {
	log := testr.New(t)
	testMessageStr := `"hello"`
	seldonMessage := proto.SeldonMessage{
		Status: &proto.Status{
			Status: proto.Status_SUCCESS,
		},
		Meta: nil,
		DataOneof: &proto.SeldonMessage_StrData{
			StrData: testMessageStr,
		},
	}
	seldonMessageEnc, _ := proto2.Marshal(&seldonMessage)
	seldonMessage.XXX_sizecache = 0 // to make test cases match

	t.Run("success", func(t *testing.T) {
		mockChan := &mockChannel{}

		mockDeliveries := make(chan amqp.Delivery, 1) // buffer 1 so that we send returns before starting consumer
		mockDeliveries <- createTestDelivery(mockChan, []byte(testMessageStr), rest.ContentTypeJSON)
		close(mockDeliveries)

		mockChan.On("Consume", queueName, consumerTag, false, false, false, false,
			amqp.Table{}).Return(mockDeliveries,
			nil)
		mockChan.On("Ack", uint64(0), false).Return(nil)

		cons := &consumer{
			connection: connection{
				log:     log,
				channel: mockChan,
			},
			queueName:   queueName,
			consumerTag: consumerTag,
		}

		payloadHandler := func(pl *SeldonPayloadWithHeaders) error {
			assert.Equal(
				t,
				&SeldonPayloadWithHeaders{
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

		errorHandler := func(err ConsumerError) error {
			assert.NoError(t, err.err, "unexpected error")
			return nil
		}

		err := cons.Consume(payloadHandler, errorHandler)

		assert.NoError(t, err)
		mockChan.AssertExpectations(t)
	})

	t.Run("on failed message don't retry, but reject method", func(t *testing.T) {
		mockChan := &mockChannel{}

		mockDeliveries := make(chan amqp.Delivery, 1) // buffer 1 so that we send returns before starting consumer
		mockDeliveries <- createTestDelivery(mockChan, []byte(testMessageStr), rest.ContentTypeJSON)
		close(mockDeliveries)

		mockChan.On("Consume", queueName, consumerTag, false, false, false, false,
			amqp.Table{}).Return(mockDeliveries,
			nil)
		mockChan.On("Reject", uint64(0), false).Return(nil)

		cons := &consumer{
			connection: connection{
				log:     log,
				channel: mockChan,
			},
			queueName:   queueName,
			consumerTag: consumerTag,
		}

		payloadHandler := func(pl *SeldonPayloadWithHeaders) error {
			return errors.New("Something bad happened during prediction")
		}

		errorHandler := func(err ConsumerError) error {
			return nil
		}

		err := cons.Consume(payloadHandler, errorHandler)

		assert.NoError(t, err)
		mockChan.AssertExpectations(t)
	})

	t.Run("encoded seldon msg", func(t *testing.T) {
		mockChan := &mockChannel{}

		mockDeliveries := make(chan amqp.Delivery, 1) // buffer 1 so that we send returns before starting consumer
		mockDeliveries <- createTestDelivery(mockChan, seldonMessageEnc, payload.APPLICATION_TYPE_PROTOBUF)
		close(mockDeliveries)

		mockChan.On("Consume", queueName, consumerTag, false, false, false, false,
			amqp.Table{}).Return(mockDeliveries,
			nil)
		mockChan.On("Ack", uint64(0), false).Return(nil)

		cons := &consumer{
			connection: connection{
				log:     log,
				channel: mockChan,
			},
			queueName:   queueName,
			consumerTag: consumerTag,
		}

		payloadHandler := func(pl *SeldonPayloadWithHeaders) error {
			assert.Equal(
				t,
				&SeldonPayloadWithHeaders{
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

		errorHandler := func(err ConsumerError) error {
			assert.NoError(t, err.err, "unexpected error")
			return nil
		}

		err := cons.Consume(payloadHandler, errorHandler)

		assert.NoError(t, err)
		mockChan.AssertExpectations(t)
	})
}

func createTestDelivery(ack amqp.Acknowledger, body []byte, contentType string) amqp.Delivery {
	return amqp.Delivery{
		Acknowledger:    ack,
		Body:            body,
		ContentType:     contentType,
		ContentEncoding: "",
	}
}

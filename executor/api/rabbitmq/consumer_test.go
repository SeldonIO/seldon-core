package rabbitmq

import (
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
	testDelivery := amqp.Delivery{
		Body:            []byte(`"hello"`),
		ContentType:     rest.ContentTypeJSON,
		ContentEncoding: "",
	}
	seldonMessage := proto.SeldonMessage{
		Status: &proto.Status{
			Status: proto.Status_SUCCESS,
		},
		Meta: nil,
		DataOneof: &proto.SeldonMessage_StrData{
			StrData: `"hello"`,
		},
	}
	seldonMessageEnc, _ := proto2.Marshal(&seldonMessage)
	seldonMessage.XXX_sizecache = 0 // to make test cases match
	testDelivery2 := amqp.Delivery{
		Body:            seldonMessageEnc,
		ContentType:     payload.APPLICATION_TYPE_PROTOBUF,
		ContentEncoding: "",
	}

	t.Run("success", func(t *testing.T) {
		mockChan := &mockChannel{}

		mockChan.On("QueueDeclare", queueName, true, false, false, false,
			queueArgs).Return(amqp.Queue{Name: queueName},
			nil)

		mockDeliveries := make(chan amqp.Delivery, 1) // buffer 1 so that we send returns before starting consumer
		mockDeliveries <- testDelivery
		close(mockDeliveries)

		mockChan.On("Consume", queueName, consumerTag, false, false, false, false,
			amqp.Table{}).Return(mockDeliveries,
			nil)

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

		mockChan.On("QueueDeclare", queueName, true, false, false, false,
			queueArgs).Return(amqp.Queue{Name: queueName},
			nil)

		mockDeliveries := make(chan amqp.Delivery, 1) // buffer 1 so that we send returns before starting consumer
		mockDeliveries <- testDelivery2
		close(mockDeliveries)

		mockChan.On("Consume", queueName, consumerTag, false, false, false, false,
			amqp.Table{}).Return(mockDeliveries,
			nil)

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

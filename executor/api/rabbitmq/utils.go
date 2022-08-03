package rabbitmq

import (
	"fmt"
	proto2 "github.com/golang/protobuf/proto"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/seldonio/seldon-core/executor/api/grpc/seldon/proto"
	"github.com/seldonio/seldon-core/executor/api/payload"
	"github.com/seldonio/seldon-core/executor/api/rest"
)

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
		var message = &proto.SeldonMessage{}
		err = proto2.Unmarshal(delivery.Body, message)
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
		err = fmt.Errorf("unknown payload type '%s'", delivery.ContentType)
	}

	return pl, err
}

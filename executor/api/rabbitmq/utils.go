package rabbitmq

import (
	"fmt"
	"github.com/golang/protobuf/jsonpb"
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

func DeliveryToPayload(delivery amqp.Delivery) (*SeldonPayloadWithHeaders, error) {
	var pl *SeldonPayloadWithHeaders = nil
	var err error = nil

	headers := TableToStringMap(delivery.Headers)

	switch delivery.ContentType {
	case payload.APPLICATION_TYPE_PROTOBUF:
		var message = &proto.SeldonMessage{}
		err = proto2.Unmarshal(delivery.Body, message)
		if err == nil {
			pl = &SeldonPayloadWithHeaders{
				&payload.ProtoPayload{Msg: message},
				headers,
			}
		}
	case rest.ContentTypeJSON:
		pl = &SeldonPayloadWithHeaders{
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

func UpdatePayloadWithPuid(reqPayload payload.SeldonPayload, oldPayload payload.SeldonPayload) (payload.SeldonPayload, error) {
	var msgBytes []byte
	switch reqPayload.GetContentType() {
	case payload.APPLICATION_TYPE_PROTOBUF:
		requestBody := &proto.SeldonMessage{}
		err := proto2.Unmarshal(reqPayload.GetPayload().([]byte), requestBody)
		if err != nil {
			return nil, err
		}

		if requestBody.Meta == nil {
			return oldPayload, nil
		}

		body := &proto.SeldonMessage{}
		err = proto2.Unmarshal(oldPayload.GetPayload().([]byte), body)
		if err != nil {
			return nil, err
		}

		if body.Meta == nil {
			body.Meta = &proto.Meta{Puid: requestBody.Meta.Puid}
		}

		msg, err := proto2.Marshal(body)
		if err != nil {
			return nil, err
		}
		msgBytes = msg
	case rest.ContentTypeJSON:
		requestBody := &proto.SeldonMessage{}
		err := jsonpb.UnmarshalString(string(reqPayload.GetPayload().([]byte)), requestBody)
		if err != nil {
			return nil, err
		}

		if requestBody.Meta == nil {
			return oldPayload, nil
		}

		body := &proto.SeldonMessage{}
		jsonpb.UnmarshalString(string(oldPayload.GetPayload().([]byte)), body)
		if err != nil {
			return nil, err
		}

		if body.Meta == nil {
			body.Meta = &proto.Meta{Puid: requestBody.Meta.Puid}
		}

		msg, err := new(jsonpb.Marshaler).MarshalToString(body)
		if err != nil {
			return nil, err
		}
		msgBytes = []byte(msg)
	default:
		err := fmt.Errorf("unknown payload type '%s'", oldPayload.GetContentType())
		return nil, err
	}

	updatedPayload := &payload.BytesPayload{
		Msg:             msgBytes,
		ContentType:     oldPayload.GetContentType(),
		ContentEncoding: oldPayload.GetContentEncoding()}
	return updatedPayload, nil
}

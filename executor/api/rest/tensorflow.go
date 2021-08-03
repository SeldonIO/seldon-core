package rest

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/seldonio/seldon-core/executor/api/payload"
)

const (
	ModelHttpPathVariable = "model"
)

func ChainTensorflow(msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	var f interface{}
	err := json.Unmarshal(msg.GetPayload().([]byte), &f)
	if err != nil {
		return nil, err
	}
	m := f.(map[string]interface{})
	if _, ok := m["instances"]; ok {
		return msg, nil
	} else if _, ok := m["inputs"]; ok {
		return msg, nil
	} else if _, ok := m["predictions"]; ok {
		m["instances"] = m["predictions"]
		delete(m, "predictions")
		b, err := json.Marshal(m)
		if err != nil {
			return nil, err
		} else {
			p := payload.BytesPayload{Msg: b, ContentType: msg.GetContentType()}
			return &p, nil
		}
	} else {
		return nil, errors.Errorf("Failed to convert tensorflow response so it could be chained to new input")
	}
}

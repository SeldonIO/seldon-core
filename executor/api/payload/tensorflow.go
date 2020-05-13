package payload

import (
	"encoding/json"
	"github.com/pkg/errors"
)

const (
	ModelHttpPathVariable = "model"
)

func ChainTensorflow(msg SeldonPayload) (SeldonPayload, error) {
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
			p := BytesPayload{Msg: b}
			return &p, nil
		}
	} else {
		return nil, errors.Errorf("Failed to convert tensorflow response so it could be chained to new input")
	}
}

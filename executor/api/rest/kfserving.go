package rest

import (
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/seldonio/seldon-core/executor/api/payload"
)

func ChainKFserving(msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	data, err := payload.DecompressSeldonPayload(msg)
	if err != nil {
		return nil, err
	}

	var f interface{}
	err = json.Unmarshal(data, &f)
	if err != nil {
		return nil, err
	}
	m := f.(map[string]interface{})
	if _, ok := m["inputs"]; ok {
		return msg, nil
	} else if _, ok := m["outputs"]; ok {
		m["inputs"] = m["outputs"]
		delete(m, "outputs")
		b, err := json.Marshal(m)
		if err != nil {
			return nil, err
		}
		p := payload.BytesPayload{Msg: b, ContentType: msg.GetContentType()}
		return &p, nil
	} else {
		return nil, errors.Errorf("Failed to convert kfserving response so it could be chained to new input")
	}
}

package rest

import (
	"compress/gzip"
	"encoding/json"
	"io/ioutil"

	"github.com/pkg/errors"
	"github.com/seldonio/seldon-core/executor/api/payload"

	"bytes"
)

func ChainKFserving(msg payload.SeldonPayload) (payload.SeldonPayload, error) {
	data, err := msg.GetBytes()
	if err != nil {
		return nil, err
	}

	if msg.GetContentEncoding() == "gzip" {
		bytesReader := bytes.NewReader(data)
		gzipReader, err := gzip.NewReader(bytesReader)
		if err != nil {
			return nil, err
		}
		output, err := ioutil.ReadAll(gzipReader)
		if err != nil {
			return nil, err
		}
		data = output
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

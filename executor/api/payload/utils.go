package payload

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

// Decompress payloads if Content-Encoding is set
func DecompressSeldonPayload(msg SeldonPayload) ([]byte, error) {
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
		return output, nil
	} else {
		return data, nil
	}
}

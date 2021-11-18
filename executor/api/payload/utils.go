package payload

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

func DecompressBytes(data []byte, contentEncoding string) ([]byte, error) {
	if contentEncoding != "gzip" {
		return data, nil
	}

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
}

// Decompress payloads if Content-Encoding is set to gzip
func DecompressSeldonPayload(msg SeldonPayload) ([]byte, error) {
	data, err := msg.GetBytes()
	if err != nil {
		return nil, err
	}

	return DecompressBytes(data, msg.GetContentEncoding())
}

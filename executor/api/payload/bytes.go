package payload

type BytesPayload struct {
	Msg             []byte
	ContentType     string
	ContentEncoding string
}

func (s *BytesPayload) GetPayload() interface{} {
	return s.Msg
}

func (s *BytesPayload) GetContentType() string {
	return s.ContentType
}

func (s *BytesPayload) GetContentEncoding() string {
	return s.ContentEncoding
}

func (s *BytesPayload) GetBytes() ([]byte, error) {
	return s.Msg, nil
}

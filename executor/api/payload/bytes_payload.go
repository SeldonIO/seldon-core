package payload

type BytesPayload struct {
	Msg         []byte
	ContentType string
}

func (s *BytesPayload) GetPayload() interface{} {
	return s.Msg
}

func (s *BytesPayload) GetContentType() string {
	return s.ContentType
}

func (s *BytesPayload) GetBytes() ([]byte, error) {
	return s.Msg, nil
}

func (s *BytesPayload) SetPayload(payload interface{}) {
	s.Msg = payload.([]byte)
}

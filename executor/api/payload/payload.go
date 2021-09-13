package payload

type SeldonPayload interface {
	GetPayload() interface{}
	GetContentType() string
	GetContentEncoding() string
	GetBytes() ([]byte, error)
}

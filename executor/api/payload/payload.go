package payload

type SeldonPayload interface {
	GetPayload() interface{}
	GetContentType() string
}

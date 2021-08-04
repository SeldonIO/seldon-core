package payload

const (
	SeldonPUIDHeader = "Seldon-Puid"
)

type MetaData struct {
	Meta map[string][]string
}

func NewFromMap(m map[string][]string) *MetaData {
	meta := MetaData{
		Meta: map[string][]string{},
	}
	for k, vv := range m {
		for _, v := range vv {
			meta.Meta[k] = append(meta.Meta[k], v)
		}
	}
	return &meta
}

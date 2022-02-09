package payload

import (
	"strings"
)

const (
	SeldonPUIDHeader        = "Seldon-Puid"
	SeldonSkipLoggingHeader = "Seldon-Skip-Logging"
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

func (m *MetaData) GetAsBoolean(key string, def bool) bool {
	values, ok := m.Meta[key]
	if !ok {
		return def
	}

	trueValues := []string{"true", "on", "1"}
	for _, val := range values {
		lowVal := strings.ToLower(val)
		for _, trueVal := range trueValues {
			if lowVal == trueVal {
				// If any element of the list matches, we'll return true
				return true
			}
		}
	}

	return false
}

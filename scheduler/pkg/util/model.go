package util

import (
	"fmt"
)

func GetVersionedModelName(model string, version uint32) string {
	return fmt.Sprintf("%s_%d", model, version)
}

func GetPinnedModelVersion() uint32 {
	return 1
}

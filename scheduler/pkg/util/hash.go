package util

import (
	"fmt"
	"hash/fnv"

	"github.com/OneOfOne/xxhash"
)

func Hash(s string) (uint32, error) {
	h := fnv.New32a()
	_, err := h.Write([]byte(s))
	if err != nil {
		return 0, err
	}
	return h.Sum32(), nil
}

func XXHash(key string) string {
	h := xxhash.New32()
	return fmt.Sprintf("%x", h.Sum([]byte(key)))
}

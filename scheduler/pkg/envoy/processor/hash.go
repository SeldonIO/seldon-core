package processor

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"sort"
	"strconv"
)

func computeHashKeyForList(list []int) string {
	var buffer bytes.Buffer
	sort.Ints(list)
	for _, v := range list {
		buffer.WriteString(
			strconv.Itoa(v))
		buffer.WriteString(",")
	}
	h := sha256.New()
	h.Write([]byte(buffer.String()))
	b := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(b)
}




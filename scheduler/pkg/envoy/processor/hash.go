/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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
	_, _ = h.Write(buffer.Bytes())
	b := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(b)
}

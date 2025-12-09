/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package steps

import (
	"crypto/rand"
	"encoding/base32"
	"strings"
)

// RandomSuffix returns a lowercase, short, k8s-safe random string.
func randomSuffix(n int) string {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err) // or return "" and handle upstream
	}

	// base32 gives A–Z2–7, so we lowercase and trim "=" padding.
	return strings.ToLower(strings.TrimRight(base32.StdEncoding.EncodeToString(b), "="))
}

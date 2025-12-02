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

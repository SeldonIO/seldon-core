package util

import (
	"fmt"
	"os"
	"strings"
)

func getEnvVarKey(prefix string, suffix string) string {
	return fmt.Sprintf("%s%s", prefix, suffix)
}

func GetEnv(prefix string, suffix string) (string, bool) {
	val, ok := os.LookupEnv(getEnvVarKey(prefix, suffix))
	if ok {
		ok = strings.TrimSpace(val) != ""
	}
	return val, ok
}

package cli

import (
	"os"
	"strconv"
)

func getEnvString(name string) (string, bool) {
	fromEnv := os.Getenv(name)
	if fromEnv == "" {
		return "", false
	}
	return fromEnv, true
}

func getEnvBool(name string) (bool, bool) {
	fromEnv := os.Getenv(name)
	if fromEnv == "" {
		return false, false
	}

	val, err := strconv.ParseBool(fromEnv)
	if err != nil {
		return false, false
	}

	return val, true
}

func getEnvUint(name string) (uint, bool) {
	fromEnv := os.Getenv(name)
	if fromEnv == "" {
		return 0, false
	}

	val, err := strconv.ParseUint(fromEnv, 10, 64)
	if err != nil {
		return 0, false
	}

	return uint(val), true
}

func getEnvInt(name string) (int, bool) {
	fromEnv := os.Getenv(name)
	if fromEnv == "" {
		return 0, false
	}

	val, err := strconv.ParseInt(fromEnv, 10, 64)
	if err != nil {
		return 0, false
	}

	return int(val), true
}

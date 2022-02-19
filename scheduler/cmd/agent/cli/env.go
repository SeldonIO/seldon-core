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

// returns value, found, parsed
func getEnvBool(name string) (bool, bool, bool) {
	fromEnv := os.Getenv(name)
	if fromEnv == "" {
		return false, false, false
	}

	val, err := strconv.ParseBool(fromEnv)
	if err != nil {
		return false, true, false
	}

	return val, true, true
}

// returns value, found, parsed
func getEnvUint(name string) (uint, bool, bool) {
	fromEnv := os.Getenv(name)
	if fromEnv == "" {
		return 0, false, false
	}

	val, err := strconv.ParseUint(fromEnv, 10, 64)
	if err != nil {
		return 0, true, false
	}

	return uint(val), true, true
}

// returns value, found, parsed
func getEnvInt(name string) (int, bool, bool) {
	fromEnv := os.Getenv(name)
	if fromEnv == "" {
		return 0, false, false
	}

	val, err := strconv.ParseInt(fromEnv, 10, 64)
	if err != nil {
		return 0, true, false
	}

	return int(val), true, true
}

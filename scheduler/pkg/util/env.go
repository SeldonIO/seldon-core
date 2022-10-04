package util

import (
	"os"
	"strconv"
)

func GetIntEnvar(key string, defaultValue int) (int, error) {
	valStr := os.Getenv(key)
	if valStr != "" {
		val, err := strconv.ParseInt(valStr, 10, 64)
		if err != nil {
			return 0, err
		}
		return int(val), nil
	}
	return defaultValue, nil
}

func GetBoolEnvar(key string, defaultValue bool) (bool, error) {
	valStr := os.Getenv(key)
	if valStr != "" {
		val, err := strconv.ParseBool(valStr)
		if err != nil {
			return false, err
		}
		return val, nil
	}
	return defaultValue, nil
}

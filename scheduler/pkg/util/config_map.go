package util

import (
	"strconv"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func GetIntConfigMapValue(configMap kafka.ConfigMap, key string, defaultValue int) (int, error) {
	configMapValue, ok := configMap[key]
	if !ok {
		return defaultValue, nil
	}

	value, err := strconv.Atoi(configMapValue.(string))
	if err != nil {
		return 0, err
	}

	return value, nil
}

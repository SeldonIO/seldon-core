/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

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

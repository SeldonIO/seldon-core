/*
Copyright 2022 Seldon Technologies Ltd.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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

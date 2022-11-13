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

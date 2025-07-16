/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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

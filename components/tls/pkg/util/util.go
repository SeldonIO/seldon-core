/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package util

import (
	"fmt"
	"os"
	"strings"
)

func getEnvVarKey(prefix string, suffix string) string {
	return fmt.Sprintf("%s%s", prefix, suffix)
}

// Return an environment value if the key is defined and the value is non-empty/non-whitespace;
// otherwise indicate no valid value is available through the Boolean return.
func GetNonEmptyEnv(prefix string, suffix string) (string, bool) {
	val, ok := os.LookupEnv(getEnvVarKey(prefix, suffix))
	if !ok || strings.TrimSpace(val) == "" {
		return "", false
	}

	return val, true
}

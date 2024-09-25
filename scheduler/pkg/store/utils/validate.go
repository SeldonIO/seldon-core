/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed BY
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package utils

import (
	"regexp"
	"strings"
)

func CheckName(name string) bool {
	ok, err := regexp.MatchString("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", name)
	if !ok || err != nil || strings.Contains(name, ".") {
		return false
	}
	return true
}

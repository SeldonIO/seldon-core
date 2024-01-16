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
	"strconv"
	"strings"
)

const (
	separator = "_"
)

func GetVersionedModelName(model string, version uint32) string {
	return fmt.Sprintf("%s%s%d", model, separator, version)
}

func GetOrignalModelNameAndVersion(versionedModel string) (string, uint32, error) {
	seperatorIndex := strings.LastIndex(versionedModel, separator)
	versionStr := versionedModel[seperatorIndex+1:]
	version, err := strconv.Atoi(versionStr)
	if err != nil {
		return "", 0, fmt.Errorf("cannot convert to original model")
	}
	return versionedModel[0:seperatorIndex], uint32(version), nil
}

func GetPinnedModelVersion() uint32 {
	return 1
}

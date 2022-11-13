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

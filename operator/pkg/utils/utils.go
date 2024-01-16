/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package utils

import (
	"fmt"
)

func ContainsStr(slice []string, value string) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func RemoveStr(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func GetStatefulSetReplicaName(name string, replicaIdx int) string {
	return fmt.Sprintf("%s-%d", name, replicaIdx)
}

func MergeMaps(child map[string]string, parent map[string]string) map[string]string {
	merged := make(map[string]string)
	for k, v := range child {
		merged[k] = v
	}
	for k, v := range parent {
		if _, ok := child[k]; !ok {
			merged[k] = v
		}

	}
	// remove any keys we want to ignore
	delete(merged, "meta.helm.sh/release-name")
	delete(merged, "meta.helm.sh/release-namespace")
	delete(merged, "kubectl.kubernetes.io/last-applied-configuration")
	return merged
}

func HasMappings(expected map[string]string, found map[string]string) bool {
	for k, v := range expected {
		if v2, ok := found[k]; ok {
			if v != v2 {
				return false
			}
		} else {
			return false
		}
	}
	return true
}

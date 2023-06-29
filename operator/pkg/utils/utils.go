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

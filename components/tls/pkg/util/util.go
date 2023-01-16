/*
Copyright 2023 Seldon Technologies Ltd.

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
	"os"
	"strings"
)

func getEnvVarKey(prefix string, suffix string) string {
	return fmt.Sprintf("%s%s", prefix, suffix)
}

func GetEnv(prefix string, suffix string) (string, bool) {
	val, ok := os.LookupEnv(getEnvVarKey(prefix, suffix))
	if ok {
		ok = strings.TrimSpace(val) != ""
	}
	return val, ok
}

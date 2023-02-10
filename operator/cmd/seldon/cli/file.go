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
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

func loadFile(filename string) []byte {
	dat, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	return dat
}

// Utility method to extract file bytes or name of resource
func extractFileOrName(flags *pflag.FlagSet, args []string) ([]byte, string, error) {
	filename, err := flags.GetString(flagFile)
	if err != nil {
		return nil, "", err
	}
	var fileBytes []byte
	name := ""
	if filename != "" {
		fileBytes = loadFile(filename)
	} else {
		if len(args) > 0 {
			name = args[0]
		}
	}
	return fileBytes, name, nil
}

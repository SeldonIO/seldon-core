/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
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

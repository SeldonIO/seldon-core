/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package main

import (
	"os"

	"github.com/seldonio/seldon-core/operator/v2/cmd/seldon/cli"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {

	return cli.GetCmd().Execute()
}

func main() {
	if err := Execute(); err != nil {
		os.Exit(-1)
	}
}

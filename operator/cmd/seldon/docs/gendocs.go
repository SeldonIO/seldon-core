/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package main

import (
	"flag"
	"log"

	"github.com/spf13/cobra/doc"

	"github.com/seldonio/seldon-core/operator/v2/cmd/seldon/cli"
)

func main() {
	var folder string
	flag.StringVar(&folder, "out", "", "folder to create docs")
	flag.Parse()

	cmd := cli.GetCmd()
	err := doc.GenMarkdownTree(cmd, folder)
	if err != nil {
		log.Fatal(err)
	}
}

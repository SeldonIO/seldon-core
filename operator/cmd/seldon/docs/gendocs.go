package main

import (
	"flag"
	"log"

	"github.com/seldonio/seldon-core/operatorv2/cmd/seldon/cli"
	"github.com/spf13/cobra/doc"
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

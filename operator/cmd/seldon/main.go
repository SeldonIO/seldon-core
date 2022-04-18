package main

import (
	"fmt"
	"os"

	"github.com/seldonio/seldon-core/operatorv2/cmd/seldon/cli"
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {

	return cli.GetCmd().Execute()
}

func main() {
	if err := Execute(); err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
}

package cli

import (
	"fmt"
	"os"
)

func loadFile(filename string) []byte {
	dat, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(-1)
	}
	return dat
}

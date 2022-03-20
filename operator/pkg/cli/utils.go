package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func printPrettyJson(data []byte) {
	prettyJson, err := prettyJson(data)
	if err == nil {
		fmt.Printf("%s\n", prettyJson)
	}
}

func prettyJson(data []byte) (string, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, data, "", "\t")
	if err != nil {
		return "", err
	}
	return prettyJSON.String(), nil
}

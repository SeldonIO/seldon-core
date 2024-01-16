/*
Copyright (c) 2024 Seldon Technologies Ltd.

Use of this software is governed by
(1) the license included in the LICENSE file or
(2) if the license included in the LICENSE file is the Business Source License 1.1,
the Change License after the Change Date as each is defined in accordance with the LICENSE file.
*/

package cli

import (
	"bytes"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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

func PrintProto(msg proto.Message) {
	resJson, err := protojson.Marshal(msg)
	if err != nil {
		fmt.Printf("Failed to print proto: %s", err.Error())
	} else {
		fmt.Printf("%s\n", string(resJson))
	}
}

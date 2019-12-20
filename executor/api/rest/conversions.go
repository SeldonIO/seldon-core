package rest

import (
	"encoding/json"
)

// Assumes the byte array is a json list of ints
func ExtractRouteAsJsonArray(msg []byte) ([]int, error) {
	var routes []int
	err := json.Unmarshal(msg, &routes)
	if err == nil {
		return routes, err
	} else {
		return nil, err
	}
}

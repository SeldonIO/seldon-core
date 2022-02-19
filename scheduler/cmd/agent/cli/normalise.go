package cli

import "strings"

func trimStrings(elems []string) []string {
	trimmed := make([]string, len(elems))
	for idx, e := range elems {
		trimmed[idx] = strings.TrimSpace(e)
	}
	return trimmed
}

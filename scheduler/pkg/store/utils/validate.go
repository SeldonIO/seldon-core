package utils

import (
	"regexp"
	"strings"
)

func CheckName(name string) bool {
	ok, err := regexp.MatchString("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", name)
	if !ok || err != nil || strings.Contains(name, ".") {
		return false
	}
	return true
}

package utils

import "regexp"

func CheckName(name string) bool {
	ok, err := regexp.Match("^[a-zA-Z0-9][a-zA-Z0-9-_]*[a-zA-Z0-9]$", []byte(name))
	if !ok || err != nil {
		return false
	}
	return true
}

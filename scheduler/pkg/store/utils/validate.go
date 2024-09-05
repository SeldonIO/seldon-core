package utils

import "regexp"

func CheckName(name string) bool {
	ok, err := regexp.Match("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$", []byte(name))
	if !ok || err != nil {
		return false
	}
	return true
}

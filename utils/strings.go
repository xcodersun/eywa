package utils

import (
	"regexp"
)

func StringSliceContains(s []string, v string) bool {
	for _, ss := range s {
		if ss == v {
			return true
		}
	}
	return false
}

func AlphaNumeric(s string) bool {
	match, _ := regexp.MatchString("^[a-z0-9_]+$", s)
	return match
}

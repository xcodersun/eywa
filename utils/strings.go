package utils

import (
	"regexp"
)

func AlphaNumeric(s string) bool {
	match, _ := regexp.MatchString("^[a-z0-9_]+$", s)
	return match
}

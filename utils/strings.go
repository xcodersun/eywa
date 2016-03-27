package utils

import (
	"regexp"
)

func AlphaNumeric(s string) bool {
	match, _ := regexp.MatchString("^[a-z0-9_A-Z]+$", s)
	return match
}

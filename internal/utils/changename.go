package utils

import "regexp"

var kebabCaseRegex = regexp.MustCompile(`^[a-z][a-z0-9]*(-[a-z0-9]+)*$`)

func ValidateChangeName(name string) bool {
	return kebabCaseRegex.MatchString(name)
}

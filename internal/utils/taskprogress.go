package utils

import (
	"regexp"
	"strings"
)

var (
	uncheckedRegex = regexp.MustCompile(`^\s*-\s*\[\s*\]`)
	checkedRegex   = regexp.MustCompile(`^\s*-\s*\[x\]`)
)

type TaskProgress struct {
	Total     int
	Completed int
}

func CountTasks(content string) TaskProgress {
	var progress TaskProgress
	for _, line := range strings.Split(content, "\n") {
		if uncheckedRegex.MatchString(line) {
			progress.Total++
		} else if checkedRegex.MatchString(strings.ToLower(line)) {
			progress.Total++
			progress.Completed++
		}
	}
	return progress
}

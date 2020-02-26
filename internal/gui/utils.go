package gui

import (
	"regexp"
	"strings"
)

func Decolorise(str string) string {
	re := regexp.MustCompile(`\[[a-z]+:[a-z]+:[a-z]+\]`)
	return re.ReplaceAllString(str, "")
}

func WithPadding(str string, padding int) string {
	uncoloredStr := Decolorise(str)
	if padding < len(uncoloredStr) {
		return str
	}
	return str + strings.Repeat(" ", padding-len(uncoloredStr))
}

func ToTitle(str string) string {
	return strings.Title(strings.ToLower(str))
}

func getIndex(names []string, name string) int {
	for k, v := range names {
		if v == name {
			return k
		}
	}

	return -1
}

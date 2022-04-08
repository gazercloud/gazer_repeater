package httpserver

import "strings"

func SplitRequest(req string) []string {
	return strings.FieldsFunc(req, func(r rune) bool {
		return r == '/'
	})
}

package generator

import (
	"html/template"
	"strings"
	"time"
)

// FuncMap returns a map of custom template functions.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"now": func() string {
			return time.Now().String()
		},
		"nowFormat": func(layout string) string {
			return time.Now().Format(layout)
		},
		"formatDate": func(t string, layout string) string {
			parsedTime, err := time.Parse(time.RFC3339, t)
			if err != nil {
				return "Invalid Date"
			}
			return parsedTime.Format(layout)
		},
		"upper": func(s string) string {
			return strings.ToUpper(s)
		},
		"lower": func(s string) string {
			return strings.ToLower(s)
		},
		"title": func(s string) string {
			return strings.Title(s)
		},
		"sub": func(a, b int) int {
			return a - b
		},
	}
}

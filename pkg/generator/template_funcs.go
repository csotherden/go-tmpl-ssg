package generator

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

// FuncMap returns a map of custom template functions.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"html": func(s string) template.HTML {
			return template.HTML(s)
		},
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
		"add": func(a, b int) int {
			return a + b
		},
		"seq": func(start, end int) []int {
			seq := []int{}
			for i := start; i <= end; i++ {
				seq = append(seq, i)
			}
			return seq
		},
		"join": func(list []interface{}, sep string) string {
			var s []string

			for _, item := range list {
				s = append(s, fmt.Sprintf("%v", item))
			}

			return strings.Join(s, sep)
		},
	}
}

package generator

import (
	"fmt"
	"html/template"
	"math"
	"net/url"
	"path/filepath"
	"regexp"
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
		"dateFromFormat": func(dateStr, fromFormat, toFormat string) string {
			parsedTime, err := time.Parse(fromFormat, dateStr)
			if err != nil {
				return "Invalid Date"
			}
			return parsedTime.Format(toFormat)
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
		"mul": func(a, b int) int {
			return a * b
		},
		"div": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"mod": func(a, b int) int {
			if b == 0 {
				return 0
			}
			return a % b
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
		// String manipulation functions
		"replace": func(s, old, new string) string {
			return strings.Replace(s, old, new, -1)
		},
		"split": func(s, sep string) []string {
			return strings.Split(s, sep)
		},
		"trim": func(s string) string {
			return strings.TrimSpace(s)
		},
		"trimPrefix": func(s, prefix string) string {
			return strings.TrimPrefix(s, prefix)
		},
		"trimSuffix": func(s, suffix string) string {
			return strings.TrimSuffix(s, suffix)
		},
		"truncate": func(s string, length int) string {
			if len(s) <= length {
				return s
			}
			return s[:length] + "..."
		},
		// URL related functions
		"urlEscape": func(s string) string {
			return url.QueryEscape(s)
		},
		"urlUnescape": func(s string) string {
			result, err := url.QueryUnescape(s)
			if err != nil {
				return s
			}
			return result
		},
		"slugify": func(s string) string {
			// Convert to lowercase
			s = strings.ToLower(s)

			// Replace spaces with hyphens
			s = strings.Replace(s, " ", "-", -1)

			// Remove all non-alphanumeric characters except hyphens
			reg := regexp.MustCompile("[^a-z0-9-]")
			s = reg.ReplaceAllString(s, "")

			// Replace multiple hyphens with a single hyphen
			reg = regexp.MustCompile("-+")
			s = reg.ReplaceAllString(s, "-")

			// Trim leading and trailing hyphens
			s = strings.Trim(s, "-")

			return s
		},
		// Path related functions
		"base": func(path string) string {
			return filepath.Base(path)
		},
		"dir": func(path string) string {
			return filepath.Dir(path)
		},
		"ext": func(path string) string {
			return filepath.Ext(path)
		},
		// Array/Slice functions
		"first": func(items []interface{}) interface{} {
			if len(items) == 0 {
				return nil
			}
			return items[0]
		},
		"last": func(items []interface{}) interface{} {
			if len(items) == 0 {
				return nil
			}
			return items[len(items)-1]
		},
		"slice": func(items []interface{}, start, end int) []interface{} {
			if len(items) == 0 {
				return []interface{}{}
			}
			if start < 0 {
				start = 0
			}
			if end > len(items) {
				end = len(items)
			}
			return items[start:end]
		},
		"contains": func(list []string, item string) bool {
			for _, v := range list {
				if v == item {
					return true
				}
			}
			return false
		},
		// Math functions
		"abs": func(n int) int {
			if n < 0 {
				return -n
			}
			return n
		},
		"max": func(a, b int) int {
			if a > b {
				return a
			}
			return b
		},
		"min": func(a, b int) int {
			if a < b {
				return a
			}
			return b
		},
		"ceil": func(n float64) int {
			return int(math.Ceil(n))
		},
		"floor": func(n float64) int {
			return int(math.Floor(n))
		},
		"round": func(n float64) int {
			return int(math.Round(n))
		},
		// Conditional functions
		"default": func(defaultValue, value interface{}) interface{} {
			if value == nil {
				return defaultValue
			}
			if s, ok := value.(string); ok && s == "" {
				return defaultValue
			}
			return value
		},
		// Content functions
		"markdownify": func(s string) template.HTML {
			// This is a simple placeholder. For proper Markdown conversion,
			// you would want to use a Markdown library like blackfriday or goldmark.
			// For now, we just wrap in paragraph tags as a basic example
			// In a real implementation, replace this with a markdown parser
			if s == "" {
				return template.HTML("")
			}
			return template.HTML("<p>" + s + "</p>") // Placeholder only
		},
	}
}

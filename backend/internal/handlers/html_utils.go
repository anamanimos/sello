package handlers

import (
	"regexp"
)

// stripHTML removes all HTML tags from the input string
func stripHTML(input string) string {
	re := regexp.MustCompile(`<[^>]*>`)
	return re.ReplaceAllString(input, "")
}

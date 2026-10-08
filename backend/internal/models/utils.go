package models

import (
	"regexp"
	"strings"
)

// GenerateSlug converts a string into a slug (lowercase, hyphens)
func GenerateSlug(input string) string {
	// Lowercase
	input = strings.ToLower(input)
	// Replace non-alphanumeric with hyphen
	reg := regexp.MustCompile("[^a-z0-9]+")
	input = reg.ReplaceAllString(input, "-")
	// Trim hyphens
	input = strings.Trim(input, "-")
	return input
}

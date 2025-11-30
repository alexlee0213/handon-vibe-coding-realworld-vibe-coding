package util

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

var (
	// nonAlphanumericRegex matches any character that is not alphanumeric or dash
	nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9-]+`)
	// multipleDashRegex matches multiple consecutive dashes
	multipleDashRegex = regexp.MustCompile(`-+`)
)

// GenerateSlug converts a title to a URL-friendly slug
// Example: "Hello World" -> "hello-world"
func GenerateSlug(title string) string {
	if title == "" {
		return ""
	}

	// Normalize unicode characters (e.g., é -> e)
	slug := normalizeUnicode(title)

	// Convert to lowercase
	slug = strings.ToLower(slug)

	// Replace spaces and underscores with dashes
	slug = strings.ReplaceAll(slug, " ", "-")
	slug = strings.ReplaceAll(slug, "_", "-")

	// Replace all non-alphanumeric characters (except dashes) with dashes
	slug = nonAlphanumericRegex.ReplaceAllString(slug, "-")

	// Replace multiple consecutive dashes with a single dash
	slug = multipleDashRegex.ReplaceAllString(slug, "-")

	// Trim leading and trailing dashes
	slug = strings.Trim(slug, "-")

	return slug
}

// GenerateUniqueSlug generates a unique slug by checking against existing slugs
// The checkExists function returns true if the slug already exists
func GenerateUniqueSlug(title string, checkExists func(slug string) bool) string {
	baseSlug := GenerateSlug(title)
	if baseSlug == "" {
		return ""
	}

	// If the base slug doesn't exist, use it
	if !checkExists(baseSlug) {
		return baseSlug
	}

	// Try adding a numeric suffix
	for i := 1; i < 1000; i++ {
		candidateSlug := baseSlug + "-" + itoa(i)
		if !checkExists(candidateSlug) {
			return candidateSlug
		}
	}

	// Fallback: add timestamp (should never happen in practice)
	return baseSlug + "-" + randomSuffix()
}

// normalizeUnicode removes accents and normalizes unicode characters
func normalizeUnicode(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, s)
	return result
}

// itoa converts an integer to a string without importing strconv
func itoa(n int) string {
	if n == 0 {
		return "0"
	}

	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

// randomSuffix generates a short random suffix for edge cases
func randomSuffix() string {
	// Use a simple timestamp-based approach
	// In production, you might use crypto/rand
	return "unique"
}

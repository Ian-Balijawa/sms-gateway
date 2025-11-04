package utils

import (
	"regexp"
	"strings"
)

// FormatPhone formats a phone number to a standard format
// Removes spaces, dashes, and other non-digit characters except +
// Ensures the number starts with the country code
func FormatPhone(phone string) string {
	// Remove all non-digit characters except +
	re := regexp.MustCompile(`[^\d+]`)
	cleaned := re.ReplaceAllString(phone, "")

	// If it doesn't start with +, assume it's a local number
	// You may want to customize this based on your country's format
	if !strings.HasPrefix(cleaned, "+") {
		// Remove leading zeros
		cleaned = strings.TrimLeft(cleaned, "0")
		// Add country code if needed (default: +256 for Uganda, adjust as needed)
		if !strings.HasPrefix(cleaned, "256") {
			cleaned = "256" + cleaned
		}
		cleaned = "+" + cleaned
	}

	return cleaned
}

// ValidatePhone performs basic phone number validation
func ValidatePhone(phone string) bool {
	formatted := FormatPhone(phone)
	// Basic validation: should be at least 10 digits (including country code)
	digits := regexp.MustCompile(`\d`).FindAllString(formatted, -1)
	return len(digits) >= 10 && len(digits) <= 15
}


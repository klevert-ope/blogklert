package middleware

import (
	"strings"
	"unicode/utf8"
)

// SanitizeInput sanitizes user input to prevent SQL injection
func SanitizeInput(input string, maxLength int) string {
	// Escape single quotes
	sanitizedInput := strings.ReplaceAll(input, "'", "''")

	// Remove potentially harmful characters
	sanitizedInput = removeSpecialCharacters(sanitizedInput)

	// Limit input length to prevent buffer overflows and DoS attacks
	sanitizedInput = truncateString(sanitizedInput, maxLength)

	return sanitizedInput
}

// removeSpecialCharacters removes potentially harmful characters from the input
func removeSpecialCharacters(input string) string {
	var safeRunes []rune
	for _, r := range input {
		switch {
		case isAlphanumeric(r):
			safeRunes = append(safeRunes, r)
		case isCommonSymbol(r):
			safeRunes = append(safeRunes, r)
		default:
			// If the rune does not match any known category, ignore it
		}
	}
	return string(safeRunes)
}

// isAlphabetic checks if a rune is an alphabetic character
func isAlphabetic(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z')
}

// isNumeric checks if a rune is a numeric character
func isNumeric(r rune) bool {
	return r >= '0' && r <= '9'
}

// isAlphanumeric checks if a rune is alphanumeric
func isAlphanumeric(r rune) bool {
	return isAlphabetic(r) || isNumeric(r)
}

var commonSymbols = map[rune]bool{
	' ':  true,
	'.':  true,
	',':  true,
	'-':  true,
	'_':  true,
	'!':  true,
	'/':  true,
	'@':  true,
	'#':  true,
	'%':  true,
	'*':  true,
	'&':  true,
	'+':  true,
	'=':  true,
	';':  true,
	':':  true,
	'?':  true,
	'<':  true,
	'>':  true,
	'(':  true,
	')':  true,
	'[':  true,
	']':  true,
	'{':  true,
	'}':  true,
	'|':  true,
	'\\': true,
	'"':  true,
	'\'': true,
}

// isCommonSymbol checks if a rune is a common symbol
func isCommonSymbol(r rune) bool {
	return commonSymbols[r]
}

// truncateString truncates the input string to the specified maximum length
func truncateString(input string, maxLength int) string {
	if utf8.RuneCountInString(input) > maxLength {
		// Truncate the string if it exceeds the maximum length
		runes := []rune(input)
		return string(runes[:maxLength])
	}
	return input
}

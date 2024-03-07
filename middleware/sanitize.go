package middleware

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// SanitizeInput sanitizes user input to prevent SQL injection
func SanitizeInput(input string, maxWordCount int) string {
	// Escape single quotes
	sanitizedInput := strings.ReplaceAll(input, "'", "''")

	// Remove potentially harmful characters
	sanitizedInput = removeSpecialCharacters(sanitizedInput)

	// Limit input length to prevent buffer overflows and DoS attacks
	sanitizedInput = truncateString(sanitizedInput, maxWordCount)

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

// truncateString truncates the input string to the specified maximum length per word
func truncateString(input string, maxWordCount int) string {
	if utf8.RuneCountInString(input) > maxWordCount {
		// Truncate the string if it exceeds the maximum length
		var truncatedRunes []rune
		var currentLength int
		for _, r := range input {
			if unicode.IsSpace(r) || unicode.IsPunct(r) {
				// Add space or punctuation to the truncated string
				truncatedRunes = append(truncatedRunes, r)
				currentLength = 0 // Reset current word length counter
			} else {
				// Add non-space and non-punctuation characters to the truncated string
				truncatedRunes = append(truncatedRunes, r)
				currentLength++
				if currentLength >= maxWordCount {
					// If the current word length exceeds the maximum length, break
					break
				}
			}
		}
		return string(truncatedRunes)
	}
	return input
}

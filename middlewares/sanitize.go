package middlewares

import (
	"strings"
	"unicode"
)

// SanitizeInput sanitizes user input by removing potentially harmful characters and limiting word count.
func SanitizeInput(input string, maxWordCount int) string {
	if maxWordCount <= 0 {
		return "" // Return empty for invalid word limit
	}

	// Remove potentially harmful characters.
	sanitizedInput := removeUnsafeCharacters(input)

	// Normalize spaces (trim and reduce multiple spaces)
	sanitizedInput = normalizeSpaces(sanitizedInput)

	// Limit input length based on word count to prevent buffer overflows and DoS attacks.
	return truncateByWordCount(sanitizedInput, maxWordCount)
}

// normalizeSpaces trims leading/trailing spaces and reduces multiple spaces to a single space.
func normalizeSpaces(input string) string {
	// Trim leading and trailing spaces
	input = strings.TrimSpace(input)
	// Reduce multiple spaces to a single space
	return strings.Join(strings.Fields(input), " ")
}

// removeUnsafeCharacters removes potentially harmful characters from the input.
func removeUnsafeCharacters(input string) string {
	var safeRunes []rune
	for _, r := range input {
		if isSafeCharacter(r) {
			safeRunes = append(safeRunes, r)
		}
	}
	return string(safeRunes)
}

// isSafeCharacter checks if a rune is a safe character (letters, digits, or safe symbols).
func isSafeCharacter(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || isSafeSymbol(r)
}

// isSafeSymbol checks if a rune is a safe symbol.
func isSafeSymbol(r rune) bool {
	safeSymbols := " .,~-!/@#%*&$+÷€£¥×=;:?<>[]{}|\\\"'()"
	return strings.ContainsRune(safeSymbols, r)
}

// truncateByWordCount truncates the input string based on word count.
func truncateByWordCount(input string, maxWordCount int) string {
	words := strings.Fields(input)
	if len(words) > maxWordCount {
		return strings.Join(words[:maxWordCount], " ")
	}
	return input
}

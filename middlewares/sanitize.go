package middlewares

import (
	"strings"
	"unicode"
)

// SanitizeInput sanitizes user input by removing potentially harmful characters and limiting word count.
func SanitizeInput(input string, maxWordCount int) string {
	// Remove potentially harmful characters.
	sanitizedInput := removeUnsafeCharacters(input)

	// Limit input length based on word count to prevent buffer overflows and DoS attacks.
	return truncateByWordCount(sanitizedInput, maxWordCount)
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

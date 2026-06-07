package service

import (
	"strings"
	"unicode/utf8"

	"github.com/even-app/even-app/services/lexicon/internal/domain"
)

// maxLetterRunes limits alphabet key size (single Unicode letter; Even uses ӈ, Ӈ, etc.).
const maxLetterRunes = 8

func normalizeLetterField(s string) string {
	return strings.TrimSpace(s)
}

// validateLetterField ensures a non-empty Unicode letter (1–8 runes).
// Even uses Cyrillic extensions, e.g. ӈ (U+04A9), Ӈ (U+04A7); standard а/б also allowed.
func validateLetterField(field, value string, required bool) (string, error) {
	v := normalizeLetterField(value)
	n := utf8.RuneCountInString(v)
	if n == 0 {
		if required {
			return "", domain.ErrValidation
		}
		return "", nil
	}
	if n > maxLetterRunes {
		return "", domain.ErrValidation
	}
	return v, nil
}

const maxTranscriptionRunes = 64

func normalizeTranscriptionField(s string) string {
	return strings.TrimSpace(s)
}

func validateOptionalTranscription(value string) (*string, error) {
	v := normalizeTranscriptionField(value)
	if v == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(v) > maxTranscriptionRunes {
		return nil, domain.ErrValidation
	}
	return &v, nil
}

func validateOptionalLetterField(field, value string) (*string, error) {
	if value == "" {
		return nil, nil
	}
	v, err := validateLetterField(field, value, true)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

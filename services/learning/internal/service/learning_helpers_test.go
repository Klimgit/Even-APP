package service

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestExtractLexemeFromConfig(t *testing.T) {
	lex := uuid.New()
	raw, _ := json.Marshal(map[string]any{"lexeme_id": lex.String()})
	got := extractLexemeFromConfig(raw)
	if got != lex {
		t.Fatalf("lexeme_id: got %v want %v", got, lex)
	}
}

func TestExtractLexemeFromConfig_promptLexeme(t *testing.T) {
	lex := uuid.New()
	raw, _ := json.Marshal(map[string]any{"prompt_lexeme_id": lex.String()})
	if got := extractLexemeFromConfig(raw); got != lex {
		t.Fatalf("prompt_lexeme_id: got %v want %v", got, lex)
	}
}

func TestExtractLexemeFromConfig_correctLexeme(t *testing.T) {
	lex := uuid.New()
	raw, _ := json.Marshal(map[string]any{"correct_lexeme_id": lex.String()})
	if got := extractLexemeFromConfig(raw); got != lex {
		t.Fatalf("correct_lexeme_id: got %v want %v", got, lex)
	}
}

func TestExtractLexemeFromConfig_empty(t *testing.T) {
	if got := extractLexemeFromConfig(nil); got != uuid.Nil {
		t.Fatalf("expected nil uuid, got %v", got)
	}
	if got := extractLexemeFromConfig(json.RawMessage(`{`)); got != uuid.Nil {
		t.Fatalf("invalid json should return nil, got %v", got)
	}
}

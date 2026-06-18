package domain

import (
	"encoding/json"
	"testing"
)

func TestGradeBlock_promptChooseWord_correctIndex(t *testing.T) {
	cfg := json.RawMessage(`{"correct_index":0,"choices":[{"lexeme_id":"a"},{"lexeme_id":"b"}]}`)
	got := GradeBlock("prompt_choose_word", cfg, 0, map[string]any{"selected_index": float64(0)})
	if !got.Correct || got.Score != 1 {
		t.Fatalf("expected correct, got %+v", got)
	}
}

func TestGradeBlock_promptChooseWord_wrongIndex(t *testing.T) {
	cfg := json.RawMessage(`{"correct_index":0,"choices":[{"lexeme_id":"a"},{"lexeme_id":"b"}]}`)
	got := GradeBlock("prompt_choose_word", cfg, 0, map[string]any{"selected_index": float64(1)})
	if got.Correct || got.Score != 0 {
		t.Fatalf("expected wrong, got %+v", got)
	}
}

func TestGradeBlock_correctLexemeID(t *testing.T) {
	lex := "550e8400-e29b-41d4-a716-446655440000"
	cfg := json.RawMessage(`{"correct_lexeme_id":"` + lex + `"}`)
	got := GradeBlock("prompt_choose_word", cfg, 0, map[string]any{"selected_lexeme_id": lex})
	if !got.Correct {
		t.Fatalf("expected correct lexeme match, got %+v", got)
	}
}

func TestGradeBlock_textAnswerCaseInsensitive(t *testing.T) {
	cfg := json.RawMessage(`{"answer":"Привет"}`)
	got := GradeBlock("prompt_type_word", cfg, 0, map[string]any{"text": "  привет "})
	if !got.Correct {
		t.Fatalf("expected case-insensitive text match, got %+v", got)
	}
}

func TestGradeBlock_wordOrder(t *testing.T) {
	cfg := json.RawMessage(`{"correct_order":["a","b","c"]}`)
	got := GradeBlock("prompt_sentence_word_order", cfg, 0, map[string]any{
		"word_order": []any{"a", "b", "c"},
	})
	if !got.Correct {
		t.Fatalf("expected order match, got %+v", got)
	}
}

func TestGradeBlock_wordOrderWrong(t *testing.T) {
	cfg := json.RawMessage(`{"correct_order":["a","b","c"]}`)
	got := GradeBlock("prompt_sentence_word_order", cfg, 0, map[string]any{
		"word_order": []any{"c", "b", "a"},
	})
	if got.Correct {
		t.Fatalf("expected wrong order, got %+v", got)
	}
}

func TestGradeBlock_subItem(t *testing.T) {
	cfg := json.RawMessage(`{"items":[{"correct_index":1},{"correct_index":0}]}`)
	got := GradeBlock("reading_yes_no", cfg, 1, map[string]any{"selected_index": float64(0)})
	if !got.Correct {
		t.Fatalf("expected sub-item 1 correct, got %+v", got)
	}
}

func TestGradeBlock_boolAnswer(t *testing.T) {
	cfg := json.RawMessage(`{"correct_answer":true}`)
	got := GradeBlock("reading_yes_no", cfg, 0, map[string]any{"answer": true})
	if !got.Correct {
		t.Fatalf("expected bool match, got %+v", got)
	}
}

func TestGradeBlock_emptyConfig(t *testing.T) {
	got := GradeBlock("prompt_choose_word", nil, 0, map[string]any{"selected_index": float64(0)})
	if got.Correct {
		t.Fatalf("expected failure on empty config, got %+v", got)
	}
}

func TestGradeBlock_singleKeyResponse(t *testing.T) {
	cfg := json.RawMessage(`{"correct_index":2}`)
	got := GradeBlock("prompt_choose_word", cfg, 0, map[string]any{"value": float64(2)})
	if !got.Correct {
		t.Fatalf("expected single-key response fallback, got %+v", got)
	}
}

func TestGradeBlock_listenChooseWord(t *testing.T) {
	lex := "11111111-1111-1111-1111-111111111111"
	cfg := json.RawMessage(`{"answer_lexeme_id":"` + lex + `"}`)
	got := GradeBlock("listen_choose_word", cfg, 0, map[string]any{"lexeme_id": lex})
	if !got.Correct {
		t.Fatalf("expected listen choose word match, got %+v", got)
	}
}

func TestReviewDueInterval(t *testing.T) {
	tests := []struct {
		failures int
		want     int
	}{
		{0, 4}, {1, 4}, {2, 24}, {3, 72}, {4, 168}, {10, 168},
	}
	for _, tc := range tests {
		if got := ReviewDueInterval(tc.failures); got != tc.want {
			t.Errorf("ReviewDueInterval(%d) = %d, want %d", tc.failures, got, tc.want)
		}
	}
}

func TestValuesEqual_uuidNotFolded(t *testing.T) {
	u := "550e8400-e29b-41d4-a716-446655440000"
	if !valuesEqual(u, u) {
		t.Fatal("uuid should match exactly")
	}
	if valuesEqual(u, "550E8400-E29B-41D4-A716-446655440000") {
		t.Fatal("uuid case should not fold")
	}
}

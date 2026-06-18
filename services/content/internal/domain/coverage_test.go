package domain

import (
	"strings"
	"testing"
)

func TestGenerateInviteCode_format(t *testing.T) {
	code, err := GenerateInviteCode()
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 8 {
		t.Fatalf("invite code length = %d, want 8", len(code))
	}
	for _, ch := range code {
		if !strings.ContainsRune(inviteAlphabet, ch) {
			t.Fatalf("invalid char %q in code %q", ch, code)
		}
	}
}

func TestGenerateInviteCode_unique(t *testing.T) {
	seen := map[string]struct{}{}
	for i := 0; i < 50; i++ {
		code, err := GenerateInviteCode()
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := seen[code]; ok {
			t.Fatalf("duplicate invite code %q", code)
		}
		seen[code] = struct{}{}
	}
}

func TestExtractLexemeRefs_vocabularySet(t *testing.T) {
	cfg := []byte(`{"lexeme_ids":["aaa-111","bbb-222"],"show_images":true}`)
	refs := ExtractLexemeRefs(BlockLexemeSource{
		BlockType: "vocabulary_set",
		Config:    cfg,
	})
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d: %+v", len(refs), refs)
	}
	for _, ref := range refs {
		if ref.UsageKind != UsageIntroduced {
			t.Fatalf("vocabulary_set should introduce, got %s", ref.UsageKind)
		}
	}
}

func TestExtractLexemeRefs_gradableExercise(t *testing.T) {
	cfg := []byte(`{"correct_lexeme_id":"lex-1","choices":[{"lexeme_id":"lex-1"},{"lexeme_id":"lex-2"}]}`)
	refs := ExtractLexemeRefs(BlockLexemeSource{
		BlockType:    "prompt_choose_word",
		Config:       cfg,
		DisplayLabel: strPtr("1.1"),
	})
	if len(refs) < 2 {
		t.Fatalf("expected at least 2 refs, got %+v", refs)
	}
	foundExercised := false
	for _, ref := range refs {
		if ref.UsageKind == UsageExercised {
			foundExercised = true
		}
		if ref.BlockDisplayLabel != "1.1" {
			t.Fatalf("expected display label 1.1, got %q", ref.BlockDisplayLabel)
		}
	}
	if !foundExercised {
		t.Fatal("expected exercised usage kind")
	}
}

func TestExtractLexemeRefs_invalidJSON(t *testing.T) {
	refs := ExtractLexemeRefs(BlockLexemeSource{BlockType: "text", Config: []byte("{")})
	if refs != nil {
		t.Fatalf("expected nil on invalid json, got %+v", refs)
	}
}

func TestExtractLexemeRefs_wordBank(t *testing.T) {
	cfg := []byte(`{"word_bank_lexeme_ids":["w1","w2"]}`)
	refs := ExtractLexemeRefs(BlockLexemeSource{BlockType: "gap_sentence_choose_word", Config: cfg})
	if len(refs) != 2 {
		t.Fatalf("expected 2 word bank refs, got %+v", refs)
	}
	for _, ref := range refs {
		if ref.UsageKind != UsageReferenced {
			t.Fatalf("word bank should be referenced, got %s", ref.UsageKind)
		}
	}
}

func strPtr(s string) *string { return &s }

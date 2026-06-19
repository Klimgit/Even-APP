package domain

import (
	"testing"
)

func TestValidateBlockConfig_promptChooseWord(t *testing.T) {
	valid := `{"prompt":{"items":[{"kind":"text","text":"Q"}]},"choices":[{"lexeme_id":"00000000-0000-4000-8000-000000000001"},{"lexeme_id":"00000000-0000-4000-8000-000000000002"}],"correct_index":0}`
	if err := ValidateBlockConfig("prompt_choose_word", []byte(valid)); err != nil {
		t.Fatalf("expected valid config: %v", err)
	}
	if err := ValidateBlockConfig("prompt_choose_word", []byte(`{"prompt":{"items":[{"kind":"text","text":"Q"}]},"choices":[],"correct_index":0}`)); err == nil {
		t.Fatal("expected error for empty choices")
	}
	if err := ValidateBlockConfig("prompt_choose_word", []byte(`{"prompt":{"items":[{"kind":"text","text":"Q"}]},"choices":[{"lexeme_id":"x"}],"correct_index":3}`)); err == nil {
		t.Fatal("expected error for out-of-range correct_index")
	}
}

func TestValidateBlockConfig_vocabularySet(t *testing.T) {
	valid := `{"lexeme_ids":["00000000-0000-4000-8000-000000000001"]}`
	if err := ValidateBlockConfig("vocabulary_set", []byte(valid)); err != nil {
		t.Fatalf("expected valid: %v", err)
	}
}

func TestValidateBlockConfig_video(t *testing.T) {
	if err := ValidateBlockConfig("video", []byte(`{"media_asset_id":"00000000-0000-4000-8000-000000000001"}`)); err != nil {
		t.Fatalf("media_asset_id: %v", err)
	}
	if err := ValidateBlockConfig("video", []byte(`{"items":[{"kind":"video","media_asset_id":"00000000-0000-4000-8000-000000000001"}]}`)); err != nil {
		t.Fatalf("items: %v", err)
	}
}

func TestValidateBlockConfig_typeWord(t *testing.T) {
	cfg := `{"prompt":{"items":[{"kind":"lexeme","lexeme_id":"00000000-0000-4000-8000-000000000001"}]},"correct_lexeme_id":"00000000-0000-4000-8000-000000000001"}`
	if err := ValidateBlockConfig("prompt_type_word", []byte(cfg)); err != nil {
		t.Fatalf("correct_lexeme_id: %v", err)
	}
}

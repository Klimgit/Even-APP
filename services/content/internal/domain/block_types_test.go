package domain

import "testing"

func TestIsGradableBlockType(t *testing.T) {
	gradable := []string{
		"prompt_choose_word", "prompt_type_word", "word_choose_image",
		"listen_choose_word", "listen_sentence_type",
	}
	content := []string{"text", "audio", "video", "vocabulary_set", "note", "images_stacked"}

	for _, bt := range gradable {
		if !IsGradableBlockType(bt) {
			t.Errorf("%q should be gradable", bt)
		}
	}
	for _, bt := range content {
		if IsGradableBlockType(bt) {
			t.Errorf("%q should not be gradable", bt)
		}
	}
}

func TestMVPBlockTypeCatalog_counts(t *testing.T) {
	cats := MVPBlockTypeCatalog()
	if len(cats) != 3 {
		t.Fatalf("expected 3 categories, got %d", len(cats))
	}
	total := 0
	gradable := 0
	for _, cat := range cats {
		total += len(cat.Types)
		for _, bt := range cat.Types {
			if bt.IsGradable != IsGradableBlockType(bt.BlockType) {
				t.Fatalf("catalog gradable flag mismatch for %q", bt.BlockType)
			}
			if bt.IsGradable {
				gradable++
			}
		}
	}
	if total != 17 {
		t.Fatalf("expected 17 block types, got %d", total)
	}
	if gradable != 11 {
		t.Fatalf("expected 11 gradable types, got %d", gradable)
	}
}

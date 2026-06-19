package domain

// GradableBlockTypes are MVP gradable block types (11 from spec).
var GradableBlockTypes = map[string]bool{
	"prompt_choose_word":         true,
	"prompt_type_word":           true,
	"word_choose_image":          true,
	"word_choose_translation":    true,
	"gap_sentence_choose_word":   true,
	"prompt_sentence_word_order": true,
	"prompt_sentence_type":       true,
	"listen_choose_word":         true,
	"listen_type_word":           true,
	"listen_sentence_word_order": true,
	"listen_sentence_type":       true,
}

func IsGradable(blockType string) bool {
	return GradableBlockTypes[blockType]
}

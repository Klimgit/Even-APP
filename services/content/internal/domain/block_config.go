package domain

import (
	"encoding/json"
	"fmt"
)

// ValidateBlockConfig checks MVP block config shape (required keys per type).
func ValidateBlockConfig(blockType string, config []byte) error {
	if blockType == "" || !IsKnownBlockType(blockType) {
		return ErrValidation
	}
	if len(config) == 0 {
		return validateEmptyConfig(blockType)
	}
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(config, &obj); err != nil {
		return ErrValidation
	}
	return validateConfigObject(blockType, obj)
}

func validateEmptyConfig(blockType string) error {
	switch blockType {
	case "text", "images_stacked", "audio", "video", "vocabulary_set", "note":
		return ErrValidation
	default:
		return nil
	}
}

func validateConfigObject(blockType string, obj map[string]json.RawMessage) error {
	has := func(key string) bool {
		raw, ok := obj[key]
		return ok && len(raw) > 0 && string(raw) != "null"
	}
	switch blockType {
	case "text", "images_stacked", "audio", "video":
		if !has("items") {
			return fmt.Errorf("%w: items required", ErrValidation)
		}
	case "vocabulary_set":
		if !has("lexeme_ids") {
			return fmt.Errorf("%w: lexeme_ids required", ErrValidation)
		}
	case "note":
		if !has("text") && !has("body") && !has("items") {
			return fmt.Errorf("%w: text, body or items required", ErrValidation)
		}
	case "prompt_choose_word", "word_choose_translation", "listen_choose_word":
		if !has("prompt") || !has("choices") || !has("correct_index") {
			return fmt.Errorf("%w: prompt, choices, correct_index required", ErrValidation)
		}
	case "prompt_type_word", "listen_type_word":
		if !has("prompt") || !has("lexeme_id") {
			return fmt.Errorf("%w: prompt and lexeme_id required", ErrValidation)
		}
	case "word_choose_image":
		if !has("prompt") || !has("choices") || !has("correct_index") {
			return fmt.Errorf("%w: prompt, choices, correct_index required", ErrValidation)
		}
	case "gap_sentence_choose_word":
		if !has("sentence") || !has("choices") || !has("correct_index") {
			return fmt.Errorf("%w: sentence, choices, correct_index required", ErrValidation)
		}
	case "prompt_sentence_word_order", "listen_sentence_word_order":
		if !has("prompt") || !has("word_bank") || !has("correct_order") {
			return fmt.Errorf("%w: prompt, word_bank, correct_order required", ErrValidation)
		}
	case "prompt_sentence_type", "listen_sentence_type":
		if !has("prompt") || !has("correct_text") {
			return fmt.Errorf("%w: prompt and correct_text required", ErrValidation)
		}
	}
	return nil
}

package domain

import (
	"encoding/json"
	"fmt"
)

// ValidateBlockConfig checks MVP block config shape (required keys and basic types).
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
	case "grammar_table", "reading_passage":
		return ErrValidation
	default:
		if IsGradableBlockType(blockType) {
			return ErrValidation
		}
		return nil
	}
}

func validateConfigObject(blockType string, obj map[string]json.RawMessage) error {
	has := func(key string) bool {
		raw, ok := obj[key]
		return ok && len(raw) > 0 && string(raw) != "null"
	}
	switch blockType {
	case "text", "images_stacked", "audio":
		if err := requireJSONArray(obj, "items"); err != nil {
			return err
		}
	case "video":
		if !has("media_asset_id") {
			if err := requireJSONArray(obj, "items"); err != nil {
				return fmt.Errorf("%w: media_asset_id or items required", ErrValidation)
			}
		}
	case "vocabulary_set":
		if err := requireNonEmptyUUIDArray(obj, "lexeme_ids"); err != nil {
			return err
		}
	case "note":
		if !has("text") && !has("body") && !has("items") {
			return fmt.Errorf("%w: text, body or items required", ErrValidation)
		}
	case "prompt_choose_word", "word_choose_translation", "listen_choose_word", "word_choose_image":
		if err := requirePrompt(obj); err != nil {
			return err
		}
		if err := requireNonEmptyArray(obj, "choices"); err != nil {
			return err
		}
		if err := requireIntInRange(obj, "correct_index", obj["choices"]); err != nil {
			return err
		}
	case "prompt_type_word", "listen_type_word":
		if err := requirePrompt(obj); err != nil {
			return err
		}
		if !has("correct_lexeme_id") && !has("lexeme_id") {
			return fmt.Errorf("%w: correct_lexeme_id or lexeme_id required", ErrValidation)
		}
	case "gap_sentence_choose_word":
		if !has("body") && !has("sentence") {
			return fmt.Errorf("%w: body or sentence required", ErrValidation)
		}
		if !has("gaps") && !has("correct_index") {
			return fmt.Errorf("%w: gaps or correct_index required", ErrValidation)
		}
		if has("choices") {
			if err := requireNonEmptyArray(obj, "choices"); err != nil {
				return err
			}
		}
	case "prompt_sentence_word_order", "listen_sentence_word_order":
		if has("sub_items") {
			return requireNonEmptyArray(obj, "sub_items")
		}
		if err := requirePrompt(obj); err != nil {
			return err
		}
		if err := requireNonEmptyArray(obj, "word_bank"); err != nil {
			return err
		}
		if !has("correct_order") {
			return fmt.Errorf("%w: correct_order or sub_items required", ErrValidation)
		}
	case "prompt_sentence_type", "listen_sentence_type":
		if err := requirePrompt(obj); err != nil {
			return err
		}
		if !has("correct_text") {
			return fmt.Errorf("%w: correct_text required", ErrValidation)
		}
	case "grammar_table", "reading_passage":
		if !has("body") && !has("items") && !has("text") {
			return fmt.Errorf("%w: body, items or text required", ErrValidation)
		}
	case "grammar_exercise", "reading_comprehension":
		if !has("prompt") && !has("items") && !has("questions") {
			return fmt.Errorf("%w: prompt, items or questions required", ErrValidation)
		}
	}
	return nil
}

func requirePrompt(obj map[string]json.RawMessage) error {
	raw, ok := obj["prompt"]
	if !ok || len(raw) == 0 || string(raw) == "null" {
		return fmt.Errorf("%w: prompt required", ErrValidation)
	}
	var prompt map[string]json.RawMessage
	if err := json.Unmarshal(raw, &prompt); err != nil {
		return fmt.Errorf("%w: prompt must be object", ErrValidation)
	}
	items, ok := prompt["items"]
	if !ok {
		return nil
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(items, &arr); err != nil || len(arr) == 0 {
		return fmt.Errorf("%w: prompt.items must be non-empty array", ErrValidation)
	}
	return nil
}

func requireJSONArray(obj map[string]json.RawMessage, key string) error {
	raw, ok := obj[key]
	if !ok || len(raw) == 0 || string(raw) == "null" {
		return fmt.Errorf("%w: %s required", ErrValidation, key)
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil || len(arr) == 0 {
		return fmt.Errorf("%w: %s must be non-empty array", ErrValidation, key)
	}
	return nil
}

func requireNonEmptyArray(obj map[string]json.RawMessage, key string) error {
	return requireJSONArray(obj, key)
}

func requireNonEmptyUUIDArray(obj map[string]json.RawMessage, key string) error {
	raw, ok := obj[key]
	if !ok || len(raw) == 0 || string(raw) == "null" {
		return fmt.Errorf("%w: %s required", ErrValidation, key)
	}
	var arr []string
	if err := json.Unmarshal(raw, &arr); err != nil || len(arr) == 0 {
		return fmt.Errorf("%w: %s must be non-empty array", ErrValidation, key)
	}
	return nil
}

func requireIntInRange(obj map[string]json.RawMessage, key string, choicesRaw json.RawMessage) error {
	raw, ok := obj[key]
	if !ok || len(raw) == 0 || string(raw) == "null" {
		return fmt.Errorf("%w: %s required", ErrValidation, key)
	}
	var idx int
	if err := json.Unmarshal(raw, &idx); err != nil {
		return fmt.Errorf("%w: %s must be integer", ErrValidation, key)
	}
	if idx < 0 {
		return fmt.Errorf("%w: %s must be >= 0", ErrValidation, key)
	}
	if len(choicesRaw) > 0 {
		var choices []json.RawMessage
		if err := json.Unmarshal(choicesRaw, &choices); err == nil && idx >= len(choices) {
			return fmt.Errorf("%w: %s out of range", ErrValidation, key)
		}
	}
	return nil
}

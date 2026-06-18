package domain

import (
	"encoding/json"
)

// LexemeUsageKind classifies how a lexeme appears in a block.
type LexemeUsageKind string

const (
	UsageIntroduced LexemeUsageKind = "introduced"
	UsageExercised  LexemeUsageKind = "exercised"
	UsageReferenced LexemeUsageKind = "referenced"
)

// LexemeRef is one lexeme (and optional form) extracted from block config.
type LexemeRef struct {
	LexemeID          string
	FormID            string
	UsageKind         LexemeUsageKind
	BlockDisplayLabel string
}

// BlockLexemeSource is input for lexeme extraction.
type BlockLexemeSource struct {
	BlockType    string
	Config       []byte
	DisplayLabel *string
}

var gradableBlockTypes = map[string]struct{}{
	"prompt_choose_word":         {},
	"prompt_type_word":           {},
	"word_choose_image":          {},
	"word_choose_translation":    {},
	"gap_sentence_choose_word":   {},
	"prompt_sentence_word_order": {},
	"prompt_sentence_type":       {},
	"listen_choose_word":         {},
	"listen_type_word":           {},
	"listen_sentence_word_order": {},
	"listen_sentence_type":       {},
}

func isGradableBlockType(blockType string) bool {
	_, ok := gradableBlockTypes[blockType]
	return ok
}

// ExtractLexemeRefs walks block config JSON and collects lexeme references.
func ExtractLexemeRefs(src BlockLexemeSource) []LexemeRef {
	var root any
	if len(src.Config) == 0 {
		return nil
	}
	if err := json.Unmarshal(src.Config, &root); err != nil {
		return nil
	}

	kind := usageKindForBlockType(src.BlockType)
	label := ""
	if src.DisplayLabel != nil {
		label = *src.DisplayLabel
	}

	seen := make(map[string]LexemeRef)
	walkJSON(root, kind, label, seen)
	out := make([]LexemeRef, 0, len(seen))
	for _, ref := range seen {
		out = append(out, ref)
	}
	return out
}

func usageKindForBlockType(blockType string) LexemeUsageKind {
	switch blockType {
	case "vocabulary_set", "text", "images_stacked", "audio", "video", "note":
		return UsageIntroduced
	default:
		if isGradableBlockType(blockType) {
			return UsageExercised
		}
		return UsageReferenced
	}
}

func walkJSON(v any, kind LexemeUsageKind, label string, out map[string]LexemeRef) {
	switch node := v.(type) {
	case map[string]any:
		var lexemeID, formID string
		if raw, ok := node["lexeme_id"].(string); ok && raw != "" {
			lexemeID = raw
		}
		if raw, ok := node["correct_lexeme_id"].(string); ok && raw != "" {
			lexemeID = raw
		}
		if raw, ok := node["correct_form_id"].(string); ok && raw != "" {
			formID = raw
		}
		if lexemeID != "" {
			key := lexemeID + "|" + formID + "|" + string(kind)
			out[key] = LexemeRef{
				LexemeID:          lexemeID,
				FormID:            formID,
				UsageKind:         kind,
				BlockDisplayLabel: label,
			}
		}
		if ids, ok := node["lexeme_ids"].([]any); ok {
			for _, id := range ids {
				if s, ok := id.(string); ok && s != "" {
					key := s + "||" + string(kind)
					out[key] = LexemeRef{LexemeID: s, UsageKind: kind, BlockDisplayLabel: label}
				}
			}
		}
		if ids, ok := node["word_bank_lexeme_ids"].([]any); ok {
			for _, id := range ids {
				if s, ok := id.(string); ok && s != "" {
					key := s + "||referenced"
					out[key] = LexemeRef{LexemeID: s, UsageKind: UsageReferenced, BlockDisplayLabel: label}
				}
			}
		}
		if gaps, ok := node["gaps"].([]any); ok {
			for _, gap := range gaps {
				if gm, ok := gap.(map[string]any); ok {
					walkJSON(gm, kind, label, out)
				}
			}
		}
		if items, ok := node["items"].([]any); ok {
			for _, item := range items {
				walkJSON(item, kind, label, out)
			}
		}
		if subs, ok := node["sub_items"].([]any); ok {
			for _, item := range subs {
				walkJSON(item, kind, label, out)
			}
		}
		for _, child := range node {
			walkJSON(child, kind, label, out)
		}
	case []any:
		for _, child := range node {
			walkJSON(child, kind, label, out)
		}
	}
}

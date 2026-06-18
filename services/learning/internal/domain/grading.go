package domain

import (
	"encoding/json"
	"strings"
)

type GradeResult struct {
	Correct       bool
	Score         float64
	CorrectAnswer any
}

// GradeBlock compares student response JSON to expected answer in block config.
func GradeBlock(blockType string, config json.RawMessage, subItemIndex int, response map[string]any) GradeResult {
	cfg := parseConfig(config)
	expected := extractExpected(cfg, blockType, subItemIndex)
	if expected == nil {
		return GradeResult{Correct: false, Score: 0}
	}
	correct := compareValues(expected, response)
	score := 0.0
	if correct {
		score = 1.0
	}
	return GradeResult{Correct: correct, Score: score, CorrectAnswer: expected}
}

func parseConfig(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil || m == nil {
		return map[string]any{}
	}
	return m
}

func extractExpected(cfg map[string]any, blockType string, subItemIndex int) any {
	if item := subItem(cfg, subItemIndex); item != nil {
		if v, ok := item["correct_answer"]; ok {
			return v
		}
		if v, ok := item["answer"]; ok {
			return v
		}
		if v, ok := item["correct_lexeme_id"]; ok {
			return v
		}
		if v, ok := item["correct_index"]; ok {
			return v
		}
		if v, ok := item["correct_text"]; ok {
			return v
		}
		if v, ok := item["correct_order"]; ok {
			return v
		}
	}

	switch blockType {
	case "prompt_choose_word", "word_choose_image", "word_choose_translation",
		"gap_sentence_choose_word", "listen_choose_word":
		if v, ok := cfg["correct_lexeme_id"]; ok {
			return v
		}
		if v, ok := cfg["answer_lexeme_id"]; ok {
			return v
		}
	case "prompt_type_word", "prompt_sentence_type", "listen_type_word":
		if v, ok := cfg["answer"]; ok {
			return v
		}
		if v, ok := cfg["correct_text"]; ok {
			return v
		}
	case "prompt_sentence_word_order", "listen_sentence_word_order":
		if v, ok := cfg["correct_order"]; ok {
			return v
		}
		if v, ok := cfg["answer_order"]; ok {
			return v
		}
	case "reading_yes_no", "true_false_unknown":
		if v, ok := cfg["correct_answer"]; ok {
			return v
		}
		if v, ok := cfg["answer"]; ok {
			return v
		}
	}

	for _, key := range []string{"correct_answer", "answer", "correct_lexeme_id", "correct_index", "correct_text", "correct_order", "value"} {
		if v, ok := cfg[key]; ok {
			return v
		}
	}
	return nil
}

func subItem(cfg map[string]any, idx int) map[string]any {
	for _, key := range []string{"items", "sub_items", "questions"} {
		raw, ok := cfg[key]
		if !ok {
			continue
		}
		arr, ok := raw.([]any)
		if !ok || idx < 0 || idx >= len(arr) {
			continue
		}
		if m, ok := arr[idx].(map[string]any); ok {
			return m
		}
	}
	return nil
}

func compareValues(expected any, response map[string]any) bool {
	for _, key := range []string{"selected_lexeme_id", "lexeme_id", "text", "answer", "value", "selected_index", "index", "order", "selected_order", "word_order"} {
		if rv, ok := response[key]; ok {
			return valuesEqual(expected, rv)
		}
	}
	if len(response) == 1 {
		for _, rv := range response {
			return valuesEqual(expected, rv)
		}
	}
	return false
}

func valuesEqual(a, b any) bool {
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		if !ok {
			return false
		}
		if isLikelyTextField(av) {
			return strings.EqualFold(strings.TrimSpace(av), strings.TrimSpace(bv))
		}
		return av == bv
	case float64:
		switch bv := b.(type) {
		case float64:
			return av == bv
		case int:
			return av == float64(bv)
		}
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	case []any:
		bArr, ok := b.([]any)
		if !ok {
			if bStr, ok := b.(string); ok {
				var parsed []any
				if json.Unmarshal([]byte(bStr), &parsed) == nil {
					return slicesEqual(av, parsed)
				}
			}
			return false
		}
		return slicesEqual(av, bArr)
	case map[string]any:
		bm, ok := b.(map[string]any)
		if !ok {
			return false
		}
		aJSON, _ := json.Marshal(av)
		bJSON, _ := json.Marshal(bm)
		return string(aJSON) == string(bJSON)
	}
	aJSON, _ := json.Marshal(a)
	bJSON, _ := json.Marshal(b)
	return string(aJSON) == string(bJSON)
}

func slicesEqual(a, b []any) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !valuesEqual(a[i], b[i]) {
			return false
		}
	}
	return true
}

func isLikelyTextField(s string) bool {
	if strings.Contains(s, "-") && len(s) > 30 {
		return false
	}
	return true
}

// ReviewDueInterval returns due_at offset hours by failure count.
func ReviewDueInterval(failureCount int) int {
	switch {
	case failureCount <= 1:
		return 4
	case failureCount == 2:
		return 24
	case failureCount == 3:
		return 72
	default:
		return 168
	}
}

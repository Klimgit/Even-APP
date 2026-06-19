package service

import (
	"context"
	"errors"

	"github.com/even-app/even-app/services/lexicon/internal/domain"
	"github.com/google/uuid"
)

type ImportLexiconResult struct {
	Created int
	Skipped int
	Failed  int
}

func (s *LexiconService) ImportLexemes(ctx context.Context, code string, items []CreateLexemeInput) (ImportLexiconResult, error) {
	if _, err := s.languageIDByCode(ctx, code); err != nil {
		return ImportLexiconResult{}, err
	}
	var out ImportLexiconResult
	for _, item := range items {
		if item.Lemma == "" {
			out.Failed++
			continue
		}
		_, err := s.CreateLexeme(ctx, code, item)
		if err != nil {
			if errors.Is(err, domain.ErrConflict) {
				out.Skipped++
				continue
			}
			out.Failed++
			continue
		}
		out.Created++
	}
	return out, nil
}

// MapImportLexemeItem converts HTTP import item to service input.
func MapImportLexemeItem(lemma string, pos, notes *string, translations []struct {
	TargetLanguageID uuid.UUID
	Text             string
}) CreateLexemeInput {
	in := CreateLexemeInput{Lemma: lemma, PartOfSpeech: pos, Notes: notes}
	for _, t := range translations {
		in.Translations = append(in.Translations, struct {
			TargetLanguageID uuid.UUID
			Text             string
		}{TargetLanguageID: t.TargetLanguageID, Text: t.Text})
	}
	return in
}

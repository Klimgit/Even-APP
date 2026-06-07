package handler

import (
	"encoding/json"

	"github.com/even-app/even-app/services/lexicon/internal/gen/query"
	http_v1 "github.com/even-app/even-app/services/lexicon/internal/gen/http/v1"
	"github.com/even-app/even-app/services/lexicon/internal/service"
)

func mapLanguage(row query.Language) http_v1.Language {
	return http_v1.Language{
		ID: row.ID, Code: row.Code, Name: row.Name, NativeName: row.NativeName,
		Direction: http_v1.LanguageDirection(row.Direction), IsActive: row.IsActive,
	}
}

func mapLanguages(rows []query.Language) []http_v1.Language {
	out := make([]http_v1.Language, len(rows))
	for i, r := range rows {
		out[i] = mapLanguage(r)
	}
	return out
}

func mapAlphabetLetter(row query.AlphabetLetter) http_v1.AlphabetLetter {
	out := http_v1.AlphabetLetter{
		ID: row.ID, LanguageID: row.LanguageID, Character: row.Character, SortOrder: int(row.SortOrder),
	}
	if row.UpperChar != nil {
		out.UpperChar = http_v1.NewOptString(*row.UpperChar)
	}
	if row.Label != nil {
		out.Label = http_v1.NewOptString(*row.Label)
	}
	if row.Transcription != nil {
		out.Transcription = http_v1.NewOptString(*row.Transcription)
	}
	return out
}

func mapAlphabetLetters(rows []query.AlphabetLetter) []http_v1.AlphabetLetter {
	out := make([]http_v1.AlphabetLetter, len(rows))
	for i, r := range rows {
		out[i] = mapAlphabetLetter(r)
	}
	return out
}

func mapSound(row query.Sound) http_v1.Sound {
	out := http_v1.Sound{ID: row.ID, LanguageID: row.LanguageID}
	if row.Ipa != nil {
		out.Ipa = http_v1.NewOptString(*row.Ipa)
	}
	if row.Description != nil {
		out.Description = http_v1.NewOptString(*row.Description)
	}
	if row.AudioKey != nil && *row.AudioKey != "" {
		out.AudioURL = http_v1.NewOptString(*row.AudioKey)
	}
	return out
}

func mapSounds(rows []query.Sound) []http_v1.Sound {
	out := make([]http_v1.Sound, len(rows))
	for i, r := range rows {
		out[i] = mapSound(r)
	}
	return out
}

func parseFormTags(raw []byte) http_v1.LexemeFormTags {
	if len(raw) == 0 {
		return http_v1.LexemeFormTags{}
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil || m == nil {
		return http_v1.LexemeFormTags{}
	}
	return http_v1.LexemeFormTags(m)
}

func mapLexemeForm(row query.LexemeForm) http_v1.LexemeForm {
	out := http_v1.LexemeForm{ID: row.ID, LexemeID: row.LexemeID, Form: row.Form}
	tags := parseFormTags(row.Tags)
	if len(tags) > 0 {
		out.Tags = http_v1.NewOptLexemeFormTags(tags)
	}
	return out
}

func mapLexemeTranslation(row query.LexemeTranslation) http_v1.LexemeTranslation {
	out := http_v1.LexemeTranslation{
		ID: row.ID, TargetLanguageID: row.TargetLanguageID, Text: row.Text,
	}
	if row.TargetLexemeID != nil {
		out.TargetLexemeID = http_v1.NewOptUUID(*row.TargetLexemeID)
	}
	return out
}

func mapLexemeMedia(row query.LexemeMedium) http_v1.LexemeMedia {
	out := http_v1.LexemeMedia{
		ID: row.ID, Kind: http_v1.LexemeMediaKind(row.Kind),
		IsPrimary: row.IsPrimary, URL: service.MediaRefURL(row.MediaAssetID),
	}
	if row.Label != nil {
		out.Label = http_v1.NewOptString(*row.Label)
	}
	if row.FormID != nil {
		out.FormID = http_v1.NewOptUUID(*row.FormID)
	}
	return out
}

func mapFullLexeme(full service.FullLexeme) http_v1.Lexeme {
	out := http_v1.Lexeme{
		ID: full.Lexeme.ID, LanguageID: full.Lexeme.LanguageID, Lemma: full.Lexeme.Lemma,
		Translations: make([]http_v1.LexemeTranslation, len(full.Translations)),
		Forms:        make([]http_v1.LexemeForm, len(full.Forms)),
		Media:        make([]http_v1.LexemeMedia, len(full.Media)),
	}
	if full.Lexeme.PartOfSpeech != nil {
		out.PartOfSpeech = http_v1.NewOptString(*full.Lexeme.PartOfSpeech)
	}
	if full.Lexeme.Notes != nil {
		out.Notes = http_v1.NewOptString(*full.Lexeme.Notes)
	}
	for i, t := range full.Translations {
		out.Translations[i] = mapLexemeTranslation(t)
	}
	for i, f := range full.Forms {
		out.Forms[i] = mapLexemeForm(f)
	}
	for i, m := range full.Media {
		out.Media[i] = mapLexemeMedia(m)
		if m.IsPrimary && m.Kind == "image" {
			out.PrimaryImageURL = http_v1.NewOptString(service.MediaRefURL(m.MediaAssetID))
		}
		if m.IsPrimary && (m.Kind == "audio_word" || m.Kind == "audio_form") {
			out.PrimaryAudioURL = http_v1.NewOptString(service.MediaRefURL(m.MediaAssetID))
		}
	}
	return out
}

func mapFullLexemes(items []service.FullLexeme) []http_v1.Lexeme {
	out := make([]http_v1.Lexeme, len(items))
	for i, item := range items {
		out[i] = mapFullLexeme(item)
	}
	return out
}

func errBody(msg string) http_v1.ErrorResponse {
	return http_v1.ErrorResponse{Message: msg, Error: msg}
}

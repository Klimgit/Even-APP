package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/even-app/even-app/services/lexicon/internal/domain"
	"github.com/even-app/even-app/services/lexicon/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type LexiconService struct {
	q *query.Queries
}

func NewLexiconService(q *query.Queries) *LexiconService {
	return &LexiconService{q: q}
}

func mediaRefURL(id uuid.UUID) string {
	return fmt.Sprintf("/api/v1/platform/media/%s", id)
}

func isUniqueViolation(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "duplicate") || strings.Contains(msg, "unique")
}

func (s *LexiconService) languageIDByCode(ctx context.Context, code string) (uuid.UUID, error) {
	id, err := s.q.GetLanguageIDByCode(ctx, query.GetLanguageIDByCodeParams{Code: code})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, domain.ErrNotFound
		}
		return uuid.Nil, err
	}
	return id, nil
}

func (s *LexiconService) ListActiveLanguages(ctx context.Context) ([]query.Language, error) {
	return s.q.ListActiveLanguages(ctx)
}

func (s *LexiconService) GetLanguageByCode(ctx context.Context, code string) (query.Language, error) {
	row, err := s.q.GetLanguageByCode(ctx, query.GetLanguageByCodeParams{Code: code})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return query.Language{}, domain.ErrNotFound
		}
		return query.Language{}, err
	}
	if !row.IsActive {
		return query.Language{}, domain.ErrNotFound
	}
	return row, nil
}

func (s *LexiconService) ListAlphabetByCode(ctx context.Context, code string) ([]query.AlphabetLetter, error) {
	langID, err := s.languageIDByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return s.q.ListAlphabetByLanguageID(ctx, query.ListAlphabetByLanguageIDParams{LanguageID: langID})
}

func (s *LexiconService) ListAllLanguages(ctx context.Context) ([]query.Language, error) {
	return s.q.ListAllLanguages(ctx)
}

func (s *LexiconService) CreateLanguage(ctx context.Context, code, name, nativeName, direction string) (query.Language, error) {
	if direction != "rtl" {
		direction = "ltr"
	}
	row, err := s.q.CreateLanguage(ctx, query.CreateLanguageParams{
		Code: code, Name: name, NativeName: nativeName, Direction: direction,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return query.Language{}, domain.ErrConflict
		}
		return query.Language{}, err
	}
	return row, nil
}

func (s *LexiconService) PatchLanguage(ctx context.Context, code string, name, nativeName, direction *string, isActive *bool) (query.Language, error) {
	row, err := s.q.UpdateLanguage(ctx, query.UpdateLanguageParams{
		Code: code, Name: name, NativeName: nativeName, Direction: direction, IsActive: isActive,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return query.Language{}, domain.ErrNotFound
		}
		return query.Language{}, err
	}
	return row, nil
}

func (s *LexiconService) ListPlatformAlphabet(ctx context.Context, code string) ([]query.AlphabetLetter, error) {
	langID, err := s.languageIDByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return s.q.ListAlphabetByLanguageID(ctx, query.ListAlphabetByLanguageIDParams{LanguageID: langID})
}

func (s *LexiconService) CreateAlphabetLetter(ctx context.Context, code, character string, upperChar *string, sortOrder int, label, transcription *string) (query.AlphabetLetter, error) {
	char, err := validateLetterField("character", character, true)
	if err != nil {
		return query.AlphabetLetter{}, err
	}
	var upper *string
	if upperChar != nil && *upperChar != "" {
		upper, err = validateOptionalLetterField("upper_char", *upperChar)
		if err != nil {
			return query.AlphabetLetter{}, err
		}
	}
	var tr *string
	if transcription != nil {
		tr, err = validateOptionalTranscription(*transcription)
		if err != nil {
			return query.AlphabetLetter{}, err
		}
	}
	langID, err := s.languageIDByCode(ctx, code)
	if err != nil {
		return query.AlphabetLetter{}, err
	}
	row, err := s.q.CreateAlphabetLetter(ctx, query.CreateAlphabetLetterParams{
		LanguageID: langID, Character: char, UpperChar: upper,
		SortOrder: int32(sortOrder), Label: label, Transcription: tr,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return query.AlphabetLetter{}, domain.ErrConflict
		}
		return query.AlphabetLetter{}, err
	}
	return row, nil
}

func (s *LexiconService) PatchAlphabetLetter(ctx context.Context, letterID uuid.UUID, character *string, upperChar, label, transcription *string, sortOrder *int) (query.AlphabetLetter, error) {
	var char, upper *string
	var err error
	if character != nil {
		c, err := validateLetterField("character", *character, true)
		if err != nil {
			return query.AlphabetLetter{}, err
		}
		char = &c
	}
	if upperChar != nil {
		if *upperChar == "" {
			empty := ""
			upper = &empty
		} else {
			upper, err = validateOptionalLetterField("upper_char", *upperChar)
			if err != nil {
				return query.AlphabetLetter{}, err
			}
		}
	}
	var tr *string
	if transcription != nil {
		tr, err = validateOptionalTranscription(*transcription)
		if err != nil {
			return query.AlphabetLetter{}, err
		}
	}
	var so *int32
	if sortOrder != nil {
		v := int32(*sortOrder)
		so = &v
	}
	row, err := s.q.UpdateAlphabetLetter(ctx, query.UpdateAlphabetLetterParams{
		ID: letterID, Character: char, UpperChar: upper, SortOrder: so, Label: label, Transcription: tr,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return query.AlphabetLetter{}, domain.ErrNotFound
		}
		return query.AlphabetLetter{}, err
	}
	return row, nil
}

func (s *LexiconService) DeleteAlphabetLetter(ctx context.Context, letterID uuid.UUID) error {
	return s.q.DeleteAlphabetLetter(ctx, query.DeleteAlphabetLetterParams{ID: letterID})
}

func (s *LexiconService) ReorderAlphabet(ctx context.Context, code string, letterIDs []uuid.UUID) ([]query.AlphabetLetter, error) {
	langID, err := s.languageIDByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	all, err := s.q.ListAlphabetByLanguageID(ctx, query.ListAlphabetByLanguageIDParams{LanguageID: langID})
	if err != nil {
		return nil, err
	}
	byID := make(map[uuid.UUID]struct{}, len(all))
	for _, row := range all {
		byID[row.ID] = struct{}{}
	}
	seen := make(map[uuid.UUID]struct{}, len(letterIDs))
	for i, id := range letterIDs {
		if _, ok := byID[id]; !ok {
			return nil, domain.ErrNotFound
		}
		seen[id] = struct{}{}
		if err := s.q.SetAlphabetLetterSortOrder(ctx, query.SetAlphabetLetterSortOrderParams{
			ID: id, SortOrder: int32(i),
		}); err != nil {
			return nil, err
		}
	}
	next := len(letterIDs)
	for _, row := range all {
		if _, ok := seen[row.ID]; ok {
			continue
		}
		if err := s.q.SetAlphabetLetterSortOrder(ctx, query.SetAlphabetLetterSortOrderParams{
			ID: row.ID, SortOrder: int32(next),
		}); err != nil {
			return nil, err
		}
		next++
	}
	return s.q.ListAlphabetByLanguageID(ctx, query.ListAlphabetByLanguageIDParams{LanguageID: langID})
}

func (s *LexiconService) ListSounds(ctx context.Context, code string) ([]query.Sound, error) {
	langID, err := s.languageIDByCode(ctx, code)
	if err != nil {
		return nil, err
	}
	return s.q.ListSoundsByLanguageID(ctx, query.ListSoundsByLanguageIDParams{LanguageID: langID})
}

func (s *LexiconService) CreateSound(ctx context.Context, code string, ipa, description, audioKey *string) (query.Sound, error) {
	langID, err := s.languageIDByCode(ctx, code)
	if err != nil {
		return query.Sound{}, err
	}
	return s.q.CreateSound(ctx, query.CreateSoundParams{
		LanguageID: langID, Ipa: ipa, Description: description, AudioKey: audioKey,
	})
}

func (s *LexiconService) PatchSound(ctx context.Context, soundID uuid.UUID, ipa, description, audioKey *string) (query.Sound, error) {
	row, err := s.q.UpdateSound(ctx, query.UpdateSoundParams{
		ID: soundID, Ipa: ipa, Description: description, AudioKey: audioKey,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return query.Sound{}, domain.ErrNotFound
		}
		return query.Sound{}, err
	}
	return row, nil
}

func (s *LexiconService) DeleteSound(ctx context.Context, soundID uuid.UUID) error {
	return s.q.DeleteSound(ctx, query.DeleteSoundParams{ID: soundID})
}

func (s *LexiconService) LinkLetterSound(ctx context.Context, letterID, soundID uuid.UUID) error {
	if _, err := s.q.GetAlphabetLetter(ctx, query.GetAlphabetLetterParams{ID: letterID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if _, err := s.q.GetSound(ctx, query.GetSoundParams{ID: soundID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	return s.q.LinkLetterSound(ctx, query.LinkLetterSoundParams{LetterID: letterID, SoundID: soundID})
}

func (s *LexiconService) UnlinkLetterSound(ctx context.Context, letterID, soundID uuid.UUID) error {
	return s.q.UnlinkLetterSound(ctx, query.UnlinkLetterSoundParams{LetterID: letterID, SoundID: soundID})
}

type LexemeList struct {
	Items []FullLexeme
	Total int
}

type FullLexeme struct {
	Lexeme       query.Lexeme
	Forms        []query.LexemeForm
	Translations []query.LexemeTranslation
	Media        []query.LexemeMedium
}

func (s *LexiconService) assembleFullLexeme(ctx context.Context, row query.Lexeme) (FullLexeme, error) {
	forms, err := s.q.ListLexemeForms(ctx, query.ListLexemeFormsParams{LexemeID: row.ID})
	if err != nil {
		return FullLexeme{}, err
	}
	trans, err := s.q.ListLexemeTranslations(ctx, query.ListLexemeTranslationsParams{SourceLexemeID: row.ID})
	if err != nil {
		return FullLexeme{}, err
	}
	media, err := s.q.ListLexemeMedia(ctx, query.ListLexemeMediaParams{LexemeID: row.ID})
	if err != nil {
		return FullLexeme{}, err
	}
	return FullLexeme{Lexeme: row, Forms: forms, Translations: trans, Media: media}, nil
}

func (s *LexiconService) ListLexemes(ctx context.Context, code, search string, page, limit int) (LexemeList, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	langID, err := s.languageIDByCode(ctx, code)
	if err != nil {
		return LexemeList{}, err
	}
	var q *string
	if search != "" {
		q = &search
	}
	total, err := s.q.CountLexemesByLanguage(ctx, query.CountLexemesByLanguageParams{
		LanguageID: langID, Search: q,
	})
	if err != nil {
		return LexemeList{}, err
	}
	rows, err := s.q.ListLexemesByLanguage(ctx, query.ListLexemesByLanguageParams{
		LanguageID: langID, Search: q,
		Offset: int32((page - 1) * limit), Limit: int32(limit),
	})
	if err != nil {
		return LexemeList{}, err
	}
	items := make([]FullLexeme, 0, len(rows))
	for _, row := range rows {
		full, err := s.assembleFullLexeme(ctx, row)
		if err != nil {
			return LexemeList{}, err
		}
		items = append(items, full)
	}
	return LexemeList{Items: items, Total: int(total)}, nil
}

type CreateLexemeInput struct {
	Lemma        string
	PartOfSpeech *string
	Notes        *string
	CreatedBy    *uuid.UUID
	Translations []struct {
		TargetLanguageID uuid.UUID
		Text             string
	}
}

func (s *LexiconService) CreateLexeme(ctx context.Context, code string, in CreateLexemeInput) (FullLexeme, error) {
	langID, err := s.languageIDByCode(ctx, code)
	if err != nil {
		return FullLexeme{}, err
	}
	row, err := s.q.CreateLexeme(ctx, query.CreateLexemeParams{
		LanguageID: langID, Lemma: in.Lemma, PartOfSpeech: in.PartOfSpeech,
		Notes: in.Notes, CreatedBy: in.CreatedBy,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return FullLexeme{}, domain.ErrConflict
		}
		return FullLexeme{}, err
	}
	for _, t := range in.Translations {
		if _, err := s.q.CreateLexemeTranslation(ctx, query.CreateLexemeTranslationParams{
			SourceLexemeID: row.ID, TargetLanguageID: t.TargetLanguageID, Text: t.Text,
		}); err != nil {
			return FullLexeme{}, err
		}
	}
	return s.assembleFullLexeme(ctx, row)
}

func (s *LexiconService) GetLexeme(ctx context.Context, lexemeID uuid.UUID) (FullLexeme, error) {
	row, err := s.q.GetLexeme(ctx, query.GetLexemeParams{ID: lexemeID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FullLexeme{}, domain.ErrNotFound
		}
		return FullLexeme{}, err
	}
	return s.assembleFullLexeme(ctx, row)
}

func (s *LexiconService) PatchLexeme(ctx context.Context, lexemeID uuid.UUID, lemma, pos, notes *string) (FullLexeme, error) {
	row, err := s.q.UpdateLexeme(ctx, query.UpdateLexemeParams{
		ID: lexemeID, Lemma: lemma, PartOfSpeech: pos, Notes: notes,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return FullLexeme{}, domain.ErrNotFound
		}
		if isUniqueViolation(err) {
			return FullLexeme{}, domain.ErrConflict
		}
		return FullLexeme{}, err
	}
	return s.assembleFullLexeme(ctx, row)
}

func (s *LexiconService) DeleteLexeme(ctx context.Context, lexemeID uuid.UUID) error {
	return s.q.DeleteLexeme(ctx, query.DeleteLexemeParams{ID: lexemeID})
}

func tagsJSON(tags map[string]string) ([]byte, error) {
	if tags == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(tags)
}

func (s *LexiconService) CreateLexemeForm(ctx context.Context, lexemeID uuid.UUID, form string, tags map[string]string) (query.LexemeForm, error) {
	if _, err := s.q.GetLexeme(ctx, query.GetLexemeParams{ID: lexemeID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return query.LexemeForm{}, domain.ErrNotFound
		}
		return query.LexemeForm{}, err
	}
	raw, err := tagsJSON(tags)
	if err != nil {
		return query.LexemeForm{}, err
	}
	row, err := s.q.CreateLexemeForm(ctx, query.CreateLexemeFormParams{
		LexemeID: lexemeID, Form: form, Tags: raw,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return query.LexemeForm{}, domain.ErrConflict
		}
		return query.LexemeForm{}, err
	}
	return row, nil
}

func (s *LexiconService) PatchLexemeForm(ctx context.Context, formID uuid.UUID, form *string, tags map[string]string) (query.LexemeForm, error) {
	var raw []byte
	if tags != nil {
		var err error
		raw, err = tagsJSON(tags)
		if err != nil {
			return query.LexemeForm{}, err
		}
	}
	row, err := s.q.UpdateLexemeForm(ctx, query.UpdateLexemeFormParams{
		ID: formID, Form: form, Tags: raw,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return query.LexemeForm{}, domain.ErrNotFound
		}
		return query.LexemeForm{}, err
	}
	return row, nil
}

func (s *LexiconService) DeleteLexemeForm(ctx context.Context, formID uuid.UUID) error {
	return s.q.DeleteLexemeForm(ctx, query.DeleteLexemeFormParams{ID: formID})
}

func (s *LexiconService) CreateLexemeTranslation(ctx context.Context, lexemeID, targetLangID uuid.UUID, text string, targetLexemeID *uuid.UUID) (query.LexemeTranslation, error) {
	if _, err := s.q.GetLexeme(ctx, query.GetLexemeParams{ID: lexemeID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return query.LexemeTranslation{}, domain.ErrNotFound
		}
		return query.LexemeTranslation{}, err
	}
	return s.q.CreateLexemeTranslation(ctx, query.CreateLexemeTranslationParams{
		SourceLexemeID: lexemeID, TargetLanguageID: targetLangID,
		Text: text, TargetLexemeID: targetLexemeID,
	})
}

func (s *LexiconService) DeleteLexemeTranslation(ctx context.Context, translationID uuid.UUID) error {
	return s.q.DeleteLexemeTranslation(ctx, query.DeleteLexemeTranslationParams{ID: translationID})
}

func (s *LexiconService) CreateLexemeMedia(ctx context.Context, lexemeID uuid.UUID, mediaAssetID uuid.UUID, kind string, label *string, isPrimary bool, formID *uuid.UUID) (query.LexemeMedium, error) {
	if _, err := s.q.GetLexeme(ctx, query.GetLexemeParams{ID: lexemeID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return query.LexemeMedium{}, domain.ErrNotFound
		}
		return query.LexemeMedium{}, err
	}
	if isPrimary {
		_ = s.q.ClearPrimaryLexemeMedia(ctx, query.ClearPrimaryLexemeMediaParams{
			LexemeID: lexemeID, Kind: kind,
		})
	}
	return s.q.CreateLexemeMedia(ctx, query.CreateLexemeMediaParams{
		LexemeID: lexemeID, FormID: formID, MediaAssetID: mediaAssetID,
		Kind: kind, Label: label, IsPrimary: isPrimary, SortOrder: 0,
	})
}

func (s *LexiconService) DeleteLexemeMedia(ctx context.Context, lexemeMediaID uuid.UUID) error {
	return s.q.DeleteLexemeMedia(ctx, query.DeleteLexemeMediaParams{ID: lexemeMediaID})
}

// MediaRefURL exposes media path for handlers.
func MediaRefURL(id uuid.UUID) string { return mediaRefURL(id) }

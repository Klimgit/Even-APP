package service

import (
	"context"
	"errors"

	"github.com/even-app/even-app/services/lexicon/internal/domain"
	"github.com/even-app/even-app/services/lexicon/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (s *LexiconService) ListPickerLexemes(ctx context.Context, code string, ownerID uuid.UUID, search string, page, limit int) (LexemeList, error) {
	langID, err := s.languageIDByCode(ctx, code)
	if err != nil {
		return LexemeList{}, err
	}
	var q *string
	if search != "" {
		q = &search
	}
	total, err := s.q.CountPickerLexemes(ctx, query.CountPickerLexemesParams{
		LanguageID: langID, OwnerID: &ownerID, Search: q,
	})
	if err != nil {
		return LexemeList{}, err
	}
	rows, err := s.q.ListPickerLexemes(ctx, query.ListPickerLexemesParams{
		LanguageID: langID, OwnerID: &ownerID, Search: q,
		Offset: int32((page - 1) * limit), Limit: int32(limit),
	})
	if err != nil {
		return LexemeList{}, err
	}
	items := make([]FullLexeme, 0, len(rows))
	for _, row := range rows {
		full, err := s.assembleFullLexeme(ctx, lexemeFromPicker(row))
		if err != nil {
			return LexemeList{}, err
		}
		items = append(items, full)
	}
	return LexemeList{Items: items, Total: int(total)}, nil
}

func (s *LexiconService) ListTeacherLexemes(ctx context.Context, code string, ownerID uuid.UUID, search string, page, limit int) (LexemeList, error) {
	langID, err := s.languageIDByCode(ctx, code)
	if err != nil {
		return LexemeList{}, err
	}
	var q *string
	if search != "" {
		q = &search
	}
	total, err := s.q.CountLexemesByLanguage(ctx, query.CountLexemesByLanguageParams{
		LanguageID: langID, Scope: LexemeScopeTeacher, OwnerID: &ownerID, Search: q,
	})
	if err != nil {
		return LexemeList{}, err
	}
	rows, err := s.q.ListLexemesByLanguage(ctx, query.ListLexemesByLanguageParams{
		LanguageID: langID, Scope: LexemeScopeTeacher, OwnerID: &ownerID, Search: q,
		Offset: int32((page - 1) * limit), Limit: int32(limit),
	})
	if err != nil {
		return LexemeList{}, err
	}
	items := make([]FullLexeme, 0, len(rows))
	for _, row := range rows {
		full, err := s.assembleFullLexeme(ctx, lexemeFromList(row))
		if err != nil {
			return LexemeList{}, err
		}
		items = append(items, full)
	}
	return LexemeList{Items: items, Total: int(total)}, nil
}

func (s *LexiconService) CreateTeacherLexeme(ctx context.Context, code string, ownerID uuid.UUID, in CreateLexemeInput) (FullLexeme, error) {
	langID, err := s.languageIDByCode(ctx, code)
	if err != nil {
		return FullLexeme{}, err
	}
	row, err := s.q.CreateLexeme(ctx, query.CreateLexemeParams{
		LanguageID: langID, Lemma: in.Lemma, PartOfSpeech: in.PartOfSpeech,
		Notes: in.Notes, Scope: LexemeScopeTeacher, OwnerID: &ownerID, CreatedBy: &ownerID,
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
	return s.assembleFullLexeme(ctx, lexemeFromCreate(row))
}

func (s *LexiconService) assertTeacherLexemeOwner(ctx context.Context, lexemeID, ownerID uuid.UUID) (LexemeCore, error) {
	row, err := s.q.GetLexeme(ctx, query.GetLexemeParams{ID: lexemeID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LexemeCore{}, domain.ErrNotFound
		}
		return LexemeCore{}, err
	}
	core := lexemeFromGet(row)
	if core.Scope != LexemeScopeTeacher || core.OwnerID == nil || *core.OwnerID != ownerID {
		return LexemeCore{}, domain.ErrForbidden
	}
	return core, nil
}

func (s *LexiconService) PatchTeacherLexeme(ctx context.Context, lexemeID, ownerID uuid.UUID, lemma, pos, notes *string) (FullLexeme, error) {
	if _, err := s.assertTeacherLexemeOwner(ctx, lexemeID, ownerID); err != nil {
		return FullLexeme{}, err
	}
	return s.PatchLexeme(ctx, lexemeID, lemma, pos, notes)
}

func (s *LexiconService) DeleteTeacherLexeme(ctx context.Context, lexemeID, ownerID uuid.UUID) error {
	if _, err := s.assertTeacherLexemeOwner(ctx, lexemeID, ownerID); err != nil {
		return err
	}
	return s.DeleteLexeme(ctx, lexemeID)
}

func (s *LexiconService) assertTeacherLexemeFormOwner(ctx context.Context, formID, ownerID uuid.UUID) error {
	form, err := s.q.GetLexemeForm(ctx, query.GetLexemeFormParams{ID: formID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	_, err = s.assertTeacherLexemeOwner(ctx, form.LexemeID, ownerID)
	return err
}

func (s *LexiconService) CreateTeacherLexemeForm(ctx context.Context, lexemeID, ownerID uuid.UUID, form string, tags map[string]string) (query.LexemeForm, error) {
	if _, err := s.assertTeacherLexemeOwner(ctx, lexemeID, ownerID); err != nil {
		return query.LexemeForm{}, err
	}
	return s.CreateLexemeForm(ctx, lexemeID, form, tags)
}

func (s *LexiconService) PatchTeacherLexemeForm(ctx context.Context, formID, ownerID uuid.UUID, form *string, tags map[string]string) (query.LexemeForm, error) {
	if err := s.assertTeacherLexemeFormOwner(ctx, formID, ownerID); err != nil {
		return query.LexemeForm{}, err
	}
	return s.PatchLexemeForm(ctx, formID, form, tags)
}

func (s *LexiconService) DeleteTeacherLexemeForm(ctx context.Context, formID, ownerID uuid.UUID) error {
	if err := s.assertTeacherLexemeFormOwner(ctx, formID, ownerID); err != nil {
		return err
	}
	return s.DeleteLexemeForm(ctx, formID)
}

func (s *LexiconService) CreateTeacherLexemeTranslation(ctx context.Context, lexemeID, ownerID, targetLangID uuid.UUID, text string, targetLexemeID *uuid.UUID) (query.LexemeTranslation, error) {
	if _, err := s.assertTeacherLexemeOwner(ctx, lexemeID, ownerID); err != nil {
		return query.LexemeTranslation{}, err
	}
	return s.CreateLexemeTranslation(ctx, lexemeID, targetLangID, text, targetLexemeID)
}

func (s *LexiconService) DeleteTeacherLexemeTranslation(ctx context.Context, translationID, ownerID uuid.UUID) error {
	tr, err := s.q.GetLexemeTranslation(ctx, query.GetLexemeTranslationParams{ID: translationID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if _, err := s.assertTeacherLexemeOwner(ctx, tr.SourceLexemeID, ownerID); err != nil {
		return err
	}
	return s.DeleteLexemeTranslation(ctx, translationID)
}

func (s *LexiconService) PatchTeacherLexemeTranslation(ctx context.Context, translationID, ownerID uuid.UUID, text *string, targetLangID *uuid.UUID, targetLexemeID *uuid.UUID) (query.LexemeTranslation, error) {
	tr, err := s.q.GetLexemeTranslation(ctx, query.GetLexemeTranslationParams{ID: translationID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return query.LexemeTranslation{}, domain.ErrNotFound
		}
		return query.LexemeTranslation{}, err
	}
	if _, err := s.assertTeacherLexemeOwner(ctx, tr.SourceLexemeID, ownerID); err != nil {
		return query.LexemeTranslation{}, err
	}
	return s.PatchLexemeTranslation(ctx, translationID, text, targetLangID, targetLexemeID)
}

func (s *LexiconService) CreateTeacherLexemeMedia(ctx context.Context, lexemeID, ownerID uuid.UUID, mediaAssetID uuid.UUID, kind string, label *string, isPrimary bool, formID *uuid.UUID) (query.LexemeMedium, error) {
	if _, err := s.assertTeacherLexemeOwner(ctx, lexemeID, ownerID); err != nil {
		return query.LexemeMedium{}, err
	}
	if s.media != nil && s.media.Available() {
		if err := s.media.AssertTeacherMediaOwner(ctx, mediaAssetID, ownerID); err != nil {
			return query.LexemeMedium{}, err
		}
	}
	return s.CreateLexemeMedia(ctx, lexemeID, mediaAssetID, kind, label, isPrimary, formID)
}

func (s *LexiconService) DeleteTeacherLexemeMedia(ctx context.Context, lexemeMediaID, ownerID uuid.UUID) error {
	m, err := s.q.GetLexemeMedia(ctx, query.GetLexemeMediaParams{ID: lexemeMediaID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		return err
	}
	if _, err := s.assertTeacherLexemeOwner(ctx, m.LexemeID, ownerID); err != nil {
		return err
	}
	return s.DeleteLexemeMedia(ctx, lexemeMediaID)
}

func (s *LexiconService) CountPlatformLexemes(ctx context.Context) (int32, error) {
	return s.q.CountLexemes(ctx)
}

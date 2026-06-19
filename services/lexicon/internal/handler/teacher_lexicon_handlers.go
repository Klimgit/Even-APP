package handler

import (
	"context"

	http_v1 "github.com/even-app/even-app/services/lexicon/internal/gen/http/v1"
	"github.com/even-app/even-app/services/lexicon/internal/service"
	"github.com/google/uuid"
)

func (h *HTTPHandler) CreateTeacherLexeme(ctx context.Context, req *http_v1.CreateLexemeRequest, params http_v1.CreateTeacherLexemeParams) (http_v1.CreateTeacherLexemeRes, error) {
	claims, err := h.requireTeacher(ctx)
	if err != nil {
		return nil, err
	}
	in := service.CreateLexemeInput{Lemma: req.Lemma, CreatedBy: &claims.UserID}
	if v, ok := req.PartOfSpeech.Get(); ok {
		in.PartOfSpeech = &v
	}
	if v, ok := req.Notes.Get(); ok {
		in.Notes = &v
	}
	for _, t := range req.Translations {
		in.Translations = append(in.Translations, struct {
			TargetLanguageID uuid.UUID
			Text             string
		}{TargetLanguageID: t.TargetLanguageID, Text: t.Text})
	}
	full, err := h.svc.CreateTeacherLexeme(ctx, params.Code, claims.UserID, in)
	if err != nil {
		return nil, err
	}
	out := mapFullLexeme(full)
	return &out, nil
}

func (h *HTTPHandler) PatchTeacherLexeme(ctx context.Context, req *http_v1.PatchLexemeRequest, params http_v1.PatchTeacherLexemeParams) (http_v1.PatchTeacherLexemeRes, error) {
	claims, err := h.requireTeacher(ctx)
	if err != nil {
		return nil, err
	}
	var lemma, pos, notes *string
	if v, ok := req.Lemma.Get(); ok {
		lemma = &v
	}
	if v, ok := req.PartOfSpeech.Get(); ok {
		pos = &v
	}
	if v, ok := req.Notes.Get(); ok {
		notes = &v
	}
	full, err := h.svc.PatchTeacherLexeme(ctx, params.LexemeId, claims.UserID, lemma, pos, notes)
	if err != nil {
		return nil, err
	}
	out := mapFullLexeme(full)
	return &out, nil
}

func (h *HTTPHandler) DeleteTeacherLexeme(ctx context.Context, params http_v1.DeleteTeacherLexemeParams) (http_v1.DeleteTeacherLexemeRes, error) {
	claims, err := h.requireTeacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.DeleteTeacherLexeme(ctx, params.LexemeId, claims.UserID); err != nil {
		return nil, err
	}
	return &http_v1.DeleteTeacherLexemeNoContent{}, nil
}

func (h *HTTPHandler) CreateTeacherLexemeForm(ctx context.Context, req *http_v1.CreateLexemeFormRequest, params http_v1.CreateTeacherLexemeFormParams) (http_v1.CreateTeacherLexemeFormRes, error) {
	claims, err := h.requireTeacher(ctx)
	if err != nil {
		return nil, err
	}
	var tags map[string]string
	if v, ok := req.Tags.Get(); ok {
		tags = map[string]string(v)
	}
	row, err := h.svc.CreateTeacherLexemeForm(ctx, params.LexemeId, claims.UserID, req.Form, tags)
	if err != nil {
		return nil, err
	}
	out := mapLexemeForm(row)
	return &out, nil
}

func (h *HTTPHandler) PatchTeacherLexemeForm(ctx context.Context, req *http_v1.PatchLexemeFormRequest, params http_v1.PatchTeacherLexemeFormParams) (http_v1.PatchTeacherLexemeFormRes, error) {
	claims, err := h.requireTeacher(ctx)
	if err != nil {
		return nil, err
	}
	var form *string
	var tags map[string]string
	if v, ok := req.Form.Get(); ok {
		form = &v
	}
	if v, ok := req.Tags.Get(); ok {
		tags = map[string]string(v)
	}
	row, err := h.svc.PatchTeacherLexemeForm(ctx, params.FormId, claims.UserID, form, tags)
	if err != nil {
		return nil, err
	}
	out := mapLexemeForm(row)
	return &out, nil
}

func (h *HTTPHandler) DeleteTeacherLexemeForm(ctx context.Context, params http_v1.DeleteTeacherLexemeFormParams) (http_v1.DeleteTeacherLexemeFormRes, error) {
	claims, err := h.requireTeacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.DeleteTeacherLexemeForm(ctx, params.FormId, claims.UserID); err != nil {
		return nil, err
	}
	return &http_v1.DeleteTeacherLexemeFormNoContent{}, nil
}

func (h *HTTPHandler) CreateTeacherLexemeTranslation(ctx context.Context, req *http_v1.CreateLexemeTranslationRequest, params http_v1.CreateTeacherLexemeTranslationParams) (http_v1.CreateTeacherLexemeTranslationRes, error) {
	claims, err := h.requireTeacher(ctx)
	if err != nil {
		return nil, err
	}
	var targetLex *uuid.UUID
	if v, ok := req.TargetLexemeID.Get(); ok {
		targetLex = &v
	}
	row, err := h.svc.CreateTeacherLexemeTranslation(ctx, params.LexemeId, claims.UserID, req.TargetLanguageID, req.Text, targetLex)
	if err != nil {
		return nil, err
	}
	out := mapLexemeTranslation(row)
	return &out, nil
}

func (h *HTTPHandler) DeleteTeacherLexemeTranslation(ctx context.Context, params http_v1.DeleteTeacherLexemeTranslationParams) (http_v1.DeleteTeacherLexemeTranslationRes, error) {
	claims, err := h.requireTeacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.DeleteTeacherLexemeTranslation(ctx, params.TranslationId, claims.UserID); err != nil {
		return nil, err
	}
	return &http_v1.DeleteTeacherLexemeTranslationNoContent{}, nil
}

func (h *HTTPHandler) PatchTeacherLexemeTranslation(ctx context.Context, req *http_v1.PatchLexemeTranslationRequest, params http_v1.PatchTeacherLexemeTranslationParams) (http_v1.PatchTeacherLexemeTranslationRes, error) {
	claims, err := h.requireTeacher(ctx)
	if err != nil {
		return nil, err
	}
	var text *string
	var targetLangID, targetLexID *uuid.UUID
	if v, ok := req.Text.Get(); ok {
		text = &v
	}
	if v, ok := req.TargetLanguageID.Get(); ok {
		targetLangID = &v
	}
	if v, ok := req.TargetLexemeID.Get(); ok {
		targetLexID = &v
	}
	row, err := h.svc.PatchTeacherLexemeTranslation(ctx, params.TranslationId, claims.UserID, text, targetLangID, targetLexID)
	if err != nil {
		return nil, err
	}
	out := mapLexemeTranslation(row)
	return &out, nil
}

func (h *HTTPHandler) CreateTeacherLexemeMedia(ctx context.Context, req *http_v1.CreateLexemeMediaRequest, params http_v1.CreateTeacherLexemeMediaParams) (http_v1.CreateTeacherLexemeMediaRes, error) {
	claims, err := h.requireTeacher(ctx)
	if err != nil {
		return nil, err
	}
	var label *string
	if v, ok := req.Label.Get(); ok {
		label = &v
	}
	var formID *uuid.UUID
	if v, ok := req.FormID.Get(); ok {
		formID = &v
	}
	isPrimary := false
	if v, ok := req.IsPrimary.Get(); ok {
		isPrimary = v
	}
	row, err := h.svc.CreateTeacherLexemeMedia(ctx, params.LexemeId, claims.UserID, req.MediaAssetID, string(req.Kind), label, isPrimary, formID)
	if err != nil {
		return nil, err
	}
	out := mapLexemeMedia(row)
	return &out, nil
}

func (h *HTTPHandler) DeleteTeacherLexemeMedia(ctx context.Context, params http_v1.DeleteTeacherLexemeMediaParams) (http_v1.DeleteTeacherLexemeMediaRes, error) {
	claims, err := h.requireTeacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.DeleteTeacherLexemeMedia(ctx, params.LexemeMediaId, claims.UserID); err != nil {
		return nil, err
	}
	return &http_v1.DeleteTeacherLexemeMediaNoContent{}, nil
}

func (h *HTTPHandler) PatchPlatformLexemeTranslation(ctx context.Context, req *http_v1.PatchLexemeTranslationRequest, params http_v1.PatchPlatformLexemeTranslationParams) (http_v1.PatchPlatformLexemeTranslationRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	var text *string
	var targetLangID, targetLexID *uuid.UUID
	if v, ok := req.Text.Get(); ok {
		text = &v
	}
	if v, ok := req.TargetLanguageID.Get(); ok {
		targetLangID = &v
	}
	if v, ok := req.TargetLexemeID.Get(); ok {
		targetLexID = &v
	}
	row, err := h.svc.PatchLexemeTranslation(ctx, params.TranslationId, text, targetLangID, targetLexID)
	if err != nil {
		return nil, err
	}
	out := mapLexemeTranslation(row)
	return &out, nil
}

func (h *HTTPHandler) PatchPlatformLexemeMedia(ctx context.Context, req *http_v1.PatchLexemeMediaRequest, params http_v1.PatchPlatformLexemeMediaParams) (http_v1.PatchPlatformLexemeMediaRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	var kind, label *string
	var isPrimary *bool
	var mediaAssetID, formID *uuid.UUID
	if v, ok := req.Kind.Get(); ok {
		s := string(v)
		kind = &s
	}
	if v, ok := req.Label.Get(); ok {
		label = &v
	}
	if v, ok := req.IsPrimary.Get(); ok {
		isPrimary = &v
	}
	if v, ok := req.MediaAssetID.Get(); ok {
		mediaAssetID = &v
	}
	if v, ok := req.FormID.Get(); ok {
		formID = &v
	}
	row, err := h.svc.PatchLexemeMedia(ctx, params.LexemeMediaId, kind, label, isPrimary, mediaAssetID, formID)
	if err != nil {
		return nil, err
	}
	out := mapLexemeMedia(row)
	return &out, nil
}

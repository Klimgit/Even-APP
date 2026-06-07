package handler

import (
	"context"
	"errors"

	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/google/uuid"
	"github.com/even-app/even-app/services/lexicon/internal/domain"
	http_v1 "github.com/even-app/even-app/services/lexicon/internal/gen/http/v1"
	"github.com/even-app/even-app/services/lexicon/internal/service"
)

var _ http_v1.Handler = (*HTTPHandler)(nil)

type HTTPHandler struct {
	svc *service.LexiconService
}

func NewHTTPHandler(svc *service.LexiconService) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

func (h *HTTPHandler) requireAdmin(ctx context.Context) error {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return domain.ErrUnauthorized
	}
	if !claims.IsAdmin {
		return domain.ErrForbidden
	}
	return nil
}

func (h *HTTPHandler) ListLanguages(ctx context.Context) ([]http_v1.Language, error) {
	rows, err := h.svc.ListActiveLanguages(ctx)
	if err != nil {
		return nil, err
	}
	return mapLanguages(rows), nil
}

func (h *HTTPHandler) GetLanguage(ctx context.Context, params http_v1.GetLanguageParams) (http_v1.GetLanguageRes, error) {
	row, err := h.svc.GetLanguageByCode(ctx, params.Code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return notFound("language not found"), nil
		}
		return nil, err
	}
	out := mapLanguage(row)
	return &out, nil
}

func (h *HTTPHandler) GetLanguageAlphabet(ctx context.Context, params http_v1.GetLanguageAlphabetParams) (http_v1.GetLanguageAlphabetRes, error) {
	rows, err := h.svc.ListAlphabetByCode(ctx, params.Code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return notFound("language not found"), nil
		}
		return nil, err
	}
	out := http_v1.GetLanguageAlphabetOKApplicationJSON(mapAlphabetLetters(rows))
	return &out, nil
}

func (h *HTTPHandler) ListPlatformLanguages(ctx context.Context) (http_v1.ListPlatformLanguagesRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	rows, err := h.svc.ListAllLanguages(ctx)
	if err != nil {
		return nil, err
	}
	out := http_v1.ListPlatformLanguagesOKApplicationJSON(mapLanguages(rows))
	return &out, nil
}

func (h *HTTPHandler) CreatePlatformLanguage(ctx context.Context, req *http_v1.CreateLanguageRequest) (http_v1.CreatePlatformLanguageRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	dir := "ltr"
	if v, ok := req.Direction.Get(); ok {
		dir = string(v)
	}
	row, err := h.svc.CreateLanguage(ctx, req.Code, req.Name, req.NativeName, dir)
	if err != nil {
		return nil, err
	}
	out := mapLanguage(row)
	return &out, nil
}

func (h *HTTPHandler) PatchPlatformLanguage(ctx context.Context, req *http_v1.PatchLanguageRequest, params http_v1.PatchPlatformLanguageParams) (http_v1.PatchPlatformLanguageRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	var name, native, dir *string
	var active *bool
	if v, ok := req.Name.Get(); ok {
		name = &v
	}
	if v, ok := req.NativeName.Get(); ok {
		native = &v
	}
	if v, ok := req.Direction.Get(); ok {
		s := string(v)
		dir = &s
	}
	if v, ok := req.IsActive.Get(); ok {
		active = &v
	}
	row, err := h.svc.PatchLanguage(ctx, params.Code, name, native, dir, active)
	if err != nil {
		return nil, err
	}
	out := mapLanguage(row)
	return &out, nil
}

func (h *HTTPHandler) ListPlatformAlphabet(ctx context.Context, params http_v1.ListPlatformAlphabetParams) (http_v1.ListPlatformAlphabetRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	rows, err := h.svc.ListPlatformAlphabet(ctx, params.Code)
	if err != nil {
		return nil, err
	}
	out := http_v1.ListPlatformAlphabetOKApplicationJSON(mapAlphabetLetters(rows))
	return &out, nil
}

func (h *HTTPHandler) CreatePlatformAlphabetLetter(ctx context.Context, req *http_v1.CreateAlphabetLetterRequest, params http_v1.CreatePlatformAlphabetLetterParams) (http_v1.CreatePlatformAlphabetLetterRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	var upper, label, transcription *string
	if v, ok := req.UpperChar.Get(); ok {
		upper = &v
	}
	if v, ok := req.Label.Get(); ok {
		label = &v
	}
	if v, ok := req.Transcription.Get(); ok {
		transcription = &v
	}
	row, err := h.svc.CreateAlphabetLetter(ctx, params.Code, req.Character, upper, req.SortOrder, label, transcription)
	if err != nil {
		return nil, err
	}
	out := mapAlphabetLetter(row)
	return &out, nil
}

func (h *HTTPHandler) PatchPlatformAlphabetLetter(ctx context.Context, req *http_v1.PatchAlphabetLetterRequest, params http_v1.PatchPlatformAlphabetLetterParams) (http_v1.PatchPlatformAlphabetLetterRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	var char, upper, label, transcription *string
	var sort *int
	if v, ok := req.Character.Get(); ok {
		char = &v
	}
	if v, ok := req.UpperChar.Get(); ok {
		upper = &v
	}
	if v, ok := req.Label.Get(); ok {
		label = &v
	}
	if v, ok := req.Transcription.Get(); ok {
		transcription = &v
	}
	if v, ok := req.SortOrder.Get(); ok {
		sort = &v
	}
	row, err := h.svc.PatchAlphabetLetter(ctx, params.LetterId, char, upper, label, transcription, sort)
	if err != nil {
		return nil, err
	}
	out := mapAlphabetLetter(row)
	return &out, nil
}

func (h *HTTPHandler) DeletePlatformAlphabetLetter(ctx context.Context, params http_v1.DeletePlatformAlphabetLetterParams) (http_v1.DeletePlatformAlphabetLetterRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := h.svc.DeleteAlphabetLetter(ctx, params.LetterId); err != nil {
		return nil, err
	}
	return &http_v1.DeletePlatformAlphabetLetterNoContent{}, nil
}

func (h *HTTPHandler) ReorderPlatformAlphabet(ctx context.Context, req *http_v1.ReorderAlphabetRequest, params http_v1.ReorderPlatformAlphabetParams) (http_v1.ReorderPlatformAlphabetRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	rows, err := h.svc.ReorderAlphabet(ctx, params.Code, req.LetterIds)
	if err != nil {
		return nil, err
	}
	out := http_v1.ReorderPlatformAlphabetOKApplicationJSON(mapAlphabetLetters(rows))
	return &out, nil
}

func (h *HTTPHandler) ListPlatformSounds(ctx context.Context, params http_v1.ListPlatformSoundsParams) (http_v1.ListPlatformSoundsRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	rows, err := h.svc.ListSounds(ctx, params.Code)
	if err != nil {
		return nil, err
	}
	out := http_v1.ListPlatformSoundsOKApplicationJSON(mapSounds(rows))
	return &out, nil
}

func (h *HTTPHandler) CreatePlatformSound(ctx context.Context, req *http_v1.CreateSoundRequest, params http_v1.CreatePlatformSoundParams) (http_v1.CreatePlatformSoundRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	var ipa, desc, key *string
	if v, ok := req.Ipa.Get(); ok {
		ipa = &v
	}
	if v, ok := req.Description.Get(); ok {
		desc = &v
	}
	if v, ok := req.AudioKey.Get(); ok {
		key = &v
	}
	row, err := h.svc.CreateSound(ctx, params.Code, ipa, desc, key)
	if err != nil {
		return nil, err
	}
	out := mapSound(row)
	return &out, nil
}

func (h *HTTPHandler) PatchPlatformSound(ctx context.Context, req *http_v1.PatchSoundRequest, params http_v1.PatchPlatformSoundParams) (http_v1.PatchPlatformSoundRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	var ipa, desc, key *string
	if v, ok := req.Ipa.Get(); ok {
		ipa = &v
	}
	if v, ok := req.Description.Get(); ok {
		desc = &v
	}
	if v, ok := req.AudioKey.Get(); ok {
		key = &v
	}
	row, err := h.svc.PatchSound(ctx, params.SoundId, ipa, desc, key)
	if err != nil {
		return nil, err
	}
	out := mapSound(row)
	return &out, nil
}

func (h *HTTPHandler) DeletePlatformSound(ctx context.Context, params http_v1.DeletePlatformSoundParams) (http_v1.DeletePlatformSoundRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := h.svc.DeleteSound(ctx, params.SoundId); err != nil {
		return nil, err
	}
	return &http_v1.DeletePlatformSoundNoContent{}, nil
}

func (h *HTTPHandler) LinkPlatformAlphabetSound(ctx context.Context, req *http_v1.LinkSoundRequest, params http_v1.LinkPlatformAlphabetSoundParams) (http_v1.LinkPlatformAlphabetSoundRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := h.svc.LinkLetterSound(ctx, params.LetterId, req.SoundID); err != nil {
		return nil, err
	}
	return &http_v1.LinkPlatformAlphabetSoundNoContent{}, nil
}

func (h *HTTPHandler) UnlinkPlatformAlphabetSound(ctx context.Context, params http_v1.UnlinkPlatformAlphabetSoundParams) (http_v1.UnlinkPlatformAlphabetSoundRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := h.svc.UnlinkLetterSound(ctx, params.LetterId, params.SoundId); err != nil {
		return nil, err
	}
	return &http_v1.UnlinkPlatformAlphabetSoundNoContent{}, nil
}

func (h *HTTPHandler) ListPlatformLexicon(ctx context.Context, params http_v1.ListPlatformLexiconParams) (http_v1.ListPlatformLexiconRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	page, limit := 1, 20
	if v, ok := params.Page.Get(); ok {
		page = v
	}
	if v, ok := params.Limit.Get(); ok {
		limit = v
	}
	q := ""
	if v, ok := params.Q.Get(); ok {
		q = v
	}
	result, err := h.svc.ListLexemes(ctx, params.Code, q, page, limit)
	if err != nil {
		return nil, err
	}
	return &http_v1.LexemeListResponse{
		Items: mapFullLexemes(result.Items),
		Total: result.Total,
		Page:  page,
		Limit: limit,
	}, nil
}

func (h *HTTPHandler) CreatePlatformLexeme(ctx context.Context, req *http_v1.CreateLexemeRequest, params http_v1.CreatePlatformLexemeParams) (http_v1.CreatePlatformLexemeRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	claims, _ := middleware.ClaimsFromContext(ctx)
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
		}{
			TargetLanguageID: t.TargetLanguageID,
			Text:             t.Text,
		})
	}
	full, err := h.svc.CreateLexeme(ctx, params.Code, in)
	if err != nil {
		return nil, err
	}
	out := mapFullLexeme(full)
	return &out, nil
}

func (h *HTTPHandler) GetPlatformLexeme(ctx context.Context, params http_v1.GetPlatformLexemeParams) (http_v1.GetPlatformLexemeRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	full, err := h.svc.GetLexeme(ctx, params.LexemeId)
	if err != nil {
		return nil, err
	}
	out := mapFullLexeme(full)
	return &out, nil
}

func (h *HTTPHandler) PatchPlatformLexeme(ctx context.Context, req *http_v1.PatchLexemeRequest, params http_v1.PatchPlatformLexemeParams) (http_v1.PatchPlatformLexemeRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
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
	full, err := h.svc.PatchLexeme(ctx, params.LexemeId, lemma, pos, notes)
	if err != nil {
		return nil, err
	}
	out := mapFullLexeme(full)
	return &out, nil
}

func (h *HTTPHandler) DeletePlatformLexeme(ctx context.Context, params http_v1.DeletePlatformLexemeParams) (http_v1.DeletePlatformLexemeRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := h.svc.DeleteLexeme(ctx, params.LexemeId); err != nil {
		return nil, err
	}
	return &http_v1.DeletePlatformLexemeNoContent{}, nil
}

func (h *HTTPHandler) CreatePlatformLexemeForm(ctx context.Context, req *http_v1.CreateLexemeFormRequest, params http_v1.CreatePlatformLexemeFormParams) (http_v1.CreatePlatformLexemeFormRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	var tags map[string]string
	if v, ok := req.Tags.Get(); ok {
		tags = map[string]string(v)
	}
	row, err := h.svc.CreateLexemeForm(ctx, params.LexemeId, req.Form, tags)
	if err != nil {
		return nil, err
	}
	out := mapLexemeForm(row)
	return &out, nil
}

func (h *HTTPHandler) PatchPlatformLexemeForm(ctx context.Context, req *http_v1.PatchLexemeFormRequest, params http_v1.PatchPlatformLexemeFormParams) (http_v1.PatchPlatformLexemeFormRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
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
	row, err := h.svc.PatchLexemeForm(ctx, params.FormId, form, tags)
	if err != nil {
		return nil, err
	}
	out := mapLexemeForm(row)
	return &out, nil
}

func (h *HTTPHandler) DeletePlatformLexemeForm(ctx context.Context, params http_v1.DeletePlatformLexemeFormParams) (http_v1.DeletePlatformLexemeFormRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := h.svc.DeleteLexemeForm(ctx, params.FormId); err != nil {
		return nil, err
	}
	return &http_v1.DeletePlatformLexemeFormNoContent{}, nil
}

func (h *HTTPHandler) CreatePlatformLexemeTranslation(ctx context.Context, req *http_v1.CreateLexemeTranslationRequest, params http_v1.CreatePlatformLexemeTranslationParams) (http_v1.CreatePlatformLexemeTranslationRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	var targetLex *uuid.UUID
	if v, ok := req.TargetLexemeID.Get(); ok {
		targetLex = &v
	}
	row, err := h.svc.CreateLexemeTranslation(ctx, params.LexemeId, req.TargetLanguageID, req.Text, targetLex)
	if err != nil {
		return nil, err
	}
	out := mapLexemeTranslation(row)
	return &out, nil
}

func (h *HTTPHandler) DeletePlatformLexemeTranslation(ctx context.Context, params http_v1.DeletePlatformLexemeTranslationParams) (http_v1.DeletePlatformLexemeTranslationRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := h.svc.DeleteLexemeTranslation(ctx, params.TranslationId); err != nil {
		return nil, err
	}
	return &http_v1.DeletePlatformLexemeTranslationNoContent{}, nil
}

func (h *HTTPHandler) CreatePlatformLexemeMedia(ctx context.Context, req *http_v1.CreateLexemeMediaRequest, params http_v1.CreatePlatformLexemeMediaParams) (http_v1.CreatePlatformLexemeMediaRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
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
	row, err := h.svc.CreateLexemeMedia(ctx, params.LexemeId, req.MediaAssetID, string(req.Kind), label, isPrimary, formID)
	if err != nil {
		return nil, err
	}
	out := mapLexemeMedia(row)
	return &out, nil
}

func (h *HTTPHandler) DeletePlatformLexemeMedia(ctx context.Context, params http_v1.DeletePlatformLexemeMediaParams) (http_v1.DeletePlatformLexemeMediaRes, error) {
	if err := h.requireAdmin(ctx); err != nil {
		return nil, err
	}
	if err := h.svc.DeleteLexemeMedia(ctx, params.LexemeMediaId); err != nil {
		return nil, err
	}
	return &http_v1.DeletePlatformLexemeMediaNoContent{}, nil
}

func notFound(msg string) *http_v1.ErrorResponse {
	r := errBody(msg)
	return &r
}

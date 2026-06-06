package handler

import (
	"context"
	"time"

	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/services/lexicon/internal/domain"
	http_v1 "github.com/even-app/even-app/services/lexicon/internal/gen/http/v1"
	"github.com/even-app/even-app/services/lexicon/internal/service"
	"github.com/google/uuid"
)

var _ http_v1.Handler = (*HTTPHandler)(nil)

type HTTPHandler struct {
	svc *service.MediaService
}

func NewHTTPHandler(svc *service.MediaService) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

func (h *HTTPHandler) PlatformMediaPresign(ctx context.Context, req *http_v1.PresignRequest) (http_v1.PlatformMediaPresignRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok || !claims.IsAdmin {
		return forbiddenPresign()
	}
	result, err := h.svc.Presign(ctx, domain.PresignInput{
		Filename: req.Filename, MimeType: req.MimeType, SizeBytes: req.SizeBytes,
		LanguageID: req.LanguageID.String(), UserID: claims.UserID, IsAdmin: claims.IsAdmin,
	})
	if err != nil {
		switch err.Error() {
		case "size_bytes required", "language_id required":
			return badRequestPresign(err.Error())
		case "media storage quota exceeded", "quota check failed":
			return forbiddenPresign()
		default:
			return nil, err
		}
	}
	id, _ := uuid.Parse(result.MediaAssetID)
	return &http_v1.PresignResponse{
		UploadURL: result.UploadURL, ObjectKey: result.ObjectKey, MediaAssetID: id,
	}, nil
}

func (h *HTTPHandler) PlatformMediaConfirm(ctx context.Context, req *http_v1.ConfirmRequest) (http_v1.PlatformMediaConfirmRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok || !claims.IsAdmin {
		return forbiddenConfirm()
	}
	a, err := h.svc.Confirm(ctx, confirmInput(req, claims.UserID, claims.IsAdmin))
	if err != nil {
		switch err.Error() {
		case "media storage quota exceeded", "quota check failed":
			return forbiddenConfirm()
		case "size_bytes required", "invalid linked_lexeme_id":
			return badRequestConfirm(err.Error())
		default:
			return badRequestConfirm(err.Error())
		}
	}
	asset, err := h.assetFromDomain(ctx, a)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (h *HTTPHandler) ListPlatformMedia(ctx context.Context, params http_v1.ListPlatformMediaParams) (http_v1.ListPlatformMediaRes, error) {
	langID, err := h.svc.LanguageIDByCode(ctx, params.Code)
	if err != nil {
		return notFoundList("language not found")
	}
	page, limit := 0, 0
	if v, ok := params.Page.Get(); ok {
		page = v
	}
	if v, ok := params.Limit.Get(); ok {
		limit = v
	}
	q, kind := "", ""
	if v, ok := params.Q.Get(); ok {
		q = v
	}
	if v, ok := params.Kind.Get(); ok {
		kind = v
	}
	items, total, err := h.svc.List(ctx, domain.ListFilter{
		LanguageID: langID, Query: q, Kind: kind, Page: page, Limit: limit,
	})
	if err != nil {
		return nil, err
	}
	dtos := make([]http_v1.MediaAsset, 0, len(items))
	for i := range items {
		asset, err := h.assetFromDomain(ctx, &items[i])
		if err != nil {
			return nil, err
		}
		dtos = append(dtos, asset)
	}
	return &http_v1.MediaListResponse{
		Items: dtos, Total: total, Page: max1(page), Limit: maxLimit(limit),
	}, nil
}

func (h *HTTPHandler) GetPlatformMedia(ctx context.Context, params http_v1.GetPlatformMediaParams) (http_v1.GetPlatformMediaRes, error) {
	a, err := h.svc.GetByID(ctx, params.ID)
	if err != nil {
		return notFoundGet("not found")
	}
	if a.IsExpired() {
		return goneGet("media expired")
	}
	asset, err := h.assetFromDomain(ctx, a)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (h *HTTPHandler) PatchPlatformMedia(ctx context.Context, req *http_v1.PatchMediaRequest, params http_v1.PatchPlatformMediaParams) (http_v1.PatchPlatformMediaRes, error) {
	if !adminFromContext(ctx) {
		return forbiddenPatch()
	}
	a, err := h.svc.Patch(ctx, params.ID, patchInput(req))
	if err != nil {
		return badRequestPatch(err.Error())
	}
	asset, err := h.assetFromDomain(ctx, a)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (h *HTTPHandler) DeletePlatformMedia(ctx context.Context, params http_v1.DeletePlatformMediaParams) (http_v1.DeletePlatformMediaRes, error) {
	if !adminFromContext(ctx) {
		return forbiddenDelete()
	}
	if err := h.svc.Delete(ctx, params.ID); err != nil {
		return nil, err
	}
	return &http_v1.DeletePlatformMediaNoContent{}, nil
}

func (h *HTTPHandler) assetFromDomain(ctx context.Context, a *domain.MediaAsset) (http_v1.MediaAsset, error) {
	dto, err := h.svc.ToDTO(ctx, a)
	if err != nil {
		return http_v1.MediaAsset{}, err
	}
	return mapDTO(*dto)
}

func mapDTO(dto domain.MediaAssetDTO) (http_v1.MediaAsset, error) {
	id, err := uuid.Parse(dto.ID)
	if err != nil {
		return http_v1.MediaAsset{}, err
	}
	langID, err := uuid.Parse(dto.LanguageID)
	if err != nil {
		return http_v1.MediaAsset{}, err
	}
	createdAt, err := time.Parse(time.RFC3339, dto.CreatedAt)
	if err != nil {
		return http_v1.MediaAsset{}, err
	}
	out := http_v1.MediaAsset{
		ID: id, Scope: dto.Scope, LanguageID: langID,
		DisplayName: dto.DisplayName, MimeType: dto.MimeType, MediaKind: dto.MediaKind,
		URL: dto.URL, SizeBytes: dto.SizeBytes, CreatedAt: createdAt,
	}
	if dto.LinkedLexemeID != nil {
		if lexID, err := uuid.Parse(*dto.LinkedLexemeID); err == nil {
			out.LinkedLexemeID = http_v1.NewOptUUID(lexID)
		}
	}
	if dto.Width != nil {
		out.Width = http_v1.NewOptInt(*dto.Width)
	}
	if dto.Height != nil {
		out.Height = http_v1.NewOptInt(*dto.Height)
	}
	if dto.DurationMs != nil {
		out.DurationMs = http_v1.NewOptInt(*dto.DurationMs)
	}
	if dto.ExpiresAt != nil {
		if t, err := time.Parse(time.RFC3339, *dto.ExpiresAt); err == nil {
			out.ExpiresAt = http_v1.NewOptDateTime(t)
		}
	}
	return out, nil
}

func confirmInput(req *http_v1.ConfirmRequest, userID uuid.UUID, isAdmin bool) domain.MediaConfirmInput {
	in := domain.MediaConfirmInput{
		ObjectKey: req.ObjectKey, MimeType: req.MimeType, SizeBytes: req.SizeBytes,
		LanguageID: req.LanguageID.String(), UserID: userID, IsAdmin: isAdmin,
	}
	if v, ok := req.DisplayName.Get(); ok {
		in.DisplayName = v
	}
	if v, ok := req.LinkedLexemeID.Get(); ok {
		s := v.String()
		in.LinkedLexemeID = &s
	}
	if v, ok := req.TTLSeconds.Get(); ok {
		in.TTLSeconds = &v
	}
	if v, ok := req.ExpiresAt.Get(); ok {
		s := v.UTC().Format(time.RFC3339)
		in.ExpiresAt = &s
	}
	if v, ok := req.Width.Get(); ok {
		in.Width = &v
	}
	if v, ok := req.Height.Get(); ok {
		in.Height = &v
	}
	if v, ok := req.DurationMs.Get(); ok {
		in.DurationMs = &v
	}
	return in
}

func patchInput(req *http_v1.PatchMediaRequest) domain.MediaPatchInput {
	var in domain.MediaPatchInput
	if v, ok := req.DisplayName.Get(); ok {
		in.DisplayName = &v
	}
	if v, ok := req.LinkedLexemeID.Get(); ok {
		s := v.String()
		in.LinkedLexemeID = &s
	}
	if v, ok := req.TTLSeconds.Get(); ok {
		in.TTLSeconds = &v
	}
	if v, ok := req.ExpiresAt.Get(); ok {
		s := v.UTC().Format(time.RFC3339)
		in.ExpiresAt = &s
	}
	return in
}

func adminFromContext(ctx context.Context) bool {
	claims, ok := middleware.ClaimsFromContext(ctx)
	return ok && claims.IsAdmin
}

func errBody(msg string) http_v1.ErrorResponse {
	return http_v1.ErrorResponse{
		Message: http_v1.NewOptString(msg),
		Error:   http_v1.NewOptString(msg),
	}
}

func forbiddenPresign() (*http_v1.PlatformMediaPresignForbidden, error) {
	r := http_v1.PlatformMediaPresignForbidden(errBody("platform admin required"))
	return &r, nil
}

func badRequestPresign(msg string) (*http_v1.PlatformMediaPresignBadRequest, error) {
	r := http_v1.PlatformMediaPresignBadRequest(errBody(msg))
	return &r, nil
}

func forbiddenConfirm() (*http_v1.PlatformMediaConfirmForbidden, error) {
	r := http_v1.PlatformMediaConfirmForbidden(errBody("platform admin required"))
	return &r, nil
}

func badRequestConfirm(msg string) (*http_v1.PlatformMediaConfirmBadRequest, error) {
	r := http_v1.PlatformMediaConfirmBadRequest(errBody(msg))
	return &r, nil
}

func notFoundList(msg string) (*http_v1.ErrorResponse, error) {
	r := errBody(msg)
	return &r, nil
}

func notFoundGet(msg string) (*http_v1.GetPlatformMediaNotFound, error) {
	r := http_v1.GetPlatformMediaNotFound(errBody(msg))
	return &r, nil
}

func goneGet(msg string) (*http_v1.GetPlatformMediaGone, error) {
	r := http_v1.GetPlatformMediaGone(errBody(msg))
	return &r, nil
}

func forbiddenPatch() (*http_v1.PatchPlatformMediaForbidden, error) {
	r := http_v1.PatchPlatformMediaForbidden(errBody("platform admin required"))
	return &r, nil
}

func badRequestPatch(msg string) (*http_v1.PatchPlatformMediaBadRequest, error) {
	r := http_v1.PatchPlatformMediaBadRequest(errBody(msg))
	return &r, nil
}

func forbiddenDelete() (*http_v1.DeletePlatformMediaForbidden, error) {
	r := http_v1.DeletePlatformMediaForbidden(errBody("platform admin required"))
	return &r, nil
}

func max1(p int) int {
	if p < 1 {
		return 1
	}
	return p
}

func maxLimit(l int) int {
	if l < 1 {
		return 20
	}
	return l
}

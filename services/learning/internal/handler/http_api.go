package handler

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/services/learning/internal/domain"
	http_v1 "github.com/even-app/even-app/services/learning/internal/gen/http/v1"
	"github.com/even-app/even-app/services/learning/internal/service"
	"github.com/go-faster/jx"
	"github.com/google/uuid"
)

var _ http_v1.Handler = (*HTTPHandler)(nil)

type HTTPHandler struct {
	svc *service.LearningService
}

func NewHTTPHandler(svc *service.LearningService) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

func (h *HTTPHandler) JoinCourse(ctx context.Context, req *http_v1.JoinCourseRequest) (http_v1.JoinCourseRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	out, err := h.svc.JoinCourse(ctx, claims.UserID, req.InviteCode)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return notFoundJoin("invite code not found")
		}
		if errors.Is(err, domain.ErrConflict) {
			return conflictJoin("already enrolled")
		}
		if errors.Is(err, domain.ErrValidation) {
			return nil, domain.ErrValidation
		}
		return nil, err
	}
	return &http_v1.JoinCourseResponse{
		CourseID: out.CourseID, EnrollmentID: out.EnrollmentID,
	}, nil
}

func (h *HTTPHandler) ListCourses(ctx context.Context) (http_v1.ListCoursesRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	rows, err := h.svc.ListCourses(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	out := make(http_v1.ListCoursesOKApplicationJSON, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapCourseListItem(r))
	}
	return &out, nil
}

func (h *HTTPHandler) GetCourse(ctx context.Context, params http_v1.GetCourseParams) (http_v1.GetCourseRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	c, err := h.svc.GetCourse(ctx, claims.UserID, params.CourseId)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenGetCourse()
		}
		if errors.Is(err, domain.ErrNotFound) {
			return notFoundGetCourse()
		}
		return nil, err
	}
	out := mapCourse(*c)
	return &out, nil
}

func (h *HTTPHandler) ListCourseLessons(ctx context.Context, params http_v1.ListCourseLessonsParams) (http_v1.ListCourseLessonsRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	rows, err := h.svc.ListCourseLessons(ctx, claims.UserID, params.CourseId)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenListCourseLessons()
		}
		return nil, err
	}
	items := make([]http_v1.CourseLessonSummary, 0, len(rows))
	for _, r := range rows {
		items = append(items, http_v1.CourseLessonSummary{
			ID: r.ID, Title: r.Title, SortOrder: r.SortOrder, CompletedPercent: r.CompletedPercent,
		})
	}
	return &http_v1.CourseLessonListResponse{Items: items}, nil
}

func (h *HTTPHandler) ListPublicCourses(ctx context.Context) ([]http_v1.PublicCourseListItem, error) {
	rows, err := h.svc.ListPublicCourses(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]http_v1.PublicCourseListItem, 0, len(rows))
	for _, r := range rows {
		item := http_v1.PublicCourseListItem{
			ID: r.ID, Title: r.Title, IsPublished: r.IsPublished,
			TargetLanguage: http_v1.Language{
				ID: r.LanguageID, Code: r.TargetLangCode, Name: r.TargetLangName,
				NativeName: r.TargetLangName, Direction: http_v1.LanguageDirectionLtr, IsActive: true,
			},
		}
		if r.InviteCode != "" {
			item.InviteCode = http_v1.NewOptString(r.InviteCode)
		}
		out = append(out, item)
	}
	return out, nil
}

func (h *HTTPHandler) GetProgressSummary(ctx context.Context) (http_v1.GetProgressSummaryRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	summary, err := h.svc.GetProgressSummary(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	return &http_v1.ProgressSummary{
		EnrolledCourses:  summary.EnrolledCourses,
		DictionaryWords:  summary.DictionaryWords,
		ReviewDue:        summary.ReviewDue,
		CompletedBlocks:  summary.CompletedBlocks,
		CompletedLessons: summary.CompletedLessons,
	}, nil
}

func (h *HTTPHandler) StartReviewSession(ctx context.Context) (http_v1.StartReviewSessionRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	session, err := h.svc.StartReviewSession(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	if session == nil {
		return &http_v1.StartReviewSessionNoContent{}, nil
	}
	return &http_v1.ReviewSessionResponse{
		Item:         mapReviewItem(session.Item),
		PendingCount: session.PendingCount,
		DueCount:     session.DueCount,
	}, nil
}

func (h *HTTPHandler) GetCourseOutline(ctx context.Context, params http_v1.GetCourseOutlineParams) (http_v1.GetCourseOutlineRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	outline, err := h.svc.GetCourseOutline(ctx, claims.UserID, params.CourseId)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenGetCourseOutline()
		}
		return nil, err
	}
	resp := mapCourseOutline(*outline)
	return &resp, nil
}

func (h *HTTPHandler) GetLesson(ctx context.Context, params http_v1.GetLessonParams) (http_v1.GetLessonRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	snap, err := h.svc.GetLesson(ctx, claims.UserID, params.LessonId)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenGetLesson()
		}
		if errors.Is(err, domain.ErrNotFound) {
			return notFoundGetLesson()
		}
		return nil, err
	}
	out := mapLesson(*snap, h.svc.LexemeLookup(ctx, snap), h.svc.MediaLookup(ctx, snap))
	return &out, nil
}

func (h *HTTPHandler) GetLessonFlow(ctx context.Context, params http_v1.GetLessonFlowParams) (http_v1.GetLessonFlowRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	flow, err := h.svc.GetLessonFlow(ctx, claims.UserID, params.LessonId)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenGetLessonFlow()
		}
		if errors.Is(err, domain.ErrNotFound) {
			return notFoundGetLessonFlow()
		}
		return nil, err
	}
	items := make([]http_v1.LessonFlowItem, 0, len(flow.Items))
	for _, it := range flow.Items {
		items = append(items, mapFlowItem(it))
	}
	return &http_v1.LessonFlow{
		LessonID: flow.LessonID, Version: flow.Version, Items: items,
	}, nil
}

func (h *HTTPHandler) SubmitBlockAttempt(ctx context.Context, req *http_v1.BlockAttemptRequest, params http_v1.SubmitBlockAttemptParams) (http_v1.SubmitBlockAttemptRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	subIdx := 0
	if v, ok := req.SubItemIndex.Get(); ok {
		subIdx = v
	}
	ctxVal := "lesson"
	if v, ok := req.Context.Get(); ok {
		ctxVal = string(v)
	}
	resp := map[string]any{}
	for k, v := range req.Response {
		resp[k] = jsonRawToAny(v)
	}
	out, err := h.svc.SubmitBlockAttempt(ctx, claims.UserID, params.BlockId, service.AttemptInput{
		SubItemIndex: subIdx, Response: resp, Context: ctxVal,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenSubmitBlockAttempt()
		}
		if errors.Is(err, domain.ErrNotFound) {
			return notFoundSubmitBlockAttempt()
		}
		if errors.Is(err, domain.ErrValidation) {
			return nil, domain.ErrValidation
		}
		return nil, err
	}
	r := mapAttemptResponse(out)
	return &r, nil
}

func (h *HTTPHandler) GetLessonProgress(ctx context.Context, params http_v1.GetLessonProgressParams) (http_v1.GetLessonProgressRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	prog, err := h.svc.GetLessonProgress(ctx, claims.UserID, params.LessonId)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenGetLessonProgress()
		}
		return nil, err
	}
	blocks := make([]http_v1.UserBlockProgress, 0, len(prog.Blocks))
	for _, b := range prog.Blocks {
		blocks = append(blocks, mapBlockProgress(b))
	}
	return &http_v1.LessonProgressResponse{LessonID: prog.LessonID, Blocks: blocks}, nil
}

func (h *HTTPHandler) ListReview(ctx context.Context, params http_v1.ListReviewParams) (http_v1.ListReviewRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	status := ""
	if v, ok := params.Status.Get(); ok {
		status = string(v)
	}
	dueOnly := false
	if v, ok := params.DueOnly.Get(); ok {
		dueOnly = v
	}
	list, err := h.svc.ListReview(ctx, claims.UserID, status, dueOnly)
	if err != nil {
		return nil, err
	}
	items := make([]http_v1.ReviewItem, 0, len(list.Items))
	for _, it := range list.Items {
		items = append(items, mapReviewItem(it))
	}
	return &http_v1.ReviewListResponse{
		PendingCount: list.PendingCount, DueCount: list.DueCount, Items: items,
	}, nil
}

func (h *HTTPHandler) ListDictionary(ctx context.Context, params http_v1.ListDictionaryParams) (http_v1.ListDictionaryRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	var courseID *uuid.UUID
	if v, ok := params.CourseID.Get(); ok {
		courseID = &v
	}
	rows, err := h.svc.ListDictionary(ctx, claims.UserID, courseID)
	if err != nil {
		return nil, err
	}
	out := make(http_v1.ListDictionaryOKApplicationJSON, 0, len(rows))
	for _, r := range rows {
		out = append(out, mapVocabularyEntry(r))
	}
	return &out, nil
}

func jsonRawToAny(v jx.Raw) any {
	var a any
	if err := json.Unmarshal(v, &a); err != nil {
		return strings.TrimSpace(string(v))
	}
	return a
}

func forbiddenGetCourse() (*http_v1.GetCourseForbidden, error) {
	r := http_v1.GetCourseForbidden(errBody("not enrolled"))
	return &r, nil
}

func notFoundGetCourse() (*http_v1.GetCourseNotFound, error) {
	r := http_v1.GetCourseNotFound(errBody("course not found"))
	return &r, nil
}

func forbiddenListCourseLessons() (*http_v1.ListCourseLessonsForbidden, error) {
	r := http_v1.ListCourseLessonsForbidden(errBody("not enrolled"))
	return &r, nil
}

func forbiddenGetLesson() (*http_v1.GetLessonForbidden, error) {
	r := http_v1.GetLessonForbidden(errBody("not enrolled"))
	return &r, nil
}

func notFoundGetLesson() (*http_v1.GetLessonNotFound, error) {
	r := http_v1.GetLessonNotFound(errBody("lesson not found"))
	return &r, nil
}

func forbiddenGetLessonFlow() (*http_v1.GetLessonFlowForbidden, error) {
	r := http_v1.GetLessonFlowForbidden(errBody("not enrolled"))
	return &r, nil
}

func notFoundGetLessonFlow() (*http_v1.GetLessonFlowNotFound, error) {
	r := http_v1.GetLessonFlowNotFound(errBody("lesson not found"))
	return &r, nil
}

func forbiddenSubmitBlockAttempt() (*http_v1.SubmitBlockAttemptForbidden, error) {
	r := http_v1.SubmitBlockAttemptForbidden(errBody("not enrolled"))
	return &r, nil
}

func notFoundSubmitBlockAttempt() (*http_v1.SubmitBlockAttemptNotFound, error) {
	r := http_v1.SubmitBlockAttemptNotFound(errBody("block not found"))
	return &r, nil
}

func forbiddenGetLessonProgress() (*http_v1.GetLessonProgressForbidden, error) {
	r := http_v1.GetLessonProgressForbidden(errBody("not enrolled"))
	return &r, nil
}

func forbiddenGetCourseOutline() (*http_v1.GetCourseOutlineForbidden, error) {
	r := http_v1.GetCourseOutlineForbidden(errBody("not enrolled"))
	return &r, nil
}

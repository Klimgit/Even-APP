package handler

import (
	"context"

	"github.com/even-app/even-app/services/content/internal/domain"
	http_v1 "github.com/even-app/even-app/services/content/internal/gen/http/v1"
	"github.com/even-app/even-app/services/content/internal/gen/query"
)

func (h *HTTPHandler) platformAdmin(ctx context.Context) (teacherClaims, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return teacherClaims{}, err
	}
	if !t.IsAdmin {
		return teacherClaims{}, domain.ErrForbidden
	}
	return t, nil
}

func mapModule(row query.CourseModule) http_v1.CourseModule {
	return http_v1.CourseModule{
		ID: row.ID, CourseID: row.CourseID, Title: row.Title, SortOrder: int(row.SortOrder),
	}
}

func mapModules(rows []query.CourseModule) []http_v1.CourseModule {
	out := make([]http_v1.CourseModule, len(rows))
	for i, r := range rows {
		out[i] = mapModule(r)
	}
	return out
}

func (h *HTTPHandler) ListTeacherCourseModules(ctx context.Context, params http_v1.ListTeacherCourseModulesParams) (http_v1.ListTeacherCourseModulesRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := h.svc.ListModules(ctx, params.CourseId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := http_v1.ListTeacherCourseModulesOKApplicationJSON(mapModules(rows))
	return &out, nil
}

func (h *HTTPHandler) CreateTeacherCourseModule(ctx context.Context, req *http_v1.CreateModuleRequest, params http_v1.CreateTeacherCourseModuleParams) (http_v1.CreateTeacherCourseModuleRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	var sortOrder *int32
	if v, ok := req.SortOrder.Get(); ok {
		s := int32(v)
		sortOrder = &s
	}
	row, err := h.svc.CreateModule(ctx, params.CourseId, t.UserID, t.IsAdmin, req.Title, sortOrder)
	if err != nil {
		return nil, err
	}
	out := mapModule(row)
	return &out, nil
}

func (h *HTTPHandler) PatchTeacherModule(ctx context.Context, req *http_v1.PatchModuleRequest, params http_v1.PatchTeacherModuleParams) (http_v1.PatchTeacherModuleRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	p := query.UpdateModuleParams{}
	if v, ok := req.Title.Get(); ok {
		p.Title = &v
	}
	if v, ok := req.SortOrder.Get(); ok {
		s := int32(v)
		p.SortOrder = &s
	}
	row, err := h.svc.PatchModule(ctx, params.ModuleId, t.UserID, t.IsAdmin, p)
	if err != nil {
		return nil, err
	}
	out := mapModule(row)
	return &out, nil
}

func (h *HTTPHandler) DeleteTeacherModule(ctx context.Context, params http_v1.DeleteTeacherModuleParams) (http_v1.DeleteTeacherModuleRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.DeleteModule(ctx, params.ModuleId, t.UserID, t.IsAdmin); err != nil {
		return nil, err
	}
	return &http_v1.DeleteTeacherModuleNoContent{}, nil
}

func (h *HTTPHandler) ReorderTeacherCourseModules(ctx context.Context, req *http_v1.ReorderIdsRequest, params http_v1.ReorderTeacherCourseModulesParams) (http_v1.ReorderTeacherCourseModulesRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.ReorderModules(ctx, params.CourseId, t.UserID, t.IsAdmin, req.Ids); err != nil {
		return nil, err
	}
	return &http_v1.ReorderTeacherCourseModulesNoContent{}, nil
}

func (h *HTTPHandler) CreateTeacherModuleLesson(ctx context.Context, req *http_v1.CreateLessonRequest, params http_v1.CreateTeacherModuleLessonParams) (http_v1.CreateTeacherModuleLessonRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	var sortOrder *int32
	if v, ok := req.SortOrder.Get(); ok {
		s := int32(v)
		sortOrder = &s
	}
	modID := params.ModuleId
	row, err := h.svc.CreateLesson(ctx, modID, t.UserID, t.IsAdmin, req.Title, sortOrder)
	if err != nil {
		return nil, err
	}
	out := mapLessonSummary(row)
	return &out, nil
}

func (h *HTTPHandler) ListPlatformCourses(ctx context.Context, params http_v1.ListPlatformCoursesParams) (http_v1.ListPlatformCoursesRes, error) {
	if _, err := h.platformAdmin(ctx); err != nil {
		return nil, err
	}
	page, limit := 1, 20
	if v, ok := params.Page.Get(); ok {
		page = v
	}
	if v, ok := params.Limit.Get(); ok {
		limit = v
	}
	var search *string
	if v, ok := params.Q.Get(); ok && v != "" {
		search = &v
	}
	rows, total, err := h.svc.ListPlatformCourses(ctx, search, page, limit)
	if err != nil {
		return nil, err
	}
	items := make([]http_v1.PlatformCourseListItem, len(rows))
	for i, r := range rows {
		items[i] = http_v1.PlatformCourseListItem{
			ID: r.ID, Title: r.Title,
			TargetLanguageID: r.TargetLanguageID,
			UILanguageID:     r.UiLanguageID,
			OwnerID:          r.OwnerID,
			IsPublished:      r.IsPublished,
			Visibility:       http_v1.PlatformCourseListItemVisibility(r.Visibility),
		}
	}
	return &http_v1.PlatformCourseListResponse{
		Items: items, Total: int(total), Page: page, Limit: limit,
	}, nil
}

func (h *HTTPHandler) GetTeacherCourseAnalytics(ctx context.Context, params http_v1.GetTeacherCourseAnalyticsParams) (http_v1.GetTeacherCourseAnalyticsRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	stats, err := h.svc.GetCourseAnalytics(ctx, params.CourseId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	return &http_v1.CourseAnalytics{
		StudentCount:         stats.StudentCount,
		LessonCount:          stats.LessonCount,
		ModuleCount:          stats.ModuleCount,
		AvgCompletionPercent: stats.AvgCompletionPercent,
	}, nil
}

func (h *HTTPHandler) EnrollPlatformStudent(ctx context.Context, req *http_v1.PlatformEnrollRequest) (http_v1.EnrollPlatformStudentRes, error) {
	t, err := h.platformAdmin(ctx)
	if err != nil {
		return nil, err
	}
	row, err := h.svc.EnrollStudent(ctx, req.CourseID, req.Email, t.UserID, true)
	if err != nil {
		return nil, err
	}
	out := mapStudent(row)
	return &out, nil
}

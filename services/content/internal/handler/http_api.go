package handler

import (
	"context"

	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/services/content/internal/domain"
	http_v1 "github.com/even-app/even-app/services/content/internal/gen/http/v1"
	"github.com/even-app/even-app/services/content/internal/gen/query"
	"github.com/even-app/even-app/services/content/internal/service"
	"github.com/google/uuid"
)

var _ http_v1.Handler = (*HTTPHandler)(nil)

type HTTPHandler struct {
	svc *service.ContentService
}

func NewHTTPHandler(svc *service.ContentService) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

type teacherClaims struct {
	UserID  uuid.UUID
	IsAdmin bool
}

func (h *HTTPHandler) teacher(ctx context.Context) (teacherClaims, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return teacherClaims{}, domain.ErrUnauthorized
	}
	if claims.Role != "teacher" && !claims.IsAdmin {
		return teacherClaims{}, domain.ErrForbidden
	}
	return teacherClaims{UserID: claims.UserID, IsAdmin: claims.IsAdmin}, nil
}

func (h *HTTPHandler) ListBlockTypes(ctx context.Context) (http_v1.ListBlockTypesRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	favs, err := h.svc.ListFavoriteBlockTypes(ctx, t.UserID)
	if err != nil {
		return nil, err
	}
	out := http_v1.ListBlockTypesOKApplicationJSON(mapBlockTypes(domain.FullBlockTypeCatalog(), favs))
	return &out, nil
}

func (h *HTTPHandler) AddTeacherBlockTypeFavorite(ctx context.Context, params http_v1.AddTeacherBlockTypeFavoriteParams) (http_v1.AddTeacherBlockTypeFavoriteRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.AddFavoriteBlockType(ctx, t.UserID, params.BlockType); err != nil {
		return nil, err
	}
	return &http_v1.AddTeacherBlockTypeFavoriteNoContent{}, nil
}

func (h *HTTPHandler) RemoveTeacherBlockTypeFavorite(ctx context.Context, params http_v1.RemoveTeacherBlockTypeFavoriteParams) (http_v1.RemoveTeacherBlockTypeFavoriteRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.RemoveFavoriteBlockType(ctx, t.UserID, params.BlockType); err != nil {
		return nil, err
	}
	return &http_v1.RemoveTeacherBlockTypeFavoriteNoContent{}, nil
}

func (h *HTTPHandler) ListTeacherCourses(ctx context.Context) (http_v1.ListTeacherCoursesRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := h.svc.ListCourses(ctx, t.UserID)
	if err != nil {
		return nil, err
	}
	out := http_v1.ListTeacherCoursesOKApplicationJSON(mapCourses(rows))
	return &out, nil
}

func (h *HTTPHandler) CreateTeacherCourse(ctx context.Context, req *http_v1.CreateCourseRequest) (http_v1.CreateTeacherCourseRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	row, err := h.svc.CreateCourse(ctx, t.UserID, req.Title, req.TargetLanguageID, req.UILanguageID, createCourseVisibility(req))
	if err != nil {
		return nil, err
	}
	out := mapCourse(row)
	return &out, nil
}

func (h *HTTPHandler) GetTeacherCourse(ctx context.Context, params http_v1.GetTeacherCourseParams) (http_v1.GetTeacherCourseRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	row, err := h.svc.GetCourse(ctx, params.CourseId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := mapCourse(row)
	return &out, nil
}

func (h *HTTPHandler) PatchTeacherCourse(ctx context.Context, req *http_v1.PatchCourseRequest, params http_v1.PatchTeacherCourseParams) (http_v1.PatchTeacherCourseRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	p := query.UpdateCourseParams{}
	if v, ok := req.Title.Get(); ok {
		p.Title = &v
	}
	if v, ok := req.TargetLanguageID.Get(); ok {
		p.TargetLanguageID = &v
	}
	if v, ok := req.UILanguageID.Get(); ok {
		p.UiLanguageID = &v
	}
	if v, ok := req.Visibility.Get(); ok {
		vis := string(v)
		p.Visibility = &vis
	}
	row, err := h.svc.PatchCourse(ctx, params.CourseId, t.UserID, t.IsAdmin, p)
	if err != nil {
		return nil, err
	}
	out := mapCourse(row)
	return &out, nil
}

func (h *HTTPHandler) DeleteTeacherCourse(ctx context.Context, params http_v1.DeleteTeacherCourseParams) (http_v1.DeleteTeacherCourseRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.DeleteCourse(ctx, params.CourseId, t.UserID, t.IsAdmin); err != nil {
		return nil, err
	}
	return &http_v1.DeleteTeacherCourseNoContent{}, nil
}

func (h *HTTPHandler) PublishTeacherCourse(ctx context.Context, params http_v1.PublishTeacherCourseParams) (http_v1.PublishTeacherCourseRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	row, err := h.svc.PublishCourse(ctx, params.CourseId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := mapCourse(row)
	return &out, nil
}

func (h *HTTPHandler) ListTeacherCourseLessons(ctx context.Context, params http_v1.ListTeacherCourseLessonsParams) (http_v1.ListTeacherCourseLessonsRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := h.svc.ListLessons(ctx, params.CourseId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := http_v1.ListTeacherCourseLessonsOKApplicationJSON(mapLessonSummaries(rows))
	return &out, nil
}

func (h *HTTPHandler) CreateTeacherCourseLesson(ctx context.Context, req *http_v1.CreateLessonRequest, params http_v1.CreateTeacherCourseLessonParams) (http_v1.CreateTeacherCourseLessonRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	var sortOrder *int32
	if v, ok := req.SortOrder.Get(); ok {
		s := int32(v)
		sortOrder = &s
	}
	row, err := h.svc.CreateLesson(ctx, params.CourseId, t.UserID, t.IsAdmin, req.Title, sortOrder)
	if err != nil {
		return nil, err
	}
	out := mapLessonSummary(row)
	return &out, nil
}

func (h *HTTPHandler) GetTeacherLesson(ctx context.Context, params http_v1.GetTeacherLessonParams) (http_v1.GetTeacherLessonRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	row, err := h.svc.GetLesson(ctx, params.LessonId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := mapLessonFull(row)
	return &out, nil
}

func (h *HTTPHandler) PatchTeacherLesson(ctx context.Context, req *http_v1.PatchLessonRequest, params http_v1.PatchTeacherLessonParams) (http_v1.PatchTeacherLessonRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	p := query.UpdateLessonParams{}
	if v, ok := req.Title.Get(); ok {
		p.Title = &v
	}
	if v, ok := req.SortOrder.Get(); ok {
		s := int32(v)
		p.SortOrder = &s
	}
	row, err := h.svc.PatchLesson(ctx, params.LessonId, t.UserID, t.IsAdmin, params.IfMatch, p)
	if err != nil {
		return nil, err
	}
	out := mapLessonFull(row)
	return &out, nil
}

func (h *HTTPHandler) DeleteTeacherLesson(ctx context.Context, params http_v1.DeleteTeacherLessonParams) (http_v1.DeleteTeacherLessonRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.DeleteLesson(ctx, params.LessonId, t.UserID, t.IsAdmin); err != nil {
		return nil, err
	}
	return &http_v1.DeleteTeacherLessonNoContent{}, nil
}

func (h *HTTPHandler) PublishTeacherLesson(ctx context.Context, params http_v1.PublishTeacherLessonParams) (http_v1.PublishTeacherLessonRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	row, err := h.svc.PublishLesson(ctx, params.LessonId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := mapLessonFull(row)
	return &out, nil
}

func (h *HTTPHandler) CreateTeacherLessonSection(ctx context.Context, req *http_v1.CreateSectionRequest, params http_v1.CreateTeacherLessonSectionParams) (http_v1.CreateTeacherLessonSectionRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	kind := "content"
	if v, ok := req.SectionKind.Get(); ok {
		kind = string(v)
	}
	var sortOrder *int32
	if v, ok := req.SortOrder.Get(); ok {
		s := int32(v)
		sortOrder = &s
	}
	row, err := h.svc.CreateSection(ctx, params.LessonId, t.UserID, t.IsAdmin, req.Title, kind, sortOrder)
	if err != nil {
		return nil, err
	}
	out := mapSection(row)
	return &out, nil
}

func (h *HTTPHandler) PatchTeacherSection(ctx context.Context, req *http_v1.PatchSectionRequest, params http_v1.PatchTeacherSectionParams) (http_v1.PatchTeacherSectionRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	p := query.UpdateSectionParams{}
	if v, ok := req.Title.Get(); ok {
		p.Title = &v
	}
	if v, ok := req.SortOrder.Get(); ok {
		s := int32(v)
		p.SortOrder = &s
	}
	if v, ok := req.SectionKind.Get(); ok {
		s := string(v)
		p.SectionKind = &s
	}
	row, err := h.svc.PatchSection(ctx, params.SectionId, t.UserID, t.IsAdmin, p)
	if err != nil {
		return nil, err
	}
	out := mapSection(row)
	return &out, nil
}

func (h *HTTPHandler) DeleteTeacherSection(ctx context.Context, params http_v1.DeleteTeacherSectionParams) (http_v1.DeleteTeacherSectionRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.DeleteSection(ctx, params.SectionId, t.UserID, t.IsAdmin); err != nil {
		return nil, err
	}
	return &http_v1.DeleteTeacherSectionNoContent{}, nil
}

func (h *HTTPHandler) ReorderTeacherLessonSections(ctx context.Context, req *http_v1.ReorderIdsRequest, params http_v1.ReorderTeacherLessonSectionsParams) (http_v1.ReorderTeacherLessonSectionsRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.ReorderSections(ctx, params.LessonId, t.UserID, t.IsAdmin, req.Ids); err != nil {
		return nil, err
	}
	return &http_v1.ReorderTeacherLessonSectionsNoContent{}, nil
}

func (h *HTTPHandler) CreateTeacherLessonBlock(ctx context.Context, req *http_v1.CreateLessonBlockRequest, params http_v1.CreateTeacherLessonBlockParams) (http_v1.CreateTeacherLessonBlockRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	cfg, err := blockConfigToBytes(req.Config)
	if err != nil {
		return nil, domain.ErrValidation
	}
	p := query.CreateBlockParams{
		SortOrder: int32(req.SortOrder), BlockType: req.BlockType, Config: cfg,
	}
	if v, ok := req.SectionID.Get(); ok {
		p.SectionID = &v
	}
	if v, ok := req.DisplayLabel.Get(); ok {
		p.DisplayLabel = &v
	}
	if v, ok := req.Title.Get(); ok {
		p.Title = &v
	}
	if v, ok := req.IsHomework.Get(); ok {
		p.IsHomework = v
	}
	row, err := h.svc.CreateBlock(ctx, params.LessonId, t.UserID, t.IsAdmin, p)
	if err != nil {
		return nil, err
	}
	out := mapBlock(row)
	return &out, nil
}

func (h *HTTPHandler) GetTeacherBlock(ctx context.Context, params http_v1.GetTeacherBlockParams) (http_v1.GetTeacherBlockRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	row, err := h.svc.GetBlock(ctx, params.BlockId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := mapBlock(row)
	return &out, nil
}

func (h *HTTPHandler) PatchTeacherBlock(ctx context.Context, req *http_v1.PatchLessonBlockRequest, params http_v1.PatchTeacherBlockParams) (http_v1.PatchTeacherBlockRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	p := query.UpdateBlockParams{}
	if v, ok := req.SectionID.Get(); ok {
		p.SectionID = &v
	}
	if v, ok := req.SortOrder.Get(); ok {
		s := int32(v)
		p.SortOrder = &s
	}
	if v, ok := req.DisplayLabel.Get(); ok {
		p.DisplayLabel = &v
	}
	if v, ok := req.Title.Get(); ok {
		p.Title = &v
	}
	if v, ok := req.BlockType.Get(); ok {
		p.BlockType = &v
	}
	if v, ok := req.Config.Get(); ok {
		raw, err := blockConfigToBytes(v)
		if err != nil {
			return nil, domain.ErrValidation
		}
		p.Config = raw
	}
	if v, ok := req.IsHomework.Get(); ok {
		p.IsHomework = &v
	}
	row, err := h.svc.PatchBlock(ctx, params.BlockId, t.UserID, t.IsAdmin, p)
	if err != nil {
		return nil, err
	}
	out := mapBlock(row)
	return &out, nil
}

func (h *HTTPHandler) DeleteTeacherBlock(ctx context.Context, params http_v1.DeleteTeacherBlockParams) (http_v1.DeleteTeacherBlockRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.DeleteBlock(ctx, params.BlockId, t.UserID, t.IsAdmin); err != nil {
		return nil, err
	}
	return &http_v1.DeleteTeacherBlockNoContent{}, nil
}

func (h *HTTPHandler) ReorderTeacherLessonBlocks(ctx context.Context, req *http_v1.ReorderIdsRequest, params http_v1.ReorderTeacherLessonBlocksParams) (http_v1.ReorderTeacherLessonBlocksRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	if err := h.svc.ReorderBlocks(ctx, params.LessonId, t.UserID, t.IsAdmin, req.Ids); err != nil {
		return nil, err
	}
	return &http_v1.ReorderTeacherLessonBlocksNoContent{}, nil
}

func (h *HTTPHandler) GetTeacherCourseLexiconCoverage(ctx context.Context, params http_v1.GetTeacherCourseLexiconCoverageParams) (http_v1.GetTeacherCourseLexiconCoverageRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	row, err := h.svc.GetCoverageSummary(ctx, params.CourseId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := mapCoverageSummary(row)
	return &out, nil
}

func (h *HTTPHandler) GetTeacherCourseLexiconByLesson(ctx context.Context, params http_v1.GetTeacherCourseLexiconByLessonParams) (http_v1.GetTeacherCourseLexiconByLessonRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := h.svc.GetCoverageByLesson(ctx, params.CourseId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := http_v1.GetTeacherCourseLexiconByLessonOKApplicationJSON(mapCoverageByLesson(rows))
	return &out, nil
}

func (h *HTTPHandler) GetTeacherCourseFormsCoverage(ctx context.Context, params http_v1.GetTeacherCourseFormsCoverageParams) (http_v1.GetTeacherCourseFormsCoverageRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := h.svc.GetFormsCoverage(ctx, params.CourseId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := http_v1.GetTeacherCourseFormsCoverageOKApplicationJSON(mapFormsCoverage(rows))
	return &out, nil
}

func (h *HTTPHandler) GetTeacherCourseInviteCode(ctx context.Context, params http_v1.GetTeacherCourseInviteCodeParams) (http_v1.GetTeacherCourseInviteCodeRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	code, err := h.svc.GetInviteCode(ctx, params.CourseId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	return &http_v1.InviteCodeResponse{InviteCode: code}, nil
}

func (h *HTTPHandler) RegenerateTeacherCourseInviteCode(ctx context.Context, params http_v1.RegenerateTeacherCourseInviteCodeParams) (http_v1.RegenerateTeacherCourseInviteCodeRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	code, err := h.svc.RegenerateInviteCode(ctx, params.CourseId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	return &http_v1.InviteCodeResponse{InviteCode: code}, nil
}

func (h *HTTPHandler) ListTeacherCourseStudents(ctx context.Context, params http_v1.ListTeacherCourseStudentsParams) (http_v1.ListTeacherCourseStudentsRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := h.svc.ListCourseStudents(ctx, params.CourseId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := http_v1.ListTeacherCourseStudentsOKApplicationJSON(mapStudents(rows))
	return &out, nil
}

func (h *HTTPHandler) EnrollTeacherStudent(ctx context.Context, req *http_v1.EnrollStudentRequest) (http_v1.EnrollTeacherStudentRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	row, err := h.svc.EnrollStudent(ctx, req.CourseID, req.Email, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := mapStudent(row)
	return &out, nil
}

func (h *HTTPHandler) GetTeacherStudentProgress(ctx context.Context, params http_v1.GetTeacherStudentProgressParams) (http_v1.GetTeacherStudentProgressRes, error) {
	t, err := h.teacher(ctx)
	if err != nil {
		return nil, err
	}
	row, err := h.svc.GetStudentProgress(ctx, params.CourseID, params.StudentId, t.UserID, t.IsAdmin)
	if err != nil {
		return nil, err
	}
	out := mapStudentProgress(row)
	return &out, nil
}

func createCourseVisibility(req *http_v1.CreateCourseRequest) *string {
	if v, ok := req.Visibility.Get(); ok {
		s := string(v)
		return &s
	}
	return nil
}

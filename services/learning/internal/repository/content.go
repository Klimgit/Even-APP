package repository

import (
	"context"
	"encoding/json"

	"github.com/even-app/even-app/services/learning/internal/domain"
	"github.com/even-app/even-app/services/learning/internal/gen/contentquery"
	"github.com/even-app/even-app/services/learning/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ContentReader struct {
	pool    *pgxpool.Pool
	queries *contentquery.Queries
}

func NewContentReader(pool *pgxpool.Pool) *ContentReader {
	if pool == nil {
		return nil
	}
	return &ContentReader{pool: pool, queries: contentquery.New(pool)}
}

func (r *ContentReader) Available() bool {
	return r != nil && r.pool != nil
}

func (r *ContentReader) GetCourseByInviteCode(ctx context.Context, code string) (domain.CourseView, error) {
	row, err := r.queries.GetCourseByInviteCode(ctx, contentquery.GetCourseByInviteCodeParams{Code: code})
	if err != nil {
		return domain.CourseView{}, err
	}
	return mapCourseInviteRow(row), nil
}

func (r *ContentReader) GetCourseByID(ctx context.Context, id uuid.UUID) (domain.CourseView, error) {
	row, err := r.queries.GetCourseByID(ctx, contentquery.GetCourseByIDParams{ID: id})
	if err != nil {
		return domain.CourseView{}, err
	}
	return mapCourseIDRow(row), nil
}

func (r *ContentReader) ListPublishedCourses(ctx context.Context, search *string) ([]domain.CourseView, error) {
	rows, err := r.queries.ListPublishedCourses(ctx, contentquery.ListPublishedCoursesParams{Search: search})
	if err != nil {
		return nil, err
	}
	out := make([]domain.CourseView, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.CourseView{
			ID: row.ID, Title: row.Title,
			TargetLanguageID: row.TargetLanguageID,
			TargetLangCode:   row.TargetLanguageCode,
			TargetLangName:   row.TargetLanguageName,
			UILanguageID:     row.UiLanguageID,
			OwnerID:          row.OwnerID,
			IsPublished:      row.IsPublished,
			Visibility:       row.Visibility,
		})
	}
	return out, nil
}

func (r *ContentReader) ListPublishedLessons(ctx context.Context, courseID uuid.UUID) ([]contentquery.Lesson, error) {
	return r.queries.ListPublishedLessonsByCourse(ctx, contentquery.ListPublishedLessonsByCourseParams{CourseID: courseID})
}

func (r *ContentReader) BuildLessonSnapshot(ctx context.Context, lessonID uuid.UUID) (domain.LessonSnapshot, error) {
	lesson, err := r.queries.GetPublishedLessonByID(ctx, contentquery.GetPublishedLessonByIDParams{ID: lessonID})
	if err != nil {
		return domain.LessonSnapshot{}, err
	}
	sections, err := r.queries.ListLessonSections(ctx, contentquery.ListLessonSectionsParams{LessonID: lessonID})
	if err != nil {
		return domain.LessonSnapshot{}, err
	}
	blocks, err := r.queries.ListLessonBlocks(ctx, contentquery.ListLessonBlocksParams{LessonID: lessonID})
	if err != nil {
		return domain.LessonSnapshot{}, err
	}

	snap := domain.LessonSnapshot{
		ID:        lesson.ID,
		CourseID:  lesson.CourseID,
		Title:     lesson.Title,
		SortOrder: int(lesson.SortOrder),
		Version:   int(lesson.Version),
		Status:    lesson.Status,
		Sections:  make([]domain.SectionSnap, 0, len(sections)),
		Blocks:    make([]domain.BlockSnap, 0, len(blocks)),
	}
	for _, s := range sections {
		snap.Sections = append(snap.Sections, domain.SectionSnap{
			ID: s.ID, Title: s.Title, SortOrder: int(s.SortOrder), SectionKind: s.SectionKind,
		})
	}
	for _, b := range blocks {
		bs := domain.BlockSnap{
			ID: b.ID, SortOrder: int(b.SortOrder), BlockType: b.BlockType,
			Config: b.Config, IsHomework: b.IsHomework,
			IsGradable: domain.IsGradable(b.BlockType),
		}
		if b.SectionID != nil {
			bs.SectionID = b.SectionID
		}
		if b.DisplayLabel != nil {
			bs.DisplayLabel = b.DisplayLabel
		}
		if b.Title != nil {
			bs.Title = b.Title
		}
		refs, err := r.queries.ListBlockLexemeRefs(ctx, contentquery.ListBlockLexemeRefsParams{LessonBlockID: b.ID})
		if err == nil {
			for _, ref := range refs {
				lr := domain.LexemeRefSnap{LexemeID: ref.LexemeID, Role: ref.Role}
				if ref.FormID != uuid.Nil {
					fid := ref.FormID
					lr.FormID = &fid
				}
				bs.LexemeRefs = append(bs.LexemeRefs, lr)
			}
		}
		snap.Blocks = append(snap.Blocks, bs)
	}
	return snap, nil
}

func (r *ContentReader) GetBlock(ctx context.Context, blockID uuid.UUID) (domain.BlockSnap, uuid.UUID, string, error) {
	row, err := r.queries.GetLessonBlock(ctx, contentquery.GetLessonBlockParams{ID: blockID})
	if err != nil {
		return domain.BlockSnap{}, uuid.Nil, "", err
	}
	bs := domain.BlockSnap{
		ID: row.ID, SortOrder: int(row.SortOrder), BlockType: row.BlockType,
		Config: row.Config, IsHomework: row.IsHomework,
		IsGradable: domain.IsGradable(row.BlockType),
	}
	if row.SectionID != nil {
		bs.SectionID = row.SectionID
	}
	if row.DisplayLabel != nil {
		bs.DisplayLabel = row.DisplayLabel
	}
	if row.Title != nil {
		bs.Title = row.Title
	}
	refs, err := r.queries.ListBlockLexemeRefs(ctx, contentquery.ListBlockLexemeRefsParams{LessonBlockID: row.ID})
	if err == nil {
		for _, ref := range refs {
			lr := domain.LexemeRefSnap{LexemeID: ref.LexemeID, Role: ref.Role}
			if ref.FormID != uuid.Nil {
				fid := ref.FormID
				lr.FormID = &fid
			}
			bs.LexemeRefs = append(bs.LexemeRefs, lr)
		}
	}
	return bs, row.CourseID, row.LessonTitle, nil
}

func (r *ContentReader) SyncPublishedLessonsForCourse(ctx context.Context, q *query.Queries, courseID uuid.UUID) error {
	lessons, err := r.queries.ListPublishedLessonsByCourse(ctx, contentquery.ListPublishedLessonsByCourseParams{CourseID: courseID})
	if err != nil {
		return err
	}
	for _, l := range lessons {
		snap, err := r.BuildLessonSnapshot(ctx, l.ID)
		if err != nil {
			return err
		}
		raw, err := snap.MarshalJSONBlob()
		if err != nil {
			return err
		}
		if err := q.UpsertPublishedLessonSnapshot(ctx, query.UpsertPublishedLessonSnapshotParams{
			LessonID: l.ID, CourseID: courseID, Version: int32(snap.Version), Snapshot: raw,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *ContentReader) GetLessonTitle(ctx context.Context, lessonID uuid.UUID) (string, error) {
	return r.queries.GetLessonTitle(ctx, contentquery.GetLessonTitleParams{ID: lessonID})
}

func (r *ContentReader) GetBlockLessonID(ctx context.Context, blockID uuid.UUID) (uuid.UUID, error) {
	row, err := r.queries.GetLessonBlock(ctx, contentquery.GetLessonBlockParams{ID: blockID})
	if err != nil {
		return uuid.Nil, err
	}
	return row.LessonID, nil
}

func (r *ContentReader) ListCourseModules(ctx context.Context, courseID uuid.UUID) ([]domain.ModuleView, error) {
	rows, err := r.queries.ListModulesByCourseID(ctx, contentquery.ListModulesByCourseIDParams{CourseID: courseID})
	if err != nil {
		return nil, err
	}
	out := make([]domain.ModuleView, 0, len(rows))
	for _, row := range rows {
		out = append(out, domain.ModuleView{
			ID: row.ID, CourseID: row.CourseID, Title: row.Title, SortOrder: int(row.SortOrder),
		})
	}
	return out, nil
}

func (r *ContentReader) LessonModuleIDs(ctx context.Context, courseID uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	rows, err := r.queries.ListPublishedLessonsByCourse(ctx, contentquery.ListPublishedLessonsByCourseParams{CourseID: courseID})
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]uuid.UUID, len(rows))
	for _, row := range rows {
		out[row.ID] = row.ModuleID
	}
	return out, nil
}

func mapCourseIDRow(row contentquery.GetCourseByIDRow) domain.CourseView {
	return domain.CourseView{
		ID: row.ID, Title: row.Title,
		TargetLanguageID: row.TargetLanguageID,
		TargetLangCode:   row.TargetLanguageCode,
		TargetLangName:   row.TargetLanguageName,
		UILanguageID:     row.UiLanguageID,
		OwnerID:          row.OwnerID,
		IsPublished:      row.IsPublished,
		Visibility:       row.Visibility,
	}
}

func mapCourseInviteRow(row contentquery.GetCourseByInviteCodeRow) domain.CourseView {
	return domain.CourseView{
		ID: row.ID, Title: row.Title,
		TargetLanguageID: row.TargetLanguageID,
		TargetLangCode:   row.TargetLanguageCode,
		TargetLangName:   row.TargetLanguageName,
		UILanguageID:     row.UiLanguageID,
		OwnerID:          row.OwnerID,
		IsPublished:      row.IsPublished,
		Visibility:       row.Visibility,
	}
}

// SnapshotFromRow parses stored snapshot JSON.
func SnapshotFromRow(raw json.RawMessage) (domain.LessonSnapshot, error) {
	return domain.ParseLessonSnapshot(raw)
}

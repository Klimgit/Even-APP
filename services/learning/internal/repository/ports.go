package repository

import (
	"context"
	"encoding/json"

	"github.com/even-app/even-app/libs/clients"
	contentclient "github.com/even-app/even-app/libs/clients/content"
	"github.com/even-app/even-app/libs/clients/dto"
	"github.com/even-app/even-app/services/learning/internal/domain"
	"github.com/even-app/even-app/services/learning/internal/gen/query"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ContentSource reads published course/lesson data from the content service.
type ContentSource interface {
	Available() bool
	GetCourseByInviteCode(ctx context.Context, code string) (domain.CourseView, error)
	GetCourseByID(ctx context.Context, id uuid.UUID) (domain.CourseView, error)
	ListPublishedCourses(ctx context.Context, search *string) ([]domain.CourseView, error)
	ListCourseModules(ctx context.Context, courseID uuid.UUID) ([]domain.ModuleView, error)
	LessonModuleIDs(ctx context.Context, courseID uuid.UUID) (map[uuid.UUID]uuid.UUID, error)
	SyncPublishedLessonsForCourse(ctx context.Context, q *query.Queries, courseID uuid.UUID) error
	BuildLessonSnapshot(ctx context.Context, lessonID uuid.UUID) (domain.LessonSnapshot, error)
	GetBlock(ctx context.Context, blockID uuid.UUID) (domain.BlockSnap, uuid.UUID, string, error)
	GetLessonTitle(ctx context.Context, lessonID uuid.UUID) (string, error)
	GetBlockLessonID(ctx context.Context, blockID uuid.UUID) (uuid.UUID, error)
}

// LexiconSource resolves lexicon metadata.
type LexiconSource interface {
	Available() bool
	GetLanguage(ctx context.Context, id uuid.UUID) (LanguageView, error)
	LanguagesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]LanguageView, error)
	LexemesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]LexemeView, error)
	FilterLexemeIDsBySearch(ctx context.Context, ids []uuid.UUID, search string) ([]uuid.UUID, error)
}

// MediaSource resolves media asset metadata.
type MediaSource interface {
	Available() bool
	MediaByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]MediaView, error)
}

var (
	_ ContentSource = (*ContentReader)(nil)
	_ LexiconSource = (*LexiconReader)(nil)
	_ MediaSource   = (*MediaReader)(nil)
)

type ContentRemote struct {
	client *contentclient.Client
}

func NewContentRemote(baseURL, token string) *ContentRemote {
	return &ContentRemote{client: contentclient.New(baseURL, token)}
}

func (r *ContentRemote) Available() bool {
	return r != nil && r.client.Available()
}

func (r *ContentRemote) GetCourseByInviteCode(ctx context.Context, code string) (domain.CourseView, error) {
	c, err := r.client.GetCourseByInviteCode(ctx, code)
	if err != nil {
		if apiErr, ok := err.(*clients.APIError); ok && apiErr.Status == 404 {
			return domain.CourseView{}, pgx.ErrNoRows
		}
		return domain.CourseView{}, err
	}
	return mapRemoteCourse(c), nil
}

func (r *ContentRemote) GetCourseByID(ctx context.Context, id uuid.UUID) (domain.CourseView, error) {
	c, err := r.client.GetCourseByID(ctx, id)
	if err != nil {
		if apiErr, ok := err.(*clients.APIError); ok && apiErr.Status == 404 {
			return domain.CourseView{}, pgx.ErrNoRows
		}
		return domain.CourseView{}, err
	}
	return mapRemoteCourse(c), nil
}

func (r *ContentRemote) ListPublishedCourses(ctx context.Context, search *string) ([]domain.CourseView, error) {
	rows, err := r.client.ListPublishedCourses(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]domain.CourseView, 0, len(rows))
	for _, c := range rows {
		out = append(out, mapRemoteCourse(c))
	}
	return out, nil
}

func (r *ContentRemote) ListCourseModules(ctx context.Context, courseID uuid.UUID) ([]domain.ModuleView, error) {
	return nil, nil
}

func (r *ContentRemote) LessonModuleIDs(ctx context.Context, courseID uuid.UUID) (map[uuid.UUID]uuid.UUID, error) {
	return nil, nil
}

func (r *ContentRemote) SyncPublishedLessonsForCourse(ctx context.Context, q *query.Queries, courseID uuid.UUID) error {
	snaps, err := r.client.PublishedLessonSnapshots(ctx, courseID)
	if err != nil {
		return err
	}
	for _, s := range snaps {
		if err := q.UpsertPublishedLessonSnapshot(ctx, query.UpsertPublishedLessonSnapshotParams{
			LessonID: s.LessonID, CourseID: s.CourseID, Version: s.Version, Snapshot: s.Snapshot,
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *ContentRemote) BuildLessonSnapshot(ctx context.Context, lessonID uuid.UUID) (domain.LessonSnapshot, error) {
	raw, err := r.client.LessonSnapshotJSON(ctx, lessonID)
	if err != nil {
		return domain.LessonSnapshot{}, err
	}
	return domain.ParseLessonSnapshot(raw)
}

func (r *ContentRemote) GetBlock(ctx context.Context, blockID uuid.UUID) (domain.BlockSnap, uuid.UUID, string, error) {
	view, err := r.client.GetBlock(ctx, blockID)
	if err != nil {
		if apiErr, ok := err.(*clients.APIError); ok && apiErr.Status == 404 {
			return domain.BlockSnap{}, uuid.Nil, "", pgx.ErrNoRows
		}
		return domain.BlockSnap{}, uuid.Nil, "", err
	}
	var bs domain.BlockSnap
	if err := json.Unmarshal(view.Block, &bs); err != nil {
		return domain.BlockSnap{}, uuid.Nil, "", err
	}
	return bs, view.CourseID, view.LessonTitle, nil
}

func (r *ContentRemote) GetLessonTitle(ctx context.Context, lessonID uuid.UUID) (string, error) {
	return r.client.GetLessonTitle(ctx, lessonID)
}

func (r *ContentRemote) GetBlockLessonID(ctx context.Context, blockID uuid.UUID) (uuid.UUID, error) {
	return r.client.GetBlockLessonID(ctx, blockID)
}

var _ ContentSource = (*ContentRemote)(nil)

func mapRemoteCourse(c dto.CourseView) domain.CourseView {
	return domain.CourseView{
		ID: c.ID, Title: c.Title, TargetLanguageID: c.TargetLanguageID,
		TargetLangCode: c.TargetLangCode, TargetLangName: c.TargetLangName,
		UILanguageID: c.UILanguageID, OwnerID: c.OwnerID, IsPublished: c.IsPublished,
		Visibility: c.Visibility, InviteCode: c.InviteCode,
	}
}

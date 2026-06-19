package repository

import (
	"context"

	contentclient "github.com/even-app/even-app/libs/clients/content"
	"github.com/even-app/even-app/services/lexicon/internal/gen/contentquery"
	"github.com/google/uuid"
)

type ContentSource interface {
	Available() bool
	ListBlocksByCourseID(ctx context.Context, courseID uuid.UUID) ([]contentquery.ListBlocksByCourseIDRow, error)
}

var (
	_ ContentSource = (*ContentReader)(nil)
	_ ContentSource = (*ContentRemote)(nil)
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

func (r *ContentRemote) ListBlocksByCourseID(ctx context.Context, courseID uuid.UUID) ([]contentquery.ListBlocksByCourseIDRow, error) {
	rows, err := r.client.ListBlocksByCourseID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	out := make([]contentquery.ListBlocksByCourseIDRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, contentquery.ListBlocksByCourseIDRow{
			ID: row.ID, LessonID: row.LessonID, SectionID: row.SectionID,
			SortOrder: row.SortOrder, DisplayLabel: row.DisplayLabel, Title: row.Title,
			BlockType: row.BlockType, Config: row.Config, IsHomework: row.IsHomework,
			LessonTitle: row.LessonTitle, LessonSortOrder: row.LessonSortOrder,
		})
	}
	return out, nil
}

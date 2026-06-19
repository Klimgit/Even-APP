package repository

import (
	"context"

	mediaclient "github.com/even-app/even-app/libs/clients/media"
	"github.com/google/uuid"
)

type MediaRemote struct {
	client *mediaclient.Client
}

func NewMediaRemote(baseURL, token string) *MediaRemote {
	return &MediaRemote{client: mediaclient.New(baseURL, token)}
}

func (r *MediaRemote) Available() bool {
	return r != nil && r.client.Available()
}

func (r *MediaRemote) MediaByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]MediaView, error) {
	rows, err := r.client.MediaByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]MediaView, len(rows))
	for id, row := range rows {
		out[id] = MediaView{
			ID: row.ID, Scope: row.Scope, DisplayName: row.DisplayName,
			MimeType: row.MimeType, MediaKind: row.MediaKind,
		}
	}
	return out, nil
}

var _ MediaSource = (*MediaRemote)(nil)

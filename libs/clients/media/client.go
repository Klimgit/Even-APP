package media

import (
	"context"

	"github.com/even-app/even-app/libs/clients"
	"github.com/even-app/even-app/libs/clients/dto"
	"github.com/google/uuid"
)

type Client struct {
	http *clients.HTTP
}

func New(baseURL, token string) *Client {
	return &Client{http: clients.NewHTTP(baseURL, token)}
}

func (c *Client) Available() bool {
	return c != nil && c.http.Available()
}

func (c *Client) MediaByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]dto.MediaView, error) {
	var rows []dto.MediaView
	err := c.http.DoJSON(ctx, "POST", "/api/v1/internal/media/by-ids", dto.IDsRequest{IDs: ids}, &rows)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]dto.MediaView, len(rows))
	for _, row := range rows {
		out[row.ID] = row
	}
	return out, nil
}

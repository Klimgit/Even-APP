package lexicon

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

func (c *Client) GetLanguage(ctx context.Context, id uuid.UUID) (dto.LanguageView, error) {
	var out dto.LanguageView
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/languages/"+id.String(), nil, &out)
	return out, err
}

func (c *Client) LanguagesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]dto.LanguageView, error) {
	var rows []dto.LanguageView
	err := c.http.DoJSON(ctx, "POST", "/api/v1/internal/languages/by-ids", dto.IDsRequest{IDs: ids}, &rows)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]dto.LanguageView, len(rows))
	for _, row := range rows {
		out[row.ID] = row
	}
	return out, nil
}

func (c *Client) LexemesByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]dto.LexemeView, error) {
	var rows []dto.LexemeView
	err := c.http.DoJSON(ctx, "POST", "/api/v1/internal/lexemes/by-ids", dto.IDsRequest{IDs: ids}, &rows)
	if err != nil {
		return nil, err
	}
	out := make(map[uuid.UUID]dto.LexemeView, len(rows))
	for _, row := range rows {
		out[row.ID] = row
	}
	return out, nil
}

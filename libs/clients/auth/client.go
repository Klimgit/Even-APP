package auth

import (
	"context"
	"net/url"

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

func (c *Client) GetUserByEmail(ctx context.Context, email string) (dto.UserView, error) {
	var out dto.UserView
	err := c.http.DoJSON(ctx, "GET", "/api/v1/internal/users/by-email?email="+url.QueryEscape(email), nil, &out)
	return out, err
}

func (c *Client) ListUsersByIDs(ctx context.Context, ids []uuid.UUID) ([]dto.UserView, error) {
	var out []dto.UserView
	err := c.http.DoJSON(ctx, "POST", "/api/v1/internal/users/by-ids", dto.IDsRequest{IDs: ids}, &out)
	return out, err
}

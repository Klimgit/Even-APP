package handler

import (
	"context"

	libjwt "github.com/even-app/even-app/libs/jwt"
	"github.com/even-app/even-app/libs/http/middleware"
	http_v1 "github.com/even-app/even-app/services/media/internal/gen/http/v1"
)

type SecurityHandler struct {
	jwt *libjwt.Manager
}

func NewSecurityHandler(jwt *libjwt.Manager) *SecurityHandler {
	return &SecurityHandler{jwt: jwt}
}

func (s *SecurityHandler) HandleBearerAuth(ctx context.Context, _ http_v1.OperationName, t http_v1.BearerAuth) (context.Context, error) {
	claims, err := s.jwt.ParseAccess(t.Token)
	if err != nil {
		return ctx, err
	}
	return middleware.WithClaims(ctx, claims), nil
}

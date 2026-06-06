package handler

import (
	"context"
	"errors"
	"strings"

	"github.com/even-app/even-app/libs/http/middleware"
	"github.com/even-app/even-app/services/auth/internal/domain"
	http_v1 "github.com/even-app/even-app/services/auth/internal/gen/http/v1"
	"github.com/even-app/even-app/services/auth/internal/service"
)

var _ http_v1.Handler = (*HTTPHandler)(nil)

type HTTPHandler struct {
	svc *service.AuthService
}

func NewHTTPHandler(svc *service.AuthService) *HTTPHandler {
	return &HTTPHandler{svc: svc}
}

func (h *HTTPHandler) Register(ctx context.Context, req *http_v1.RegisterRequest) (http_v1.RegisterRes, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || len(req.Password) < 8 {
		return badRegisterRequest("email and password (min 8) required")
	}
	var dn *string
	if v, ok := req.DisplayName.Get(); ok && v != "" {
		dn = &v
	}
	role := ""
	if v, ok := req.Role.Get(); ok {
		role = string(v)
	}
	outcome, err := h.svc.Register(ctx, email, req.Password, role, dn)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			return conflictRegister("email already registered")
		}
		return nil, err
	}
	return authResponse(outcome), nil
}

func (h *HTTPHandler) Login(ctx context.Context, req *http_v1.LoginRequest) (http_v1.LoginRes, error) {
	outcome, err := h.svc.Login(ctx, strings.TrimSpace(strings.ToLower(req.Email)), req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return unauthorizedLogin("invalid credentials")
		}
		return nil, err
	}
	return authResponse(outcome), nil
}

func (h *HTTPHandler) Refresh(ctx context.Context, req *http_v1.RefreshRequest) (http_v1.RefreshRes, error) {
	tokens, err := h.svc.Refresh(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, domain.ErrUnauthorized) {
			return unauthorizedRefresh("invalid refresh token")
		}
		return nil, err
	}
	return &http_v1.TokensResponse{
		AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken,
	}, nil
}

func (h *HTTPHandler) GetMe(ctx context.Context) (http_v1.GetMeRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	u, err := h.svc.Me(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	user := mapUser(*u)
	return &user, nil
}

func authResponse(outcome *domain.AuthOutcome) *http_v1.AuthResponse {
	return &http_v1.AuthResponse{
		AccessToken: outcome.Tokens.AccessToken, RefreshToken: outcome.Tokens.RefreshToken,
		User: mapUser(*outcome.User),
	}
}

func mapUser(u domain.User) http_v1.User {
	out := http_v1.User{
		ID: u.ID, Email: u.Email, Role: http_v1.UserRole(u.Role),
		IsAdmin: u.IsAdmin, CreatedAt: u.CreatedAt,
	}
	if u.DisplayName != nil {
		out.DisplayName = http_v1.NewOptString(*u.DisplayName)
	}
	return out
}

func badRegisterRequest(msg string) (*http_v1.RegisterBadRequest, error) {
	r := http_v1.RegisterBadRequest(errBody(msg))
	return &r, nil
}

func conflictRegister(msg string) (*http_v1.RegisterConflict, error) {
	r := http_v1.RegisterConflict(errBody(msg))
	return &r, nil
}

func unauthorizedLogin(msg string) (*http_v1.ErrorResponse, error) {
	r := errBody(msg)
	return &r, nil
}

func unauthorizedRefresh(msg string) (*http_v1.ErrorResponse, error) {
	r := errBody(msg)
	return &r, nil
}

func errBody(msg string) http_v1.ErrorResponse {
	return http_v1.ErrorResponse{
		Message: http_v1.NewOptString(msg),
		Error:   http_v1.NewOptString(msg),
	}
}

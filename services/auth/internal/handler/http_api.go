package handler

import (
	"context"
	"encoding/json"
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

func (h *HTTPHandler) ListPlatformUsers(ctx context.Context, params http_v1.ListPlatformUsersParams) (http_v1.ListPlatformUsersRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	q, role := "", ""
	if v, ok := params.Q.Get(); ok {
		q = v
	}
	if v, ok := params.Role.Get(); ok {
		role = string(v)
	}
	page, limit := 1, 20
	if v, ok := params.Page.Get(); ok {
		page = v
	}
	if v, ok := params.Limit.Get(); ok {
		limit = v
	}
	out, err := h.svc.ListPlatformUsers(ctx, claims.IsAdmin, q, role, page, limit)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenListPlatformUsers()
		}
		return nil, err
	}
	items := make([]http_v1.User, 0, len(out.Items))
	for _, u := range out.Items {
		items = append(items, mapUser(u))
	}
	return &http_v1.UserListResponse{Items: items, Total: out.Total}, nil
}

func (h *HTTPHandler) PatchPlatformUser(ctx context.Context, req *http_v1.PatchPlatformUserRequest, params http_v1.PatchPlatformUserParams) (http_v1.PatchPlatformUserRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	in := service.PatchPlatformUserInput{}
	if v, ok := req.Role.Get(); ok {
		s := string(v)
		in.Role = &s
	}
	if v, ok := req.IsAdmin.Get(); ok {
		in.IsAdmin = &v
	}
	u, err := h.svc.PatchPlatformUser(ctx, claims.IsAdmin, claims.UserID, params.UserId, in)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenPatchPlatformUser()
		}
		if errors.Is(err, domain.ErrNotFound) {
			return notFoundPatchPlatformUser()
		}
		return nil, err
	}
	user := mapUser(*u)
	return &user, nil
}

func (h *HTTPHandler) CreatePlatformUser(ctx context.Context, req *http_v1.CreatePlatformUserRequest) (http_v1.CreatePlatformUserRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if email == "" || len(req.Password) < 8 {
		return nil, domain.ErrValidation
	}
	var dn *string
	if v, ok := req.DisplayName.Get(); ok && v != "" {
		dn = &v
	}
	isAdmin := false
	if v, ok := req.IsAdmin.Get(); ok {
		isAdmin = v
	}
	u, err := h.svc.CreatePlatformUser(ctx, claims.IsAdmin, claims.UserID, service.CreatePlatformUserInput{
		Email: email, Password: req.Password, DisplayName: dn,
		Role: string(req.Role), IsAdmin: isAdmin,
	})
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenCreatePlatformUser()
		}
		if errors.Is(err, domain.ErrConflict) {
			return conflictCreatePlatformUser()
		}
		return nil, err
	}
	user := mapUser(*u)
	return &user, nil
}

func (h *HTTPHandler) GetPlatformUser(ctx context.Context, params http_v1.GetPlatformUserParams) (http_v1.GetPlatformUserRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	u, err := h.svc.GetPlatformUser(ctx, claims.IsAdmin, params.UserId)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenGetPlatformUser()
		}
		if errors.Is(err, domain.ErrNotFound) {
			return notFoundGetPlatformUser()
		}
		return nil, err
	}
	user := mapUser(*u)
	return &user, nil
}

func (h *HTTPHandler) DeletePlatformUser(ctx context.Context, params http_v1.DeletePlatformUserParams) (http_v1.DeletePlatformUserRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	if err := h.svc.DeletePlatformUser(ctx, claims.IsAdmin, claims.UserID, params.UserId); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenDeletePlatformUser()
		}
		if errors.Is(err, domain.ErrNotFound) {
			return notFoundDeletePlatformUser()
		}
		if errors.Is(err, domain.ErrValidation) {
			return forbiddenDeletePlatformUser()
		}
		return nil, err
	}
	return &http_v1.DeletePlatformUserNoContent{}, nil
}

func (h *HTTPHandler) ResetPlatformUserPassword(ctx context.Context, req *http_v1.ResetPlatformUserPasswordRequest, params http_v1.ResetPlatformUserPasswordParams) (http_v1.ResetPlatformUserPasswordRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	if len(req.Password) < 8 {
		return nil, domain.ErrValidation
	}
	if err := h.svc.ResetPlatformUserPassword(ctx, claims.IsAdmin, claims.UserID, params.UserId, req.Password); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenResetPlatformUserPassword()
		}
		if errors.Is(err, domain.ErrNotFound) {
			return notFoundResetPlatformUserPassword()
		}
		return nil, err
	}
	return &http_v1.ResetPlatformUserPasswordNoContent{}, nil
}

func (h *HTTPHandler) ListPlatformAudit(ctx context.Context, params http_v1.ListPlatformAuditParams) (http_v1.ListPlatformAuditRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	action := ""
	if v, ok := params.Action.Get(); ok {
		action = v
	}
	page, limit := 1, 50
	if v, ok := params.Page.Get(); ok && v > 0 {
		page = v
	}
	if v, ok := params.Limit.Get(); ok && v > 0 {
		limit = v
	}
	out, err := h.svc.ListPlatformAudit(ctx, claims.IsAdmin, action, page, limit)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenListPlatformAudit()
		}
		return nil, err
	}
	items := make([]http_v1.AuditEvent, 0, len(out.Items))
	for _, e := range out.Items {
		ev := http_v1.AuditEvent{
			ID: e.ID, Action: e.Action, CreatedAt: e.CreatedAt,
		}
		if e.ActorID != nil {
			ev.ActorID = http_v1.NewOptUUID(*e.ActorID)
		}
		if e.TargetType != nil {
			ev.TargetType = http_v1.NewOptString(*e.TargetType)
		}
		if e.TargetID != nil {
			ev.TargetID = http_v1.NewOptUUID(*e.TargetID)
		}
		if e.Details != nil {
			if b, err := json.Marshal(e.Details); err == nil {
				var raw http_v1.AuditEventDetails
				if err := json.Unmarshal(b, &raw); err == nil {
					ev.Details = http_v1.NewOptAuditEventDetails(raw)
				}
			}
		}
		items = append(items, ev)
	}
	return &http_v1.AuditListResponse{Items: items, Total: out.Total}, nil
}

func (h *HTTPHandler) GetPlatformStats(ctx context.Context) (http_v1.GetPlatformStatsRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	stats, err := h.svc.GetPlatformStats(ctx, claims.IsAdmin)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenGetPlatformStats()
		}
		return nil, err
	}
	return &http_v1.PlatformStatsResponse{
		Users: http_v1.PlatformStatsResponseUsers{
			Total: stats.TotalUsers, Students: stats.Students,
			Teachers: stats.Teachers, Admins: stats.Admins,
		},
		PublishedCourses:  stats.PublishedCourses,
		ActiveEnrollments: stats.ActiveEnrollments,
		TotalCourses:      stats.TotalCourses,
		PlatformLexemes:   stats.PlatformLexemes,
	}, nil
}

func (h *HTTPHandler) DemoNotes(ctx context.Context) (*http_v1.DemoNotesResponse, error) {
	notes, err := h.svc.ListDemoNotes(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]http_v1.DemoNote, 0, len(notes))
	for _, n := range notes {
		items = append(items, http_v1.DemoNote{
			ID: n.ID, Text: n.Text, CreatedAt: n.CreatedAt,
		})
	}
	return &http_v1.DemoNotesResponse{Items: items}, nil
}

func (h *HTTPHandler) DemoPublic(ctx context.Context) (*http_v1.DemoPublicResponse, error) {
	count, err := h.svc.DemoPublic(ctx)
	if err != nil {
		return nil, err
	}
	return &http_v1.DemoPublicResponse{
		Message:   "public demo: user count from even_auth.users",
		UserCount: int32(count),
	}, nil
}

func (h *HTTPHandler) DemoMe(ctx context.Context) (http_v1.DemoMeRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	u, err := h.svc.DemoMe(ctx, claims.UserID)
	if err != nil {
		return nil, err
	}
	return &http_v1.DemoMeResponse{
		User:         mapUser(*u),
		TokenRole:    http_v1.DemoMeResponseTokenRole(claims.Role),
		TokenIsAdmin: claims.IsAdmin,
	}, nil
}

func (h *HTTPHandler) DemoTeacher(ctx context.Context) (http_v1.DemoTeacherRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	u, err := h.svc.DemoTeacher(ctx, claims.UserID, claims.Role, claims.IsAdmin)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenDemoTeacher()
		}
		return nil, err
	}
	name := u.Email
	if u.DisplayName != nil && *u.DisplayName != "" {
		name = *u.DisplayName
	}
	return &http_v1.DemoTeacherResponse{
		Message:     "teacher demo: role check passed, profile loaded from DB",
		Role:        http_v1.DemoTeacherResponseRole(u.Role),
		DisplayName: name,
	}, nil
}

func (h *HTTPHandler) DemoAdminStats(ctx context.Context) (http_v1.DemoAdminStatsRes, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, domain.ErrUnauthorized
	}
	stats, err := h.svc.DemoAdminStats(ctx, claims.IsAdmin)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			return forbiddenDemoAdminStats()
		}
		return nil, err
	}
	return &http_v1.DemoAdminStatsResponse{
		TotalUsers: int32(stats.TotalUsers),
		Students:   int32(stats.Students),
		Teachers:   int32(stats.Teachers),
		Admins:     int32(stats.Admins),
	}, nil
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

func forbiddenDemoTeacher() (*http_v1.DemoTeacherForbidden, error) {
	r := http_v1.DemoTeacherForbidden(errBody("teacher role required"))
	return &r, nil
}

func forbiddenDemoAdminStats() (*http_v1.DemoAdminStatsForbidden, error) {
	r := http_v1.DemoAdminStatsForbidden(errBody("platform admin required"))
	return &r, nil
}

func forbiddenListPlatformUsers() (*http_v1.ListPlatformUsersForbidden, error) {
	r := http_v1.ListPlatformUsersForbidden(errBody("platform admin required"))
	return &r, nil
}

func forbiddenPatchPlatformUser() (*http_v1.PatchPlatformUserForbidden, error) {
	r := http_v1.PatchPlatformUserForbidden(errBody("platform admin required"))
	return &r, nil
}

func forbiddenGetPlatformStats() (*http_v1.GetPlatformStatsForbidden, error) {
	r := http_v1.GetPlatformStatsForbidden(errBody("platform admin required"))
	return &r, nil
}

func notFoundPatchPlatformUser() (*http_v1.PatchPlatformUserNotFound, error) {
	r := http_v1.PatchPlatformUserNotFound(errBody("user not found"))
	return &r, nil
}

func forbiddenCreatePlatformUser() (*http_v1.CreatePlatformUserForbidden, error) {
	r := http_v1.CreatePlatformUserForbidden(errBody("platform admin required"))
	return &r, nil
}

func conflictCreatePlatformUser() (*http_v1.CreatePlatformUserConflict, error) {
	r := http_v1.CreatePlatformUserConflict(errBody("email already registered"))
	return &r, nil
}

func forbiddenGetPlatformUser() (*http_v1.GetPlatformUserForbidden, error) {
	r := http_v1.GetPlatformUserForbidden(errBody("platform admin required"))
	return &r, nil
}

func notFoundGetPlatformUser() (*http_v1.GetPlatformUserNotFound, error) {
	r := http_v1.GetPlatformUserNotFound(errBody("user not found"))
	return &r, nil
}

func forbiddenDeletePlatformUser() (*http_v1.DeletePlatformUserForbidden, error) {
	r := http_v1.DeletePlatformUserForbidden(errBody("platform admin required"))
	return &r, nil
}

func notFoundDeletePlatformUser() (*http_v1.DeletePlatformUserNotFound, error) {
	r := http_v1.DeletePlatformUserNotFound(errBody("user not found"))
	return &r, nil
}

func forbiddenResetPlatformUserPassword() (*http_v1.ResetPlatformUserPasswordForbidden, error) {
	r := http_v1.ResetPlatformUserPasswordForbidden(errBody("platform admin required"))
	return &r, nil
}

func notFoundResetPlatformUserPassword() (*http_v1.ResetPlatformUserPasswordNotFound, error) {
	r := http_v1.ResetPlatformUserPasswordNotFound(errBody("user not found"))
	return &r, nil
}

func forbiddenListPlatformAudit() (*http_v1.ListPlatformAuditForbidden, error) {
	r := http_v1.ListPlatformAuditForbidden(errBody("platform admin required"))
	return &r, nil
}

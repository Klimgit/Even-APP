package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/even-app/even-app/services/auth/internal/domain"
	"github.com/even-app/even-app/services/auth/internal/gen/query"
	"github.com/google/uuid"
)

type AuditEvent struct {
	ID         uuid.UUID
	ActorID    *uuid.UUID
	Action     string
	TargetType *string
	TargetID   *uuid.UUID
	Details    map[string]any
	CreatedAt  time.Time
}

type AuditListResult struct {
	Items []AuditEvent
	Total int
}

func (s *AuthService) LogAudit(ctx context.Context, actorID *uuid.UUID, action, targetType string, targetID *uuid.UUID, details map[string]any) error {
	var detailsJSON []byte
	if details != nil {
		var err error
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			return err
		}
	}
	var tt *string
	if targetType != "" {
		tt = &targetType
	}
	return s.q.InsertAuditEvent(ctx, query.InsertAuditEventParams{
		ActorID: actorID, Action: action, TargetType: tt, TargetID: targetID, Details: detailsJSON,
	})
}

func (s *AuthService) ListPlatformAudit(ctx context.Context, isAdmin bool, action string, page, limit int) (*AuditListResult, error) {
	if !isAdmin {
		return nil, domain.ErrForbidden
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	offset := (page - 1) * limit
	total, err := s.q.CountAuditEvents(ctx, query.CountAuditEventsParams{Column1: action})
	if err != nil {
		return nil, err
	}
	rows, err := s.q.ListAuditEvents(ctx, query.ListAuditEventsParams{
		Column1: action, Offset: int32(offset), Limit: int32(limit),
	})
	if err != nil {
		return nil, err
	}
	items := make([]AuditEvent, 0, len(rows))
	for _, row := range rows {
		var details map[string]any
		if len(row.Details) > 0 {
			_ = json.Unmarshal(row.Details, &details)
		}
		items = append(items, AuditEvent{
			ID: row.ID, ActorID: row.ActorID, Action: row.Action,
			TargetType: row.TargetType, TargetID: row.TargetID,
			Details: details, CreatedAt: row.CreatedAt,
		})
	}
	return &AuditListResult{Items: items, Total: int(total)}, nil
}

package handler

import (
	"context"
	"errors"

	"github.com/even-app/even-app/services/auth/internal/domain"
	http_v1 "github.com/even-app/even-app/services/auth/internal/gen/http/v1"
)

const (
	defaultHTTPErrorCode = 500
	defaultUserMessage   = "internal error"
)

var errToHTTPStatus = map[error]int{
	domain.ErrNotFound:     404,
	domain.ErrConflict:     409,
	domain.ErrUnauthorized: 401,
}

func (h *HTTPHandler) NewError(ctx context.Context, err error) *http_v1.DefaultErrorStatusCode {
	status := defaultHTTPErrorCode
	for target, code := range errToHTTPStatus {
		if errors.Is(err, target) {
			status = code
			break
		}
	}
	msg := err.Error()
	if status == defaultHTTPErrorCode {
		msg = defaultUserMessage
	}
	return &http_v1.DefaultErrorStatusCode{
		StatusCode: status,
		Response: http_v1.ErrorResponse{
			Message: http_v1.NewOptString(msg),
			Error:   http_v1.NewOptString(msg),
			Code:    http_v1.NewOptInt32(int32(status)),
		},
	}
}

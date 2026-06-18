package handler

import (
	"context"
	"errors"

	"github.com/even-app/even-app/services/learning/internal/domain"
	http_v1 "github.com/even-app/even-app/services/learning/internal/gen/http/v1"
)

const (
	defaultHTTPErrorCode = 500
	defaultUserMessage   = "internal error"
)

var errToHTTPStatus = map[error]int{
	domain.ErrNotFound:     404,
	domain.ErrConflict:     409,
	domain.ErrUnauthorized: 401,
	domain.ErrForbidden:    403,
	domain.ErrValidation:   400,
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

func errBody(msg string) http_v1.ErrorResponse {
	return http_v1.ErrorResponse{
		Message: http_v1.NewOptString(msg),
		Error:   http_v1.NewOptString(msg),
	}
}

func notFoundJoin(msg string) (*http_v1.JoinCourseNotFound, error) {
	r := http_v1.JoinCourseNotFound(errBody(msg))
	return &r, nil
}

func conflictJoin(msg string) (*http_v1.JoinCourseConflict, error) {
	r := http_v1.JoinCourseConflict(errBody(msg))
	return &r, nil
}

func notFound(msg string) (*http_v1.ErrorResponse, error) {
	r := errBody(msg)
	return &r, nil
}

func forbidden(msg string) (*http_v1.ErrorResponse, error) {
	r := errBody(msg)
	return &r, nil
}

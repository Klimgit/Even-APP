package handler

import (
	"context"

	http_v1 "github.com/even-app/even-app/services/media/internal/gen/http/v1"
)

const (
	defaultHTTPErrorCode = 500
	defaultUserMessage   = "internal error"
)

func (h *HTTPHandler) NewError(ctx context.Context, err error) *http_v1.DefaultErrorStatusCode {
	msg := defaultUserMessage
	if err != nil {
		msg = err.Error()
	}
	return &http_v1.DefaultErrorStatusCode{
		StatusCode: defaultHTTPErrorCode,
		Response: http_v1.ErrorResponse{
			Message: http_v1.NewOptString(msg),
			Error:   http_v1.NewOptString(msg),
			Code:    http_v1.NewOptInt32(int32(defaultHTTPErrorCode)),
		},
	}
}

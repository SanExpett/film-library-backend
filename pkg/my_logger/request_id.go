package my_logger

import "context"

type keyCtx string

const (
	requestIDKey keyCtx = "req_id"

	minRequestID = 100000
	maxRequestID = 999999
)

func GetRequestIDFromCtx(ctx context.Context) string {
	requestID, ok := ctx.Value(requestIDKey).(string)
	if !ok {
		return ""
	}

	return requestID
}

package observability

import (
	"context"
	"log/slog"
)

type ctxKey int

const (
	ctxRequestID ctxKey = iota
	ctxLogger
	ctxHandle
)

func withRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, ctxRequestID, id)
}

func withLogger(ctx context.Context, logger *slog.Logger) context.Context {
	return context.WithValue(ctx, ctxLogger, logger)
}

func withHandle(ctx context.Context, h *Handle) context.Context {
	return context.WithValue(ctx, ctxHandle, h)
}

func handleFrom(ctx context.Context) *Handle {
	if ctx == nil {
		return nil
	}
	h, _ := ctx.Value(ctxHandle).(*Handle)
	return h
}

// RequestID returns the request ID stored on ctx, or empty.
func RequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	id, _ := ctx.Value(ctxRequestID).(string)
	return id
}

// Logger returns the request-scoped slog logger, or slog.Default().
func Logger(ctx context.Context) *slog.Logger {
	if ctx != nil {
		if l, ok := ctx.Value(ctxLogger).(*slog.Logger); ok && l != nil {
			return l
		}
	}
	return slog.Default()
}

package observability

import (
	"context"
	"log/slog"
)

const errorReportingType = "type.googleapis.com/google.devtools.clouderrorreporting.v1beta1.ReportedErrorEvent"

// ReportError logs at ERROR with GCP Error Reporting payload fields.
// Does not include request bodies or PII. Pass only safe slog attrs.
func ReportError(ctx context.Context, err error, attrs ...any) {
	if err == nil {
		return
	}
	logger := Logger(ctx)
	args := []any{
		slog.String("@type", errorReportingType),
		slog.String("message", err.Error()),
		slog.String("error", err.Error()),
	}
	args = append(args, attrs...)
	logger.ErrorContext(ctx, "reported_error", args...)
}

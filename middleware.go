package observability

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *statusRecorder) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(b)
}

func (w *statusRecorder) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w *statusRecorder) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// Handler wraps next with request ID, slog fields, and otelhttp RED/traces.
func (h *Handle) Handler(next http.Handler) http.Handler {
	if h == nil {
		return next
	}
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		id := incomingRequestID(r)
		w.Header().Set(HeaderRequestID, id)

		span := trace.SpanFromContext(r.Context())
		traceID := ""
		if span.SpanContext().HasTraceID() {
			traceID = span.SpanContext().TraceID().String()
			span.SetAttributes(attribute.String("request.id", id))
		}

		logger := h.Logger.With(
			slog.String("request_id", id),
			slog.String("trace_id", traceID),
			slog.String("method", r.Method),
			slog.String("path", routeLabel(r)),
		)
		ctx := withHandle(withLogger(withRequestID(r.Context(), id), logger), h)
		rec := &statusRecorder{ResponseWriter: w, status: 0}
		next.ServeHTTP(rec, r.WithContext(ctx))
		status := rec.status
		if status == 0 {
			status = http.StatusOK
		}
		logger.InfoContext(ctx, "http_request",
			slog.Int("status", status),
			slog.String("status_class", statusClass(status)),
			slog.Int64("duration_ms", time.Since(start).Milliseconds()),
			slog.String("platform", strings.TrimSpace(r.Header.Get("X-Client-Platform"))),
		)
		if status >= 500 {
			ReportError(ctx, errStatus(status), slog.Int("status", status))
		}
	})

	return otelhttp.NewHandler(inner, h.cfg.ServiceName,
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			return r.Method + " " + routeLabel(r)
		}),
		otelhttp.WithMetricAttributesFn(func(r *http.Request) []attribute.KeyValue {
			return []attribute.KeyValue{
				attribute.String("http.route", routeLabel(r)),
				attribute.String("platform", platformLabel(r)),
			}
		}),
	)
}

func platformLabel(r *http.Request) string {
	p := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Client-Platform")))
	switch p {
	case "web", "expo_web", "ios", "android":
		return p
	default:
		return "unknown"
	}
}

func routeLabel(r *http.Request) string {
	if r.Pattern != "" {
		return r.Pattern
	}
	p := r.URL.Path
	switch {
	case p == "/health":
		return "/health"
	case p == "/api/contact":
		return "/api/contact"
	case p == "/analytics-config.js":
		return "/analytics-config.js"
	case p == "/privacy" || p == "/privacy.html":
		return "/privacy"
	case strings.HasPrefix(p, "/assets/"):
		return "/assets/*"
	case p == "/" || p == "":
		return "/"
	default:
		return "/_other"
	}
}

func statusClass(status int) string {
	switch {
	case status >= 500:
		return "5xx"
	case status >= 400:
		return "4xx"
	case status >= 300:
		return "3xx"
	case status >= 200:
		return "2xx"
	default:
		return "0xx"
	}
}

type statusError int

func errStatus(code int) error {
	return statusError(code)
}

func (e statusError) Error() string {
	return http.StatusText(int(e))
}

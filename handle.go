package observability

import (
	"context"
	"log/slog"
	"os"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// Handle holds process-wide observability state.
type Handle struct {
	cfg    Config
	Logger *slog.Logger

	tp *sdktrace.TracerProvider
	mp *sdkmetric.MeterProvider

	shutdownOnce sync.Once
}

// Init configures slog, optional OTel exporters, and W3C propagation.
func Init(ctx context.Context, cfg Config) (*Handle, error) {
	if cfg.ServiceName == "" {
		cfg.ServiceName = "unknown-service"
	}
	if cfg.Exporter == "" {
		cfg.Exporter = ExporterNone
	}

	logger := newJSONLogger(cfg)
	slog.SetDefault(logger)

	h := &Handle{cfg: cfg, Logger: logger}

	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	if err := h.initOTel(ctx, cfg); err != nil {
		return nil, err
	}

	h.Logger.InfoContext(ctx, "observability_init",
		slog.String("service", cfg.ServiceName),
		slog.String("env", cfg.Env),
		slog.String("exporter", cfg.Exporter),
	)
	return h, nil
}

func newJSONLogger(cfg Config) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	return slog.New(handler).With(
		slog.String("service", cfg.ServiceName),
		slog.String("env", cfg.Env),
	)
}

// Shutdown flushes OTel exporters. Safe to call more than once.
func (h *Handle) Shutdown(ctx context.Context) error {
	if h == nil {
		return nil
	}
	var err error
	h.shutdownOnce.Do(func() {
		if h.tp != nil {
			err = h.tp.Shutdown(ctx)
		}
		if h.mp != nil {
			if e := h.mp.Shutdown(ctx); e != nil && err == nil {
				err = e
			}
		}
	})
	return err
}

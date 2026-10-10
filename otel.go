package observability

import (
	"context"
	"fmt"
	"strings"

	mexporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/metric"
	texporter "github.com/GoogleCloudPlatform/opentelemetry-operations-go/exporter/trace"
	"go.opentelemetry.io/contrib/detectors/gcp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

func (h *Handle) initOTel(ctx context.Context, cfg Config) error {
	switch strings.ToLower(cfg.Exporter) {
	case ExporterNone, "":
		return nil
	case ExporterGCP:
		if cfg.ProjectID == "" {
			return fmt.Errorf("OTEL_EXPORTER=gcp requires GCP_PROJECT_ID")
		}
		res, err := newResource(ctx, cfg)
		if err != nil {
			return err
		}
		tExp, err := texporter.New(texporter.WithProjectID(cfg.ProjectID))
		if err != nil {
			return fmt.Errorf("trace exporter: %w", err)
		}
		h.tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(tExp),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(h.tp)

		mExp, err := mexporter.New(mexporter.WithProjectID(cfg.ProjectID))
		if err != nil {
			return fmt.Errorf("metric exporter: %w", err)
		}
		h.mp = metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(mExp)),
			metric.WithResource(res),
		)
		otel.SetMeterProvider(h.mp)
		return nil
	case ExporterStdout:
		res, err := newResource(ctx, cfg)
		if err != nil {
			return err
		}
		tExp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return err
		}
		h.tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(tExp),
			sdktrace.WithResource(res),
		)
		otel.SetTracerProvider(h.tp)
		mExp, err := stdoutmetric.New()
		if err != nil {
			return err
		}
		h.mp = metric.NewMeterProvider(
			metric.WithReader(metric.NewPeriodicReader(mExp)),
			metric.WithResource(res),
		)
		otel.SetMeterProvider(h.mp)
		return nil
	default:
		return fmt.Errorf("unknown OTEL_EXPORTER %q (use gcp, stdout, or none)", cfg.Exporter)
	}
}

func newResource(ctx context.Context, cfg Config) (*resource.Resource, error) {
	return resource.New(ctx,
		resource.WithDetectors(gcp.NewDetector()),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.DeploymentEnvironment(cfg.Env),
		),
	)
}

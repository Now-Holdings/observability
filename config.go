package observability

import (
	"os"
	"strings"
)

const (
	ExporterGCP    = "gcp"
	ExporterStdout = "stdout"
	ExporterNone   = "none"
)

// Config is read from process environment (no secrets).
type Config struct {
	ServiceName string
	Env         string
	Exporter    string
	ProjectID   string
}

func ConfigFromEnv() Config {
	service := firstNonEmpty(
		os.Getenv("OTEL_SERVICE_NAME"),
		os.Getenv("K_SERVICE"),
		"unknown-service",
	)
	env := firstNonEmpty(os.Getenv("APP_ENV"), os.Getenv("ENV"), "development")
	exporter := strings.ToLower(strings.TrimSpace(os.Getenv("OTEL_EXPORTER")))
	if exporter == "" {
		exporter = ExporterNone
	}
	return Config{
		ServiceName: service,
		Env:         env,
		Exporter:    exporter,
		ProjectID:   strings.TrimSpace(os.Getenv("GCP_PROJECT_ID")),
	}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

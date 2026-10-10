# observability

Go module `github.com/now-holdings/observability` for Now Holdings Cloud Run services.

- Request ID (`X-Request-Id`) + JSON `slog` correlation (`service`, `env`, `request_id`, `trace_id`)
- OpenTelemetry export to **Cloud Trace** + **Cloud Monitoring** (RED via `otelhttp`)
- GCP **Error Reporting** hooks via structured logs (no Sentry)

## Usage

```go
obs, err := observability.Init(ctx, observability.ConfigFromEnv())
if err != nil {
    log.Fatal(err)
}
defer obs.Shutdown(context.Background())

mux := http.NewServeMux()
handler := obs.Handler(mux)
```

## Environment

| Variable | Meaning |
| --- | --- |
| `OTEL_SERVICE_NAME` | Service name (e.g. `aundar-customer-prod`) |
| `OTEL_EXPORTER` | `gcp` \| `stdout` \| `none` (local default `none`) |
| `GCP_PROJECT_ID` | Required when exporter is `gcp` |
| `APP_ENV` or `ENV` | `development` / `production` |

Never log secrets. Request IDs are UUIDs; do not put PII in span attributes.

## Versions

- **v0.1** — request ID + slog
- **v0.2** — OTel GCP exporters + Error Reporting helper

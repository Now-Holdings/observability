package observability

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMiddlewareRequestIDAndSlog(t *testing.T) {
	rOut, wOut, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = wOut
	t.Cleanup(func() { os.Stdout = old })

	h, err := Init(context.Background(), Config{
		ServiceName: "test-api",
		Env:         "test",
		Exporter:    ExporterNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = h.Shutdown(context.Background()) })

	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if RequestID(r.Context()) == "" {
			t.Error("missing request id on context")
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := httptest.NewServer(h.Handler(mux))
	t.Cleanup(srv.Close)

	req, err := http.NewRequest(http.MethodGet, srv.URL+"/health", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set(HeaderRequestID, "abc-fixed-id")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if got := resp.Header.Get(HeaderRequestID); got != "abc-fixed-id" {
		t.Fatalf("X-Request-Id = %q", got)
	}
	if string(body) != "ok" {
		t.Fatalf("body %q", body)
	}

	_ = wOut.Close()
	logged, _ := io.ReadAll(rOut)
	if !strings.Contains(string(logged), `"request_id":"abc-fixed-id"`) {
		t.Fatalf("log missing request_id: %s", logged)
	}
	if !strings.Contains(string(logged), `"service":"test-api"`) {
		t.Fatalf("log missing service: %s", logged)
	}
}

func TestReportErrorJSONType(t *testing.T) {
	var buf strings.Builder
	logger := slog.New(slog.NewJSONHandler(&buf, nil))
	ctx := withLogger(context.Background(), logger)
	ReportError(ctx, io.EOF)
	var payload map[string]any
	if err := json.Unmarshal([]byte(strings.TrimSpace(buf.String())), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["@type"] != errorReportingType {
		t.Fatalf("@type = %v", payload["@type"])
	}
}

func TestConfigFromEnvDefaultExporterNone(t *testing.T) {
	t.Setenv("OTEL_SERVICE_NAME", "svc")
	t.Setenv("OTEL_EXPORTER", "")
	t.Setenv("APP_ENV", "development")
	cfg := ConfigFromEnv()
	if cfg.Exporter != ExporterNone {
		t.Fatalf("exporter %q", cfg.Exporter)
	}
	if cfg.ServiceName != "svc" {
		t.Fatalf("service %q", cfg.ServiceName)
	}
}

package telego

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	apiTrace "go.opentelemetry.io/otel/trace"
)

func TestSpanExporterCapturesAndSanitizesRequest(t *testing.T) {
	store := NewStore(10)
	provider := trace.NewTracerProvider(trace.WithSyncer(NewSpanExporter(store)))
	tracer := provider.Tracer("github.com/goravel/framework/telemetry/instrumentation/http")

	_, span := tracer.Start(context.Background(), "GET /users", apiTrace.WithSpanKind(apiTrace.SpanKindServer))
	span.SetAttributes(
		attribute.String("http.request.method", "GET"),
		attribute.String("url.full", "https://example.test/users?api_key=secret&active=true"),
		attribute.Int("http.response.status_code", 500),
		attribute.String("http.request.header.authorization", "Bearer secret"),
	)
	span.SetStatus(codes.Error, "request failed")
	span.End()

	entries := store.List(Filters{Limit: 10})
	require.Len(t, entries, 2)
	require.Equal(t, "exception", entries[0].Type)
	require.Equal(t, "request", entries[1].Type)
	require.Equal(t, entries[0].TraceID, entries[1].TraceID)
	require.Equal(t, "GET", entries[1].Method)
	require.Equal(t, 500, entries[1].StatusCode)
	require.Contains(t, entries[1].URI, "api_key=%5BREDACTED%5D")

	attributes := entries[1].Payload["attributes"].(map[string]any)
	require.Equal(t, "[REDACTED]", attributes["http.request.header.authorization"])
}

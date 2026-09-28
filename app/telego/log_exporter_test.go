package telego

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

func TestLogExporterCapturesAndRedactsRecord(t *testing.T) {
	store := NewStore(10)
	provider := sdklog.NewLoggerProvider(sdklog.WithProcessor(sdklog.NewSimpleProcessor(NewLogExporter(store))))
	logger := provider.Logger("telego-test")
	record := log.Record{}
	record.SetTimestamp(time.Now())
	record.SetSeverity(log.SeverityError)
	record.SetSeverityText("ERROR")
	record.SetBody(log.StringValue("payment failed"))
	record.AddAttributes(log.KeyValue{Key: "api_token", Value: log.StringValue("secret")})

	logger.Emit(context.Background(), record)

	entries := store.List(Filters{Type: "log", Limit: 10})
	require.Len(t, entries, 1)
	require.Equal(t, "payment failed", entries[0].Name)
	require.Equal(t, "ERROR", entries[0].Method)
	require.Equal(t, "error", entries[0].Status)
	attributes := entries[0].Payload["attributes"].(map[string]any)
	require.Equal(t, "[REDACTED]", attributes["api_token"])
}

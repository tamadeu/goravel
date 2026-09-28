package telego

import (
	"context"
	"strings"
	"time"

	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
)

type LogExporter struct {
	store *Store
}

func NewLogExporter(store *Store) *LogExporter {
	return &LogExporter{store: store}
}

func (r *LogExporter) Export(_ context.Context, records []sdklog.Record) error {
	for _, record := range records {
		attributes := make(map[string]any, record.AttributesLen())
		record.WalkAttributes(func(attribute log.KeyValue) bool {
			attributes[attribute.Key] = sanitize(attribute.Key, logValue(attribute.Value))
			return true
		})

		body := logValue(record.Body())
		name, ok := body.(string)
		if !ok || name == "" {
			name = record.EventName()
		}
		if name == "" {
			name = "log"
		}

		status := "ok"
		if record.Severity() >= log.SeverityError {
			status = "error"
		}
		timestamp := record.Timestamp()
		if timestamp.IsZero() {
			timestamp = record.ObservedTimestamp()
		}
		if timestamp.IsZero() {
			timestamp = time.Now()
		}

		r.store.Add(Entry{
			Type:      "log",
			Name:      name,
			TraceID:   record.TraceID().String(),
			SpanID:    record.SpanID().String(),
			Method:    strings.ToUpper(record.SeverityText()),
			Status:    status,
			Timestamp: timestamp.UTC(),
			Payload: map[string]any{
				"attributes":      Redact(attributes),
				"body":            Redact(body),
				"event_name":      record.EventName(),
				"instrumentation": record.InstrumentationScope().Name,
				"severity":        record.Severity().String(),
			},
		})
	}

	return nil
}

func (r *LogExporter) Shutdown(context.Context) error {
	return nil
}

func (r *LogExporter) ForceFlush(context.Context) error {
	return nil
}

func logValue(value log.Value) any {
	switch value.Kind() {
	case log.KindBool:
		return value.AsBool()
	case log.KindFloat64:
		return value.AsFloat64()
	case log.KindInt64:
		return value.AsInt64()
	case log.KindString:
		return value.AsString()
	case log.KindBytes:
		return string(value.AsBytes())
	case log.KindSlice:
		items := value.AsSlice()
		converted := make([]any, len(items))
		for index, item := range items {
			converted[index] = logValue(item)
		}
		return converted
	case log.KindMap:
		converted := make(map[string]any)
		for _, item := range value.AsMap() {
			converted[item.Key] = logValue(item.Value)
		}
		return converted
	default:
		return nil
	}
}

package telego

import (
	"context"
	"net/url"
	"strings"

	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/trace"
	apiTrace "go.opentelemetry.io/otel/trace"
)

type SpanExporter struct {
	store *Store
}

func NewSpanExporter(store *Store) *SpanExporter {
	return &SpanExporter{store: store}
}

func (r *SpanExporter) ExportSpans(_ context.Context, spans []trace.ReadOnlySpan) error {
	for _, span := range spans {
		entry := spanEntry(span)
		if entry.Type == "" {
			continue
		}

		r.store.Add(entry)
		if entry.Status == "error" && entry.Type != "exception" {
			exception := entry
			exception.ID = ""
			exception.Type = "exception"
			exception.Name = errorName(span, entry.Name)
			r.store.Add(exception)
		}
	}

	return nil
}

func (r *SpanExporter) Shutdown(context.Context) error {
	return nil
}

func spanEntry(span trace.ReadOnlySpan) Entry {
	attributes := make(map[string]any, len(span.Attributes()))
	for _, attribute := range span.Attributes() {
		key := string(attribute.Key)
		attributes[key] = sanitize(key, attribute.Value.AsInterface())
	}

	status := "ok"
	if span.Status().Code == codes.Error {
		status = "error"
	}

	return Entry{
		Type:         spanType(span, attributes),
		Name:         span.Name(),
		TraceID:      span.SpanContext().TraceID().String(),
		SpanID:       span.SpanContext().SpanID().String(),
		ParentSpanID: span.Parent().SpanID().String(),
		Method:       stringAttribute(attributes, "http.request.method", "http.method"),
		URI:          stringAttribute(attributes, "url.full", "http.url", "http.route", "server.address"),
		Status:       status,
		StatusCode:   intAttribute(attributes, "http.response.status_code", "http.status_code"),
		Duration:     span.EndTime().Sub(span.StartTime()).Microseconds(),
		Timestamp:    span.StartTime().UTC(),
		Payload: map[string]any{
			"attributes":      attributes,
			"events":          spanEvents(span),
			"instrumentation": span.InstrumentationScope().Name,
			"status_message":  span.Status().Description,
		},
	}
}

func spanType(span trace.ReadOnlySpan, attributes map[string]any) string {
	instrumentation := span.InstrumentationScope().Name

	switch {
	case span.SpanKind() == apiTrace.SpanKindServer:
		return "request"
	case strings.Contains(instrumentation, "/database/") || hasAttribute(attributes, "db.system.name", "db.system", "db.query.text", "db.statement"):
		return "query"
	case span.SpanKind() == apiTrace.SpanKindClient && hasAttribute(attributes, "http.request.method", "http.method", "url.full", "http.url"):
		return "http"
	case strings.Contains(instrumentation, "cache"):
		return "cache"
	case strings.Contains(instrumentation, "event"):
		return "event"
	case strings.Contains(instrumentation, "log"):
		return "log"
	default:
		return "trace"
	}
}

func spanEvents(span trace.ReadOnlySpan) []map[string]any {
	events := make([]map[string]any, 0, len(span.Events()))
	for _, event := range span.Events() {
		attributes := make(map[string]any, len(event.Attributes))
		for _, attribute := range event.Attributes {
			key := string(attribute.Key)
			attributes[key] = sanitize(key, attribute.Value.AsInterface())
		}
		events = append(events, map[string]any{
			"name":       event.Name,
			"timestamp":  event.Time.UTC(),
			"attributes": attributes,
		})
	}

	return events
}

func errorName(span trace.ReadOnlySpan, fallback string) string {
	if span.Status().Description != "" {
		return span.Status().Description
	}
	for _, event := range span.Events() {
		for _, attribute := range event.Attributes {
			if attribute.Key == "exception.message" {
				return attribute.Value.AsString()
			}
		}
	}

	return fallback
}

func sanitize(key string, value any) any {
	lowerKey := strings.ToLower(key)
	for _, sensitive := range []string{"authorization", "cookie", "password", "passwd", "secret", "token", "api_key", "apikey"} {
		if strings.Contains(lowerKey, sensitive) {
			return "[REDACTED]"
		}
	}

	if stringValue, ok := value.(string); ok && (strings.Contains(lowerKey, "url") || strings.Contains(lowerKey, "uri")) {
		return sanitizeURL(stringValue)
	}

	return value
}

func sanitizeURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	query := parsed.Query()
	for key := range query {
		lowerKey := strings.ToLower(key)
		if strings.Contains(lowerKey, "key") || strings.Contains(lowerKey, "token") || strings.Contains(lowerKey, "secret") || strings.Contains(lowerKey, "password") {
			query.Set(key, "[REDACTED]")
		}
	}
	parsed.RawQuery = query.Encode()

	return parsed.String()
}

func hasAttribute(attributes map[string]any, keys ...string) bool {
	for _, key := range keys {
		if _, ok := attributes[key]; ok {
			return true
		}
	}

	return false
}

func stringAttribute(attributes map[string]any, keys ...string) string {
	for _, key := range keys {
		if value, ok := attributes[key].(string); ok {
			return value
		}
	}

	return ""
}

func intAttribute(attributes map[string]any, keys ...string) int {
	for _, key := range keys {
		switch value := attributes[key].(type) {
		case int:
			return value
		case int64:
			return int(value)
		}
	}

	return 0
}

package telego

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/goravel/framework/contracts/event"
	"go.opentelemetry.io/otel/trace"
)

type observedEvents struct {
	event.Instance
	store *Store
}

type observedEventTask struct {
	event.Task
	event   event.Event
	args    []event.Arg
	store   *Store
	context context.Context
}

func ObserveEvents(instance event.Instance, store *Store) event.Instance {
	return &observedEvents{Instance: instance, store: store}
}

func (r *observedEvents) Job(dispatchedEvent event.Event, args []event.Arg) event.Task {
	return &observedEventTask{
		Task:    r.Instance.Job(dispatchedEvent, args),
		event:   dispatchedEvent,
		args:    args,
		store:   r.store,
		context: eventContext(args),
	}
}

func (r *observedEventTask) Dispatch() error {
	started := time.Now()
	err := r.Task.Dispatch()
	if !Enabled() {
		return err
	}

	status := "ok"
	if err != nil {
		status = "error"
	}
	entry := r.store.Add(Entry{
		Type:      "event",
		Name:      eventName(r.event),
		TraceID:   trace.SpanContextFromContext(r.context).TraceID().String(),
		Method:    "DISPATCH",
		Status:    status,
		Duration:  time.Since(started).Microseconds(),
		Timestamp: started.UTC(),
		Payload: map[string]any{
			"argument_types": eventArgumentTypes(r.args),
		},
	})
	if err != nil {
		entry.ID = ""
		entry.Type = "exception"
		entry.Name = err.Error()
		entry.Payload["error"] = err.Error()
		r.store.Add(entry)
	}

	return err
}

func eventContext(args []event.Arg) context.Context {
	for _, arg := range args {
		if ctx, ok := arg.Value.(context.Context); ok {
			return ctx
		}
	}

	return context.Background()
}

func eventName(dispatchedEvent event.Event) string {
	typeOf := reflect.TypeOf(dispatchedEvent)
	if typeOf == nil {
		return "event"
	}

	return strings.TrimPrefix(typeOf.String(), "*")
}

func eventArgumentTypes(args []event.Arg) []string {
	types := make([]string, 0, len(args))
	for _, arg := range args {
		if arg.Type != "" {
			types = append(types, arg.Type)
			continue
		}
		types = append(types, fmt.Sprintf("%T", arg.Value))
	}

	return types
}

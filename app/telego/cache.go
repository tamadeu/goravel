package telego

import (
	"context"
	"time"

	"github.com/goravel/framework/contracts/cache"
	"go.opentelemetry.io/otel/trace"
)

type observedCache struct {
	*observedCacheDriver
	cache cache.Cache
}

type observedCacheDriver struct {
	cache.Driver
	context context.Context
	store   *Store
	name    string
}

func ObserveCache(driver cache.Cache, store *Store) cache.Cache {
	return &observedCache{
		observedCacheDriver: observeCacheDriver(driver, store, "default", context.Background()),
		cache:               driver,
	}
}

func (r *observedCache) Store(name string) cache.Driver {
	return observeCacheDriver(r.cache.Store(name), r.store, name, r.context)
}

func observeCacheDriver(driver cache.Driver, store *Store, name string, ctx context.Context) *observedCacheDriver {
	return &observedCacheDriver{Driver: driver, context: ctx, store: store, name: name}
}

func (r *observedCacheDriver) Add(key string, value any, ttl time.Duration) bool {
	started := time.Now()
	added := r.Driver.Add(key, value, ttl)
	r.record("ADD", key, started, nil, map[string]any{"added": added, "ttl": ttl.String()})
	return added
}

func (r *observedCacheDriver) Decrement(key string, value ...int64) (int64, error) {
	started := time.Now()
	result, err := r.Driver.Decrement(key, value...)
	r.record("DECREMENT", key, started, err, nil)
	return result, err
}

func (r *observedCacheDriver) Forever(key string, value any) bool {
	started := time.Now()
	stored := r.Driver.Forever(key, value)
	r.record("FOREVER", key, started, nil, map[string]any{"stored": stored})
	return stored
}

func (r *observedCacheDriver) Forget(key string) bool {
	started := time.Now()
	forgotten := r.Driver.Forget(key)
	r.record("FORGET", key, started, nil, map[string]any{"forgotten": forgotten})
	return forgotten
}

func (r *observedCacheDriver) Flush() bool {
	started := time.Now()
	flushed := r.Driver.Flush()
	r.record("FLUSH", "*", started, nil, map[string]any{"flushed": flushed})
	return flushed
}

func (r *observedCacheDriver) Get(key string, defaults ...any) any {
	started := time.Now()
	hit := r.Driver.Has(key)
	value := r.Driver.Get(key, defaults...)
	r.record("GET", key, started, nil, map[string]any{"hit": hit})
	return value
}

func (r *observedCacheDriver) GetBool(key string, defaults ...bool) bool {
	started := time.Now()
	hit := r.Driver.Has(key)
	value := r.Driver.GetBool(key, defaults...)
	r.record("GET", key, started, nil, map[string]any{"hit": hit, "value_type": "bool"})
	return value
}

func (r *observedCacheDriver) GetInt(key string, defaults ...int) int {
	started := time.Now()
	hit := r.Driver.Has(key)
	value := r.Driver.GetInt(key, defaults...)
	r.record("GET", key, started, nil, map[string]any{"hit": hit, "value_type": "int"})
	return value
}

func (r *observedCacheDriver) GetInt64(key string, defaults ...int64) int64 {
	started := time.Now()
	hit := r.Driver.Has(key)
	value := r.Driver.GetInt64(key, defaults...)
	r.record("GET", key, started, nil, map[string]any{"hit": hit, "value_type": "int64"})
	return value
}

func (r *observedCacheDriver) GetString(key string, defaults ...string) string {
	started := time.Now()
	hit := r.Driver.Has(key)
	value := r.Driver.GetString(key, defaults...)
	r.record("GET", key, started, nil, map[string]any{"hit": hit, "value_type": "string"})
	return value
}

func (r *observedCacheDriver) Has(key string) bool {
	started := time.Now()
	hit := r.Driver.Has(key)
	r.record("HAS", key, started, nil, map[string]any{"hit": hit})
	return hit
}

func (r *observedCacheDriver) Increment(key string, value ...int64) (int64, error) {
	started := time.Now()
	result, err := r.Driver.Increment(key, value...)
	r.record("INCREMENT", key, started, err, nil)
	return result, err
}

func (r *observedCacheDriver) Put(key string, value any, ttl time.Duration) error {
	started := time.Now()
	err := r.Driver.Put(key, value, ttl)
	r.record("PUT", key, started, err, map[string]any{"ttl": ttl.String()})
	return err
}

func (r *observedCacheDriver) Pull(key string, defaults ...any) any {
	started := time.Now()
	hit := r.Driver.Has(key)
	value := r.Driver.Pull(key, defaults...)
	r.record("PULL", key, started, nil, map[string]any{"hit": hit})
	return value
}

func (r *observedCacheDriver) Remember(key string, ttl time.Duration, callback func() (any, error)) (any, error) {
	started := time.Now()
	hit := r.Driver.Has(key)
	value, err := r.Driver.Remember(key, ttl, callback)
	r.record("REMEMBER", key, started, err, map[string]any{"hit": hit, "ttl": ttl.String()})
	return value, err
}

func (r *observedCacheDriver) RememberForever(key string, callback func() (any, error)) (any, error) {
	started := time.Now()
	hit := r.Driver.Has(key)
	value, err := r.Driver.RememberForever(key, callback)
	r.record("REMEMBER_FOREVER", key, started, err, map[string]any{"hit": hit})
	return value, err
}

func (r *observedCacheDriver) WithContext(ctx context.Context) cache.Driver {
	return observeCacheDriver(r.Driver.WithContext(ctx), r.store, r.name, ctx)
}

func (r *observedCacheDriver) record(operation, key string, started time.Time, err error, payload map[string]any) {
	if !Enabled() {
		return
	}
	if payload == nil {
		payload = make(map[string]any)
	}
	payload["store"] = r.name
	if err != nil {
		payload["error"] = err.Error()
	}

	status := "ok"
	if err != nil {
		status = "error"
	}
	entry := r.store.Add(Entry{
		Type:      "cache",
		Name:      "cache." + operation,
		TraceID:   trace.SpanContextFromContext(r.context).TraceID().String(),
		Method:    operation,
		URI:       key,
		Status:    status,
		Duration:  time.Since(started).Microseconds(),
		Timestamp: started.UTC(),
		Payload:   payload,
	})
	if err != nil {
		entry.ID = ""
		entry.Type = "exception"
		entry.Name = err.Error()
		r.store.Add(entry)
	}
}

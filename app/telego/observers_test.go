package telego

import (
	"testing"

	"github.com/goravel/framework/contracts/event"
	"github.com/goravel/framework/testing/mock"
	"github.com/stretchr/testify/require"
)

type observerTestEvent struct{}

func (observerTestEvent) Handle(args []event.Arg) ([]event.Arg, error) {
	return args, nil
}

func TestObservedCacheRecordsHitWithoutValue(t *testing.T) {
	store := NewStore(10)
	driver := mock.Factory().CacheDriver()
	driver.EXPECT().Has("profile:1").Return(true).Once()
	driver.EXPECT().Get("profile:1").Return("sensitive value").Once()

	value := observeCacheDriver(driver, store, "memory", t.Context()).Get("profile:1")

	require.Equal(t, "sensitive value", value)
	entries := store.List(Filters{Type: "cache", Limit: 10})
	require.Len(t, entries, 1)
	require.Equal(t, "profile:1", entries[0].URI)
	require.Equal(t, true, entries[0].Payload["hit"])
	require.NotContains(t, entries[0].Payload, "value")
}

func TestObservedEventRecordsDispatch(t *testing.T) {
	store := NewStore(10)
	factory := mock.Factory()
	instance := factory.Event()
	task := factory.EventTask()
	dispatchedEvent := observerTestEvent{}
	args := []event.Arg{{Type: "user_id", Value: 1}}
	instance.EXPECT().Job(dispatchedEvent, args).Return(task).Once()
	task.EXPECT().Dispatch().Return(nil).Once()

	err := ObserveEvents(instance, store).Job(dispatchedEvent, args).Dispatch()

	require.NoError(t, err)
	entries := store.List(Filters{Type: "event", Limit: 10})
	require.Len(t, entries, 1)
	require.Equal(t, "telego.observerTestEvent", entries[0].Name)
	require.Equal(t, []string{"user_id"}, entries[0].Payload["argument_types"])
}

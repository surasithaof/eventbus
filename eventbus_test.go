package eventbus_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/surasithaof/eventbus"
)

type testKey int8

const testCtxKey testKey = iota
const (
	TestEventType      eventbus.EventType = "test.event"
	TestPanicEventType eventbus.EventType = "test.panic.event"
)

func TestEventBus(t *testing.T) {
	bus := eventbus.NewEventBus(eventbus.WithMaxWorkers[string](0))

	ctx := context.Background()
	waitTimeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	count := 0
	bus.Subscribe(TestEventType, func(ctx context.Context, data string) {
		if ctx.Value(testCtxKey) == nil {
			t.Errorf("expected context value to be set, got nil")
		}
		if data == "" {
			t.Errorf("expected data to be set, got empty string")
		}
		t.Logf("Received data: %s", data)
		t.Logf("Context value: %v", ctx.Value(testCtxKey))
		count++
	})

	t.Run("success", func(t *testing.T) {
		count = 0 // reset count for this test
		ctx = context.WithValue(ctx, testCtxKey, "value 1")
		err := bus.Publish(ctx, TestEventType, "test message 1")
		require.NoError(t, err, "expected no error when publishing")
		bus.Wait()
		require.Equal(t, 1, count, "expected count to be 1")
	})

	t.Run("context cancelled", func(t *testing.T) {
		count = 0 // reset count for this test
		ctx := context.WithValue(ctx, testCtxKey, "value 2")
		ctx, cancel := context.WithTimeout(ctx, waitTimeout)
		cancel()
		err := bus.Publish(ctx, TestEventType, "test message 2")
		require.Error(t, err, "expected error when publishing to cancelled context")
		bus.Wait()
		require.Equal(t, count, 0, "expected count to be 1")
	})

	t.Run("recovery from panic", func(t *testing.T) {
		err := bus.Subscribe(TestPanicEventType, func(ctx context.Context, data string) {
			panic("test panic")
		})
		require.NoError(t, err, "expected no error when subscribing to panic event")
		require.NotPanics(t, func() {
			err = bus.Publish(ctx, TestPanicEventType, "test message")
			require.NoError(t, err, "expected no error when publishing to panic event")
			bus.Wait()
		})
	})

	t.Run("eventbus stopped", func(t *testing.T) {
		count = 0 // reset count for this test
		bus.StopAndWait()
		ctx = context.WithValue(ctx, testCtxKey, "value 3")
		err := bus.Publish(ctx, TestEventType, "test message 3")
		require.Error(t, err, "expected error when publishing to stopped eventbus")
		require.Equal(t, count, 0, "expected count to be 0")

		err = bus.Subscribe(TestEventType, func(ctx context.Context, data string) {
			t.Log("this should not be called")
		})
		require.Error(t, err, "expected error when subscribing to stopped eventbus")
		bus.Stop() // stop again to ensure no panic
		require.Equal(t, count, 0, "expected count to be 0")
	})

	select {
	case <-ctx.Done():
		if ctx.Err() != nil {
			t.Errorf("expected no error, got '%v'", ctx.Err())
		}
	case <-time.After(waitTimeout):
		// Do nothing, test passed
		t.Logf("Test passed, received all messages")
	}
}

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

func TestEventBus(t *testing.T) {
	eventbus := eventbus.NewEventBus(eventbus.WithMaxWorkers[string](0))

	ctx := context.Background()
	waitTimeout := 3 * time.Second
	ctx, cancel := context.WithTimeout(ctx, waitTimeout)
	defer cancel()

	count := 0
	eventbus.Subscribe("test", func(ctx context.Context, data string) {
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
		err := eventbus.Publish(ctx, "test", "test message 1")
		require.NoError(t, err, "expected no error when publishing")
		eventbus.Wait()
		require.Equal(t, 1, count, "expected count to be 1")
	})

	t.Run("context cancelled", func(t *testing.T) {
		count = 0 // reset count for this test
		ctx := context.WithValue(ctx, testCtxKey, "value 2")
		ctx, cancel := context.WithTimeout(ctx, waitTimeout)
		cancel()
		err := eventbus.Publish(ctx, "test", "test message 2")
		require.Error(t, err, "expected error when publishing to cancelled context")
		eventbus.Wait()
		require.Equal(t, count, 0, "expected count to be 1")
	})

	t.Run("eventbus stopped", func(t *testing.T) {
		count = 0 // reset count for this test
		eventbus.StopAndWait()
		ctx = context.WithValue(ctx, testCtxKey, "value 3")
		err := eventbus.Publish(ctx, "test", "test message 3")
		require.Error(t, err, "expected error when publishing to stopped eventbus")
		require.Equal(t, count, 0, "expected count to be 0")

		err = eventbus.Subscribe("test", func(ctx context.Context, data string) {
			t.Log("this should not be called")
		})
		require.Error(t, err, "expected error when subscribing to stopped eventbus")
		eventbus.Stop() // stop again to ensure no panic
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

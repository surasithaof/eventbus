package eventbus_test

import (
	"context"
	"testing"

	"github.com/surasithaof/eventbus"
)

func BenchmarkEventBus(b *testing.B) {
	eventbus := eventbus.NewEventBus[string]()
	defer eventbus.StopAndWait()
	ctx := context.Background()
	ctx = context.WithValue(ctx, testCtxKey, "value 1 for benchmark")

	eventbus.Subscribe("test", func(ctx context.Context, data string) {
		b.Logf("Received data: %s", data)
		b.Logf("Context value: %v", ctx.Value(testCtxKey))
	})
	for i := 0; i < b.N; i++ {
		eventbus.Publish(ctx, "test", "test message")
	}
}

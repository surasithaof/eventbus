package eventbus_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/surasithaof/eventbus"
)

func BenchmarkEventBus(b *testing.B) {
	eventbus := eventbus.NewEventBus[string]()
	defer eventbus.StopAndWait()
	ctx := context.Background()

	eventbus.Subscribe(TestEventType, func(ctx context.Context, data string) {
		b.Logf("Received data: %s", data)
		b.Logf("Context value: %v", ctx.Value(testCtxKey))
	})
	for i := 0; i < b.N; i++ {
		ctx = context.WithValue(ctx, testCtxKey, i)
		eventbus.Publish(ctx, TestEventType, fmt.Sprintf("test message %d", i))
	}
}

package eventbus_test

import (
	"context"
	"testing"

	"github.com/surasithaof/eventbus"
)

func BenchmarkSubscribe(b *testing.B) {
	b.ReportAllocs()

	eventbus := eventbus.NewEventBus[string]()
	defer eventbus.StopAndWait()

	for i := 0; i < b.N; i++ {
		eventbus.Subscribe(TestEventType, func(ctx context.Context, data string) {})
	}
}

func BenchmarkPublish(b *testing.B) {
	b.ReportAllocs()

	eventbus := eventbus.NewEventBus[string]()
	defer eventbus.StopAndWait()
	ctx := context.Background()

	eventbus.Subscribe(TestEventType, func(ctx context.Context, data string) {})

	for i := 0; i < b.N; i++ {
		eventbus.Publish(ctx, TestEventType, "test message")
	}
}

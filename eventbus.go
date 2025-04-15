package eventbus

import (
	"context"
	"errors"
	"runtime"
	"sync"
)

var (
	ErrorEventStopped = errors.New("event stopped")
)

type EventType string

type Option[T any] func(*Event[T])

type Event[T any] struct {
	limiter  chan bool
	stopped  bool
	mu       sync.RWMutex
	wg       sync.WaitGroup
	handlers map[EventType][]HandlerFunc[T]
}

type HandlerFunc[T any] func(ctx context.Context, payload T)

type EventBus[T any] interface {
	// publishes an event to all subscribers.
	// it will call the handler with the payload.
	Publish(ctx context.Context, event EventType, payload T) error
	// registers a handler for an event type.
	// the handler will be called with the payload when the event is published.
	Subscribe(event EventType, handler HandlerFunc[T]) error

	// waits for all handlers to finish.
	// this will block until all handlers are finished.
	Wait()
	// stops the event bus.
	// this will stop all handlers and close the event bus.
	Stop()
	// stops the event bus and waits for all handlers to finish.
	// this will block until all handlers are finished.
	// this is a convenience method for stopping the event bus and waiting for all handlers to finish.
	// it will call Stop() and Wait() in order.
	// this is useful for cleaning up resources and stopping the event bus gracefully.
	StopAndWait()
}

func NewEventBus[T any](opts ...Option[T]) EventBus[T] {
	// Set maxWorkers to a default value as 4 times the number of CPU cores.
	maxWorkers := 4 * runtime.GOMAXPROCS(0)
	e := &Event[T]{
		limiter:  make(chan bool, maxWorkers),
		mu:       sync.RWMutex{},
		wg:       sync.WaitGroup{},
		handlers: make(map[EventType][]HandlerFunc[T]),
	}

	for _, opt := range opts {
		opt(e)
	}

	return e
}

func WithMaxWorkers[T any](maxWorkers int) Option[T] {
	return func(e *Event[T]) {
		if maxWorkers <= 0 {
			maxWorkers = 4 * runtime.GOMAXPROCS(0)
		}
		e.limiter = make(chan bool, maxWorkers)
	}
}

func (s *Event[T]) Publish(ctx context.Context, event EventType, payload T) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if ctx.Err() != nil {
		return ctx.Err()
	}

	if s.stopped {
		return ErrorEventStopped
	}

	for _, handler := range s.handlers[event] {
		s.wg.Add(1)
		s.limiter <- true
		go func(ctx context.Context, handler HandlerFunc[T]) {
			defer func() {
				<-s.limiter
				s.wg.Done()
			}()

			handler(ctx, payload)
		}(ctx, handler)
	}
	return nil
}

func (s *Event[T]) Subscribe(event EventType, handler HandlerFunc[T]) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return ErrorEventStopped
	}

	if s.handlers[event] == nil {
		s.handlers[event] = make([]HandlerFunc[T], 0)
	}
	handlers := s.handlers[event]
	handlers = append(handlers, handler)
	s.handlers[event] = handlers
	return nil
}

func (s *Event[T]) Wait() {
	s.wg.Wait()
}

func (s *Event[T]) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.stopped {
		return
	}
	s.stopped = true

	s.handlers = make(map[EventType][]HandlerFunc[T])
	close(s.limiter)
}

func (s *Event[T]) StopAndWait() {
	s.Stop()
	s.Wait()
}

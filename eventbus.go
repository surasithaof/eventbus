package eventbus

import (
	"context"
	"errors"
	"log"
	"runtime"
	"sync"
)

var (
	ErrorEventStopped = errors.New("event stopped")
)

type (
	Option[T any] func(*Event[T])

	// EventType is the type of the event of topic name
	EventType string

	// HandlerFunc is a handler function that will be called when an event is published
	HandlerFunc[T any] func(ctx context.Context, payload T)

	// Event is the main struct for the event bus
	Event[T any] struct {
		limiter  chan bool
		stopped  bool
		mu       sync.RWMutex
		wg       sync.WaitGroup
		handlers map[EventType][]HandlerFunc[T]
	}

	// EventBus is the interface for the event bus
	EventBus[T any] interface {
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
)

// NewEventBus initializes the handlers map and the limiter channel.
// the default maxWorkers is 4 times the number of CPU cores.
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

// WithMaxWorkers sets the maximum number of workers for initializing the event bus.
// this will limit the number of concurrent handlers that can be executed.
// if maxWorkers is less than or equal to 0, it will be set to 4 times the number of CPU cores.
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
				// recover from panic in the handler
				if r := recover(); r != nil {
					// Log or handle the panic as needed
					log.Printf("Recovered from panic: %v", r)
				}

				// release the limiter when the handler is done
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
	s.handlers[event] = append(s.handlers[event], handler)
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

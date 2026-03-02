package events

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Event represents a domain event
type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Payload   json.RawMessage `json:"payload"`
	Timestamp time.Time       `json:"timestamp"`
}

// Handler processes an event
type Handler func(ctx context.Context, event Event) error

// Bus is the event bus interface - swap implementation for Kafka later
type Bus interface {
	Publish(ctx context.Context, eventType string, payload interface{}) error
	Subscribe(eventType string, handler Handler)
	Start(ctx context.Context) error
	Stop()
}

// ─── In-Memory Channel Bus (Demo) ───────────────
type ChannelBus struct {
	handlers map[string][]Handler
	ch       chan Event
	mu       sync.RWMutex
	done     chan struct{}
}

func NewChannelBus(bufferSize int) *ChannelBus {
	return &ChannelBus{
		handlers: make(map[string][]Handler),
		ch:       make(chan Event, bufferSize),
		done:     make(chan struct{}),
	}
}

func (b *ChannelBus) Publish(ctx context.Context, eventType string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	event := Event{
		ID:        uuid.New().String(),
		Type:      eventType,
		Payload:   data,
		Timestamp: time.Now(),
	}

	select {
	case b.ch <- event:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (b *ChannelBus) Subscribe(eventType string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *ChannelBus) Start(ctx context.Context) error {
	for {
		select {
		case event := <-b.ch:
			b.mu.RLock()
			handlers := b.handlers[event.Type]
			// Also notify wildcard subscribers
			allHandlers := b.handlers["*"]
			b.mu.RUnlock()

			for _, h := range append(handlers, allHandlers...) {
				go h(ctx, event)
			}
		case <-ctx.Done():
			return nil
		case <-b.done:
			return nil
		}
	}
}

func (b *ChannelBus) Stop() {
	close(b.done)
}

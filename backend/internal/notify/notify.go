package notify

import (
	"context"
	"sync"
)

type EventType string

const (
	EventNotificationCreated EventType = "notification.created"
	EventChatMessageCreated  EventType = "chat.message.created"
	EventOrderStatusChanged  EventType = "order.status.changed"
)

type Event struct {
	Type   EventType
	UserID string
	Data   map[string]any
}

type Subscription struct {
	Events <-chan Event
	cancel func()
}

func (s Subscription) Cancel() {
	if s.cancel != nil {
		s.cancel()
	}
}

type Publisher interface {
	Publish(ctx context.Context, evt Event)
	Subscribe(userID string, buffer int) Subscription
}

// InProcessPublisher is a simple, best-effort pub/sub for the API-first phase.
// It does not persist events; persistence happens via database writes in services.
type InProcessPublisher struct {
	mu    sync.RWMutex
	subs  map[string]map[chan Event]struct{} // user_id -> set(ch)
}

func NewInProcessPublisher() *InProcessPublisher {
	return &InProcessPublisher{
		subs: map[string]map[chan Event]struct{}{},
	}
}

func (p *InProcessPublisher) Subscribe(userID string, buffer int) Subscription {
	if buffer < 1 {
		buffer = 32
	}
	ch := make(chan Event, buffer)
	p.mu.Lock()
	if _, ok := p.subs[userID]; !ok {
		p.subs[userID] = map[chan Event]struct{}{}
	}
	p.subs[userID][ch] = struct{}{}
	p.mu.Unlock()

	cancel := func() {
		p.mu.Lock()
		if m, ok := p.subs[userID]; ok {
			delete(m, ch)
			if len(m) == 0 {
				delete(p.subs, userID)
			}
		}
		p.mu.Unlock()
		close(ch)
	}
	return Subscription{Events: ch, cancel: cancel}
}

func (p *InProcessPublisher) Publish(_ context.Context, evt Event) {
	p.mu.RLock()
	m := p.subs[evt.UserID]
	p.mu.RUnlock()
	for ch := range m {
		select {
		case ch <- evt:
		default:
			// Best-effort: drop if subscriber is slow.
		}
	}
}


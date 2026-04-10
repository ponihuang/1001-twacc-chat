package chat

import "sync"

const subscriberBufferSize = 16

// Hub keeps per-user realtime subscriptions for WebSocket delivery.
type Hub struct {
	mu          sync.RWMutex
	subscribers map[int64]map[chan RealtimeEvent]struct{}
}

// NewHub creates an empty realtime hub.
func NewHub() *Hub {
	return &Hub{
		subscribers: make(map[int64]map[chan RealtimeEvent]struct{}),
	}
}

// Subscribe registers a user-level realtime subscription.
func (h *Hub) Subscribe(userID int64) (<-chan RealtimeEvent, func()) {
	ch := make(chan RealtimeEvent, subscriberBufferSize)

	h.mu.Lock()
	if _, ok := h.subscribers[userID]; !ok {
		h.subscribers[userID] = make(map[chan RealtimeEvent]struct{})
	}
	h.subscribers[userID][ch] = struct{}{}
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		if listeners, ok := h.subscribers[userID]; ok {
			if _, exists := listeners[ch]; exists {
				delete(listeners, ch)
				close(ch)
			}
			if len(listeners) == 0 {
				delete(h.subscribers, userID)
			}
		}
		h.mu.Unlock()
	}

	return ch, cancel
}

// PublishToUsers pushes an event to the target users' current subscribers.
func (h *Hub) PublishToUsers(userIDs []int64, event RealtimeEvent) {
	seen := make(map[int64]struct{}, len(userIDs))

	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, userID := range userIDs {
		if userID <= 0 {
			continue
		}
		if _, ok := seen[userID]; ok {
			continue
		}
		seen[userID] = struct{}{}

		for ch := range h.subscribers[userID] {
			select {
			case ch <- event:
			default:
			}
		}
	}
}

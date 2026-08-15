package realtime

import "sync"

// Hub is a minimal in-process pub/sub broker keyed by an arbitrary topic
// string. One Hub instance == one class of events (e.g. all chat messages,
// or all notifications). Safe for concurrent Subscribe/Publish/unsubscribe.
type Hub[T any] struct {
	mu   sync.RWMutex
	subs map[string]map[chan T]struct{}
}

func NewHub[T any]() *Hub[T] {
	return &Hub[T]{subs: make(map[string]map[chan T]struct{})}
}

// Subscribe registers a new subscriber for topic. The caller MUST call the
// returned unsubscribe func exactly once (e.g. via defer) when it's done
// reading, typically when the client's request context is cancelled.
func (h *Hub[T]) Subscribe(topic string) (<-chan T, func()) {
	ch := make(chan T, 8) // small buffer: publisher never blocks on a slow reader
	h.mu.Lock()
	if h.subs[topic] == nil {
		h.subs[topic] = make(map[chan T]struct{})
	}
	h.subs[topic][ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	unsubscribe := func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			if set, ok := h.subs[topic]; ok {
				delete(set, ch)
				close(ch)
				if len(set) == 0 {
					delete(h.subs, topic)
				}
			}
		})
	}
	return ch, unsubscribe
}

// Publish fans payload out to every current subscriber of topic. Non-blocking:
// a full/stuck subscriber channel drops the event rather than blocking the
// publisher (acceptable — the DB row is already the source of truth; the
// client's own reconnect + REST refetch covers it).
func (h *Hub[T]) Publish(topic string, payload T) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[topic] {
		select {
		case ch <- payload:
		default:
		}
	}
}

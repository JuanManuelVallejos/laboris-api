package realtime

import "sync"

// ConnLimiter caps how many concurrent long-lived connections (SSE streams)
// a single key (typically a user ID) may hold open at once — basic abuse
// protection against a single account exhausting server resources by
// opening an unbounded number of streaming connections.
type ConnLimiter struct {
	mu    sync.Mutex
	count map[string]int
	max   int
}

func NewConnLimiter(max int) *ConnLimiter {
	return &ConnLimiter{count: make(map[string]int), max: max}
}

// Acquire tries to reserve a connection slot for key. On success it returns
// a release func the caller must call (typically via defer) when the
// connection closes, and true. On failure (key already at the limit) it
// returns false and a nil func.
func (l *ConnLimiter) Acquire(key string) (func(), bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.count[key] >= l.max {
		return nil, false
	}
	l.count[key]++

	var once sync.Once
	release := func() {
		once.Do(func() {
			l.mu.Lock()
			defer l.mu.Unlock()
			l.count[key]--
			if l.count[key] <= 0 {
				delete(l.count, key)
			}
		})
	}
	return release, true
}

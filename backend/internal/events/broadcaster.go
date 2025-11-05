package events

import (
	"sync"
)

type Broadcaster struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		clients: make(map[chan []byte]struct{}),
	}
}

func (b *Broadcaster) Subscribe(buffer int) (chan []byte, func()) {
	ch := make(chan []byte, buffer)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()

	cancel := func() {
		b.mu.Lock()
		if _, ok := b.clients[ch]; ok {
			delete(b.clients, ch)
			close(ch)
		}
		b.mu.Unlock()
	}
	return ch, cancel
}

func (b *Broadcaster) Publish(payload []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- payload:
		default:
			// Drop when subscriber is slow to avoid blocking others.
		}
	}
}

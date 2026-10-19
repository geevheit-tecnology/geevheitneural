package events

import (
	"context"
	"sync"
)

type InMemoryPublisher struct {
	mu          sync.RWMutex
	subscribers map[string][]chan any
}

func NewInMemoryPublisher() *InMemoryPublisher {
	return &InMemoryPublisher{
		subscribers: make(map[string][]chan any),
	}
}

func (p *InMemoryPublisher) Publish(_ context.Context, topic string, payload any) error {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, ch := range p.subscribers[topic] {
		select {
		case ch <- payload:
		default:
			// Não bloqueia caso o consumidor esteja ocupado
		}
	}
	return nil
}

func (p *InMemoryPublisher) Subscribe(topic string) <-chan any {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan any, 100)
	p.subscribers[topic] = append(p.subscribers[topic], ch)
	return ch
}

func (p *InMemoryPublisher) Unsubscribe(topic string, ch <-chan any) {
	p.mu.Lock()
	defer p.mu.Unlock()

	subs := p.subscribers[topic]
	for i, sub := range subs {
		if sub == ch {
			p.subscribers[topic] = append(subs[:i], subs[i+1:]...)
			close(sub)
			break
		}
	}
}

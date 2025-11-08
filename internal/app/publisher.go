package app

import (
	"context"
	"sync"
)

// ContentPublisher emits events whenever a content record is created.
type ContentPublisher interface {
	Publish(ctx context.Context, item api) error
	Close() error
}

type noopPublisher struct{}

func (n *noopPublisher) Publish(_ context.Context, _ api) error {
	return nil
}

func (n *noopPublisher) Close() error {
	return nil
}

var (
	publisherMu sync.RWMutex
	publisher   ContentPublisher = &noopPublisher{}
)

func getContentPublisher() ContentPublisher {
	publisherMu.RLock()
	defer publisherMu.RUnlock()
	return publisher
}

func setContentPublisher(p ContentPublisher) {
	publisherMu.Lock()
	publisher = p
	publisherMu.Unlock()
}

func resetContentPublisher() {
	setContentPublisher(&noopPublisher{})
}

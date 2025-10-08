package app

import (
	"context"
	"sort"
	"sync"
)

type inMemoryRepository struct {
	mu    sync.RWMutex
	items map[string]api
}

func newInMemoryRepository(seed allContent) *inMemoryRepository {
	repo := &inMemoryRepository{
		items: make(map[string]api, len(seed)),
	}
	for _, item := range seed {
		repo.items[item.ID] = item
	}
	return repo
}

func (r *inMemoryRepository) ListContent(_ context.Context) (allContent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(allContent, 0, len(r.items))
	for _, item := range r.items {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func (r *inMemoryRepository) GetContent(_ context.Context, id string) (*api, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	item, ok := r.items[id]
	if !ok {
		return nil, ErrContentNotFound
	}
	copy := item
	return &copy, nil
}

func (r *inMemoryRepository) CreateContent(_ context.Context, item api) (*api, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[item.ID]; exists {
		return nil, ErrContentAlreadyExists
	}
	r.items[item.ID] = item
	copy := item
	return &copy, nil
}

func (r *inMemoryRepository) UpdateContent(_ context.Context, id string, name string) (*api, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	item, exists := r.items[id]
	if !exists {
		return nil, ErrContentNotFound
	}
	item.Name = name
	r.items[id] = item
	copy := item
	return &copy, nil
}

func (r *inMemoryRepository) DeleteContent(_ context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.items[id]; !exists {
		return ErrContentNotFound
	}
	delete(r.items, id)
	return nil
}

package app

import (
	"context"
	"errors"
	"sync"
)

// ErrContentNotFound signals that the requested content could not be located in the backing store.
var ErrContentNotFound = errors.New("content not found")

// ErrContentAlreadyExists signals that the requested content id already exists in the backing store.
var ErrContentAlreadyExists = errors.New("content already exists")

// ContentRepository defines the required behaviour for interacting with the content data store.
type ContentRepository interface {
	ListContent(ctx context.Context) (allContent, error)
	GetContent(ctx context.Context, id string) (*api, error)
	CreateContent(ctx context.Context, item api) (*api, error)
	UpdateContent(ctx context.Context, id string, name string) (*api, error)
	DeleteContent(ctx context.Context, id string) error
}

var (
	contentRepoMu sync.RWMutex
	contentRepo   ContentRepository = newInMemoryRepository(nil)
)

func getContentRepository() ContentRepository {
	contentRepoMu.RLock()
	defer contentRepoMu.RUnlock()
	return contentRepo
}

func setContentRepository(repo ContentRepository) {
	contentRepoMu.Lock()
	contentRepo = repo
	contentRepoMu.Unlock()
}

package repository

import (
	"slices"
	"sync"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
)

type Repository struct {
	mu      sync.RWMutex
	chats   map[int64]bool
	links   map[int64][]*domain.Link
	counter int64
}

func New() *Repository {
	return &Repository{
		chats:   make(map[int64]bool),
		links:   make(map[int64][]*domain.Link),
		counter: 1,
	}
}

func (r *Repository) RegisterChat(chatID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.chats[chatID] = true

	return nil
}

func (r *Repository) DeleteChat(chatID int64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.chats[chatID] {
		return NewErrUnregister(chatID)
	}

	delete(r.links, chatID)
	delete(r.chats, chatID)

	return nil
}

func (r *Repository) TrackLink(link *domain.Link) (*domain.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.chats[link.ChatID] {
		return nil, NewErrUnregister(link.ChatID)
	}

	link.ID = r.counter
	r.links[link.ChatID] = append(r.links[link.ChatID], link)
	r.counter++

	return link, nil
}

func (r *Repository) UntrackLink(chatID int64, url string) (*domain.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.chats[chatID] {
		return nil, NewErrUnregister(chatID)
	}

	links := r.links[chatID]
	idx := slices.IndexFunc(links, func(l *domain.Link) bool {
		return l.URL == url
	})

	if idx == -1 {
		return nil, NewErLinkNotFound(url)
	}

	link := links[idx]
	r.links[chatID] = slices.Delete(links, idx, idx+1)

	return link, nil
}

func (r *Repository) ListLinks(chatID int64) ([]*domain.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.chats[chatID] {
		return nil, NewErrUnregister(chatID)
	}

	return r.links[chatID], nil
}

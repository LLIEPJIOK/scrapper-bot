package bot

import (
	"maps"
	"slices"
	"sync"
)

type Repository struct {
	mu    sync.Mutex
	chats map[int64]*UpdateChat
}

func New() *Repository {
	return &Repository{
		mu:    sync.Mutex{},
		chats: make(map[int64]*UpdateChat),
	}
}

func (r *Repository) AddLink(chatID int64, link string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.chats[chatID]; !ok {
		r.chats[chatID] = &UpdateChat{
			ID: chatID,
		}
	}

	links := r.chats[chatID].Links
	if slices.Contains(links, link) {
		return nil
	}

	links = append(links, link)
	r.chats[chatID].Links = links

	return nil
}

func (r *Repository) GetUpdates() ([]*UpdateChat, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	defer clear(r.chats)

	return slices.Collect(maps.Values(r.chats)), nil
}

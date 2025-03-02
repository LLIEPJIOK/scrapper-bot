package repository

import (
	"slices"
	"sync"
	"time"

	"github.com/es-debug/backend-academy-2024-go-template/internal/domain"
)

type Repository struct {
	mu        sync.RWMutex
	chats     map[int64]bool
	links     map[string]*Link
	chatLinks map[int64][]string
	counter   int64
}

func New() *Repository {
	return &Repository{
		chats:     make(map[int64]bool),
		links:     make(map[string]*Link),
		chatLinks: make(map[int64][]string),
		counter:   1,
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

	delete(r.chats, chatID)
	delete(r.chatLinks, chatID)

	for url, link := range r.links {
		idx := slices.Index(link.Chats, chatID)
		if idx != -1 {
			link.Chats = slices.Delete(link.Chats, idx, idx+1)
		}

		r.links[url] = link
	}

	return nil
}

func (r *Repository) TrackLink(link *domain.Link) (*domain.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.chats[link.ChatID] {
		return nil, NewErrUnregister(link.ChatID)
	}

	if _, ok := r.links[link.URL]; !ok {
		link.ID = r.counter
		repoLink := domainToRepoLink(link)
		repoLink.CheckedAt = time.Now()

		r.links[link.URL] = repoLink
		r.counter++
	}

	if slices.Contains(r.links[link.URL].Chats, link.ChatID) {
		return nil, NewErrLinkExists(link.URL)
	}

	r.chatLinks[link.ChatID] = append(r.chatLinks[link.ChatID], link.URL)

	repoLink := r.links[link.URL]
	repoLink.Chats = append(repoLink.Chats, link.ChatID)
	r.links[link.URL] = repoLink

	link.ID = repoLink.ID

	return link, nil
}

func (r *Repository) UntrackLink(chatID int64, url string) (*domain.Link, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.chats[chatID] {
		return nil, NewErrUnregister(chatID)
	}

	links := r.chatLinks[chatID]
	idx := slices.Index(links, url)

	if idx == -1 {
		return nil, NewErrLinkNotFound(url)
	}

	r.chatLinks[chatID] = slices.Delete(links, idx, idx+1)

	link := r.links[url]

	idx = slices.Index(link.Chats, chatID)
	if idx == -1 {
		return nil, NewErrLinkNotFound(url)
	}

	link.Chats = slices.Delete(link.Chats, idx, idx+1)
	r.links[url] = link

	return repoToDomainLink(link), nil
}

func (r *Repository) ListLinks(chatID int64) ([]*domain.Link, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.chats[chatID] {
		return nil, NewErrUnregister(chatID)
	}

	urls := r.chatLinks[chatID]
	links := make([]*domain.Link, len(urls))

	for i, url := range urls {
		links[i] = repoToDomainLink(r.links[url])
		links[i].ChatID = chatID
	}

	return links, nil
}

func (r *Repository) GetCheckLinks() []*domain.CheckLink {
	r.mu.RLock()
	defer r.mu.RUnlock()

	links := make([]*domain.CheckLink, 0, len(r.links))

	for _, link := range r.links {
		links = append(links, repoLinkToCheckLink(link))
	}

	return links
}

func (r *Repository) UpdateCheckTime(url string, checkedAt time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	link, ok := r.links[url]
	if !ok {
		return NewErrLinkNotFound(url)
	}

	link.CheckedAt = checkedAt
	r.links[url] = link

	return nil
}

func domainToRepoLink(link *domain.Link) *Link {
	return &Link{
		ID:      link.ID,
		URL:     link.URL,
		Filters: link.Filters,
		Tags:    link.Tags,
	}
}

func repoToDomainLink(link *Link) *domain.Link {
	return &domain.Link{
		ID:      link.ID,
		URL:     link.URL,
		Filters: link.Filters,
		Tags:    link.Tags,
	}
}

func repoLinkToCheckLink(link *Link) *domain.CheckLink {
	checkLink := &domain.CheckLink{
		ID:        link.ID,
		URL:       link.URL,
		Filters:   link.Filters,
		Tags:      link.Tags,
		CheckedAt: link.CheckedAt,
	}

	checkLink.Chats = make([]int64, len(link.Chats))
	copy(checkLink.Chats, link.Chats)

	return checkLink
}

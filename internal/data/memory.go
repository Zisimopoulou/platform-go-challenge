package data

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Zisimopoulou/platform-go-challenge/internal/models"
)

type InMemoryStore struct {
	mu       sync.RWMutex
	data     map[string]map[string]models.RawAsset
	counters map[string]int64
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{
		data:     make(map[string]map[string]models.RawAsset),
		counters: make(map[string]int64),
	}
}

func (s *InMemoryStore) nextID(userID string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	c := s.counters[userID] + 1
	s.counters[userID] = c
	return fmt.Sprintf("%s-%s-%d", time.Now().UTC().Format("20060102T150405"), userID, c)
}

func (s *InMemoryStore) Add(userID string, asset models.RawAsset) (string, error) {
	favID := s.nextID(userID)
	asset.ID = favID
	asset.CreatedAt = time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.data[userID]; !ok {
		s.data[userID] = make(map[string]models.RawAsset)
	}
	s.data[userID][favID] = asset
	return favID, nil
}

func (s *InMemoryStore) List(userID string) ([]models.Favorite, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.data[userID]
	if !ok {
		return []models.Favorite{}, nil
	}
	out := make([]models.Favorite, 0, len(m))
	for k, v := range m {
		out = append(out, models.Favorite{FavoriteID: k, Asset: v})
	}
	return out, nil
}

func (s *InMemoryStore) Delete(userID, favID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.data[userID]
	if !ok {
		return errors.New("not found")
	}
	if _, ok := m[favID]; !ok {
		return errors.New("not found")
	}
	delete(m, favID)
	return nil
}

func (s *InMemoryStore) UpdateDescription(userID, favID, desc string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m, ok := s.data[userID]
	if !ok {
		return errors.New("not found")
	}
	asset, ok := m[favID]
	if !ok {
		return errors.New("not found")
	}
	asset.Description = desc
	m[favID] = asset
	return nil
}
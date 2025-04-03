package storage

import (
	"fmt"
	"sync"
	"time"
)

type Value struct {
	Data      string
	ExpiresAt time.Time
}

type Storage struct {
	data map[string]Value
	mu   sync.RWMutex
}

func NewStorage() *Storage {
	s := &Storage{
		data: make(map[string]Value),
	}

	go s.startAutoPurge()

	return s
}
func (s *Storage) Set(key, value string, ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	expiresAt := time.Time{}
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	fmt.Println(expiresAt)

	s.data[key] = Value{Data: value, ExpiresAt: expiresAt}
}

func (s *Storage) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exist := s.data[key]

	if !exist || (value.ExpiresAt.After(time.Time{}) && time.Now().After(value.ExpiresAt)) {
		return "", false
	}

	return value.Data, true
}

func (s *Storage) Del(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *Storage) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exist := s.data[key]
	return exist
}

func (s *Storage) Expire(key string, ttl time.Duration) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, exists := s.data[key]
	if !exists {
		return false
	}

	fmt.Println(time.Now().Add(ttl))

	val.ExpiresAt = time.Now().Add(ttl)
	s.data[key] = val
	return true
}

func (s *Storage) TTL(key string) time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, exists := s.data[key]

	if !exists || val.ExpiresAt.IsZero() {
		return -1
	}

	duration := time.Until(val.ExpiresAt)
	if duration < 0 {
		return -1
	}
	return duration
}

func (s *Storage) PurgeExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, val := range s.data {
		if val.ExpiresAt.After(time.Time{}) && now.After(val.ExpiresAt) {
			delete(s.data, key)
		}
	}
}

func (s *Storage) startAutoPurge() {
	ticker := time.NewTicker(10 * time.Second) // Интервал очистки
	defer ticker.Stop()

	for range ticker.C {
		s.PurgeExpired()
	}
}

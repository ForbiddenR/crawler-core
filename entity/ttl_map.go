package entity

import (
	"sync"
	"time"
)

type TTLMap struct {
	TTL time.Duration

	data sync.Map
}

type expireEntry struct {
	ExpiresAt time.Time
	Value     any
}

func (t *TTLMap) Store(key string, val any) {
	t.data.Store(key, expireEntry{
		ExpiresAt: time.Now().Add(t.TTL),
		Value:     val,
	})
}

func (t *TTLMap) Load(key string) (val any) {
	entry, ok := t.data.Load(key)
	if !ok {
		return nil
	}

	expireEntry := entry.(expireEntry)
	if expireEntry.ExpiresAt.After(time.Now()) {
		return nil
	}

	return expireEntry.Value
}

func NewTTLMap(ttl time.Duration) (m *TTLMap) {
	m = &TTLMap{
		TTL: ttl,
	}

	go func() {
		for now := range time.Tick(time.Second) {
			m.data.Range(func(k, v any) bool {
				expiresAt := v.(expireEntry).ExpiresAt
				if expiresAt.Before(now) {
					m.data.Delete(k)
				}
				return true
			})
		}
	}()

	return
}

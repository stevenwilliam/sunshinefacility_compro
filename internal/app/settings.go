package app

import (
	"context"
	"strconv"
	"sync"
	"time"
)

// Settings reads parameters with a short cache. A parameter write calls
// Invalidate, so a change is live on the next request; the TTL only bounds
// how stale a value can be if a write happens outside the app.
type Settings struct {
	store Store
	ttl   time.Duration

	mu     sync.Mutex
	at     time.Time
	values map[string]Parameter
}

func NewSettings(store Store) *Settings {
	return &Settings{store: store, ttl: 30 * time.Second}
}

func (s *Settings) Invalidate() {
	s.mu.Lock()
	s.at = time.Time{}
	s.mu.Unlock()
}

func (s *Settings) load(ctx context.Context) map[string]Parameter {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.values != nil && time.Since(s.at) < s.ttl {
		return s.values
	}
	all, err := s.store.AllParameters(ctx)
	if err != nil {
		// Keep serving the last good values rather than failing every
		// request because one read failed. Without any, callers get their
		// defaults.
		if s.values != nil {
			return s.values
		}
		return map[string]Parameter{}
	}
	m := make(map[string]Parameter, len(all))
	for _, p := range all {
		m[p.Group+"."+p.Key] = p
	}
	s.values, s.at = m, time.Now()
	return m
}

func (s *Settings) String(ctx context.Context, key, def string) string {
	if p, ok := s.load(ctx)[key]; ok && p.ValueEN != "" {
		return p.ValueEN
	}
	return def
}

// Int returns the parameter, or def when it is missing or unparseable. A
// non-positive value also falls back: a misconfigured threshold must not
// silently switch a control off (impeccable §2).
func (s *Settings) Int(ctx context.Context, key string, def int) int {
	p, ok := s.load(ctx)[key]
	if !ok {
		return def
	}
	n, err := strconv.Atoi(p.ValueEN)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

// IntAllowZero is Int for thresholds where 0 is meaningful (a minimum fill
// time of 0 seconds turns the timing check off deliberately).
func (s *Settings) IntAllowZero(ctx context.Context, key string, def int) int {
	p, ok := s.load(ctx)[key]
	if !ok {
		return def
	}
	n, err := strconv.Atoi(p.ValueEN)
	if err != nil || n < 0 {
		return def
	}
	return n
}

func (s *Settings) Keys(ctx context.Context, group string) []Parameter {
	var out []Parameter
	for _, p := range s.load(ctx) {
		if p.Group == group {
			out = append(out, p)
		}
	}
	return out
}

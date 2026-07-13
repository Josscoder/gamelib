package gamelib

import (
	"sync"
	"time"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

type Session struct {
	mu       sync.RWMutex
	xuid     string
	name     string
	handle   *world.EntityHandle
	match    *Match
	metadata map[string]any
}

// Match returns the current match, or nil.
func (s *Session) Match() *Match {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.match
}

// SetMatch assigns the session to a match (or nil to clear).
func (s *Session) setMatch(m *Match) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.match = m
}

// Player resolves the *player.Player from a Tx.
func (s *Session) Player(tx *world.Tx) (*player.Player, bool) {
	h := s.Handle()
	if h == nil {
		return nil, false
	}
	e, ok := h.Entity(tx)
	if !ok {
		return nil, false
	}
	return e.(*player.Player), true
}

// Handle returns the entity handle. Safe to call from any goroutine; may
// return nil once the session has closed (e.g. after the player quit).
func (s *Session) Handle() *world.EntityHandle {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.handle
}

// Do schedules f to run with the player on their current world owner and
// returns immediately, safe to call from any goroutine (timers, webhooks,
// background workers, etc.) — including from inside a match/component
// callback that is already on some other world's owner, where a blocking
// call would risk a deadlock. If the session has since closed (the player
// quit), the returned *world.Task fails with world.ErrEntityClosed.
func (s *Session) Do(f func(tx *world.Tx, p *player.Player)) *world.Task {
	return player.Do(s.Handle(), f)
}

// DoAfter schedules f to run with the player after delay, following them
// across world changes (e.g. portals) in the meantime. Safe to call from
// any goroutine, for the same reasons as Do.
func (s *Session) DoAfter(delay time.Duration, f func(tx *world.Tx, p *player.Player)) *world.Task {
	return player.DoAfter(s.Handle(), delay, f)
}

// XUID returns the player's XUID.
func (s *Session) XUID() string { return s.xuid }

// Name returns the player's name.
func (s *Session) Name() string { return s.name }

// Metadata returns the metadata map. Thread-safe via Session's mutex.
func (s *Session) SetMeta(key string, val any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.metadata[key] = val
}

func (s *Session) Meta(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.metadata[key]
	return v, ok
}

func (s *Session) close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.match = nil
	s.handle = nil
	s.metadata = nil
}

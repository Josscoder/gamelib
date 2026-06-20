package gamelib

import (
	"sync"

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
	e, ok := s.handle.Entity(tx)
	if !ok {
		return nil, false
	}
	return e.(*player.Player), true
}

// Handle returns the entity handle.
func (s *Session) Handle() *world.EntityHandle {
	return s.handle
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

package gamelib

import (
	"sync"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

type ParticipantState uint8

const (
	ParticipantAlive ParticipantState = iota
	ParticipantSpectating
	ParticipantEliminated
)

type Participant struct {
	mu      sync.RWMutex
	session *Session
	state   ParticipantState
	data    map[string]any // game-specific data (kills, score, etc.)
}

// Session returns the underlying session.
func (p *Participant) Session() *Session { return p.session }

// State returns the participant state.
func (p *Participant) State() ParticipantState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.state
}

// SetState changes the participant state.
func (p *Participant) SetState(s ParticipantState) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.state = s
}

// Alive returns true if the participant is alive.
func (p *Participant) Alive() bool { return p.State() == ParticipantAlive }

// Player resolves the *player.Player from a Tx.
func (p *Participant) Player(tx *world.Tx) (*player.Player, bool) {
	return p.session.Player(tx)
}

// XUID returns the participant's XUID (delegate to session).
func (p *Participant) XUID() string { return p.session.XUID() }

// Name returns the participant's name.
func (p *Participant) Name() string { return p.session.Name() }

// SetData / Data for game-specific participant data.
func (p *Participant) SetData(key string, val any) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data[key] = val
}

func (p *Participant) Data(key string) (any, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	v, ok := p.data[key]
	return v, ok
}

// DataAs is a typed convenience wrapper.
func DataAs[T any](p *Participant, key string) (T, bool) {
	v, ok := p.Data(key)
	if !ok {
		var zero T
		return zero, false
	}
	t, ok := v.(T)
	return t, ok
}

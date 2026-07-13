package gamelib

import (
	"sync"
	"time"

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

// Do schedules f to run with the participant's player on their current world
// owner and returns immediately. Safe to call from any goroutine, including
// from a callback already running on a different world's owner — the case
// that used to risk a self-deadlock with the old World.Exec/ExecWorld APIs.
// If the player has since quit, the returned *world.Task fails with
// world.ErrEntityClosed.
func (p *Participant) Do(f func(tx *world.Tx, pl *player.Player)) *world.Task {
	return p.session.Do(f)
}

// DoAfter schedules f to run with the participant's player after delay,
// following them across world changes in the meantime. Safe to call from
// any goroutine, for the same reasons as Do.
func (p *Participant) DoAfter(delay time.Duration, f func(tx *world.Tx, pl *player.Player)) *world.Task {
	return p.session.DoAfter(delay, f)
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

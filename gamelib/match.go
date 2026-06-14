package gamelib

import (
	"errors"
	"fmt"
	"iter"
	"log/slog"
	"math/rand"
	"sync"
	"sync/atomic"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/josscoder/fsmgo/state"
)

type MatchState uint8

const (
	MatchStateCreated MatchState = iota
	MatchStateWaiting
	MatchStatePlaying
	MatchStateEnding
	MatchStateClosed
)

type Match struct {
	mu sync.RWMutex

	id           uuid.UUID
	engine       *Engine
	definition   *GameDefinition
	participants *SyncMap[string, *Participant]
	state        MatchState
	metadata     map[string]any
	log          *slog.Logger

	// World
	arena         *Arena
	availableMaps []*GameMap
	selectedMap   *GameMap
	mapLoader     MapLoader

	// Handlers — per-match, scoped
	playerHandler player.Handler
	worldHandler  world.Handler

	// Lifecycle
	components  []Component
	stateSeries *state.ScheduledStateSeries

	// Hooks
	onClose func()

	closed atomic.Bool
}

// ID returns a unique identifier for this match.
func (m *Match) ID() uuid.UUID { return m.id }

// Engine returns the parent engine.
func (m *Match) Engine() *Engine { return m.engine }

// Definition returns the game definition.
func (m *Match) Definition() *GameDefinition { return m.definition }

// State returns the current match state.
func (m *Match) State() MatchState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state
}

func (m *Match) setState(s MatchState) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state = s
}

// Arena returns the match's arena (world wrapper).
func (m *Match) Arena() *Arena { return m.arena }

// World returns the underlying Dragonfly *world.World.
func (m *Match) World() *world.World {
	if m.arena == nil {
		return nil
	}
	return m.arena.World()
}

// StateSeries returns the match's state series.
func (m *Match) StateSeries() *state.ScheduledStateSeries { return m.stateSeries }

// CurrentState returns the active state.
func (m *Match) CurrentState() state.State {
	return m.stateSeries.Current()
}

//SkipState, FreezeState, UnFreezeState, etc

// SelectedMap returns the map chosen for this match.
func (m *Match) SelectedMap() *GameMap { return m.selectedMap }

// PlayerHandler returns the per-match player handler.
func (m *Match) PlayerHandler() player.Handler { return m.playerHandler }

// WorldHandler returns the per-match world handler.
func (m *Match) WorldHandler() world.Handler { return m.worldHandler }

// --- Player Management ---

// Join adds a player to this match.
func (m *Match) Join(p *player.Player) error {
	if m.closed.Load() {
		return errors.New("match is closed")
	}
	if m.State() != MatchStateWaiting {
		return errors.New("match is not accepting players")
	}
	if m.isFull() {
		return errors.New("match is full")
	}

	sess, ok := m.engine.Session(p.XUID())
	if !ok {
		return errors.New("no session found")
	}
	if sess.Match() != nil {
		return errors.New("player is already in a match")
	}

	par := &Participant{
		session: sess,
		state:   ParticipantAlive,
		data:    make(map[string]any),
	}
	m.participants.Store(p.XUID(), par)
	sess.setMatch(m)

	// Notify components.
	for _, c := range m.components {
		c.OnJoin(p, par)
	}

	return nil
}

// Leave removes a player from the match.
func (m *Match) leave(p *player.Player, disconnected bool) {
	par, ok := m.participants.Load(p.XUID())
	if !ok {
		return
	}

	// Notify components.
	for _, c := range m.components {
		c.OnQuit(p, par, disconnected)
	}

	m.participants.Delete(p.XUID())
	par.session.setMatch(nil)
}

func (m *Match) Leave(p *player.Player) {
	m.leave(p, false)
}

// Participants iterates all participants.
func (m *Match) Participants() iter.Seq[*Participant] {
	return func(yield func(*Participant) bool) {
		for _, par := range m.participants.Map() {
			if !yield(par) {
				return
			}
		}
	}
}

// AliveParticipants returns only alive participants.
func (m *Match) AliveParticipants() iter.Seq[*Participant] {
	return func(yield func(*Participant) bool) {
		for _, par := range m.participants.Map() {
			if par.state == ParticipantAlive {
				if !yield(par) {
					return
				}
			}
		}
	}
}

// ParticipantCount returns the number of participants.
func (m *Match) ParticipantCount() int {
	return m.participants.Len()
}

func (m *Match) isFull() bool {
	return m.ParticipantCount() >= m.definition.MaxPlayers
}

// AliveCount returns the number of alive participants.
func (m *Match) AliveCount() int {
	n := 0
	for _, p := range m.participants.Map() {
		if p.state == ParticipantAlive {
			n++
		}
	}
	return n
}

// Players calls fn for each participant that can be resolved in tx.
func (m *Match) Players(tx *world.Tx, fn func(*player.Player, *Participant)) {
	for _, par := range m.participants.Map() {
		p, ok := par.session.Player(tx)
		if !ok {
			continue
		}
		fn(p, par)
	}
}

// Broadcast sends a message to all players via tx.
func (m *Match) Broadcast(tx *world.Tx, format string, args ...any) {
	m.Players(tx, func(p *player.Player, _ *Participant) {
		p.Messagef(format, args...)
	})
}

// SetMeta / Meta for arbitrary match data.
func (m *Match) SetMeta(key string, val any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata[key] = val
}

func (m *Match) Meta(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.metadata[key]
	return v, ok
}

// --- Lifecycle ---

// Open prepares the match: loads a map, creates the world,
// installs handlers, and enters the Waiting state.
func (m *Match) Open() error {
	// Select a random map (or allow voting later).
	if len(m.availableMaps) == 0 {
		return errors.New("no maps available")
	}
	m.selectedMap = m.availableMaps[rand.Intn(len(m.availableMaps))]

	// Load world.
	arena, err := m.selectedMap.LoadArena(m.id)
	if err != nil {
		return fmt.Errorf("loading arena: %w", err)
	}
	m.arena = arena

	// Create handlers.
	if m.definition.NewPlayerHandler != nil {
		m.playerHandler = m.definition.NewPlayerHandler(m)
	} else {
		m.playerHandler = player.NopHandler{}
	}

	if m.definition.NewWorldHandler != nil {
		m.worldHandler = m.definition.NewWorldHandler(m)
	} else {
		m.worldHandler = world.NopHandler{}
	}

	// Install the world handler.
	m.arena.World().Handle(&matchWorldHandler{match: m})

	// Enable components.
	for _, c := range m.components {
		c.Enable(m)
	}

	// Start state series.
	m.stateSeries.Start()

	m.setState(MatchStateWaiting)
	m.log.Info("match opened", "map", m.selectedMap.Name)
	return nil
}

// Start transitions from Waiting → Playing.
func (m *Match) Start(tx *world.Tx) {
	if m.State() != MatchStateWaiting {
		return
	}
	m.setState(MatchStatePlaying)

	// Notify components.
	for _, c := range m.components {
		c.OnStart(tx)
	}

	m.log.Info("match started")
}

// End transitions to Ending state.
func (m *Match) End(tx *world.Tx) {
	if m.State() != MatchStatePlaying {
		return
	}
	m.setState(MatchStateEnding)

	// Notify components.
	for _, c := range m.components {
		c.OnEnd(tx)
	}

	m.log.Info("match ending")
}

// Close tears down everything.
func (m *Match) Close(tx *world.Tx) {
	if m.closed.CompareAndSwap(false, true) {
		// Disable components.
		for _, c := range m.components {
			c.Disable()
		}

		// Stop state series.
		m.stateSeries.End()

		// Remove all players.
		m.Players(tx, func(p *player.Player, par *Participant) {
			m.leave(p, false)
		})

		m.setState(MatchStateClosed)

		// Close arena.
		if m.arena != nil {
			m.arena.Close()
		}

		if m.onClose != nil {
			m.onClose()
		}

		m.log.Info("match closed")
	}
}

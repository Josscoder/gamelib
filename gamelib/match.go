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
	teams        *SyncMap[string, *Team]
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

// ShortID returns the first 5 characters of the UUID string.
func (m *Match) ShortID() string { return m.id.String()[:5] }

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

// CurrentState returns the active state, or nil if no series is configured.
func (m *Match) CurrentState() state.State {
	if m.stateSeries == nil {
		return nil
	}
	return m.stateSeries.Current()
}

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

// leave removes a player from the match.
func (m *Match) leave(p *player.Player, disconnected bool) {
	par, ok := m.participants.Load(p.XUID())
	if !ok {
		return
	}

	for _, c := range m.components {
		c.OnPreQuit(p, par, disconnected)
	}

	m.participants.Delete(p.XUID())
	par.session.setMatch(nil)

	for _, c := range m.components {
		c.OnPostQuit(p, disconnected)
	}
}

// Leave removes a player voluntarily from the match.
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

func (m *Match) HasEnoughPlayers() bool {
	return m.ParticipantCount() >= m.definition.MinPlayers
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

// ForEachPlayer iterates participants without allowing mutation of m.participants.
func (m *Match) ForEachPlayer(tx *world.Tx, fn func(*player.Player, *Participant)) {
	for _, par := range m.participants.Map() {
		p, ok := par.session.Player(tx)
		if !ok {
			continue
		}
		fn(p, par)
	}
}

// Players returns a snapshot slice, safe to use while mutating m.participants.
func (m *Match) Players(tx *world.Tx) []*player.Player {
	participants := m.participants.Map()
	players := make([]*player.Player, 0, len(participants))

	for _, par := range participants {
		p, ok := par.session.Player(tx)
		if !ok {
			continue
		}
		players = append(players, p)
	}

	return players
}

// SetMeta stores an arbitrary value for the given key.
func (m *Match) SetMeta(key string, val any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.metadata[key] = val
}

// Meta retrieves an arbitrary value by key.
func (m *Match) Meta(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.metadata[key]
	return v, ok
}

// --- Lifecycle ---

// Open selects a random map and opens the arena.
func (m *Match) Open() error {
	if len(m.availableMaps) == 0 {
		return errors.New("no maps available")
	}
	m.selectedMap = m.availableMaps[rand.Intn(len(m.availableMaps))]
	return m.openArena()
}

// OpenWithMap forces a specific map by name and opens the arena.
func (m *Match) OpenWithMap(mapName string) error {
	for _, gm := range m.availableMaps {
		if gm.Name == mapName {
			m.selectedMap = gm
			return m.openArena()
		}
	}
	return fmt.Errorf("map %q not found in definition %q", mapName, m.definition.Name)
}

// openArena is the shared setup logic for Open and OpenWithMap.
func (m *Match) openArena() error {
	arena, err := m.selectedMap.LoadArena(m.id)
	if err != nil {
		return fmt.Errorf("loading arena: %w", err)
	}
	m.arena = arena

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

	m.arena.World().Handle(&matchWorldHandler{match: m})

	for _, c := range m.components {
		c.Enable(m)
	}

	if m.stateSeries != nil {
		m.stateSeries.Start()
	}

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

// Close tears down everything. Safe to call once.
func (m *Match) Close(tx *world.Tx) {
	if !m.closed.CompareAndSwap(false, true) {
		return
	}

	for _, c := range m.components {
		c.Disable()
	}

	// Stop state series.
	if m.stateSeries != nil {
		m.stateSeries.End()
	}

	// Remove all players
	for _, p := range m.Players(tx) {
		m.Leave(p)
	}

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

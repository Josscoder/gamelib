package gamelib

import (
	"fmt"
	"sync"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/google/uuid"
)

// MatchmakingStrategy is the pluggable decision-making layer of the Matchmaker.
// Implement this to provide custom logic (ELO, region, party grouping, etc.).
type MatchmakingStrategy interface {
	// PickMatch receives all valid waiting candidates and returns the best one,
	// or nil to signal that a new match should be created instead.
	PickMatch(p *player.Player, def *GameDefinition, candidates []*Match) *Match
}

// DefaultMatchmakingStrategy fills the most populated waiting match first,
// so matches reach MinPlayers and start as fast as possible.
type DefaultMatchmakingStrategy struct{}

func (DefaultMatchmakingStrategy) PickMatch(_ *player.Player, _ *GameDefinition, candidates []*Match) *Match {
	var best *Match
	for _, m := range candidates {
		if best == nil || m.ParticipantCount() > best.ParticipantCount() {
			best = m
		}
	}
	return best
}

// QueueOptions configure how the Matchmaker places a single player.
type QueueOptions struct {
	// MapName forces the player into a match using this specific map.
	// If no waiting match with this map exists, a new one is created with it.
	// Leave empty for automatic (random) map selection.
	MapName string
}

// Matchmaker manages active matches and queues players into the best available match.
type Matchmaker struct {
	mu            sync.RWMutex
	engine        *Engine
	strategy      MatchmakingStrategy
	activeMatches map[string]map[uuid.UUID]*Match
}

// newMatchmaker initializes a new matchmaker linked to the passed engine.
func newMatchmaker(e *Engine) *Matchmaker {
	return &Matchmaker{
		engine:        e,
		strategy:      DefaultMatchmakingStrategy{},
		activeMatches: make(map[string]map[uuid.UUID]*Match),
	}
}

// SetStrategy replaces the matchmaking strategy with a custom one.
func (mm *Matchmaker) SetStrategy(s MatchmakingStrategy) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	mm.strategy = s
}

// ActiveMatches returns a snapshot of all active matches for a game type.
func (mm *Matchmaker) ActiveMatches(gameName string) []*Match {
	mm.mu.RLock()
	defer mm.mu.RUnlock()
	pool := mm.activeMatches[gameName]
	out := make([]*Match, 0, len(pool))
	for _, m := range pool {
		out = append(out, m)
	}
	return out
}

// Queue attempts to place a player into the best available match.
// If no matches are available (or all are full), it dynamically creates a new one.
func (mm *Matchmaker) Queue(p *player.Player, gameName string) (*Match, error) {
	return mm.QueueWithOptions(p, gameName, QueueOptions{})
}

// QueueWithOptions queues a player with full control over match selection.
func (mm *Matchmaker) QueueWithOptions(p *player.Player, gameName string, opts QueueOptions) (*Match, error) {
	def, ok := mm.engine.definitions[gameName]
	if !ok {
		return nil, fmt.Errorf("unknown game definition: %q", gameName)
	}

	// 1. Build candidates under a read lock.
	mm.mu.RLock()
	pool := mm.activeMatches[gameName]
	candidates := make([]*Match, 0, len(pool))
	for _, m := range pool {
		if m.State() != MatchStateWaiting || m.isFull() {
			continue
		}
		if opts.MapName != "" {
			sel := m.SelectedMap()
			if sel == nil || sel.Name != opts.MapName {
				continue
			}
		}
		candidates = append(candidates, m)
	}
	strategy := mm.strategy
	mm.mu.RUnlock()

	// 2. Let the strategy pick.
	if best := strategy.PickMatch(p, def, candidates); best != nil {
		if err := best.Join(p); err == nil {
			return best, nil
		}
		// Slot was taken between picking and joining — fall through to create a new match.
	}

	// 3. Create a new match.
	newMatch, err := def.NewMatch(mm.engine)
	if err != nil {
		return nil, fmt.Errorf("creating match: %w", err)
	}

	newMatch.onClose = func() {
		mm.unregisterMatch(def.Name, newMatch.ID())
	}

	if opts.MapName != "" {
		err = newMatch.OpenWithMap(opts.MapName)
	} else {
		err = newMatch.Open()
	}
	if err != nil {
		return nil, fmt.Errorf("opening match: %w", err)
	}

	mm.registerMatch(def.Name, newMatch)

	if err := newMatch.Join(p); err != nil {
		return nil, fmt.Errorf("joining new match: %w", err)
	}

	return newMatch, nil
}

// registerMatch adds a newly created match to the matchmaking pool.
func (mm *Matchmaker) registerMatch(gameName string, m *Match) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	if mm.activeMatches[gameName] == nil {
		mm.activeMatches[gameName] = make(map[uuid.UUID]*Match)
	}
	mm.activeMatches[gameName][m.ID()] = m
}

// unregisterMatch removes a match from the matchmaking pool (usually when it closes).
func (mm *Matchmaker) unregisterMatch(gameName string, mID uuid.UUID) {
	mm.mu.Lock()
	defer mm.mu.Unlock()
	if mm.activeMatches[gameName] != nil {
		delete(mm.activeMatches[gameName], mID)
	}
}

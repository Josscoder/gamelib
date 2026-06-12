package gamelib

import (
	"fmt"
	"sync"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/google/uuid"
)

// Matchmaker manages active matches and queues players into the best available match.
type Matchmaker struct {
	mu            sync.RWMutex
	engine        *Engine
	activeMatches map[string]map[uuid.UUID]*Match // Map definition Name -> Match ID -> Match instance
}

// newMatchmaker initializes a new matchmaker linked to the passed engine.
func newMatchmaker(e *Engine) *Matchmaker {
	return &Matchmaker{
		engine:        e,
		activeMatches: make(map[string]map[uuid.UUID]*Match),
	}
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

// Queue attempts to place a player into the best available match.
// If no matches are available (or all are full), it dynamically creates a new one.
func (mm *Matchmaker) Queue(p *player.Player, gameName string) (*Match, error) {
	def, ok := mm.engine.definitions[gameName]
	if !ok {
		return nil, fmt.Errorf("unknown game definition: %s", gameName)
	}

	mm.mu.RLock()
	matches := mm.activeMatches[gameName]

	var bestMatch *Match
	// Look through all active matches for this game type
	for _, m := range matches {
		// We only want matches in the Waiting state that have room
		if m.State() == MatchStateWaiting && m.ParticipantCount() < def.MaxPlayers {
			// Prioritize the match that is closest to starting (highest player count)
			if bestMatch == nil || m.ParticipantCount() > bestMatch.ParticipantCount() {
				bestMatch = m
			}
		}
	}
	mm.mu.RUnlock() // Unlock before attempting to join inside the Match, to avoid deadlocks

	// If we found a suitable match, try to join it
	if bestMatch != nil {
		err := bestMatch.Join(p)
		if err == nil {
			return bestMatch, nil
		}
		// If joining failed (e.g., someone else filled the last spot a millisecond ago),
		// we skip the return and fall through to create a new match.
	}

	// CREATE A NEW MATCH (Fallback)
	newMatch, err := def.NewMatch(mm.engine)
	if err != nil {
		return nil, fmt.Errorf("failed to create new match blueprint: %w", err)
	}

	// Register a hook so the match removes itself from Matchmaker when it closes
	newMatch.onClose = func() {
		mm.unregisterMatch(def.Name, newMatch.ID())
	}

	// Initialize the map, zip extraction, worlds and handlers
	if err := newMatch.Open(); err != nil {
		return nil, fmt.Errorf("failed to open new match arena: %w", err)
	}

	// Register in the active pool so other Queuing players can find it
	mm.registerMatch(def.Name, newMatch)

	// Deposit the player into their freshly generated match
	if err := newMatch.Join(p); err != nil {
		return nil, fmt.Errorf("failed joining newly created match: %w", err)
	}

	return newMatch, nil
}

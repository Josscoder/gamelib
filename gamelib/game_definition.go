package gamelib

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/scoreboard"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
)

type GameDefinition struct {
	Name       string
	MapsDir    string // path to maps directory
	MinPlayers int
	MaxPlayers int

	// Callbacks for creating per-match implementations.
	NewPlayerHandler func(m *Match) player.Handler
	NewWorldHandler  func(m *Match) world.Handler

	// Components to attach to every match.
	ComponentFactories []func(m *Match) Component

	// Phases define the ordered lifecycle of a match.
	PhaseFactories []func(m *Match) Phase

	// Optional: custom map loader.
	MapLoader MapLoader

	// Optional: custom scoring.
	NewScoreboard func(m *Match) scoreboard.Scoreboard
}

// NewMatch creates and loads a new Match from this definition.
func (def *GameDefinition) NewMatch(engine *Engine) (*Match, error) {
	m := &Match{
		id:           uuid.New(),
		engine:       engine,
		definition:   def,
		participants: NewSyncMap[string, *Participant](),
		state:        MatchStateCreated,
		metadata:     make(map[string]any),
		log:          engine.log.With("match", def.Name),
	}

	// Load available maps.
	loader := def.MapLoader
	if loader == nil {
		loader = &DefaultMapLoader{Dir: def.MapsDir}
	}
	maps, err := loader.LoadMaps()
	if err != nil {
		return nil, fmt.Errorf("loading maps for %s: %w", def.Name, err)
	}
	m.availableMaps = maps
	m.mapLoader = loader

	// Create components.
	for _, factory := range def.ComponentFactories {
		c := factory(m)
		m.components = append(m.components, c)
	}

	// Create phases.
	for _, factory := range def.PhaseFactories {
		p := factory(m)
		m.phases = append(m.phases, p)
	}

	// Create scheduler.
	m.scheduler = newScheduler(m)

	return m, nil
}

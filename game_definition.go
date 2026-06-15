package gamelib

import (
	"fmt"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/google/uuid"
	"github.com/josscoder/fsmgo/state"
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

	// StateSeries define the ordered lifecycle of a match.
	StateSeries func(m *Match) *state.ScheduledStateSeries

	// Optional: custom map loader.
	MapLoader MapLoader
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

	// Build the state series after the match is fully initialized.
	if def.StateSeries != nil {
		m.stateSeries = def.StateSeries(m)
	}

	return m, nil
}

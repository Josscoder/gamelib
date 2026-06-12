package gamelib

import "github.com/df-mc/dragonfly/server/world"

// Phase represents an ordered stage of a match (e.g., Grace Period,
// PvP, Deathmatch, Ending).
type Phase interface {
	// Name returns the display name.
	Name() string

	// Duration returns the phase duration in seconds. 0 = infinite
	// (must be advanced manually).
	Duration() int

	// Start is called when the phase begins.
	Start(tx *world.Tx, m *Match)

	// End is called when the phase ends (either timeout or manual).
	End(tx *world.Tx, m *Match)

	// Tick is called every tick while this phase is active.
	Tick(tx *world.Tx, m *Match, tick uint64)
}

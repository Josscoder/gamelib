package gamelib

import (
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

// Component is a modular, lifecycle-managed piece of game logic
// that can be attached to a Match.
type Component interface {
	// Enable is called when the match opens. Use this to register
	// scheduler tasks, initial state, etc.
	Enable(m *Match)

	// Disable is called when the match closes. Cleanup here.
	Disable()

	// OnJoin is called when a player joins the match.
	OnJoin(p *player.Player, par *Participant)

	// OnQuit is called when a player leaves.
	OnQuit(p *player.Player, par *Participant, disconnected bool)

	// OnStart is called when the match transitions to Playing.
	OnStart(tx *world.Tx)

	// OnEnd is called when the match transitions to Ending.
	OnEnd(tx *world.Tx)
}

// BaseComponent provides default no-op implementations.
// Embed this to avoid implementing every method.
type BaseComponent struct{}

func (BaseComponent) Enable(*Match)                             {}
func (BaseComponent) Disable()                                  {}
func (BaseComponent) OnJoin(*player.Player, *Participant)       {}
func (BaseComponent) OnQuit(*player.Player, *Participant, bool) {}
func (BaseComponent) OnStart(*world.Tx)                         {}
func (BaseComponent) OnEnd(*world.Tx)                           {}

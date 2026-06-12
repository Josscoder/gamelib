package example

import (
	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

// CageComponent — traps players in cages until game starts.
type CageComponent struct {
	gamelib.BaseComponent
	match *gamelib.Match
}

func NewCageComponent() func(m *gamelib.Match) gamelib.Component {
	return func(m *gamelib.Match) gamelib.Component {
		return &CageComponent{match: m}
	}
}

func (c *CageComponent) OnJoin(p *player.Player, par *gamelib.Participant) {
	// Teleport player to cage spawn from map config
	// Place cage blocks around spawn

	println("Se dio la jaula")
}

func (c *CageComponent) OnQuit(p *player.Player, par *gamelib.Participant, disconnected bool) {
	println("Se quito la jaula")

}

func (c *CageComponent) OnStart(tx *world.Tx) {
	// Destroy all cages — release players
	println("Se quito la jaula")
}

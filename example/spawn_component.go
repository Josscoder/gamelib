package example

import (
	"log"

	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

type SpawnComponent struct {
	gamelib.BaseComponent
	match *gamelib.Match
}

func NewSpawnComponent() func(m *gamelib.Match) gamelib.Component {
	return func(m *gamelib.Match) gamelib.Component {
		return &SpawnComponent{match: m}
	}
}

func (s *SpawnComponent) Enable(m *gamelib.Match) {
	if err := m.SelectedMap().LoadConfig(&ExampleMapData{}); err != nil {
		log.Fatal(err)
	}
	
}

func (s *SpawnComponent) OnJoin(p *player.Player, par *gamelib.Participant) {
	sm := s.match.SelectedMap()
	cfg := gamelib.GetConfig[*ExampleMapData](sm)
	p.Teleport(cfg.Mid.Vec3)
}

func (s *SpawnComponent) OnQuit(p *player.Player, par *gamelib.Participant, disconnected bool) {

}

func (s *SpawnComponent) OnStart(tx *world.Tx) {

}

package example_components

import (
	"log"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/josscoder/gamelib/example/example_config"
	"github.com/josscoder/gamelib/gamelib"
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
	if err := m.SelectedMap().LoadConfig(&example_config.ExampleMapData{}); err != nil {
		log.Fatal(err)
	}
}

func (s *SpawnComponent) OnJoin(p *player.Player, par *gamelib.Participant) {
	oldP := p.Tx().RemoveEntity(p)
	s.match.World().Exec(func(tx *world.Tx) {
		newP := tx.AddEntity(oldP).(*player.Player)
		sm := s.match.SelectedMap()
		cfg := gamelib.GetConfig[*example_config.ExampleMapData](sm)
		newP.Teleport(cfg.Mid.Vec3)
	})
}

func (s *SpawnComponent) OnQuit(p *player.Player, par *gamelib.Participant, disconnected bool) {

}

func (s *SpawnComponent) OnStart(tx *world.Tx) {
	sm := s.match.SelectedMap()
	cfg := gamelib.GetConfig[*example_config.ExampleMapData](sm)

	spawns := make([]example_config.Location, 0, len(cfg.Spawns))
	for _, spawn := range cfg.Spawns {
		spawns = append(spawns, spawn.SpawnPoint)
	}

	index := 0

	s.match.ForEachPlayer(tx, func(p *player.Player, pa *gamelib.Participant) {
		if len(spawns) == 0 {
			return
		}

		spawn := spawns[index%len(spawns)]

		p.Teleport(spawn.Vec3)
		//p.SetRotation(spawn.Rotation)

		index++
	})
}

func (s *SpawnComponent) OnEnd(tx *world.Tx) {

}

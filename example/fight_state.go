package example

import (
	"time"

	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

type FightState struct {
	*gamelib.MatchStateBase
}

func NewFightState(m *gamelib.Match) *FightState {
	s := &FightState{}
	s.MatchStateBase = gamelib.NewMatchStateBase(m, s)
	return s
}

func (fs *FightState) OnStart() {
	m := fs.Match()
	m.World().Exec(func(tx *world.Tx) {
		m.Start(tx)
	})
}

func (fs *FightState) OnUpdate(time.Duration) {
	m := fs.Match()
	m.World().Exec(func(tx *world.Tx) {
		m.Players(tx, func(p *player.Player, pa *gamelib.Participant) {
			if sb := m.Definition().NewScoreboard; sb != nil {
				p.SendScoreboard(sb(m, p, pa))
			}
		})
	})
}

func (fs *FightState) OnEnd() {

}

func (fs *FightState) GetDuration() time.Duration {
	return time.Minute * 10
}

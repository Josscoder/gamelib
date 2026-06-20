package example

import (
	"time"

	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

type EndGameState struct {
	*gamelib.MatchStateBase
}

func NewEndGameState(m *gamelib.Match) *EndGameState {
	s := &EndGameState{}
	s.MatchStateBase = gamelib.NewMatchStateBase(m, s)
	return s
}

func (es *EndGameState) OnStart() {
	m := es.Match()
	m.World().Exec(func(tx *world.Tx) {
		m.End(tx)
	})
}

func (es *EndGameState) OnUpdate(time.Duration) {
	m := es.Match()
	m.World().Exec(func(tx *world.Tx) {
		m.Players(tx, func(p *player.Player, pa *gamelib.Participant) {
			if sb := m.Definition().Scoreboard; sb != nil {
				p.SendScoreboard(sb(m, p, pa))
			}
		})
	})
}

func (es *EndGameState) OnEnd() {
	m := es.Match()
	m.World().Exec(func(tx *world.Tx) {
		m.Close(tx)
	})
}

func (es *EndGameState) GetDuration() time.Duration {
	return time.Second * 10
}

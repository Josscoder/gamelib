package example_states

import (
	"time"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/josscoder/gamelib/gamelib"
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
}

func (es *EndGameState) OnUpdate(time.Duration) {
	m := es.Match()
	m.World().Exec(func(tx *world.Tx) {
		m.ForEachPlayer(tx, func(p *player.Player, pa *gamelib.Participant) {
			if sb := m.Definition().NewScoreboard; sb != nil {
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

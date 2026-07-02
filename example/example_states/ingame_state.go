package example_states

import (
	"time"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/josscoder/gamelib/gamelib"
)

type InGameState struct {
	*gamelib.MatchStateBase
}

func NewInGameState(m *gamelib.Match) *InGameState {
	s := &InGameState{}
	s.MatchStateBase = gamelib.NewMatchStateBase(m, s)
	return s
}

func (is *InGameState) OnStart() {
	m := is.Match()
	m.World().Exec(func(tx *world.Tx) {
		m.Start(tx)
	})
}

func (is *InGameState) OnUpdate(time.Duration) {
	m := is.Match()
	m.World().Exec(func(tx *world.Tx) {
		m.ForEachPlayer(tx, func(p *player.Player, pa *gamelib.Participant) {
			if sb := m.Definition().NewScoreboard; sb != nil {
				p.SendScoreboard(sb(m, p, pa))
			}
		})
	})
}

func (is *InGameState) OnEnd() {
	m := is.Match()
	m.World().Exec(func(tx *world.Tx) {
		m.End(tx)
	})
}

func (is *InGameState) GetDuration() time.Duration {
	return time.Minute * 10
}

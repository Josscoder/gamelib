package example

import (
	"math"
	"time"

	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

type PreGameState struct {
	*gamelib.MatchStateBase
}

func NewPreGameState(m *gamelib.Match) *PreGameState {
	s := &PreGameState{}
	s.MatchStateBase = gamelib.NewMatchStateBase(m, s)
	return s
}

func (ps *PreGameState) OnStart() {
	ps.Pause()
}

func (ps *PreGameState) OnUpdate(time.Duration) {
	m := ps.Match()
	r := ps.GetRemainingTime()

	m.World().Exec(func(tx *world.Tx) {
		m.Players(tx, func(p *player.Player, pa *gamelib.Participant) {
			p.SetExperienceLevel(int(r.Seconds()))
			progress := math.Max(0, math.Min(1, r.Seconds()/ps.GetDuration().Seconds()))
			p.SetExperienceProgress(progress)

			if sb := m.Definition().NewScoreboard; sb != nil {
				p.SendScoreboard(sb(m, p, pa))
			}
		})

		if m.StateSeries().IsPaused() {
			return
		}

		if !m.HasEnoughPlayers() {
			if r != ps.GetDuration() {
				ps.SetRemainingTime(ps.GetDuration())
				ps.Pause()

				m.Players(tx, func(p *player.Player, pa *gamelib.Participant) {
					p.Message("Juego cancelado")
				})
			}
			return
		}

		ps.Resume()

		if int(r.Seconds()) != 0 && int(r.Seconds())%10 == 0 {
			m.Players(tx, func(p *player.Player, pa *gamelib.Participant) {
				p.Messagef("El juego comienza en %s", r.String())
			})
		}

		if int(r.Seconds()) <= 5 {
			m.Players(tx, func(p *player.Player, pa *gamelib.Participant) {
				p.SendTitle(title.New(text.Colourf("<bold>%d", int(r.Seconds()))))
			})
		}
	})
}

func (ps *PreGameState) OnEnd() {

}

func (ps *PreGameState) GetDuration() time.Duration {
	return time.Second * 31
}

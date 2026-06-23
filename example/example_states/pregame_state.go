package example_states

import (
	"math"
	"time"

	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/title"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/josscoder/gamelib/gamelib"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

const (
	PreGameIdleTimeout = 3 * time.Minute
	preGameIdleWarning = 30 * time.Second
)

type PreGameState struct {
	*gamelib.MatchStateBase

	IdleElapsed time.Duration
	warnedClose bool
}

func NewPreGameState(m *gamelib.Match) *PreGameState {
	s := &PreGameState{}
	s.MatchStateBase = gamelib.NewMatchStateBase(m, s)
	return s
}

func (ps *PreGameState) OnStart() {
	ps.IdleElapsed = 0
	ps.warnedClose = false
	ps.Pause()
}

func (ps *PreGameState) OnUpdate(delta time.Duration) {
	m := ps.Match()

	m.World().Exec(func(tx *world.Tx) {
		r := ps.GetRemainingTime()
		secs := int(r.Seconds())

		m.Players(tx, func(p *player.Player, pa *gamelib.Participant) {
			p.SetExperienceLevel(secs)

			progress := math.Max(
				0,
				math.Min(1, r.Seconds()/ps.GetDuration().Seconds()),
			)

			p.SetExperienceProgress(progress)

			if sb := m.Definition().NewScoreboard; sb != nil {
				p.SendScoreboard(sb(m, p, pa))
			}
		})

		if !m.HasEnoughPlayers() {
			ps.IdleElapsed += delta

			remaining := PreGameIdleTimeout - ps.IdleElapsed

			if !ps.warnedClose && remaining <= preGameIdleWarning {
				ps.warnedClose = true

				m.Players(tx, func(p *player.Player, _ *gamelib.Participant) {
					p.Message(text.Colourf(
						"<red>La partida cerrará en <bold>%.0f segundos</bold> por falta de jugadores.",
						remaining.Seconds(),
					))
				})
			}

			if ps.IdleElapsed >= PreGameIdleTimeout {
				m.Close(tx)
				return
			}

			if r != ps.GetDuration() {
				ps.SetRemainingTime(ps.GetDuration())
				ps.Pause()

				m.Players(tx, func(p *player.Player, _ *gamelib.Participant) {
					p.Message(text.Colourf(
						"<red>Juego pausado: no hay suficientes jugadores.",
					))
				})
			}

			return
		}

		ps.IdleElapsed = 0
		ps.warnedClose = false
		ps.Resume()

		if m.StateSeries().IsPaused() {
			return
		}

		switch {
		case secs > 5 && secs%10 == 0:
			m.Players(tx, func(p *player.Player, _ *gamelib.Participant) {
				p.Message(text.Colourf(
					"<yellow>El juego comienza en <bold>%d</bold> segundos.",
					secs,
				))
			})

		case secs > 0 && secs <= 5:
			m.Players(tx, func(p *player.Player, _ *gamelib.Participant) {
				p.SendTitle(title.New(
					text.Colourf("<red><bold>%d", secs),
				))
			})
		}
	})
}

func (ps *PreGameState) OnEnd() {
	m := ps.Match()
	m.World().Exec(func(tx *world.Tx) {
		m.Players(tx, func(p *player.Player, _ *gamelib.Participant) {
			p.SendTitle(title.New(
				text.Colourf("<green><bold>¡YA!"),
			))
		})
	})
}

func (ps *PreGameState) GetDuration() time.Duration {
	return 31 * time.Second
}

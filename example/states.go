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

const (
	preGameIdleTimeout = 3 * time.Minute
	preGameIdleWarning = 30 * time.Second
)

type PreGameState struct {
	*gamelib.MatchStateBase

	idleElapsed time.Duration
	warnedClose bool
}

func NewPreGameState(m *gamelib.Match) *PreGameState {
	s := &PreGameState{}
	s.MatchStateBase = gamelib.NewMatchStateBase(m, s)
	return s
}

func (ps *PreGameState) OnStart() {
	ps.idleElapsed = 0
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
			ps.idleElapsed += delta

			remaining := preGameIdleTimeout - ps.idleElapsed

			if !ps.warnedClose && remaining <= preGameIdleWarning {
				ps.warnedClose = true

				m.Players(tx, func(p *player.Player, _ *gamelib.Participant) {
					p.Message(text.Colourf(
						"<red>La partida cerrará en <bold>%.0f segundos</bold> por falta de jugadores.",
						remaining.Seconds(),
					))
				})
			}

			if ps.idleElapsed >= preGameIdleTimeout {
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

		ps.idleElapsed = 0
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
		m.Players(tx, func(p *player.Player, pa *gamelib.Participant) {
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
		m.Players(tx, func(p *player.Player, pa *gamelib.Participant) {
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

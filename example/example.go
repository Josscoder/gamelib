package example

import (
	"context"
	"strings"
	"time"

	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/scoreboard"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/josscoder/fsmgo/state"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

func Definition() *gamelib.GameDefinition {
	return &gamelib.GameDefinition{
		Name:       "example",
		MapsDir:    "example/maps",
		MinPlayers: 1,
		MaxPlayers: 12,
		NewPlayerHandler: func(m *gamelib.Match) player.Handler {
			return NewExamplePlayerHandler(m)
		},
		NewWorldHandler: func(m *gamelib.Match) world.Handler {
			return NewExampleWorldHandler(m)
		},
		ComponentFactories: []func(m *gamelib.Match) gamelib.Component{
			NewSpawnComponent(),
		},
		StateSeries: func(m *gamelib.Match) *state.ScheduledStateSeries {
			states := []state.State{
				NewPreGameState(m),
				NewFightState(m),
				NewEndGameState(m),
			}
			return state.NewScheduledStateSeries(context.Background(), states, time.Second)
		},
		Scoreboard: func(m *gamelib.Match, p *player.Player, pa *gamelib.Participant) *scoreboard.Scoreboard {
			sb := scoreboard.New(text.Colourf("<white><b>Example</b></white>"))
			sb.RemovePadding()

			dateStr := time.Now().Format("02.01.06")

			lines := []string{
				text.Colourf("<grey>%s <dark-grey>(%s)", dateStr, m.ShortID()),
				"",
			}

			cs := m.CurrentState()

			if _, ok := cs.(*PreGameState); ok {
				lines = append(lines,
					text.Colourf("<blue>Mapa:</blue>"),
					m.SelectedMap().Name,
					" ",
					text.Colourf("<blue>Jugadores:</blue>"),
					text.Colourf("%d/%d",
						m.AliveCount(), 12),
					"  ",
				)
				if !m.HasEnoughPlayers() {
					lines = append(lines, text.Colourf("<red>Esperando"))
				} else {
					lines = append(lines, text.Colourf("<blue>Inicia en:</blue>"), cs.GetRemainingTime().String())
				}
			} else if _, ok := cs.(*FightState); ok {
				lines = append(lines, text.Colourf("<blue>Termina en:</blue>"), cs.GetRemainingTime().String())
			} else if _, ok := cs.(*EndGameState); ok {
				lines = append(lines, text.Colourf("<blue>Termina en:</blue>"), cs.GetRemainingTime().String())
			}

			lines = append(lines, "   ")
			lines = append(lines, text.Colourf("<gold>mc.blockbrawn.net</gold>"))

			_, _ = sb.WriteString(strings.Join(lines, "\n"))

			return sb
		},
	}
}

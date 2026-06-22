package example

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/blockbrawn/gamelib/example/example_components"
	"github.com/blockbrawn/gamelib/example/example_handlers"
	"github.com/blockbrawn/gamelib/example/example_states"
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
			return example_handlers.NewExamplePlayerHandler(m)
		},
		NewWorldHandler: func(m *gamelib.Match) world.Handler {
			return example_handlers.NewExampleWorldHandler(m)
		},
		ComponentFactories: []func(m *gamelib.Match) gamelib.Component{
			example_components.NewSpawnComponent(),
		},
		StateSeries: func(m *gamelib.Match) *state.ScheduledStateSeries {
			states := []state.State{
				example_states.NewPreGameState(m),
				example_states.NewInGameState(m),
				example_states.NewEndGameState(m),
			}
			return state.NewScheduledStateSeries(context.Background(), states, time.Second)
		},
		NewScoreboard: func(m *gamelib.Match, p *player.Player, pa *gamelib.Participant) *scoreboard.Scoreboard {
			return exampleScoreboard(m, p, pa)
		},
	}
}

func exampleScoreboard(m *gamelib.Match, p *player.Player, pa *gamelib.Participant) *scoreboard.Scoreboard {
	sb := scoreboard.New(text.Colourf("<white><b>SkyWars Solo</b></white>"))
	sb.RemovePadding()

	dateStr := time.Now().Format("02.01.06")

	lines := []string{
		text.Colourf("<grey>%s <dark-grey>(%s)", dateStr, m.ShortID()),
		"",
	}

	cs := m.CurrentState()
	if _, ok := cs.(*example_states.PreGameState); ok {
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
			lines = append(lines,
				text.Colourf("<red>Esperando..."),
				"   ",
			)
		} else {
			lines = append(lines,
				text.Colourf("<blue>Inicia en:</blue>"),
				cs.GetRemainingTime().String(),
				"   ",
			)
		}
	} else if _, ok := cs.(*example_states.InGameState); ok {
		lines = append(lines,
			text.Colourf("<blue>Termina en:</blue>"),
			cs.GetRemainingTime().String(),
			"   ",
			text.Colourf("<blue>Relleno en:</blue>"),
			"4:00",
			"    ",
			text.Colourf("<blue>Jugadores vivos:</blue>"),
			strconv.Itoa(m.AliveCount()),
			"     ",
		)
	} else if _, ok := cs.(*example_states.EndGameState); ok {
		lines = append(lines,
			text.Colourf("<blue>Termina en:</blue>"),
			cs.GetRemainingTime().String(),
			"   ",
		)
	}

	lines = append(lines, text.Colourf("<gold>mc.blockbrawn.net</gold>"))
	_, _ = sb.WriteString(strings.Join(lines, "\n"))

	return sb
}

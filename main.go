package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/blockbrawn/gamelib/example"
	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/josscoder/fsmgo/state"
	_ "github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)
	chat.Global.Subscribe(chat.StdoutSubscriber{})

	conf, err := server.DefaultConfig().Config(slog.Default())
	if err != nil {
		panic(err)
	}

	engine := gamelib.EngineConfig{
		Log: slog.Default(),
	}.New()

	exampleDef := &gamelib.GameDefinition{
		Name:       "example",
		MapsDir:    "example/maps",
		MinPlayers: 2,
		MaxPlayers: 12,
		NewPlayerHandler: func(m *gamelib.Match) player.Handler {
			return example.NewExamplePlayerHandler(m)
		},
		NewWorldHandler: func(m *gamelib.Match) world.Handler {
			return example.NewExampleWorldHandler(m)
		},
		ComponentFactories: []func(m *gamelib.Match) gamelib.Component{
			example.NewCageComponent(),
		},
		StateSeries: func(m *gamelib.Match) *state.ScheduledStateSeries {
			return state.NewScheduledStateSeries(context.Background(), []state.State{
				example.NewPreGameState(m),
			}, time.Second)
		},
	}
	engine.Register(exampleDef)

	srv := conf.New()
	srv.CloseOnProgramEnd()
	srv.Listen()

	for p := range srv.Accept() {
		_ = engine.HandleJoin(p)
		p.Handle(engine.PlayerHandler())

		m, err := engine.Matchmaker().Queue(p, "example")
		if err != nil {
			slog.Default().Error("failed to queue matchmaker", "error", err)
			continue
		}

		p.Message(text.Colourf("<yellow>Conectando a %s-%s", m.SelectedMap().Name, m.ID()))

		h := p.Tx().RemoveEntity(p)

		m.World().Exec(func(tx *world.Tx) {
			newP := tx.AddEntity(h).(*player.Player)
			newP.Teleport(mgl64.Vec3{0, 100, 0})
			newP.SetGameMode(world.GameModeCreative)

			m.Start(tx)
		})
	}
}

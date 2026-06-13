package main

import (
	"log/slog"

	"github.com/blockbrawn/gamelib/example"
	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	_ "github.com/pelletier/go-toml"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

func main() {
	chat.Global.Subscribe(chat.StdoutSubscriber{})

	conf, err := server.DefaultConfig().Config(slog.Default())
	if err != nil {
		panic(err)
	}

	// 1. Create engine.
	engine := gamelib.EngineConfig{
		Log: slog.Default(),
	}.New()
	// 2. Define SkyWars.
	skywarsDef := &gamelib.GameDefinition{
		Name:       "skywars",
		MapsDir:    "maps/skywars",
		MinPlayers: 2,
		MaxPlayers: 12,
		NewPlayerHandler: func(m *gamelib.Match) player.Handler {
			return example.NewSkyWarsPlayerHandler(m)
		},
		NewWorldHandler: func(m *gamelib.Match) world.Handler {
			return example.NewSkyWarsWorldHandler(m)
		},
		ComponentFactories: []func(m *gamelib.Match) gamelib.Component{
			example.NewCountdownComponent(10),
			example.NewCageComponent(),
			//NewChestFillComponent(),
		},
		PhaseFactories: []func(m *gamelib.Match) gamelib.Phase{
			func(m *gamelib.Match) gamelib.Phase {
				return &example.GracePeriodPhase{}
			},
			/*func(m *gamelib.Match) gamelib.Phase {
				return &PvPPhase{}
			},
			func(m *gamelib.Match) gamelib.Phase {
				return &DeathmatchPhase{}
			},*/
		},
	}
	engine.Register(skywarsDef)

	srv := conf.New()
	srv.CloseOnProgramEnd()
	srv.Listen()

	for p := range srv.Accept() {
		_ = engine.HandleJoin(p)
		p.Handle(engine.PlayerHandler())

		m, err := engine.Matchmaker().Queue(p, "skywars")
		if err != nil {
			slog.Default().Error("failed to queue matchmaker", "error", err)
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

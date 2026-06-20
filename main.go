package main

import (
	"log"
	"log/slog"

	"github.com/blockbrawn/gamelib/example"
	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player/chat"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

func main() {
	chat.Global.Subscribe(chat.StdoutSubscriber{})

	conf, err := server.DefaultConfig().Config(slog.Default())
	if err != nil {
		panic(err)
	}

	srv := conf.New()
	srv.CloseOnProgramEnd()
	srv.Listen()

	engine := gamelib.EngineConfig{
		Log: slog.Default(),
	}.New()

	def := example.Definition()
	engine.Register(def)

	commands := []cmd.Command{
		cmd.New("pause", "Pause state", []string{}, gamelib.NewPauseCommand(engine)),
		cmd.New("resume", "Resume state", []string{}, gamelib.NewResumeCommand(engine)),
		cmd.New("skip", "Skip the current state", []string{}, gamelib.NewSkipCommand(engine)),
		cmd.New("previous", "Return to the previous state", []string{}, gamelib.NewPreviousCommand(engine)),
		cmd.New("rewind", "Rewind to the first state", []string{}, gamelib.NewRewindCommand(engine)),
	}

	for _, c := range commands {
		cmd.Register(c)
	}

	for p := range srv.Accept() {
		_ = engine.HandleJoin(p)

		m, err := engine.Matchmaker().Queue(p, def.Name)
		if err != nil {
			log.Fatal("failed to queue matchmaker", "error", err)
		}

		p.Handle(engine.PlayerHandler())
		p.Message(text.Colourf("<green>Conectando al mapa %s <grey>(%s)", m.SelectedMap().Name, m.ShortID()))
	}
}

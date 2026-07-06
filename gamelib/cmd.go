package gamelib

import (
	"errors"

	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

func matchFromSource(engine *Engine, src cmd.Source) (*Match, error) {
	p, ok := src.(*player.Player)
	if !ok {
		return nil, errors.New("only players can execute this command")
	}

	sess, ok := engine.Session(p.XUID())
	if !ok {
		return nil, errors.New("you are not playing")
	}

	m := sess.Match()
	if m == nil {
		return nil, errors.New("you are not playing")
	}

	return m, nil
}

type PauseCommand struct {
	engine *Engine
}

func NewPauseCommand(engine *Engine) *PauseCommand {
	return &PauseCommand{engine: engine}
}

func (pc PauseCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	m, err := matchFromSource(pc.engine, src)
	if err != nil {
		output.Error(err.Error())
		return
	}

	series := m.StateSeries()
	if series.IsPaused() {
		output.Error("The match is already paused")
		return
	}
	series.Pause()

	output.Print(text.Colourf(
		"<green>The match has been successfully paused</green>",
	))
}

type ResumeCommand struct {
	engine *Engine
}

func NewResumeCommand(engine *Engine) *ResumeCommand {
	return &ResumeCommand{engine: engine}
}

func (rc ResumeCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	m, err := matchFromSource(rc.engine, src)
	if err != nil {
		output.Error(err.Error())
		return
	}

	series := m.StateSeries()
	if !series.IsPaused() {
		output.Error("The match is not paused")
		return
	}
	series.Resume()

	output.Print(text.Colourf(
		"<green>The match has successfully resumed</green>",
	))
}

type SkipCommand struct {
	engine *Engine
}

func NewSkipCommand(engine *Engine) *SkipCommand {
	return &SkipCommand{engine: engine}
}

func (sc SkipCommand) Run(src cmd.Source, output *cmd.Output, _ *world.Tx) {
	m, err := matchFromSource(sc.engine, src)
	if err != nil {
		output.Error(err.Error())
		return
	}

	m.StateSeries().Skip()

	output.Print(text.Colourf(
		"<green>The match has successfully advanced to the next state</green>",
	))
}

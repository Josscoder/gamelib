package main

import (
	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/world"
)

type ExampleWorldHandler struct {
	world.NopHandler

	match *gamelib.Match
}

func NewExampleWorldHandler(m *gamelib.Match) *ExampleWorldHandler {
	return &ExampleWorldHandler{match: m}
}

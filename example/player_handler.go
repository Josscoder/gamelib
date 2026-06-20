package example

import (
	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/player"
)

type ExamplePlayerHandler struct {
	player.NopHandler
	match *gamelib.Match
}

func NewExamplePlayerHandler(m *gamelib.Match) *ExamplePlayerHandler {
	return &ExamplePlayerHandler{match: m}
}

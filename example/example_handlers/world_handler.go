package example_handlers

import (
	"github.com/df-mc/dragonfly/server/world"
	"github.com/josscoder/gamelib/gamelib"
)

type ExampleWorldHandler struct {
	world.NopHandler
	match *gamelib.Match
}

func NewExampleWorldHandler(m *gamelib.Match) *ExampleWorldHandler {
	return &ExampleWorldHandler{match: m}
}

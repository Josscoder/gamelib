package example

import (
	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/world"
)

type SkyWarsWorldHandler struct {
	world.NopHandler

	match *gamelib.Match
}

func NewSkyWarsWorldHandler(m *gamelib.Match) *SkyWarsWorldHandler {
	return &SkyWarsWorldHandler{match: m}
}

package example

import (
	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
)

type ExamplePlayerHandler struct {
	player.NopHandler

	match *gamelib.Match
}

func NewExamplePlayerHandler(m *gamelib.Match) *ExamplePlayerHandler {
	return &ExamplePlayerHandler{match: m}
}

func (h *ExamplePlayerHandler) HandleDeath(p *player.Player, src world.DamageSource, keepInv *bool) {
	*keepInv = false
	//par, ok := h.match.Engine().Session(p.XUID())
}

func (h *ExamplePlayerHandler) HandleBlockPlace(ctx *player.Context, pos cube.Pos, b world.Block) {
	if pos.Y() > 128 {
		ctx.Cancel()
	}
}

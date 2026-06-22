package example_handlers

import (
	"github.com/blockbrawn/gamelib/example/example_config"
	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/go-gl/mathgl/mgl64"
)

type ExamplePlayerHandler struct {
	player.NopHandler
	match *gamelib.Match
}

func NewExamplePlayerHandler(m *gamelib.Match) *ExamplePlayerHandler {
	return &ExamplePlayerHandler{match: m}
}

func (eh ExamplePlayerHandler) HandleMove(ctx *player.Context, newPos mgl64.Vec3, _ cube.Rotation) {
	if eh.match.State() != gamelib.MatchStatePlaying && newPos.Y() < 0 {
		cfg := gamelib.GetConfig[*example_config.ExampleMapData](eh.match.SelectedMap())
		ctx.Val().Teleport(cfg.Mid.Vec3)
	}
}

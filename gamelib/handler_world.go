package gamelib

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// matchWorldHandler wraps the user-provided world.Handler with
// state-based guards.
type matchWorldHandler struct {
	match *Match
}

var _ world.Handler = (*matchWorldHandler)(nil)

func (wh *matchWorldHandler) HandleLiquidFlow(ctx *world.Context, from, into cube.Pos, liquid world.Liquid, replaced world.Block) {
	if wh.match.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	wh.match.worldHandler.HandleLiquidFlow(ctx, from, into, liquid, replaced)
}

func (wh *matchWorldHandler) HandleEntitySpawn(tx *world.Tx, e world.Entity) {
	wh.match.worldHandler.HandleEntitySpawn(tx, e)
}

func (wh *matchWorldHandler) HandleEntityDespawn(tx *world.Tx, e world.Entity) {
	wh.match.worldHandler.HandleEntityDespawn(tx, e)
}

func (wh *matchWorldHandler) HandleLiquidDecay(ctx *world.Context, pos cube.Pos, before, after world.Liquid) {

}

func (wh *matchWorldHandler) HandleLiquidHarden(ctx *world.Context, hardenedPos cube.Pos, liquidHardened, otherLiquid, newBlock world.Block) {

}

func (wh *matchWorldHandler) HandleSound(ctx *world.Context, s world.Sound, pos mgl64.Vec3) {

}

func (wh *matchWorldHandler) HandleFireSpread(ctx *world.Context, from, to cube.Pos) {

}

func (wh *matchWorldHandler) HandleBlockBurn(ctx *world.Context, pos cube.Pos) {

}

func (wh *matchWorldHandler) HandleCropTrample(ctx *world.Context, pos cube.Pos) {

}

func (wh *matchWorldHandler) HandleLeavesDecay(ctx *world.Context, pos cube.Pos) {

}

func (wh *matchWorldHandler) HandleExplosion(ctx *world.Context, position mgl64.Vec3, entities *[]world.Entity, blocks *[]cube.Pos, itemDropChance *float64, spawnFire *bool) {

}

func (wh *matchWorldHandler) HandleRedstoneUpdate(ctx *world.Context, pos cube.Pos) {

}

func (wh *matchWorldHandler) HandleClose(tx *world.Tx) {

}

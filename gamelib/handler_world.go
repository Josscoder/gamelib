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
	if wh.match.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}

	wh.match.worldHandler.HandleLiquidDecay(ctx, pos, before, after)
}

func (wh *matchWorldHandler) HandleLiquidHarden(ctx *world.Context, hardenedPos cube.Pos, liquidHardened, otherLiquid, newBlock world.Block) {
	if wh.match.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}

	wh.match.worldHandler.HandleLiquidHarden(ctx, hardenedPos, liquidHardened, otherLiquid, newBlock)
}

func (wh *matchWorldHandler) HandleSound(ctx *world.Context, s world.Sound, pos mgl64.Vec3) {
	wh.match.worldHandler.HandleSound(ctx, s, pos)
}

func (wh *matchWorldHandler) HandleFireSpread(ctx *world.Context, from, to cube.Pos) {
	if wh.match.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}

	wh.match.worldHandler.HandleFireSpread(ctx, from, to)
}

func (wh *matchWorldHandler) HandleBlockBurn(ctx *world.Context, pos cube.Pos) {
	if wh.match.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}

	wh.match.worldHandler.HandleBlockBurn(ctx, pos)
}

func (wh *matchWorldHandler) HandleCropTrample(ctx *world.Context, pos cube.Pos) {
	if wh.match.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}

	wh.match.worldHandler.HandleCropTrample(ctx, pos)
}

func (wh *matchWorldHandler) HandleLeavesDecay(ctx *world.Context, pos cube.Pos) {
	if wh.match.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}

	wh.match.worldHandler.HandleLeavesDecay(ctx, pos)
}

func (wh *matchWorldHandler) HandleExplosion(ctx *world.Context, position mgl64.Vec3, entities *[]world.Entity, blocks *[]cube.Pos, itemDropChance *float64, spawnFire *bool) {
	if wh.match.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}

	wh.match.worldHandler.HandleExplosion(ctx, position, entities, blocks, itemDropChance, spawnFire)
}

func (wh *matchWorldHandler) HandleRedstoneUpdate(ctx *world.Context, pos cube.Pos) {
	if wh.match.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}

	wh.match.worldHandler.HandleRedstoneUpdate(ctx, pos)
}

func (wh *matchWorldHandler) HandleClose(tx *world.Tx) {
	wh.match.worldHandler.HandleClose(tx)
}

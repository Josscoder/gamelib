package gamelib

import (
	"net"
	"time"

	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

// enginePlayerHandler routes all player events to the correct match handler.
type enginePlayerHandler struct {
	engine *Engine
}

var _ player.Handler = (*enginePlayerHandler)(nil)

// resolve finds the match and its handler for a player.
func (h *enginePlayerHandler) resolve(p *player.Player) (player.Handler, *Match, bool) {
	sess, ok := h.engine.Session(p.XUID())
	if !ok {
		return nil, nil, false
	}
	m := sess.Match()
	if m == nil {
		return nil, nil, false
	}
	return m.playerHandler, m, true
}

func (h *enginePlayerHandler) HandleMove(ctx *player.Context, newPos mgl64.Vec3, newRot cube.Rotation) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleMove(ctx, newPos, newRot)
	}
}

func (h *enginePlayerHandler) HandleHurt(ctx *player.Context, damage *float64, immune bool, attackImmunity *time.Duration, src world.DamageSource) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	// Auto-cancel damage outside of Playing state.
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleHurt(ctx, damage, immune, attackImmunity, src)
}

func (h *enginePlayerHandler) HandleDeath(p *player.Player, src world.DamageSource, keepInv *bool) {
	if ph, _, ok := h.resolve(p); ok {
		ph.HandleDeath(p, src, keepInv)
	}
}

func (h *enginePlayerHandler) HandleBlockBreak(ctx *player.Context, pos cube.Pos, drops *[]item.Stack, xp *int) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleBlockBreak(ctx, pos, drops, xp)
}

func (h *enginePlayerHandler) HandleBlockPlace(ctx *player.Context, pos cube.Pos, b world.Block) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleBlockPlace(ctx, pos, b)
}

func (h *enginePlayerHandler) HandleQuit(p *player.Player) {
	h.engine.HandleQuit(p)
}

func (h *enginePlayerHandler) HandleJump(p *player.Player) {

}

func (h *enginePlayerHandler) HandleTeleport(ctx *player.Context, pos mgl64.Vec3) {

}

func (h *enginePlayerHandler) HandleChangeWorld(p *player.Player, before, after *world.World) {

}

func (h *enginePlayerHandler) HandleToggleSprint(ctx *player.Context, after bool) {

}

func (h *enginePlayerHandler) HandleToggleSneak(ctx *player.Context, after bool) {

}

func (h *enginePlayerHandler) HandleChat(ctx *player.Context, message *string) {

}

func (h *enginePlayerHandler) HandleFoodLoss(ctx *player.Context, from int, to *int) {

}

func (h *enginePlayerHandler) HandleHeal(ctx *player.Context, health *float64, src world.HealingSource) {

}

func (h *enginePlayerHandler) HandleRespawn(p *player.Player, pos *mgl64.Vec3, w **world.World) {

}

func (h *enginePlayerHandler) HandleSkinChange(ctx *player.Context, skin *skin.Skin) {

}

func (h *enginePlayerHandler) HandleFireExtinguish(ctx *player.Context, pos cube.Pos) {

}

func (h *enginePlayerHandler) HandleStartBreak(ctx *player.Context, pos cube.Pos) {

}

func (h *enginePlayerHandler) HandleBlockPick(ctx *player.Context, pos cube.Pos, b world.Block) {

}

func (h *enginePlayerHandler) HandleItemUse(ctx *player.Context) {

}

func (h *enginePlayerHandler) HandleItemUseOnBlock(ctx *player.Context, pos cube.Pos, face cube.Face, clickPos mgl64.Vec3) {

}

func (h *enginePlayerHandler) HandleItemUseOnEntity(ctx *player.Context, e world.Entity) {

}

func (h *enginePlayerHandler) HandleItemRelease(ctx *player.Context, item item.Stack, dur time.Duration) {

}

func (h *enginePlayerHandler) HandleItemConsume(ctx *player.Context, item item.Stack) {

}

func (h *enginePlayerHandler) HandleAttackEntity(ctx *player.Context, e world.Entity, force, height *float64, critical *bool) {

}

func (h *enginePlayerHandler) HandleExperienceGain(ctx *player.Context, amount *int) {

}

func (h *enginePlayerHandler) HandlePunchAir(ctx *player.Context) {

}

func (h *enginePlayerHandler) HandleSignEdit(ctx *player.Context, pos cube.Pos, frontSide bool, oldText, newText string) {

}

func (h *enginePlayerHandler) HandleSleep(ctx *player.Context, sendReminder *bool) {

}

func (h *enginePlayerHandler) HandleLecternPageTurn(ctx *player.Context, pos cube.Pos, oldPage int, newPage *int) {

}

func (h *enginePlayerHandler) HandleItemDamage(ctx *player.Context, i item.Stack, damage *int) {

}

func (h *enginePlayerHandler) HandleItemPickup(ctx *player.Context, i *item.Stack) {

}

func (h *enginePlayerHandler) HandleHeldSlotChange(ctx *player.Context, from, to int) {

}

func (h *enginePlayerHandler) HandleItemDrop(ctx *player.Context, s item.Stack) {

}

func (h *enginePlayerHandler) HandleTransfer(ctx *player.Context, addr *net.UDPAddr) {

}

func (h *enginePlayerHandler) HandleCommandExecution(ctx *player.Context, command cmd.Command, args []string) {

}

func (h *enginePlayerHandler) HandleDiagnostics(p *player.Player, d session.Diagnostics) {

}

package gamelib

import (
	"image/color"
	"net"
	"time"

	"github.com/blockbrawn/gamelib/reflect"
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/df-mc/dragonfly/server/cmd"
	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/player/skin"
	"github.com/df-mc/dragonfly/server/session"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
)

// enginePlayerHandler routes all player events to the correct match handler.
type enginePlayerHandler struct {
	engine *Engine

	lastWorldChangeAnimation time.Time
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
	if ph, _, ok := h.resolve(p); ok {
		ph.HandleJump(p)
	}
}

func (h *enginePlayerHandler) HandleTeleport(ctx *player.Context, pos mgl64.Vec3) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleTeleport(ctx, pos)
	}
}

func (h *enginePlayerHandler) HandleChangeWorld(p *player.Player, before, after *world.World) {
	if time.Since(h.lastWorldChangeAnimation) > time.Second*2 {
		reflect.WritePacket(reflect.Session(p), &packet.CameraInstruction{
			Fade: protocol.Option(protocol.CameraInstructionFade{
				TimeData: protocol.Option(protocol.CameraFadeTimeData{FadeInDuration: 0, WaitDuration: 1, FadeOutDuration: 0.5}),
				Colour:   protocol.Option(color.RGBA{G: 55, B: 128, A: 255}),
			}),
		})

		h.lastWorldChangeAnimation = time.Now()
	}

	if ph, _, ok := h.resolve(p); ok {
		ph.HandleChangeWorld(p, before, after)
	}
}

func (h *enginePlayerHandler) HandleToggleSprint(ctx *player.Context, after bool) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleToggleSprint(ctx, after)
	}
}

func (h *enginePlayerHandler) HandleToggleSneak(ctx *player.Context, after bool) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleToggleSneak(ctx, after)
	}
}

func (h *enginePlayerHandler) HandleChat(ctx *player.Context, message *string) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleChat(ctx, message)
	}
}

func (h *enginePlayerHandler) HandleFoodLoss(ctx *player.Context, from int, to *int) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleFoodLoss(ctx, from, to)
}

func (h *enginePlayerHandler) HandleHeal(ctx *player.Context, health *float64, src world.HealingSource) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleHeal(ctx, health, src)
	}
}

func (h *enginePlayerHandler) HandleRespawn(p *player.Player, pos *mgl64.Vec3, w **world.World) {
	if ph, _, ok := h.resolve(p); ok {
		ph.HandleRespawn(p, pos, w)
	}
}

func (h *enginePlayerHandler) HandleSkinChange(ctx *player.Context, skin *skin.Skin) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleSkinChange(ctx, skin)
	}
}

func (h *enginePlayerHandler) HandleFireExtinguish(ctx *player.Context, pos cube.Pos) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleFireExtinguish(ctx, pos)
	}
}

func (h *enginePlayerHandler) HandleStartBreak(ctx *player.Context, pos cube.Pos) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleStartBreak(ctx, pos)
}

func (h *enginePlayerHandler) HandleBlockPick(ctx *player.Context, pos cube.Pos, b world.Block) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleBlockPick(ctx, pos, b)
}

func (h *enginePlayerHandler) HandleItemUse(ctx *player.Context) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleItemUse(ctx)
	}
}

func (h *enginePlayerHandler) HandleItemUseOnBlock(ctx *player.Context, pos cube.Pos, face cube.Face, clickPos mgl64.Vec3) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleItemUseOnBlock(ctx, pos, face, clickPos)
	}
}

func (h *enginePlayerHandler) HandleItemUseOnEntity(ctx *player.Context, e world.Entity) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleItemUseOnEntity(ctx, e)
	}
}

func (h *enginePlayerHandler) HandleItemRelease(ctx *player.Context, item item.Stack, dur time.Duration) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleItemRelease(ctx, item, dur)
	}
}

func (h *enginePlayerHandler) HandleItemConsume(ctx *player.Context, item item.Stack) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleItemConsume(ctx, item)
	}
}

func (h *enginePlayerHandler) HandleAttackEntity(ctx *player.Context, e world.Entity, force, height *float64, critical *bool) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleAttackEntity(ctx, e, force, height, critical)
}

func (h *enginePlayerHandler) HandleExperienceGain(ctx *player.Context, amount *int) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleExperienceGain(ctx, amount)
}

func (h *enginePlayerHandler) HandlePunchAir(ctx *player.Context) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandlePunchAir(ctx)
	}
}

func (h *enginePlayerHandler) HandleSignEdit(ctx *player.Context, pos cube.Pos, frontSide bool, oldText, newText string) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleSignEdit(ctx, pos, frontSide, oldText, newText)
}

func (h *enginePlayerHandler) HandleSleep(ctx *player.Context, sendReminder *bool) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleSleep(ctx, sendReminder)
}

func (h *enginePlayerHandler) HandleLecternPageTurn(ctx *player.Context, pos cube.Pos, oldPage int, newPage *int) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleLecternPageTurn(ctx, pos, oldPage, newPage)
}

func (h *enginePlayerHandler) HandleItemDamage(ctx *player.Context, i item.Stack, damage *int) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleItemDamage(ctx, i, damage)
	}
}

func (h *enginePlayerHandler) HandleItemPickup(ctx *player.Context, i *item.Stack) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleItemPickup(ctx, i)
}

func (h *enginePlayerHandler) HandleHeldSlotChange(ctx *player.Context, from, to int) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleHeldSlotChange(ctx, from, to)
	}
}

func (h *enginePlayerHandler) HandleItemDrop(ctx *player.Context, s item.Stack) {
	ph, m, ok := h.resolve(ctx.Val())
	if !ok {
		return
	}
	if m.State() != MatchStatePlaying {
		ctx.Cancel()
		return
	}
	ph.HandleItemDrop(ctx, s)
}

func (h *enginePlayerHandler) HandleTransfer(ctx *player.Context, addr *net.UDPAddr) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleTransfer(ctx, addr)
	}
}

func (h *enginePlayerHandler) HandleCommandExecution(ctx *player.Context, command cmd.Command, args []string) {
	if ph, _, ok := h.resolve(ctx.Val()); ok {
		ph.HandleCommandExecution(ctx, command, args)
	}
}

func (h *enginePlayerHandler) HandleDiagnostics(p *player.Player, d session.Diagnostics) {
	if ph, _, ok := h.resolve(p); ok {
		ph.HandleDiagnostics(p, d)
	}
}

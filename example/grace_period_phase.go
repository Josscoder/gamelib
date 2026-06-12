package example

import (
	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/world"
)

type GracePeriodPhase struct {
	remaining int
}

func (p *GracePeriodPhase) Name() string  { return "Grace Period" }
func (p *GracePeriodPhase) Duration() int { return 30 }

func (p *GracePeriodPhase) Start(tx *world.Tx, m *gamelib.Match) {
	p.remaining = p.Duration()
	m.Broadcast(tx, "§a30 second grace period — no PvP!")
	println("grace period started")
}

func (p *GracePeriodPhase) End(tx *world.Tx, m *gamelib.Match) {
	m.Broadcast(tx, "§cGrace period over! PvP is now enabled!")
	println("grace period end")
}

func (p *GracePeriodPhase) Tick(tx *world.Tx, m *gamelib.Match, tick uint64) {
	if tick%20 == 0 {
		p.remaining--
		if p.remaining <= 0 {
			m.NextPhase(tx)
			println("grace period next phase")
		}
	}
}

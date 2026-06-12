package example

import (
	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/world"
)

// CountdownComponent — a countdown before match starts.
type CountdownComponent struct {
	gamelib.BaseComponent
	match    *gamelib.Match
	duration int // seconds
	taskID   gamelib.TaskID
}

func NewCountdownComponent(seconds int) func(m *gamelib.Match) gamelib.Component {
	return func(m *gamelib.Match) gamelib.Component {
		return &CountdownComponent{match: m, duration: seconds}
	}
}

func (c *CountdownComponent) Enable(m *gamelib.Match) {
	c.match = m
}

func (c *CountdownComponent) OnStart(tx *world.Tx) {
	remaining := c.duration
	c.taskID = c.match.Scheduler().EverySecond(func(tx *world.Tx) {
		remaining--
		c.match.Broadcast(tx, "§eGame starts in §c%d§e seconds!", remaining)
		println("inicia en ", remaining)
		if remaining <= 0 {
			c.match.Scheduler().Cancel(c.taskID)
			// Trigger actual game start logic
		}
	})
}

func (c *CountdownComponent) Disable() {
	c.match.Scheduler().Cancel(c.taskID)
}

package gamelib

import (
	"os"

	"github.com/df-mc/dragonfly/server/world"
)

type Arena struct {
	w    *world.World
	path string // on-disk path for cleanup
}

// World returns the Dragonfly world.
func (a *Arena) World() *world.World { return a.w }

// Close shuts down and cleans up the arena.
func (a *Arena) Close() {
	if a.w != nil {
		_ = a.w.Close()
	}
	_ = os.RemoveAll(a.path)
}

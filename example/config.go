package example

import (
	"github.com/df-mc/dragonfly/server/block/cube"
	"github.com/go-gl/mathgl/mgl64"
)

type Location struct {
	mgl64.Vec3
	cube.Rotation
}

type ExampleMapData struct {
	Mid    Location
	Spawns map[string]struct {
		SpawnPoint Location
	}
}

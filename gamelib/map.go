package gamelib

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/mcdb"
	"github.com/google/uuid"
)

type GameMap struct {
	Name      string
	SourceZip string // path to world.zip
	ConfigRaw []byte // raw config bytes (JSON)
}

// UnmarshalConfig deserializes the map config into the given type.
func (gm *GameMap) UnmarshalConfig(v any) error {
	return json.Unmarshal(gm.ConfigRaw, v)
}

// LoadArena unzips the world, opens it, and returns an Arena.
func (gm *GameMap) LoadArena(matchID uuid.UUID) (*Arena, error) {
	destPath := filepath.Join("match_worlds", matchID.String())

	// Create the directory for the unzipped world
	if err := os.MkdirAll(destPath, os.ModePerm); err != nil {
		return nil, fmt.Errorf("creating match world dir: %w", err)
	}

	// Unzip the world.zip into the destination path
	if err := UnZipFile(gm.SourceZip, destPath); err != nil {
		return nil, fmt.Errorf("unzipping world: %w", err)
	}

	prov, err := mcdb.Open(destPath)
	if err != nil {
		return nil, fmt.Errorf("opening world: %w", err)
	}

	cfg := world.Config{
		Dim:       world.Overworld,
		Provider:  prov,
		Generator: world.NopGenerator{},
		ReadOnly:  false,
		// Entities:  entity.DefaultRegistry, (Configure as needed)
	}

	w := cfg.New()
	w.StopTime()
	w.StopWeatherCycle()
	w.StopRaining()
	w.SetDifficulty(world.DifficultyNormal)

	return &Arena{w: w, path: destPath}, nil
}

// MapLoader interface for custom map loading strategies.
type MapLoader interface {
	LoadMaps() ([]*GameMap, error)
}

// DefaultMapLoader reads from a directory structure:
//
//	mapsDir/
//	  mapName/
//	    world.zip
//	    config.json
type DefaultMapLoader struct {
	Dir string
}

func (l *DefaultMapLoader) LoadMaps() ([]*GameMap, error) {
	entries, err := os.ReadDir(l.Dir)
	if err != nil {
		return nil, err
	}
	var maps []*GameMap
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		zipPath := filepath.Join(l.Dir, entry.Name(), "world.zip")
		if stat, err := os.Stat(zipPath); err != nil || stat.IsDir() {
			continue
		}
		configRaw, err := os.ReadFile(filepath.Join(l.Dir, entry.Name(), "config.json"))
		if err != nil {
			continue
		}
		maps = append(maps, &GameMap{
			Name:      entry.Name(),
			SourceZip: zipPath,
			ConfigRaw: configRaw,
		})
	}
	return maps, nil
}

package gamelib

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/df-mc/dragonfly/server/world/mcdb"
	"github.com/google/uuid"
)

type MapConfig interface{}

type GameMap struct {
	Name      string
	SourceZip string // path to world.zip
	ConfigRaw []byte // raw config bytes (JSON)
	mapConfig MapConfig
}

func (gm *GameMap) LoadConfig(v MapConfig) error {
	if err := json.Unmarshal(gm.ConfigRaw, v); err != nil {
		return err
	}
	gm.mapConfig = v
	return nil
}

func GetConfig[C MapConfig](gm *GameMap) C {
	if gm.mapConfig == nil {
		panic("config not loaded, call LoadConfig first")
	}
	return gm.mapConfig.(C)
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
		Dim:          world.Overworld,
		Provider:     prov,
		Generator:    world.NopGenerator{},
		ReadOnly:     false,
		SaveInterval: time.Hour,
		Entities:     entity.DefaultRegistry,
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

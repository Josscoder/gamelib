package gamelib

import (
	"log/slog"
	"sync"

	"github.com/df-mc/dragonfly/server/player"
)

// Engine is the root of the GameLib. One per server.
type Engine struct {
	mu          sync.RWMutex
	definitions map[string]*GameDefinition // keyed by game name
	sessions    *SyncMap[string, *Session] // keyed by XUID
	matchmaker  *Matchmaker
	log         *slog.Logger
}

type EngineConfig struct {
	Log *slog.Logger
}

func (c EngineConfig) New() *Engine {
	e := &Engine{
		definitions: make(map[string]*GameDefinition),
		sessions:    NewSyncMap[string, *Session](),
		log:         c.Log,
	}
	e.matchmaker = newMatchmaker(e)
	return e
}

// Register adds a GameDefinition to the engine.
func (e *Engine) Register(def *GameDefinition) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.definitions[def.Name] = def
}

// GameNames returns the names of all registered games.
func (e *Engine) GameNames() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()
	names := make([]string, 0, len(e.definitions))
	for name := range e.definitions {
		names = append(names, name)
	}
	return names
}

// PlayerHandler returns the global player.Handler to assign
// to every player that connects. This handler routes events
// to the correct match.
func (e *Engine) PlayerHandler() player.Handler {
	return &enginePlayerHandler{engine: e}
}

// HandleJoin creates or retrieves a Session for the player.
func (e *Engine) HandleJoin(p *player.Player) *Session {
	sess := &Session{
		xuid:     p.XUID(),
		name:     p.Name(),
		handle:   p.H(),
		metadata: make(map[string]any),
	}
	e.sessions.Store(p.XUID(), sess)
	return sess
}

// HandleQuit cleans up the session and removes from match.
func (e *Engine) HandleQuit(p *player.Player) {
	sess, ok := e.sessions.Load(p.XUID())
	if !ok {
		return
	}
	if m := sess.Match(); m != nil {
		m.leave(p, true)
	}
	sess.close()
	e.sessions.Delete(p.XUID())
}

// Session returns the session for the given XUID.
func (e *Engine) Session(xuid string) (*Session, bool) {
	return e.sessions.Load(xuid)
}

// Matchmaker returns the engine's matchmaker.
func (e *Engine) Matchmaker() *Matchmaker {
	return e.matchmaker
}

package example

import (
	"time"

	"github.com/blockbrawn/gamelib/gamelib"
	"github.com/df-mc/dragonfly/server/world"
)

type PreGameState struct {
	*gamelib.MatchStateBase
}

func NewPreGameState(m *gamelib.Match) *PreGameState {
	s := &PreGameState{}
	s.MatchStateBase = gamelib.NewMatchStateBase(m, s)
	return s
}

func (s *PreGameState) OnStart() {
	match := s.Match()
	match.World().Exec(func(tx *world.Tx) {
		match.Broadcast(tx, "§eEsperando jugadores... (%d/%d)",
			match.ParticipantCount(),
			match.Definition().MaxPlayers,
		)
	})
}

func (s *PreGameState) OnUpdate(delta time.Duration) {
	match := s.Match()
	remaining := s.GetRemainingTime().Round(time.Second)
	match.World().Exec(func(tx *world.Tx) {
		match.Broadcast(tx, "§eComienza en §f%v §7(δ %v) §e| Jugadores: §f%d/%d",
			remaining,
			delta.Round(time.Millisecond),
			match.ParticipantCount(),
			match.Definition().MaxPlayers,
		)
	})
}

func (s *PreGameState) OnEnd() {
	s.Match().World().Exec(func(tx *world.Tx) {
		s.Match().Broadcast(tx, "§a¡El juego comienza!")
	})
}

func (s *PreGameState) GetDuration() time.Duration {
	return 30 * time.Second
}

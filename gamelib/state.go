package gamelib

import (
	"github.com/josscoder/fsmgo/state"
)

// MatchStateBase is a BaseState that carries a reference to the Match,
// so every concrete state has direct access to participants, world, metadata, etc.
type MatchStateBase struct {
	*state.BaseState
	match *Match
}

func NewMatchStateBase(m *Match, lifecycle state.Lifecycle) *MatchStateBase {
	ms := &MatchStateBase{match: m}
	ms.BaseState = state.NewBaseState(lifecycle)
	return ms
}

// Match returns the associated match.
func (ms *MatchStateBase) Match() *Match { return ms.match }

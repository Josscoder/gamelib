package gamelib

import (
	"errors"
	"iter"
	"sync/atomic"

	"github.com/df-mc/dragonfly/server/item"
	"github.com/df-mc/dragonfly/server/player"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/sandertv/gophertunnel/minecraft/text"
)

type TeamColor int

const (
	Black TeamColor = iota
	DarkBlue
	DarkGreen
	DarkAqua
	DarkRed
	DarkPurple
	Gold
	Grey
	DarkGrey
	Blue
	Green
	Aqua
	Red
	Purple
	Yellow
	White
	DarkYellow
	Quartz
	Iron
	Netherite
	Redstone
	Copper
	Emerald
	Diamond
	Lapis
	Amethyst
)

// teamColorTextTags maps each TeamColor to its gophertunnel text tag name.
// Indexed by TeamColor constant — O(1) lookup, no switch needed.
var teamColorTextTags = [...]string{
	Black:      "black",
	DarkBlue:   "dark_blue",
	DarkGreen:  "dark_green",
	DarkAqua:   "dark_aqua",
	DarkRed:    "dark_red",
	DarkPurple: "dark_purple",
	Gold:       "gold",
	Grey:       "gray",
	DarkGrey:   "dark_gray",
	Blue:       "blue",
	Green:      "green",
	Aqua:       "aqua",
	Red:        "red",
	Purple:     "light_purple",
	Yellow:     "yellow",
	White:      "white",
	DarkYellow: "dark_yellow",
	Quartz:     "quartz",
	Iron:       "iron",
	Netherite:  "netherite",
	Redstone:   "redstone",
	Copper:     "copper",
	Emerald:    "emerald",
	Diamond:    "diamond",
	Lapis:      "lapis",
	Amethyst:   "amethyst",
}

// teamColorItemColours maps each TeamColor to its item.Colour equivalent.
var teamColorItemColours = [...]item.Colour{
	Black:      item.ColourBlack(),
	DarkBlue:   item.ColourBlue(),
	DarkGreen:  item.ColourGreen(),
	DarkAqua:   item.ColourCyan(),
	DarkRed:    item.ColourRed(),
	DarkPurple: item.ColourPurple(),
	Gold:       item.ColourOrange(),
	Grey:       item.ColourLightGrey(),
	DarkGrey:   item.ColourGrey(),
	Blue:       item.ColourBlue(),
	Green:      item.ColourLime(),
	Aqua:       item.ColourCyan(),
	Red:        item.ColourRed(),
	Purple:     item.ColourMagenta(),
	Yellow:     item.ColourYellow(),
	White:      item.ColourWhite(),
	DarkYellow: item.ColourYellow(),
	Quartz:     item.ColourLightGrey(),
	Iron:       item.ColourGrey(),
	Netherite:  item.ColourBrown(),
	Redstone:   item.ColourRed(),
	Copper:     item.ColourOrange(),
	Emerald:    item.ColourLime(),
	Diamond:    item.ColourLightBlue(),
	Lapis:      item.ColourBlue(),
	Amethyst:   item.ColourPurple(),
}

// AsTextColour wraps text in the color's tag.
func (tc TeamColor) AsTextColour(t string) string {
	if int(tc) >= len(teamColorTextTags) {
		return t
	}
	tag := teamColorTextTags[tc]
	return text.Colourf("<%s>%s</%s>", tag, t, tag)
}

// AsItemColour returns the closest item.Colour for this TeamColor.
func (tc TeamColor) AsItemColour() item.Colour {
	if int(tc) >= len(teamColorItemColours) {
		return item.ColourWhite()
	}
	return teamColorItemColours[tc]
}

// Team represents a group of participants working together in a match.
type Team struct {
	ID          string
	Name        string
	Color       TeamColor
	MaxPlayers  int
	members     *SyncMap[string, *Participant] // keyed by XUID
	memberCount atomic.Int32                   // O(1) Size() / IsFull()
}

// NewTeam creates a new team.
func NewTeam(id, name string, color TeamColor, max int) *Team {
	return &Team{
		ID:         id,
		Name:       name,
		Color:      color,
		MaxPlayers: max,
		members:    NewSyncMap[string, *Participant](),
	}
}

// Members returns an iterator over all participants in the team.
func (t *Team) Members() iter.Seq[*Participant] {
	return func(yield func(*Participant) bool) {
		for _, par := range t.members.Map() {
			if !yield(par) {
				return
			}
		}
	}
}

// Size returns the current number of members in O(1).
func (t *Team) Size() int { return int(t.memberCount.Load()) }

// IsFull returns true when the team has reached MaxPlayers.
func (t *Team) IsFull() bool { return t.Size() >= t.MaxPlayers }

// AliveCount returns how many members are still alive.
func (t *Team) AliveCount() int {
	n := 0
	for _, p := range t.members.Map() {
		if p.Alive() {
			n++
		}
	}
	return n
}

// IsEliminated returns true when all members are dead or spectating.
func (t *Team) IsEliminated() bool { return t.AliveCount() == 0 }

// Players calls fn for every team member resolvable in the transaction.
func (t *Team) Players(tx *world.Tx, fn func(*player.Player, *Participant)) {
	for _, par := range t.members.Map() {
		p, ok := par.Player(tx)
		if !ok {
			continue
		}
		fn(p, par)
	}
}

// --- Match integration ---

// RegisterTeam adds a team to the match.
func (m *Match) RegisterTeam(t *Team) {
	m.teams.Store(t.ID, t)
}

// Teams returns a snapshot of all teams in the match.
func (m *Match) Teams() []*Team {
	all := m.teams.Map()
	out := make([]*Team, 0, len(all))
	for _, t := range all {
		out = append(out, t)
	}
	return out
}

// Team returns a team by ID.
func (m *Match) Team(id string) (*Team, bool) {
	return m.teams.Load(id)
}

// AssignTeam puts a participant into a team.
// Returns an error if the team is full.
func (m *Match) AssignTeam(par *Participant, t *Team) error {
	if t.IsFull() {
		return errors.New("team is full")
	}
	t.members.Store(par.XUID(), par)
	t.memberCount.Add(1)
	par.SetData("team_id", t.ID)
	return nil
}

// RemoveFromTeam removes a participant from their current team.
func (m *Match) RemoveFromTeam(par *Participant) {
	t := m.TeamOf(par)
	if t == nil {
		return
	}
	t.members.Delete(par.XUID())
	t.memberCount.Add(-1)
}

// TeamOf returns the team a participant belongs to, or nil for FFA.
func (m *Match) TeamOf(par *Participant) *Team {
	teamID, ok := DataAs[string](par, "team_id")
	if !ok {
		return nil
	}
	t, _ := m.teams.Load(teamID)
	return t
}

// AreEnemies returns true if p1 and p2 are on different teams,
// or if either has no team (FFA mode).
func (m *Match) AreEnemies(p1, p2 *Participant) bool {
	t1 := m.TeamOf(p1)
	t2 := m.TeamOf(p2)
	if t1 == nil || t2 == nil {
		return true
	}
	return t1.ID != t2.ID
}

// AutoAssignTeam places a participant into the team with the fewest members.
// Returns an error if all teams are full.
func (m *Match) AutoAssignTeam(par *Participant) error {
	var best *Team
	for _, t := range m.teams.Map() {
		if t.IsFull() {
			continue
		}
		if best == nil || t.Size() < best.Size() {
			best = t
		}
	}
	if best == nil {
		return errors.New("all teams are full")
	}
	return m.AssignTeam(par, best)
}

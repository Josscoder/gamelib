package gamelib

import (
	"sync/atomic"
	"time"

	"github.com/df-mc/dragonfly/server/world"
)

type TaskID uint64

type scheduledTask struct {
	id       TaskID
	fn       func(tx *world.Tx)
	interval uint64 // in ticks (0 = one-shot)
	delay    uint64 // ticks until first execution
	repeat   bool
	cancel   atomic.Bool
}

type Scheduler struct {
	match   *Match
	tasks   *SyncMap[TaskID, *scheduledTask]
	nextID  atomic.Uint64
	running atomic.Bool
}

func newScheduler(m *Match) *Scheduler {
	return &Scheduler{
		match: m,
		tasks: NewSyncMap[TaskID, *scheduledTask](),
	}
}

func (s *Scheduler) Start() {
	ticker := time.NewTicker(time.Second / 20) // 20 TPS
	go func() {
		defer ticker.Stop()

		for range ticker.C {
			s.match.World().Exec(func(tx *world.Tx) {
				s.tick(tx)
			})
		}
	}()
}

// EveryTick runs fn every tick (50ms).
func (s *Scheduler) EveryTick(fn func(tx *world.Tx)) TaskID {
	return s.schedule(fn, 1, true)
}

// EverySecond runs fn every 20 ticks.
func (s *Scheduler) EverySecond(fn func(tx *world.Tx)) TaskID {
	return s.schedule(fn, 20, true)
}

// Every runs fn every N ticks.
func (s *Scheduler) Every(ticks uint64, fn func(tx *world.Tx)) TaskID {
	return s.schedule(fn, ticks, true)
}

// Delay runs fn once after N ticks.
func (s *Scheduler) Delay(ticks uint64, fn func(tx *world.Tx)) TaskID {
	return s.scheduleOnce(fn, ticks)
}

// DelaySeconds runs fn once after N seconds.
func (s *Scheduler) DelaySeconds(seconds int, fn func(tx *world.Tx)) TaskID {
	return s.scheduleOnce(fn, uint64(seconds)*20)
}

// Cancel cancels a scheduled task.
func (s *Scheduler) Cancel(id TaskID) {
	if t, ok := s.tasks.Load(id); ok {
		t.cancel.Store(true)
		s.tasks.Delete(id)
	}
}

func (s *Scheduler) schedule(fn func(tx *world.Tx), interval uint64, repeat bool) TaskID {
	id := TaskID(s.nextID.Add(1))
	s.tasks.Store(id, &scheduledTask{
		id:       id,
		fn:       fn,
		interval: interval,
		delay:    interval,
		repeat:   repeat,
	})
	return id
}

func (s *Scheduler) scheduleOnce(fn func(tx *world.Tx), delay uint64) TaskID {
	id := TaskID(s.nextID.Add(1))
	s.tasks.Store(id, &scheduledTask{
		id:       id,
		fn:       fn,
		interval: 0,
		delay:    delay,
		repeat:   false,
	})
	return id
}

// tick is called from the match's tick loop.
func (s *Scheduler) tick(tx *world.Tx) {
	for _, t := range s.tasks.Map() {
		if t.cancel.Load() {
			s.tasks.Delete(t.id)
			continue
		}
		t.delay--
		if t.delay <= 0 {
			t.fn(tx)
			if t.repeat {
				t.delay = t.interval
			} else {
				s.tasks.Delete(t.id)
			}
		}
	}
}

func (s *Scheduler) stopAll() {
	for _, t := range s.tasks.Map() {
		t.cancel.Store(true)
	}
	s.tasks = NewSyncMap[TaskID, *scheduledTask]()
}

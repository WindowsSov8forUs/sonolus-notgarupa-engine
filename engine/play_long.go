//go:build play

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus/play"
)

// PlayLongTail is the end of a Long; its judgement is decided by the head.
type PlayLongTail struct {
	PlayNoteBase   `archetype:"base"`
	play.Archetype `archetype:"name=LongTail,hasInput=true"`
	Head           sonolus.EntityRef[PlayLongHead] `archetype:"imported,name=head"`
	Kind           float64                         `archetype:"imported,name=kind"`
	IconStart      float64                         `archetype:"memory"`
}

func (n *PlayLongTail) Preprocess() {
	n.prepare()
	// the tail lives as long as its head
	n.SpawnTime = math.Min(n.SpawnTime, n.Head.Get().SpawnTime)
	scheduleAuto(kindClip(n.Kind, 1), n.Time, n.Skill)
}
func (n *PlayLongTail) SpawnOrder() float64 { return n.SpawnTime }
func (n *PlayLongTail) ShouldSpawn() bool   { return play.Time.Now() >= n.SpawnTime }
func (n *PlayLongTail) Initialize() {
	n.IconStart = play.Time.Now()
}
func (n *PlayLongTail) UpdateParallel() {
	if n.State == stDone {
		writeResult(n.Grade, n.Error, Buckets.LongTail)
		play.Entity.SetDespawn(true)
		return
	}
	now := play.Time.Now()
	pr := n.progress(now)
	if pr < 0 {
		return
	}
	drawTerminal(n.look(), n.Kind, n.Width, n.Lane, curveOf(pr), now-n.IconStart, noteZ(zNote, n.Head.Get().Time, n.Head.Get().Lane))
}

// PlayLongHead owns the Long: hold state, body, flash, loop SE and TapKeep.
type PlayLongHead struct {
	PlayNoteBase   `archetype:"base"`
	play.Archetype `archetype:"name=LongHead,hasInput=true"`
	Tail           sonolus.EntityRef[PlayLongTail] `archetype:"imported,name=tail"`
	TailTime       float64                         `archetype:"data"`
	HoldStart      float64                         `archetype:"shared"`
	Grace          float64                         `archetype:"shared"`
	OX             float64                         `archetype:"shared"`
	OY             float64                         `archetype:"shared"`
	TailDone       bool                            `archetype:"shared"`
	Loop           sonolus.LoopedEffectHandle      `archetype:"memory"`
	Keep           sonolus.ParticleHandle          `archetype:"memory"`
	KeepPlane      sonolus.ParticleHandle          `archetype:"memory"`
}

func (n *PlayLongHead) Preprocess() {
	n.prepare()
	n.TailTime = play.Time.BeatToTime(n.Tail.Get().Beat)
	scheduleAuto(clipPerfect, n.Time, n.Skill)
	autoHold(&n.PlayNoteBase, n.Time, n.TailTime)
}
func (n *PlayLongHead) SpawnOrder() float64 { return n.SpawnTime }
func (n *PlayLongHead) ShouldSpawn() bool   { return play.Time.Now() >= n.SpawnTime }

func (n *PlayLongHead) tailNote() *PlayLongTail { return n.Tail.Get() }

func (n *PlayLongHead) UpdateSequential() {
	if n.TailDone {
		return
	}
	now := play.Time.Now()
	switch n.State {
	case stMove, stWait:
		// Wait begins once the head has crossed the line (progress > 1, the
		// visual arrival under SV); only a waiting head times out
		if n.State == stMove && n.progress(now) > 1 {
			n.State = stWait
		}
		if n.State == stWait && now-n.Time > missSeconds {
			// execOverWaitState: two Miss entries (head and tail), half damage each
			n.Grade, n.Error = gradeMiss, missSeconds
			lane := receiverLane(n.SpanStart, n.Width)
			j := judgement{Width: n.Width, Grade: gradeMiss, Type: jtLongMiss, X: laneX(lane), Lane: lane, Key: n.Time, Divisor: 2}
			applyJudgement(now, j)
			applyJudgement(now, j)
			n.finishTail(gradeMiss, missSeconds)
			return
		}
		if math.Abs(now-n.Time) <= 9*frameSeconds {
			y := notePos(n.Lane, curveOf(n.progress(now))).Y
			registerCandidate(play.Entity.Info().Index, n.SpanStart, n.Width, n.Time, y, false, n.Finger >= 0)
		}
	case stHold:
		// a held Long still competes (by its head time) and consumes the touch
		if math.Abs(now-n.Time) <= 9*frameSeconds {
			registerCandidate(play.Entity.Info().Index, n.SpanStart, n.Width, n.Time, Layout.G, false, true)
		}
		if now-n.TailTime > missSeconds {
			n.endHold(now, gradeMiss, timingNone, missSeconds, jtLongMiss)
		}
	}
}

func (n *PlayLongHead) Touch() {
	if n.TailDone {
		return
	}
	now := play.Time.Now()
	if n.Assigned >= 0 && (n.State == stMove || n.State == stWait) {
		touchID := n.Assigned
		n.Assigned = -1
		if n.AGrade <= gradeMiss {
			return
		}
		n.Finger = touchID
		n.State = stHold
		n.HoldStart = now
		n.Grade, n.Error = n.AGrade, n.ATime-n.Time
		n.OX, n.OY = n.AX, n.AY
		n.Grace = leaveGrace
		lane := receiverLane(n.SpanStart, n.Width)
		if Config.SFX && !Config.AutoSFX {
			n.Loop = startHoldLoop()
		}
		applyJudgement(now, judgement{Width: n.Width, Grade: n.AGrade, Type: jtLongFront, Timing: n.ATiming, Skill: n.Skill,
			X: laneX(lane), Lane: lane, Key: n.Time, Divisor: 1})
		if Config.NoteEffects {
			// a Long hold uses the TapKeep of its width (HABAHIRO)
			w := int(sonolus.Clamp(n.Width, 1, 7))
			n.Keep = pickW(Particles.Keep, Particles.KeepW, w).Spawn(particleQuad(laneX(lane), 1), durKeep, true)
			n.KeepPlane = pickW(Particles.KeepPlane, Particles.KeepPlaneW, w).Spawn(particleQuad(laneX(lane), planeCos48), durKeepPlane, true)
		}
	}
	if n.State != stHold {
		return
	}
	tail := n.tailNote()
	for touch := range play.Touches.Values() {
		// a touch that began this frame is only processed when it also ended
		if touch.ID != n.Finger || (touch.Started && !touch.Ended) {
			continue
		}
		p := touch.Position
		if touch.Ended {
			lane := receiverLane(n.SpanStart, n.Width)
			laneOnReserve(lane, now)
			grade := gradeMiss
			timing := timingNone
			jt := jtLongEnd
			if tail.Kind == NodePlain {
				if insideSpan(p, n.SpanStart, n.Width, false) {
					grade = timeGrade(n.TailTime-touch.Time, 1)
					timing = timeTiming(grade, n.TailTime-touch.Time)
				}
				if grade == gradeNone {
					grade = gradeMiss
				}
			}
			if grade == gradeMiss {
				jt = jtLongMiss
			}
			n.endHold(now, grade, timing, touch.Time-n.TailTime, jt)
			return
		}
		if tail.Kind == NodePlain {
			continue // the finger position is not tracked during a plain hold
		}
		r := timeGrade(n.TailTime-now, 1)
		if r == gradeNone {
			n.OX, n.OY = p.X, p.Y
		}
		if insideSpan(p, n.SpanStart, n.Width, false) {
			n.Grace = leaveGrace
		} else {
			n.Grace -= frameStep()
		}
		dir, count := 0.0, 1.0
		jt := jtLongEndFlick
		if tail.Kind == NodeDirectional {
			dir, jt = tail.Direction, jtLongEndDir
		}
		if gestureDone(p.X-n.OX, p.Y-n.OY, dir, count) && r != gradeNone && n.Grace > 0 {
			n.endHold(now, r, timeTiming(r, n.TailTime-now), now-n.TailTime, jt)
			return
		}
	}
}

// endHold judges the tail and releases the hold feedback.
func (n *PlayLongHead) endHold(now, grade, timing, err, jt float64) {
	tail := n.tailNote()
	lane := receiverLane(tail.SpanStart, tail.Width)
	applyJudgement(now, judgement{Width: tail.Width, Grade: grade, Type: jt, Timing: timing, Skill: tail.Skill,
		X: laneX(lane), Lane: lane, Key: n.TailTime, Direction: tail.Direction, Count: 1, Divisor: 1})
	n.finishTail(grade, err)
	n.stopFeedback()
}

func (n *PlayLongHead) stopFeedback() {
	if n.Loop.ID > 0 {
		releaseHoldLoop(n.Loop)
		n.Loop.ID = 0
	}
	if n.Keep.ID > 0 {
		n.Keep.Destroy()
		n.KeepPlane.Destroy()
		n.Keep.ID, n.KeepPlane.ID = 0, 0
	}
}

func (n *PlayLongHead) finishTail(grade, err float64) {
	tail := n.tailNote()
	tail.State = stDone
	tail.Grade, tail.Error = grade, err
	n.TailDone = true
	n.State = stDone
}

func (n *PlayLongHead) UpdateParallel() {
	if n.TailDone {
		writeResult(n.Grade, n.Error, Buckets.LongHead)
		play.Entity.SetDespawn(true)
		return
	}
	now := play.Time.Now()
	tail := n.tailNote()
	held := n.State == stHold
	var c float64
	if held {
		c = math.Min(curveOf(n.progress(now)), 1)
	} else {
		pr := n.progress(now)
		if pr < 0 {
			return
		}
		c = curveOf(pr)
		if n.State == stWait {
			c = math.Min(c, 1) // snapped onto the line while waiting
		}
	}
	tc := 0.0
	if tpr := tail.progress(now); tpr >= 0 {
		tc = math.Min(curveOf(tpr), 1)
	}
	drawLong(n.look(), c, tail.Lane, tail.Width, tc, held, now-n.HoldStart, n.z(), lineZ(zBelt, n.Time))
}

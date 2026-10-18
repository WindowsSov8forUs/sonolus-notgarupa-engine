//go:build play

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus/play"
)

// ---------------------------------------------------------------- timing ----

type PlayBPMChange struct {
	play.Archetype `archetype:"name=#BPM_CHANGE"`
	Beat           float64 `archetype:"imported,name=#BEAT"`
	BPM            float64 `archetype:"imported,name=#BPM"`
}

func (*PlayBPMChange) SpawnOrder() float64 { return 1e8 }
func (*PlayBPMChange) ShouldSpawn() bool   { return false }

// PlayTimingChange is one step of a timing group's scroll speed (SV).
type PlayTimingChange struct {
	play.Archetype `archetype:"name=TimingChange"`
	Time           float64 `archetype:"imported,name=time"`
	Speed          float64 `archetype:"imported,name=speed"`
	Axis           float64 `archetype:"imported,name=axis"`
	Max            float64 `archetype:"imported,name=max"`
}

func (*PlayTimingChange) SpawnOrder() float64 { return 1e8 }
func (*PlayTimingChange) ShouldSpawn() bool   { return false }

// PlayTimingGroup keeps C(now) of its scroll axis up to date each frame.
type PlayTimingGroup struct {
	play.Archetype      `archetype:"name=TimingGroup"`
	play.CallbackOrders `archetype:"updateSequential=-50"`
	First               sonolus.EntityRef[PlayTimingChange] `archetype:"imported,name=first"`
	Count               float64                             `archetype:"imported,name=count"`
	Cursor              float64                             `archetype:"shared"`
	AxisNow             float64                             `archetype:"shared"`
}

func (*PlayTimingGroup) SpawnOrder() float64 { return -1e8 }
func (*PlayTimingGroup) ShouldSpawn() bool   { return true }

func (g *PlayTimingGroup) change(k float64) *PlayTimingChange {
	return sonolus.EntityRef[PlayTimingChange]{Index: g.First.Index + k}.Get()
}

func (g *PlayTimingGroup) UpdateSequential() {
	now := play.Time.Now()
	c := g.Cursor
	for c > 0 && now < g.change(c).Time {
		c--
	}
	for c+1 < g.Count && g.change(c+1).Time <= now {
		c++
	}
	g.Cursor = c
	first := g.change(0)
	if now < first.Time {
		g.AxisNow = now - first.Time + first.Axis
	} else {
		ch := g.change(c)
		g.AxisNow = ch.Axis + ch.Speed*(now-ch.Time)
	}
}

// visualSpawn is the first time the axis brings a note at axisHit within
// one arrival of the judge line: Cmax(t) ≥ axisHit − arrival.
func groupVisualSpawn(group sonolus.EntityRef[PlayTimingGroup], axisHit float64) float64 {
	g := group.Get()
	x := axisHit - Layout.Arrival
	first := g.change(0)
	if x <= first.Axis {
		return x - first.Axis + first.Time
	}
	lo, hi := 1.0, g.Count
	for lo < hi {
		mid := math.Floor((lo + hi) / 2)
		if g.change(mid).Max >= x {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	if lo < g.Count {
		p := g.change(lo - 1)
		return p.Time + (x-p.Axis)/p.Speed
	}
	last := g.change(g.Count - 1)
	if last.Speed > 0 {
		return last.Time + (x-last.Axis)/last.Speed
	}
	return 1e8
}

// ----------------------------------------------------------------- notes ----

// Note states.
const (
	stMove    = 0
	stWait    = 1 // touched flick waiting for the gesture
	stHold    = 2 // Long / Slide held
	stDone    = 3
	stLongEnd = 4
)

// PlayNoteBase holds what every note shares: geometry, scroll axis and the
// input hand-over written by the InputManager.
type PlayNoteBase struct {
	play.Archetype `archetype:"abstract"`
	Beat           float64                            `archetype:"imported,name=#BEAT"`
	Lane           float64                            `archetype:"imported,name=lane"`
	SpriteLane     float64                            `archetype:"imported,name=spriteLane"`
	Width          float64                            `archetype:"imported,name=width"`
	SpanStart      float64                            `archetype:"imported,name=spanStart"`
	Group          sonolus.EntityRef[PlayTimingGroup] `archetype:"imported,name=group"`
	AxisHit        float64                            `archetype:"imported,name=axisHit"`
	Skill          bool                               `archetype:"imported,name=skill"`
	Direction      float64                            `archetype:"imported,name=direction"`
	ScoreNext      sonolus.EntityRef[PlayNoteBase]    `archetype:"imported,name=scoreNext"`
	Time           float64                            `archetype:"data"`
	SpawnTime      float64                            `archetype:"data"`

	HoldUntil float64 `archetype:"shared"` // #EFFECT_AUTO hold loop starting here ends then
	State     float64 `archetype:"shared"`
	Finger    float64 `archetype:"shared"` // bound touch id, −1 when free
	Assigned  float64 `archetype:"shared"` // touch id handed over this frame, −1
	AGrade    float64 `archetype:"shared"`
	ATiming   float64 `archetype:"shared"`
	ATime     float64 `archetype:"shared"`
	AX        float64 `archetype:"shared"`
	AY        float64 `archetype:"shared"`
	Frames    float64 `archetype:"shared"`
	Grade     float64 `archetype:"shared"`
	Error     float64 `archetype:"shared"`
}

func (n *PlayNoteBase) prepare() {
	if Config.Mirror {
		n.Lane = 6 - n.Lane
		n.SpanStart = 6 - (n.SpanStart + n.Width - 1)
		n.SpriteLane = 6 - n.SpriteLane
		n.Direction = -n.Direction
	}
	n.Time = play.Time.BeatToTime(n.Beat)
	visual := n.Time - Layout.Arrival
	if n.Group.Index > 0 {
		visual = groupVisualSpawn(n.Group, n.AxisHit)
	}
	n.SpawnTime = math.Min(visual, n.Time-0.25)
	n.Finger = -1
	n.Assigned = -1
	n.State = stMove
	n.Frames = 0
	n.Grade = gradeNone
}

// progress returns the approach progress at now (1 = on the judge line).
func (n *PlayNoteBase) progress(now float64) float64 {
	if n.Group.Index > 0 {
		return 1 - (n.AxisHit-n.Group.Get().AxisNow)/Layout.Arrival
	}
	return 1 - (n.Time-now)/Layout.Arrival
}

// z orders this note's sprites by its own time (a Long / Slide child passes
// its head's time to noteZ instead, as the original does).
func (n *PlayNoteBase) z() sonolus.ZIndex { return noteZ(zNote, n.Time, n.Lane) }

// containsLane reports whether an in-field note covers integer lane l.
func (n *PlayNoteBase) containsLane(l float64) bool {
	return n.SpanStart == math.Floor(n.SpanStart) && l >= n.SpanStart && l <= n.SpanStart+n.Width-1
}

// insideSpan is the original IsInsideTargetNoteButtons: a touch is inside if
// it is within the collision circle of any member lane's judge-line target.
func insideSpan(p sonolus.Vec2, spanStart, width float64, sweet bool) bool {
	r2 := Layout.Collision2
	if sweet {
		r2 *= 1.0800000429153442
	}
	dy := p.Y - Layout.G
	for k := 0.0; k < width && k < 16; k++ {
		dx := p.X - laneX(spanStart+k)
		if dx*dx+dy*dy < r2 {
			return true
		}
	}
	return false
}

// writeResult stores the Sonolus result of a judged note.
func writeResult(grade, errorSeconds float64, bucket sonolus.Bucket) {
	e := sonolus.Clamp(errorSeconds, -0.15, 0.15)
	play.Entity.SetResult(play.InputResult{
		Judgment: sonolusJudgment(grade), Accuracy: e, Bucket: bucket, BucketValue: e * 1000,
	})
}

// frameStep is the original ExecuteFrame increment (capped at 1 per update).
func frameStep() float64 { return math.Min(play.Time.Delta()*60, 1) }

// candidate registers a note for this frame's touch arbitration.
func registerCandidate(index, start, width, t, y float64, slide, owned bool) {
	if Input.Candidates.IsFull() {
		return
	}
	Input.Candidates.Append(candidate{Index: index, Start: start, Width: width, Time: t, Y: y, Slide: slide, Owned: owned})
}

// look returns the drawing attributes of the note.
func (n *PlayNoteBase) look() noteLook {
	return noteLook{SpriteLane: n.SpriteLane, Lane: n.Lane, Width: n.Width, Direction: n.Direction, SpanStart: n.SpanStart}
}

// applyJudgement reflects one judgement: HUD state and life gauge, then the
// particle and SE in the original order, then lane lights.
func applyJudgement(now float64, j judgement) {
	hudApply(now, j)
	if Config.NoteEffects {
		playJudgementParticles(j)
	}
	clip := judgementClip(now, j)
	if Config.SFX && !Config.AutoSFX {
		if clip != clipNone {
			clipOf(clip).Play(0)
		}
	}
	judgementLights(now, j)
}

// autoSFX reports whether SE follow the chart (#EFFECT_AUTO) instead of the
// judgements.
func autoSFX() bool { return Config.SFX && Config.AutoSFX }

// scheduleAuto schedules a note's Perfect SE at its chart time (#EFFECT_AUTO);
// the minimum distance merges simultaneous notes like the original's
// same-position silence.
func scheduleAuto(code, t float64, skill bool) {
	if !autoSFX() {
		return
	}
	if skill {
		code = clipSkill // the skill SE replaces the cue
	}
	clipOf(code).PlayScheduled(t, frameSeconds)
}

// startHoldLoop starts a hold loop, stopping the oldest live one when the
// cue is full (holdLoopLimit).
func startHoldLoop() sonolus.LoopedEffectHandle {
	slot := 0
	for i := 1; i < holdLoopLimit; i++ {
		if Input.Loops[slot] > 0 && (Input.Loops[i] <= 0 || Input.LoopOrder[i] < Input.LoopOrder[slot]) {
			slot = i
		}
	}
	if Input.Loops[slot] > 0 {
		sonolus.LoopedEffectHandle{ID: Input.Loops[slot]}.Stop()
	}
	loop := Effects.HoldLoop.PlayLooped()
	Input.LoopSeq++
	Input.Loops[slot], Input.LoopOrder[slot] = loop.ID, Input.LoopSeq
	return loop
}

// releaseHoldLoop stops a hold loop and plays its release, unless the cue
// limit already stopped it (its owner's later commands are then no-ops).
func releaseHoldLoop(loop sonolus.LoopedEffectHandle) {
	for i := 0; i < holdLoopLimit; i++ {
		if Input.Loops[i] == loop.ID {
			Input.Loops[i] = 0
			loop.Stop()
			Effects.HoldRelease.Play(0)
		}
	}
}

// playChained: the level has the scoring chain (firstScored); levels from
// the old converter do not.
func playChained() bool {
	return sonolus.EntityRef[PlayInitialization]{Index: 0}.Get().FirstScored.Index > 0
}

// autoHold records an #EFFECT_AUTO hold loop for PlayStage to schedule under
// the cue limit; without a scoring chain it schedules the loop directly.
func autoHold(n *PlayNoteBase, start, end float64) {
	if !autoSFX() {
		return
	}
	if playChained() {
		n.HoldUntil = end
		return
	}
	loop := Effects.HoldLoop.PlayLoopedScheduled(start)
	loop.Stop(end)
	Effects.HoldRelease.PlayScheduled(end, 0)
}

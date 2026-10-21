//go:build play

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus/play"
)

// ------------------------------------------------------------------ tap ----

// PlayTapNote is a Single or Skill note.
type PlayTapNote struct {
	PlayNoteBase   `archetype:"base"`
	play.Archetype `archetype:"name=TapNote,hasInput=true"`
	Offbeat        bool `archetype:"imported,name=offbeat"`
}

func (n *PlayTapNote) Preprocess() {
	n.prepare()
	n.Single = true
	scheduleAuto(clipPerfect, n.Time, n.Skill)
}
func (n *PlayTapNote) SpawnOrder() float64 { return n.SpawnTime }
func (n *PlayTapNote) ShouldSpawn() bool   { return play.Time.Now() >= n.SpawnTime }
func (n *PlayTapNote) UpdateSequential()   { singleUpdate(&n.PlayNoteBase, jtNormal) }
func (n *PlayTapNote) Touch()              { singleTouch(&n.PlayNoteBase, jtNormal) }
func (n *PlayTapNote) UpdateParallel() {
	if n.State == stDone {
		writeResult(n.Grade, n.Error, Buckets.Tap)
		play.Entity.SetDespawn(true)
		return
	}
	now := play.Time.Now()
	pr := n.progress(now)
	if pr < 0 {
		return
	}
	drawTap(n.look(), n.Skill, n.Offbeat, curveOf(pr), n.z())
}

// singleUpdate runs the single-note miss timer and registers the candidate.
func singleUpdate(n *PlayNoteBase, jt float64) {
	if n.State == stDone {
		return
	}
	now := play.Time.Now()
	if n.State == stMove {
		if n.missDue(now) {
			finishSingle(n, gradeMiss, jtNormal, timingNone, now-n.Time, now)
			return
		}
		if math.Abs(now-n.Time) <= 9*frameSeconds {
			y := notePos(n.Lane, curveOf(n.progress(now))).Y
			registerCandidate(play.Entity.Info().Index, n.SpanStart, n.Width, n.Time, y, false, n.Finger >= 0)
		}
	}
}

func singleTouch(n *PlayNoteBase, jt float64) {
	if n.Assigned < 0 || n.State != stMove {
		return
	}
	n.Assigned = -1
	finishSingle(n, n.AGrade, jt, n.ATiming, n.ATime-n.Time, play.Time.Now())
}

func finishSingle(n *PlayNoteBase, grade, jt, timing, err, now float64) {
	n.State = stDone
	n.Grade = grade
	n.Error = err
	lane := receiverLane(n.SpanStart, n.Width)
	applyJudgement(now, judgement{
		Width: n.Width,
		Grade: grade, Type: jt, Timing: timing, Skill: n.Skill, X: laneX(lane), Lane: lane,
		Count: 1, Direction: n.Direction, Key: n.Time, Divisor: 1,
	})
}

// ---------------------------------------------------------------- flick ----

// PlayFlickNote: the touch-down grade is cached; a 2-D motion > 0.04 rate
// judges it, otherwise it is a Miss after 7 ExecuteFrames.
type PlayFlickNote struct {
	PlayNoteBase   `archetype:"base"`
	play.Archetype `archetype:"name=FlickNote,hasInput=true"`
	IconStart      float64 `archetype:"memory"`
}

func (n *PlayFlickNote) Preprocess() {
	n.prepare()
	n.Single = true
	scheduleAuto(clipFlick, n.Time, n.Skill)
}
func (n *PlayFlickNote) SpawnOrder() float64 { return n.SpawnTime }
func (n *PlayFlickNote) ShouldSpawn() bool   { return play.Time.Now() >= n.SpawnTime }
func (n *PlayFlickNote) Initialize() {
	n.IconStart = play.Time.Now()
}
func (n *PlayFlickNote) UpdateSequential() { flickUpdate(&n.PlayNoteBase) }
func (n *PlayFlickNote) Touch()            { flickTouch(&n.PlayNoteBase, 0, 1, jtFlick) }
func (n *PlayFlickNote) UpdateParallel() {
	if n.State == stDone {
		writeResult(n.Grade, n.Error, Buckets.Flick)
		play.Entity.SetDespawn(true)
		return
	}
	if n.State == stWait {
		return // the sprite is disabled while waiting for the gesture
	}
	now := play.Time.Now()
	pr := n.progress(now)
	if pr < 0 {
		return
	}
	l := n.look()
	l.Skill = n.Skill
	drawFlick(l, curveOf(pr), now-n.IconStart, n.z())
}

func flickUpdate(n *PlayNoteBase) {
	switch n.State {
	case stMove:
		singleUpdate(n, jtFlick)
	case stWait:
		n.Frames += frameStep()
		if n.Frames >= flickWaitFrames {
			finishSingle(n, gradeMiss, jtNormal, timingNone, play.Time.Now()-n.Time, play.Time.Now())
			return
		}
		registerOwned(n)
	}
}

// registerOwned keeps a note waiting for its gesture in the arbitration: if
// it wins, the touch is consumed (original GetMoveEndTimeNearestZeroNote has
// no state filter).
func registerOwned(n *PlayNoteBase) {
	if math.Abs(play.Time.Now()-n.Time) <= 9*frameSeconds {
		registerCandidate(play.Entity.Info().Index, n.SpanStart, n.Width, n.Time, n.WaitY, false, true)
	}
}

// flickTouch handles touch-down caching and the gesture of a bound finger.
// dir: 0 any, ±1 directional; count: multiple-directional group size.
func flickTouch(n *PlayNoteBase, dir, count, jt float64) {
	now := play.Time.Now()
	if n.Assigned >= 0 && n.State == stMove {
		n.Finger = n.Assigned
		n.Assigned = -1
		n.State = stWait
		n.WaitY = notePos(n.Lane, math.Min(curveOf(n.progress(now)), 1)).Y
		n.Frames = 0
		return
	}
	if n.State != stWait {
		return
	}
	for touch := range play.Touches.Values() {
		if touch.ID != n.Finger || touch.Started {
			continue
		}
		if touch.Ended {
			continue // releasing does not finish a waiting flick
		}
		if gestureDone(touch.Position.X-n.AX, touch.Position.Y-n.AY, dir, count) {
			finishSingle(n, n.AGrade, jt, n.ATiming, n.ATime-n.Time, now)
			return
		}
	}
}

// gestureDone tests a finger displacement against the original flick rates.
func gestureDone(dx, dy, dir, count float64) bool {
	if dir == 0 {
		return sq(dx)+sq(dy) > sq(Layout.FlickWorld)
	}
	if (dir < 0 && dx >= 0) || (dir > 0 && dx <= 0) {
		return false
	}
	if math.Abs(dx) <= Layout.DirWorld {
		return false
	}
	if count > 1 {
		return sq(dx)+sq(dy) > sq(((count-1)*0.01+0.01)*uiWidth/750)
	}
	return true
}

// ---------------------------------------------------------- directional ----

// PlayDirectionalMember is one lane of a directional flick. Each member
// binds its own finger; the first completed gesture (or the first timeout)
// decides the whole group, which the leader reports.
type PlayDirectionalMember struct {
	PlayNoteBase   `archetype:"base"`
	play.Archetype `archetype:"name=DirectionalMember"`
	Leader         sonolus.EntityRef[PlayDirectionalNote]   `archetype:"imported,name=leader"`
	Left           sonolus.EntityRef[PlayDirectionalMember] `archetype:"imported,name=left"`
	Right          sonolus.EntityRef[PlayDirectionalMember] `archetype:"imported,name=right"`
	Count          float64                                  `archetype:"imported,name=count"`
	Arrow          bool                                     `archetype:"imported,name=arrow"`
	IconStart      float64                                  `archetype:"memory"`
}

func (n *PlayDirectionalMember) Preprocess() {
	n.prepare()
	n.Single = true
}
func (n *PlayDirectionalMember) SpawnOrder() float64 { return n.SpawnTime }
func (n *PlayDirectionalMember) ShouldSpawn() bool   { return play.Time.Now() >= n.SpawnTime }
func (n *PlayDirectionalMember) Initialize() {
	n.IconStart = play.Time.Now()
}
func (n *PlayDirectionalMember) UpdateSequential() {
	directionalUpdate(&n.PlayNoteBase, n.Leader)
}
func (n *PlayDirectionalMember) Touch() {
	directionalTouch(&n.PlayNoteBase, n.Leader, n.Count)
}
func (n *PlayDirectionalMember) UpdateParallel() {
	if n.State == stDone || n.Leader.Get().GroupDone {
		play.Entity.SetDespawn(true)
		return
	}
	if n.Leader.Get().GroupTouched {
		return
	}
	drawDirectional(&n.PlayNoteBase, n.Arrow, n.IconStart, n.Left.Index > 0 || n.Right.Index > 0, groupFirst(n.Leader.Get().Lane, n.Count), n.Right.Index)
}

// PlayDirectionalNote is the group leader (one scoring unit per group).
type PlayDirectionalNote struct {
	PlayNoteBase   `archetype:"base"`
	play.Archetype `archetype:"name=DirectionalNote,hasInput=true"`
	Leader         sonolus.EntityRef[PlayDirectionalNote]   `archetype:"imported,name=leader"`
	Left           sonolus.EntityRef[PlayDirectionalMember] `archetype:"imported,name=left"`
	Right          sonolus.EntityRef[PlayDirectionalMember] `archetype:"imported,name=right"`
	Count          float64                                  `archetype:"imported,name=count"`
	Arrow          bool                                     `archetype:"imported,name=arrow"`
	GroupDone      bool                                     `archetype:"shared"`
	GroupGrade     float64                                  `archetype:"shared"`
	GroupTiming    float64                                  `archetype:"shared"`
	GroupError     float64                                  `archetype:"shared"`
	GroupLane      float64                                  `archetype:"shared"`
	// GroupTouched: a member was touched down; NoteMultipleDirectionalFlick's
	// ExecTouchBegan hides every member's sprite and back line
	// (setSpriteEnabledForRight/LeftSideNotes)
	GroupTouched bool    `archetype:"shared"`
	IconStart    float64 `archetype:"memory"`
}

func (n *PlayDirectionalNote) Preprocess() {
	n.prepare()
	n.Single = true
	kind := NodeDirectional
	if n.Count > 1 {
		kind = NodeMultiDir
	}
	scheduleAuto(kindClip(kind, n.Count), n.Time, false)
	n.Leader = sonolus.EntityRef[PlayDirectionalNote]{Index: play.Entity.Info().Index}
}
func (n *PlayDirectionalNote) SpawnOrder() float64 { return n.SpawnTime }
func (n *PlayDirectionalNote) ShouldSpawn() bool   { return play.Time.Now() >= n.SpawnTime }
func (n *PlayDirectionalNote) Initialize() {
	n.IconStart = play.Time.Now()
}
func (n *PlayDirectionalNote) UpdateSequential() {
	directionalUpdate(&n.PlayNoteBase, n.Leader)
}
func (n *PlayDirectionalNote) Touch() { directionalTouch(&n.PlayNoteBase, n.Leader, n.Count) }
func (n *PlayDirectionalNote) UpdateParallel() {
	if n.GroupDone {
		writeResult(n.GroupGrade, n.GroupError, Buckets.Directional)
		play.Entity.SetDespawn(true)
		return
	}
	if n.State == stDone || n.GroupTouched {
		return
	}
	drawDirectional(&n.PlayNoteBase, n.Arrow, n.IconStart, n.Left.Index > 0 || n.Right.Index > 0, groupFirst(n.Leader.Get().Lane, n.Count), n.Right.Index)
}

func groupFinish(leader sonolus.EntityRef[PlayDirectionalNote], grade, timing, err, lane float64) {
	l := leader.Get()
	if l.GroupDone {
		return
	}
	l.GroupDone = true
	l.GroupGrade, l.GroupTiming, l.GroupError, l.GroupLane = grade, timing, err, lane
	// judged in the callback that decides it (judgeFrontNote runs inside the
	// input dispatch or the member's own update), not on the leader's next one
	jt := jtDirectional
	if l.Count > 1 {
		jt = jtMultiDir
	}
	if grade == gradeMiss {
		jt = jtNormal
	}
	// the body particle starts at the group's flick-start end, the finger
	// particle on the member that was flicked
	left := l.Lane
	if Config.Mirror {
		left -= l.Count - 1 // the leader is the group's first (leftmost) member
	}
	body := dirBodyLane(left, l.Count, l.Direction)
	applyJudgement(play.Time.Now(), judgement{
		Grade: grade, Type: jt, Timing: timing, X: laneX(body), Lane: body,
		Count: l.Count, Direction: l.Direction, Key: l.Time, Divisor: 1,
		FingerX: laneX(lane), HasFinger: true,
	})
}

func directionalUpdate(n *PlayNoteBase, leader sonolus.EntityRef[PlayDirectionalNote]) {
	if n.State == stDone || leader.Get().GroupDone {
		return
	}
	now := play.Time.Now()
	switch n.State {
	case stMove:
		if n.missDue(now) {
			n.State = stDone
			groupFinish(leader, gradeMiss, timingNone, now-n.Time, n.Lane)
			return
		}
		if math.Abs(now-n.Time) <= 9*frameSeconds {
			y := notePos(n.Lane, curveOf(n.progress(now))).Y
			registerCandidate(play.Entity.Info().Index, n.SpanStart, n.Width, n.Time, y, false, n.Finger >= 0)
		}
	case stWait:
		n.Frames += frameStep()
		if n.Frames >= flickWaitFrames {
			n.State = stDone
			groupFinish(leader, gradeMiss, timingNone, now-n.Time, n.Lane)
			return
		}
		registerOwned(n)
	}
}

func directionalTouch(n *PlayNoteBase, leader sonolus.EntityRef[PlayDirectionalNote], count float64) {
	if leader.Get().GroupDone {
		return
	}
	if n.Assigned >= 0 && n.State == stMove {
		n.Finger = n.Assigned
		n.Assigned = -1
		n.State = stWait
		n.WaitY = notePos(n.Lane, math.Min(curveOf(n.progress(play.Time.Now())), 1)).Y
		n.Frames = 0
		leader.Get().GroupTouched = true
		return
	}
	if n.State != stWait {
		return
	}
	for touch := range play.Touches.Values() {
		if touch.ID != n.Finger || touch.Started || touch.Ended {
			continue
		}
		if gestureDone(touch.Position.X-n.AX, touch.Position.Y-n.AY, n.Direction, count) {
			n.State = stDone
			groupFinish(leader, n.AGrade, n.ATiming, n.ATime-n.Time, n.Lane)
			return
		}
	}
}

func drawDirectional(n *PlayNoteBase, arrow bool, iconStart float64, grouped bool, first, rightIndex float64) {
	if n.State == stWait {
		return
	}
	now := play.Time.Now()
	pr := n.progress(now)
	if pr < 0 {
		return
	}
	p := drawDirectionalBody(n.look(), curveOf(pr), now-iconStart, arrow, grouped, first, n.z())
	if rightIndex > 0 {
		other := sonolus.EntityRef[PlayNoteBase]{Index: rightIndex}.Get()
		if other.State != stWait && other.State != stDone {
			drawBackLine(p, notePos(other.Lane, curveOf(other.progress(now))), n.Direction, n.Time)
		}
	}
}

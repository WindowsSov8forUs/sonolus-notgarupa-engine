//go:build watch

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus/watch"
)

// WatchNoteBase holds a note's geometry, its judgement (Auto Live or
// replay) and the HUD snapshot taken right after that judgement.
type WatchNoteBase struct {
	watch.Archetype `archetype:"abstract"`
	Beat            float64                             `archetype:"imported,name=#BEAT"`
	Lane            float64                             `archetype:"imported,name=lane"`
	SpriteLane      float64                             `archetype:"imported,name=spriteLane"`
	Width           float64                             `archetype:"imported,name=width"`
	SpanStart       float64                             `archetype:"imported,name=spanStart"`
	Group           sonolus.EntityRef[WatchTimingGroup] `archetype:"imported,name=group"`
	AxisHit         float64                             `archetype:"imported,name=axisHit"`
	Skill           bool                                `archetype:"imported,name=skill"`
	Direction       float64                             `archetype:"imported,name=direction"`
	ScoreNext       sonolus.EntityRef[WatchNoteBase]    `archetype:"imported,name=scoreNext"`
	Judgment        float64                             `archetype:"imported,name=#JUDGMENT"`
	Accuracy        float64                             `archetype:"imported,name=#ACCURACY"`

	Time     float64 `archetype:"data"`
	Spawn    float64 `archetype:"data"`
	JTime    float64 `archetype:"data"`
	JGrade   float64 `archetype:"data"`
	JType    float64 `archetype:"data"`
	JTiming  float64 `archetype:"data"`
	JLane    float64 `archetype:"data"`
	JCount   float64 `archetype:"data"`
	JDivisor float64 `archetype:"data"`
	JMid     bool    `archetype:"shared"` // Slide head / intermediate node
	// hold loop starting at this unit's judgement ends then (0 = none);
	// WatchStage schedules them in chain order under the cue limit
	HoldUntil float64 `archetype:"shared"`
	JSlide    bool    `archetype:"shared"`

	SCombo      float64 `archetype:"shared"`
	SComboStart float64 `archetype:"shared"`
	SScore      float64 `archetype:"shared"`
	SLife       float64 `archetype:"shared"`
	SAllPerfect bool    `archetype:"shared"`
	SJudge      float64 `archetype:"shared"`
	STiming     float64 `archetype:"shared"`
	SAddValue   float64 `archetype:"shared"`
	SAddStart   float64 `archetype:"shared"`
	SJudgeStart float64 `archetype:"shared"` // the frame representative's time
	SGameOver   bool    `archetype:"shared"`
}

// replayGrade recovers the original grade from a replay's Sonolus judgment
// (Bad was recorded as Good) and accuracy.
func replayGrade(judgment, accuracy, sweet float64) float64 {
	switch judgment {
	case float64(sonolus.JudgmentPerfect):
		return gradePerfect
	case float64(sonolus.JudgmentGreat):
		return gradeGreat
	case float64(sonolus.JudgmentGood):
		if timeGrade(accuracy, sweet) == gradeBad {
			return gradeBad
		}
		return gradeGood
	}
	return gradeMiss
}

// prepare computes timing, spawn and the judgement of a scoring unit.
func (n *WatchNoteBase) prepare(sweet, jt float64, scored bool, bucket sonolus.Bucket) {
	if Config.Mirror {
		n.Lane = 6 - n.Lane
		n.SpanStart = 6 - (n.SpanStart + n.Width - 1)
		n.SpriteLane = 6 - n.SpriteLane
		n.Direction = -n.Direction
	}
	n.Time = watch.Time.BeatToTime(n.Beat)
	n.Spawn = watchSpawnTimeOf(n.Beat, n.Group, n.AxisHit)
	n.JTime = n.Time
	n.JGrade = gradePerfect
	n.JTiming = timingNone
	n.JType = jt
	n.JLane = receiverLane(n.SpanStart, n.Width)
	n.JCount = 1
	n.JDivisor = 1
	if !isAuto() {
		n.JGrade = replayGrade(n.Judgment, n.Accuracy, sweet)
		n.JTime = n.Time + n.Accuracy
		n.JTiming = timeTiming(n.JGrade, -n.Accuracy)
	}
	if scored {
		watch.Entity.SetResult(watch.InputResult{TargetTime: n.Time, Bucket: bucket, BucketValue: n.Accuracy * 1000})
	}
}

func (n *WatchNoteBase) judgement() judgement {
	return judgement{Width: n.Width, Grade: n.JGrade, Type: n.JType, Timing: n.JTiming, Skill: n.Skill, X: laneX(n.JLane), Lane: n.JLane,
		Count: n.JCount, Direction: n.Direction, Key: n.Time, Divisor: n.JDivisor, Auto: isAuto(), Mid: n.JMid, Slide: n.JSlide}
}

func (n *WatchNoteBase) progress(now float64) float64 {
	if n.Group.Index > 0 {
		return 1 - (n.AxisHit-n.Group.Get().AxisNow)/Layout.Arrival
	}
	return 1 - (n.Time-now)/Layout.Arrival
}

// z orders this note's sprites by its own time (see PlayNoteBase.z).
func (n *WatchNoteBase) z() sonolus.ZIndex { return noteZ(zNote, n.Time, n.Lane) }

func (n *WatchNoteBase) look() noteLook {
	return noteLook{SpriteLane: n.SpriteLane, Lane: n.Lane, Width: n.Width, Direction: n.Direction, SpanStart: n.SpanStart}
}

// curve returns the approach curve at now (clamped onto the line when
// stop is set), or −1 while the note is not on screen yet.
func (n *WatchNoteBase) curve(now float64, stop bool) float64 {
	pr := n.progress(now)
	if pr < 0 {
		return -1
	}
	c := curveOf(pr)
	if stop {
		c = math.Min(c, 1)
	}
	return c
}

// ------------------------------------------------------------------ tap ----

type WatchTapNote struct {
	WatchNoteBase   `archetype:"base"`
	watch.Archetype `archetype:"name=TapNote,hasInput=true"`
	Offbeat         bool `archetype:"imported,name=offbeat"`
}

func (n *WatchTapNote) Preprocess()          { n.prepare(0, jtNormal, true, Buckets.Tap) }
func (n *WatchTapNote) SpawnTime() float64   { return n.Spawn }
func (n *WatchTapNote) DespawnTime() float64 { return n.JTime }
func (n *WatchTapNote) UpdateParallel() {
	if c := n.curve(watch.Time.Now(), false); c >= 0 {
		drawTap(n.look(), n.Skill, n.Offbeat, c, n.z())
	}
}

type WatchFlickNote struct {
	WatchNoteBase   `archetype:"base"`
	watch.Archetype `archetype:"name=FlickNote,hasInput=true"`
}

func (n *WatchFlickNote) Preprocess() {
	n.prepare(0, jtFlick, true, Buckets.Flick)
	if n.JGrade == gradeMiss {
		n.JType = jtNormal
	}
}
func (n *WatchFlickNote) SpawnTime() float64   { return n.Spawn }
func (n *WatchFlickNote) DespawnTime() float64 { return n.JTime }
func (n *WatchFlickNote) UpdateParallel() {
	now := watch.Time.Now()
	if c := n.curve(now, false); c >= 0 {
		drawFlick(n.look(), c, now-n.Spawn, n.z())
	}
}

// ---------------------------------------------------------- directional ----

type WatchDirectionalBase struct {
	WatchNoteBase   `archetype:"base"`
	watch.Archetype `archetype:"abstract"`
	Leader          sonolus.EntityRef[WatchDirectionalBase] `archetype:"imported,name=leader"`
	Left            sonolus.EntityRef[WatchDirectionalBase] `archetype:"imported,name=left"`
	Right           sonolus.EntityRef[WatchDirectionalBase] `archetype:"imported,name=right"`
	Count           float64                                 `archetype:"imported,name=count"`
	Arrow           bool                                    `archetype:"imported,name=arrow"`
}

type WatchDirectionalNote struct {
	WatchDirectionalBase `archetype:"base"`
	watch.Archetype      `archetype:"name=DirectionalNote,hasInput=true"`
}

type WatchDirectionalMember struct {
	WatchDirectionalBase `archetype:"base"`
	watch.Archetype      `archetype:"name=DirectionalMember"`
}

func (n *WatchDirectionalNote) Preprocess() {
	jt := jtDirectional
	if n.Count > 1 {
		jt = jtMultiDir
	}
	n.prepare(0, jt, true, Buckets.Directional)
	n.JCount = n.Count
	left := n.Lane
	if Config.Mirror {
		left -= n.Count - 1 // the leader is the group's first (leftmost) member
	}
	n.JLane = dirBodyLane(left, n.Count, n.Direction)
	if n.JGrade == gradeMiss {
		n.JType = jtNormal
	}
}
func (n *WatchDirectionalMember) Preprocess() {
	n.prepare(0, jtDirectional, false, Buckets.Directional)
}

func (n *WatchDirectionalBase) SpawnTime() float64 { return n.Spawn }
func (n *WatchDirectionalBase) DespawnTime() float64 {
	if n.Leader.Index > 0 {
		return n.Leader.Get().JTime
	}
	return n.JTime
}
func (n *WatchDirectionalBase) UpdateParallel() {
	now := watch.Time.Now()
	c := n.curve(now, false)
	if c < 0 {
		return
	}
	p := drawDirectionalBody(n.look(), c, now-n.Spawn, n.Arrow, n.Left.Index > 0 || n.Right.Index > 0, groupFirst(n.Leader.Get().Lane, n.Count), n.z())
	if n.Right.Index > 0 {
		o := n.Right.Get()
		if oc := o.curve(now, false); oc >= 0 {
			drawBackLine(p, notePos(o.Lane, oc), n.Direction, n.Time)
		}
	}
}

// ----------------------------------------------------------------- long ----

type WatchLongHead struct {
	WatchNoteBase        `archetype:"base"`
	watch.Archetype      `archetype:"name=LongHead,hasInput=true"`
	watch.CallbackOrders `archetype:"preprocess=50"`
	Tail                 sonolus.EntityRef[WatchLongTail] `archetype:"imported,name=tail"`
	Keep                 sonolus.ParticleHandle           `archetype:"memory"`
	KeepPlane            sonolus.ParticleHandle           `archetype:"memory"`
	KeepPlane2           sonolus.ParticleHandle           `archetype:"memory"`
}

type WatchLongTail struct {
	WatchNoteBase   `archetype:"base"`
	watch.Archetype `archetype:"name=LongTail,hasInput=true"`
	Head            sonolus.EntityRef[WatchLongHead] `archetype:"imported,name=head"`
	Kind            float64                          `archetype:"imported,name=kind"`
}

func (n *WatchLongTail) Preprocess() {
	jt := jtLongEnd
	switch n.Kind {
	case NodeFlick:
		jt = jtLongEndFlick
	case NodeDirectional:
		jt = jtLongEndDir
	case NodeMultiDir:
		jt = jtLongEndMulti
	}
	n.prepare(1, jt, true, Buckets.LongTail)
	h := n.Head.Get()
	if !isAuto() && replayGrade(h.Judgment, h.Accuracy, 0) == gradeMiss {
		// head timeout: both entries at once, half damage each
		n.JGrade = gradeMiss
		n.JTime = watch.Time.BeatToTime(h.Beat) + h.Accuracy + 1e-6
		n.JDivisor = 2
	}
	if n.JGrade == gradeMiss {
		n.JType = jtLongMiss
	}
}
func (n *WatchLongTail) SpawnTime() float64   { return n.Spawn }
func (n *WatchLongTail) DespawnTime() float64 { return n.JTime }
func (n *WatchLongTail) UpdateParallel() {
	now := watch.Time.Now()
	if c := n.curve(now, true); c >= 0 {
		drawTerminal(n.look(), n.Kind, n.Width, n.Lane, c, now-n.Spawn, noteZ(zNote, n.Head.Get().Time, n.Head.Get().Lane))
	}
}

func (n *WatchLongHead) Preprocess() {
	n.prepare(0, jtLongFront, true, Buckets.LongHead)
	if n.JGrade == gradeMiss {
		n.JType = jtLongMiss
		n.JDivisor = 2
		return
	}
	watchHold(&n.WatchNoteBase, n.JTime, n.Tail.Get().JTime)
}
func (n *WatchLongHead) SpawnTime() float64 { return n.Spawn }
func (n *WatchLongHead) DespawnTime() float64 {
	if n.JGrade == gradeMiss {
		return n.JTime
	}
	return n.Tail.Get().JTime
}
func (n *WatchLongHead) held(now float64) bool { return n.JGrade != gradeMiss && now >= n.JTime }

func (n *WatchLongHead) UpdateSequential() {
	if !Config.NoteEffects || !n.held(watch.Time.Now()) || fxAlive(n.Keep, n.KeepPlane, n.KeepPlane2) {
		return
	}
	x := laneX(n.JLane)
	w := int(sonolus.Clamp(n.Width, 1, 7))
	n.Keep, n.KeepPlane, n.KeepPlane2 = spawnKeep(w, x)
}

func (n *WatchLongHead) Terminate() {
	destroyFx(n.Keep, n.KeepPlane, n.KeepPlane2)
	n.Keep.ID, n.KeepPlane.ID, n.KeepPlane2.ID = 0, 0, 0
}

func (n *WatchLongHead) UpdateParallel() {
	now := watch.Time.Now()
	held := n.held(now)
	c := n.curve(now, held)
	if c < 0 {
		return
	}
	tail := n.Tail.Get()
	tc := math.Max(tail.curve(now, true), 0)
	drawLong(n.look(), c, tail.Lane, tail.Width, tc, held, now-n.JTime, n.z(), lineZ(zBelt, n.Time))
}

// ---------------------------------------------------------------- slide ----

type WatchSlideNodeBase struct {
	WatchNoteBase   `archetype:"base"`
	watch.Archetype `archetype:"abstract"`
	Index           float64                               `archetype:"imported,name=index"`
	Kind            float64                               `archetype:"imported,name=kind"`
	KindWidth       float64                               `archetype:"imported,name=kindWidth"`
	Exit            float64                               `archetype:"imported,name=exit"`
	Hidden          bool                                  `archetype:"imported,name=hidden"`
	Terminal        bool                                  `archetype:"imported,name=terminal"`
	Head            sonolus.EntityRef[WatchSlideBase]     `archetype:"imported,name=head"`
	Next            sonolus.EntityRef[WatchSlideNodeBase] `archetype:"imported,name=next"`
}

type WatchSlideNode struct {
	WatchSlideNodeBase `archetype:"base"`
	watch.Archetype    `archetype:"name=SlideNode,hasInput=true"`
}

type WatchSlideHidden struct {
	WatchSlideNodeBase `archetype:"base"`
	watch.Archetype    `archetype:"name=SlideHidden"`
}

func (n *WatchSlideNode) Preprocess() {
	n.Skill = false // only front judgements and Long ends play the skill SE
	// gesture nodes judge as their original Slide-tail family (types 5/6/7)
	bucket := Buckets.SlideNode
	if n.Terminal {
		bucket = Buckets.SlideTail
	}
	n.prepare(0, endType(n.Kind), true, bucket)
	n.JMid, n.JSlide = !n.Terminal, true
	if n.Kind == NodeDirectional || n.Kind == NodeMultiDir {
		n.JLane = dirBodyLane(n.SpanStart, n.Width, n.Direction)
	}
	if Config.Mirror {
		n.Exit = -n.Exit
	}
	n.JCount = gestureCount(n.Kind, n.KindWidth)
	if n.JGrade == gradeMiss {
		n.JType = jtSlide
		if !n.Terminal {
			n.JDivisor = 5
		}
	}
	if !n.Terminal {
		life := watch.Entity.Life()
		life.Miss = math.Trunc(lifeDamage(gradeMiss, jtSlide, Config.FullLength) / 5)
		watch.Entity.SetLife(life)
	}
}

func (n *WatchSlideHidden) Preprocess() {
	n.prepare(0, jtSlide, false, Buckets.SlideNode)
	if Config.Mirror {
		n.Exit = -n.Exit
	}
}
func (*WatchSlideHidden) SpawnTime() float64   { return -1e9 }
func (*WatchSlideHidden) DespawnTime() float64 { return -1e9 }

func (n *WatchSlideNodeBase) SpawnTime() float64 {
	return math.Max(n.Spawn, n.Head.Get().Spawn)
}
func (n *WatchSlideNodeBase) DespawnTime() float64 { return n.JTime }

func (n *WatchSlideNodeBase) UpdateParallel() {
	now := watch.Time.Now()
	h := n.Head.Get()
	lane, c, stopped := n.Lane, 1.0, false
	if now < n.Time {
		c = n.curve(now, true)
		if c < 0 {
			return
		}
	} else {
		stopped = true
		if h.HasContact {
			lane = h.ContactLane
			if !h.SegHead && h.Seg.Index == watch.Entity.Info().Index {
				lane -= n.Exit // as Play: the exit offset stays behind the gesture
			}
		}
	}
	z := noteZ(zNote, h.Time, h.Lane)
	if !n.Terminal && (n.Kind == NodePlain || n.Kind == NodeFlick) {
		drawAmong(lane, c, n.Width, n.Kind, now-n.Spawn, stopped, z)
		return
	}
	l := n.look()
	l.Stopped = stopped // a stopped terminal takes the target-button scale
	drawTerminal(l, n.Kind, n.KindWidth, lane, c, now-n.Spawn, z)
}

type WatchSlideBase struct {
	WatchNoteBase        `archetype:"base"`
	watch.Archetype      `archetype:"abstract"`
	watch.CallbackOrders `archetype:"preprocess=50"`
	Kind                 float64                               `archetype:"imported,name=kind"`
	KindWidth            float64                               `archetype:"imported,name=kindWidth"`
	Exit                 float64                               `archetype:"imported,name=exit"`
	Hidden               bool                                  `archetype:"imported,name=hidden"`
	Curve                bool                                  `archetype:"imported,name=curve"`
	AllHidden            bool                                  `archetype:"imported,name=allHidden"`
	First                sonolus.EntityRef[WatchSlideNodeBase] `archetype:"imported,name=first"`
	Last                 sonolus.EntityRef[WatchSlideNodeBase] `archetype:"imported,name=last"`
	HoldStart            float64                               `archetype:"data"`
	HoldEnd              float64                               `archetype:"data"`

	SegHead      bool                                  `archetype:"shared"`
	Seg          sonolus.EntityRef[WatchSlideNodeBase] `archetype:"shared"`
	HasContact   bool                                  `archetype:"shared"`
	ContactLane  float64                               `archetype:"shared"`
	ContactWidth float64                               `archetype:"shared"`
	Broken       float64                               `archetype:"shared"` // as PlaySlideBase.Broken
	Keep         sonolus.ParticleHandle                `archetype:"memory"`
	KeepPlane    sonolus.ParticleHandle                `archetype:"memory"`
	KeepPlane2   sonolus.ParticleHandle                `archetype:"memory"`
}

type WatchSlideHead struct {
	WatchSlideBase  `archetype:"base"`
	watch.Archetype `archetype:"name=SlideHead,hasInput=true"`
}

type WatchSlideStart struct {
	WatchSlideBase  `archetype:"base"`
	watch.Archetype `archetype:"name=SlideStart"`
}

func watchFirstVisible(node sonolus.EntityRef[WatchSlideNodeBase]) sonolus.EntityRef[WatchSlideNodeBase] {
	for k := 0; k < 4096 && node.Index > 0; k++ {
		d := node.Get()
		if !d.Hidden {
			return node
		}
		node = d.Next
	}
	return sonolus.EntityRef[WatchSlideNodeBase]{Index: 0}
}

func (n *WatchSlideBase) Preprocess() {
	n.prepare(0, endType(n.Kind), !n.Hidden, Buckets.SlideHead)
	n.JMid, n.JSlide = true, true
	if n.Kind == NodeDirectional || n.Kind == NodeMultiDir {
		n.JLane = dirBodyLane(n.SpanStart, n.Width, n.Direction)
	}
	if Config.Mirror {
		n.Exit = -n.Exit
	}
	n.JCount = gestureCount(n.Kind, n.KindWidth)
	if n.JGrade == gradeMiss {
		n.JType = jtSlide
	}
	node := n.First
	lastVisible, firstMiss := -1e9, 1e9
	for k := 0; k < 4096 && node.Index > 0; k++ {
		d := node.Get()
		n.Spawn = math.Min(n.Spawn, d.Spawn)
		if !d.Hidden {
			lastVisible = d.JTime
			if d.JGrade == gradeMiss {
				firstMiss = math.Min(firstMiss, d.JTime)
			}
		}
		node = d.Next
	}
	// a Miss stops the hold feedback; the replay cannot tell a new touch
	lastVisible = math.Min(lastVisible, firstMiss)
	n.SegHead = true
	n.HoldStart = n.JTime
	if n.Hidden {
		n.HoldStart = 1e9
		if fv := watchFirstVisible(n.First); fv.Index > 0 {
			n.HoldStart = fv.Get().JTime
		}
	} else if n.JGrade == gradeMiss {
		n.HoldStart = 1e9
	}
	n.HoldEnd = lastVisible
	if n.AllHidden || n.HoldEnd <= n.HoldStart {
		n.HoldStart, n.HoldEnd = 1e9, 1e9
	}
	if n.HoldStart < n.HoldEnd {
		// the hold starts on the first visible node
		if n.Hidden {
			watchHold(&watchFirstVisible(n.First).Get().WatchNoteBase, n.HoldStart, n.HoldEnd)
		} else {
			watchHold(&n.WatchNoteBase, n.HoldStart, n.HoldEnd)
		}
	}
}
func (n *WatchSlideBase) SpawnTime() float64 { return n.Spawn }
func (n *WatchSlideBase) DespawnTime() float64 {
	return math.Max(n.Last.Get().Time, n.Last.Get().JTime)
}

func (n *WatchSlideBase) holding(now float64) bool {
	return now >= n.HoldStart && now < n.HoldEnd
}

func (n *WatchSlideBase) UpdateSequential() {
	now := watch.Time.Now()
	if watch.Time.Skip() {
		n.SegHead = true
	}
	for k := 0; k < 64; k++ {
		next := n.First
		if !n.SegHead {
			next = n.Seg.Get().Next
		}
		if next.Index <= 0 || now < next.Get().Time {
			break
		}
		n.Seg = next
		n.SegHead = false
	}
	// as Play: the head starts along its segment at its stop on the virtual
	// perfect line, and again from its own lane at its touch (its judgement
	// in the replay)
	ft, fl, fw := n.Time, n.Lane+n.Exit, n.Width
	start := n.Time + Layout.VPLDelay
	if !n.Hidden && n.JGrade != gradeMiss && now >= n.JTime {
		start = n.JTime
	}
	next := n.First
	if !n.SegHead {
		f := n.Seg.Get()
		ft, fl, fw = f.Time, f.Lane+f.Exit, f.Width
		start = ft
		next = f.Next
	}
	n.HasContact = now >= start
	n.ContactWidth = fw
	n.ContactLane = fl
	if n.HasContact && next.Index > 0 {
		a := next.Get()
		k := 1.0
		if a.Time > ft {
			k = sonolus.Clamp((now-start)/(a.Time-ft), 0, 1)
		}
		n.ContactLane = fl + (a.Lane-fl)*k
	}
	n.Broken = -1
	if !n.Hidden && n.JGrade == gradeMiss && now >= n.JTime {
		n.Broken = 0
	}
	node := n.First
	for k := 0; k < 4096 && node.Index > 0; k++ {
		d := node.Get()
		if !d.Hidden && d.JGrade == gradeMiss && now >= d.JTime {
			n.Broken = d.Index
		}
		node = d.Next
	}
	if Config.NoteEffects && n.holding(now) && !fxAlive(n.Keep, n.KeepPlane, n.KeepPlane2) {
		n.Keep, n.KeepPlane, n.KeepPlane2 = spawnFx(fxKeep, laneX(n.ContactLane), durKeep, durKeepPlane, true)
	}
	if fxAlive(n.Keep, n.KeepPlane, n.KeepPlane2) && !n.holding(now) {
		destroyFx(n.Keep, n.KeepPlane, n.KeepPlane2)
		n.Keep.ID, n.KeepPlane.ID, n.KeepPlane2.ID = 0, 0, 0
	}
}

func (n *WatchSlideBase) Terminate() {
	destroyFx(n.Keep, n.KeepPlane, n.KeepPlane2)
	n.Keep.ID, n.KeepPlane.ID, n.KeepPlane2.ID = 0, 0, 0
}

func (n *WatchSlideBase) UpdateParallel() {
	now := watch.Time.Now()
	n.drawBody(now)
	if !n.Hidden && now < n.JTime {
		if c := n.curve(now, true); c >= 0 {
			// an unheld head snaps onto the line at the target-button scale
			l := n.look()
			l.Stopped = n.progress(now) >= 1
			lane := n.Lane
			if l.Stopped {
				lane = n.anchorLane(now) // as Play: the stopped head rides the contact point
			}
			drawTerminal(l, n.Kind, n.KindWidth, lane, c, now-n.Spawn, n.z())
		}
	}
	if n.holding(now) {
		lane := n.anchorLane(now)
		p := heldPos(lane)
		chain, passed := n.chainSprite(now)
		body, flash, px := slideHoldSprites(n.Width, lane)
		drawHoldFlash(body, flash, p, px, !n.Hidden && chain == chainHead, now-n.HoldStart, n.z())
		if chain == chainPassed {
			drawAmong(lane, 1, passed.Get().Width, NodePlain, 0, true, n.z())
		}
		moveFx(fxKeep, p.X, n.Keep, n.KeepPlane, n.KeepPlane2)
	}
}

// anchorLane is PlaySlideBase.anchorLane; a gesture shows until its judgement.
func (n *WatchSlideBase) anchorLane(now float64) float64 {
	if !n.HasContact {
		return n.Lane
	}
	lane := n.ContactLane
	if n.SegHead {
		if now < n.JTime {
			lane -= n.Exit
		}
	} else if f := n.Seg.Get(); now < f.JTime {
		lane -= f.Exit
	}
	return lane
}

// chainSprite is PlaySlideBase.chainSprite, with the passed node taken from
// the replay: the last visible middle node judged (not Missed) by now.
func (n *WatchSlideBase) chainSprite(now float64) (float64, sonolus.EntityRef[WatchSlideNodeBase]) {
	passed := sonolus.EntityRef[WatchSlideNodeBase]{Index: 0}
	node := n.First
	for k := 0; k < 4096 && node.Index > 0; k++ {
		d := node.Get()
		if !d.Hidden && !d.Terminal && d.JGrade != gradeMiss && now >= d.JTime {
			passed = node
		}
		node = d.Next
	}
	from, chain := n.First, chainHead
	if passed.Index > 0 {
		from, chain = passed.Get().Next, chainPassed
	}
	for k := 0; k < 4096 && from.Index > 0; k++ {
		d := from.Get()
		if !d.Hidden {
			if chain == chainPassed && now >= d.stopTime() {
				return chainNone, passed
			}
			return chain, passed
		}
		if now >= d.Time {
			return chainNone, passed
		}
		from = d.Next
	}
	return chain, passed
}

func (n *WatchSlideNodeBase) stopTime() float64 {
	if n.Terminal {
		return n.Time + Layout.VPLDelay
	}
	return n.Time
}

func (n *WatchSlideBase) drawBody(now float64) {
	alpha := bodyAlpha()
	if n.AllHidden {
		alpha *= 0.5
	}
	z := lineZ(zBelt, n.Time)
	var fp sonolus.Vec2
	var fh float64
	next := n.First
	front := 0.0
	if n.HasContact {
		fp, fh = heldPos(n.ContactLane), heldHalf(n.ContactWidth)
		if !n.SegHead {
			next = n.Seg.Get().Next
			front = n.Seg.Get().Index
		}
	} else {
		c := n.curve(now, true)
		if c < 0 {
			return
		}
		fp = notePos(n.Lane+n.Exit, c)
		fh = bodyHalf(fp.Y, n.Width)
		if c >= 1 {
			fh = heldHalf(n.Width)
		}
	}
	for k := 0; k < 64 && next.Index > 0; k++ {
		a := next.Get()
		c := a.curve(now, true)
		ap := notePos(a.Lane, math.Max(c, 0))
		ah := bodyHalf(ap.Y, a.Width)
		switch {
		case c < 0:
			ah = waitHalf(a.Width)
		case c >= 1:
			ah = heldHalf(a.Width)
		}
		if front > n.Broken {
			drawBelt(fp, fh, ap, ah, n.Curve, z, alpha)
		}
		if c < 0 {
			return
		}
		fp, fh = notePos(a.Lane+a.Exit, c), ah
		front = a.Index
		next = a.Next
	}
}

// ------------------------------------------------------------- sync line ----

type WatchSimLine struct {
	watch.Archetype `archetype:"name=SimLine"`
	A               sonolus.EntityRef[WatchNoteBase] `archetype:"imported,name=a"`
	B               sonolus.EntityRef[WatchNoteBase] `archetype:"imported,name=b"`
	ALane           float64                          `archetype:"imported,name=aLane,default=-1000"`
	BLane           float64                          `archetype:"imported,name=bLane,default=-1000"`
	SpeedMin        float64                          `archetype:"imported,name=speedMin,default=0"` // note speeds (tenths) the line exists at
	SpeedMax        float64                          `archetype:"imported,name=speedMax,default=1000"`
	ANoMargin       bool                             `archetype:"imported,name=aNoMargin"` // directional-note end: no edge margin
	BNoMargin       bool                             `archetype:"imported,name=bNoMargin"`
}

func (l *WatchSimLine) Preprocess() {
	l.ALane, l.BLane = simLane(l.ALane, l.A.Get().Lane), simLane(l.BLane, l.B.Get().Lane)
}

func (l *WatchSimLine) SpawnTime() float64 {
	if !Config.SyncLine || !syncLineAtSpeed(l.SpeedMin, l.SpeedMax) {
		return 1e9
	}
	return math.Max(l.A.Get().Spawn, l.B.Get().Spawn)
}
func (l *WatchSimLine) DespawnTime() float64 {
	return math.Min(math.Min(l.A.Get().JTime, l.B.Get().JTime), math.Min(l.A.Get().Time, l.B.Get().Time))
}
func (l *WatchSimLine) UpdateParallel() {
	now := watch.Time.Now()
	a, b := l.A.Get(), l.B.Get()
	pa, pb := a.progress(now), b.progress(now)
	if pa < 0 || pb < 0 || pa > 1 || pb > 1 {
		return
	}
	drawSimLine(notePos(l.ALane, curveOf(pa)), notePos(l.BLane, curveOf(pb)), math.Min(a.Time, b.Time), l.ANoMargin, l.BNoMargin)
}

// watchChained: the level has the scoring chain (firstScored); levels from
// the old converter do not.
func watchChained() bool {
	return sonolus.EntityRef[WatchInitialization]{Index: 0}.Get().FirstScored.Index > 0
}

// watchHold records a hold loop for WatchStage to schedule under the cue
// limit; without a scoring chain it schedules the loop directly (uncapped).
func watchHold(n *WatchNoteBase, start, end float64) {
	if watchChained() {
		n.HoldUntil = end
		return
	}
	if Config.SFX {
		scheduleHoldSound(start, end)
	}
}

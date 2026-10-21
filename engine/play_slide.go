//go:build play

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus/play"
)

// A Slide is one head entity (SlideHead, or SlideStart when the head is
// Hidden) that owns the hold: cursor, finger, contact point, body, flash and
// hold feedback; plus one entity per connection (SlideNode, or SlideHidden).
// Nodes draw their own sprite and carry their own result; the head judges
// them (original NoteSlide with its NoteSlideAfter children).

// PlaySlideNodeBase is one connection after the head.
type PlaySlideNodeBase struct {
	PlayNoteBase   `archetype:"base"`
	play.Archetype `archetype:"abstract"`
	Index          float64                              `archetype:"imported,name=index"`
	Kind           float64                              `archetype:"imported,name=kind"`
	KindWidth      float64                              `archetype:"imported,name=kindWidth"`
	Exit           float64                              `archetype:"imported,name=exit"`
	Hidden         bool                                 `archetype:"imported,name=hidden"`
	Terminal       bool                                 `archetype:"imported,name=terminal"`
	Head           sonolus.EntityRef[PlaySlideBase]     `archetype:"imported,name=head"`
	Prev           sonolus.EntityRef[PlaySlideNodeBase] `archetype:"imported,name=prev"`
	Next           sonolus.EntityRef[PlaySlideNodeBase] `archetype:"imported,name=next"`
	Judged         bool                                 `archetype:"shared"`
	IconStart      float64                              `archetype:"memory"`
}

type PlaySlideNode struct {
	PlaySlideNodeBase `archetype:"base"`
	play.Archetype    `archetype:"name=SlideNode,hasInput=true"`
}

type PlaySlideHidden struct {
	PlaySlideNodeBase `archetype:"base"`
	play.Archetype    `archetype:"name=SlideHidden"`
}

func (n *PlaySlideNodeBase) Preprocess() {
	n.Skill = false // only front judgements and Long ends play the skill SE
	n.prepare()
	if Config.Mirror {
		n.Exit = -n.Exit
	}
	if !n.Terminal && !n.Hidden {
		// a middle node's Miss costs a fifth of the Slide damage (Entity Life)
		life := play.Entity.Life()
		life.Miss = math.Trunc(lifeDamage(gradeMiss, jtSlide, Config.FullLength) / 5)
		play.Entity.SetLife(life)
	}
	if !n.Hidden {
		scheduleAuto(kindClip(n.Kind, n.KindWidth), n.Time, n.Skill)
	}
}
func (n *PlaySlideNodeBase) SpawnOrder() float64 { return n.SpawnTime }
func (n *PlaySlideNodeBase) ShouldSpawn() bool {
	return n.Judged || play.Time.Now() >= math.Max(n.SpawnTime, n.Head.Get().SpawnTime)
}
func (n *PlaySlideNodeBase) Initialize() { n.IconStart = play.Time.Now() }

// Hidden nodes never spawn; they must sort last or they block the queue.
func (*PlaySlideHidden) SpawnOrder() float64 { return 1e9 }
func (*PlaySlideHidden) ShouldSpawn() bool   { return false }

func (n *PlaySlideNodeBase) UpdateParallel() {
	if n.Judged {
		bucket := Buckets.SlideNode
		if n.Terminal {
			bucket = Buckets.SlideTail
		}
		writeResult(n.Grade, n.Error, bucket)
		play.Entity.SetDespawn(true)
		return
	}
	h := n.Head.Get()
	if h.Ended {
		play.Entity.SetDespawn(true)
		return
	}
	if n.State == stWait {
		return // a picked-up flick end hides its sprite
	}
	now := play.Time.Now()
	lane, c, stopped := n.Lane, 1.0, false
	if now < n.Time {
		pr := n.progress(now)
		if pr < 0 {
			return
		}
		c = math.Min(curveOf(pr), 1)
	} else {
		stopped = true
		if h.HasContact {
			// a stopped node rides the contact point, gesture or not; the
			// front's own exit offset stays behind it while its gesture shows
			lane = h.ContactLane
			if !h.SegHead && h.Seg.Index == play.Entity.Info().Index {
				lane -= n.Exit
			}
		}
	}
	// a child sorts with its head (head time and lane)
	z := noteZ(zNote, h.Time, h.Lane)
	if !n.Terminal && (n.Kind == NodePlain || n.Kind == NodeFlick) {
		drawAmong(lane, c, n.Width, n.Kind, now-n.IconStart, stopped, z)
		return
	}
	l := n.look()
	l.Stopped = stopped // a stopped terminal takes the target-button scale
	drawTerminal(l, n.Kind, n.KindWidth, lane, c, now-n.IconStart, z)
}

// PlaySlideBase is the head and owner of a Slide.
type PlaySlideBase struct {
	PlayNoteBase   `archetype:"base"`
	play.Archetype `archetype:"abstract"`
	Kind           float64                              `archetype:"imported,name=kind"`
	KindWidth      float64                              `archetype:"imported,name=kindWidth"`
	Exit           float64                              `archetype:"imported,name=exit"`
	Hidden         bool                                 `archetype:"imported,name=hidden"`
	Curve          bool                                 `archetype:"imported,name=curve"`
	AllHidden      bool                                 `archetype:"imported,name=allHidden"`
	First          sonolus.EntityRef[PlaySlideNodeBase] `archetype:"imported,name=first"`
	Last           sonolus.EntityRef[PlaySlideNodeBase] `archetype:"imported,name=last"`

	HeadDone     bool                                 `archetype:"shared"`
	OnHead       bool                                 `archetype:"shared"` // the cursor is the head
	Cursor       sonolus.EntityRef[PlaySlideNodeBase] `archetype:"shared"`
	JudgeDone    bool                                 `archetype:"shared"`
	Ended        bool                                 `archetype:"shared"`
	Touched      bool                                 `archetype:"shared"`
	SegHead      bool                                 `archetype:"shared"`
	Seg          sonolus.EntityRef[PlaySlideNodeBase] `archetype:"shared"`
	HasContact   bool                                 `archetype:"shared"`
	ContactLane  float64                              `archetype:"shared"`
	ContactWidth float64                              `archetype:"shared"`
	Cached       float64                              `archetype:"shared"`
	CachedTiming float64                              `archetype:"shared"`
	CachedError  float64                              `archetype:"shared"`
	Grace        float64                              `archetype:"shared"`
	OX           float64                              `archetype:"shared"`
	OY           float64                              `archetype:"shared"`
	Feedback     bool                                 `archetype:"shared"`
	// Broken is the order (head 0, node Index) of the last missed node: the
	// original hides the segments up to and out of it (hideNote(root,
	// current, true)); −1 while nothing is missed.
	Broken     float64 `archetype:"shared"`
	FlashStart float64 `archetype:"shared"`
	GestureAt  float64 `archetype:"memory"` // cursor whose gesture state is set
	// A gesture node's own detection state (origin, grace) and touch-down
	// cache; the original's tail origin/grace (OX, OY, Grace) and tail cache
	// (Cached…) stay untouched by gesture nodes.
	GX           float64 `archetype:"memory"`
	GY           float64 `archetype:"memory"`
	GGrace       float64 `archetype:"memory"`
	GCached      float64 `archetype:"memory"`
	GCachedTime  float64 `archetype:"memory"`
	GCachedError float64 `archetype:"memory"`
	// Passed is the last middle node the hold has judged; the original keeps
	// its sprite riding at the front (ExecTouchMoved hides only the note
	// before it), so the held head body shows only until the first one.
	Passed     sonolus.EntityRef[PlaySlideNodeBase] `archetype:"memory"`
	Loop       sonolus.LoopedEffectHandle           `archetype:"memory"`
	Keep       sonolus.ParticleHandle               `archetype:"memory"`
	KeepPlane  sonolus.ParticleHandle               `archetype:"memory"`
	KeepPlane2 sonolus.ParticleHandle               `archetype:"memory"`
}

type PlaySlideHead struct {
	PlaySlideBase  `archetype:"base"`
	play.Archetype `archetype:"name=SlideHead,hasInput=true"`
}

type PlaySlideStart struct {
	PlaySlideBase  `archetype:"base"`
	play.Archetype `archetype:"name=SlideStart"`
}

// spawnTimeOf is prepare()'s spawn time of any note, without side effects.
func spawnTimeOf(beat float64, group sonolus.EntityRef[PlayTimingGroup], axisHit float64) float64 {
	t := play.Time.BeatToTime(beat)
	visual := t - Layout.Arrival
	if group.Index > 0 {
		visual = groupVisualSpawn(group, axisHit)
	}
	return math.Min(visual, t-0.25)
}

func (n *PlaySlideBase) Preprocess() {
	n.prepare()
	if Config.Mirror {
		n.Exit = -n.Exit
	}
	// the head must be alive before any of its nodes appears
	node := n.First
	holdStart, holdEnd := n.Time, -1e9
	if n.Hidden {
		holdStart = 1e9
	} else {
		scheduleAuto(kindClip(n.Kind, n.KindWidth), n.Time, n.Skill)
	}
	for k := 0; k < 4096 && node.Index > 0; k++ {
		d := node.Get()
		n.SpawnTime = math.Min(n.SpawnTime, spawnTimeOf(d.Beat, d.Group, d.AxisHit))
		if !d.Hidden {
			t := play.Time.BeatToTime(d.Beat)
			holdStart = math.Min(holdStart, t)
			holdEnd = t
		}
		node = d.Next
	}
	if !n.AllHidden && holdEnd > holdStart {
		// the hold starts on the first visible node
		if n.Hidden {
			autoHold(&firstVisible(n.First).Get().PlayNoteBase, holdStart, holdEnd)
		} else {
			autoHold(&n.PlayNoteBase, holdStart, holdEnd)
		}
	}
	n.SegHead = true
	n.Broken = -1
	n.Cached = -1
	n.GCached = -1
	n.OnHead = !n.Hidden
	n.HeadDone = n.Hidden
	if n.Hidden {
		n.Cursor = firstVisible(n.First)
		n.JudgeDone = n.Cursor.Index <= 0 || n.AllHidden
	}
}
func (n *PlaySlideBase) SpawnOrder() float64 { return n.SpawnTime }
func (n *PlaySlideBase) ShouldSpawn() bool   { return play.Time.Now() >= n.SpawnTime }

// firstVisible returns node or the first visible node after it (Index 0: none).
func firstVisible(node sonolus.EntityRef[PlaySlideNodeBase]) sonolus.EntityRef[PlaySlideNodeBase] {
	for k := 0; k < 4096 && node.Index > 0; k++ {
		d := node.Get()
		if !d.Hidden {
			return node
		}
		node = d.Next
	}
	return sonolus.EntityRef[PlaySlideNodeBase]{Index: 0}
}

// ------------------------------------------------------------ sequential ----

func (n *PlaySlideBase) UpdateSequential() {
	if n.Ended {
		return
	}
	now := play.Time.Now()
	n.advanceSegment(now)
	n.updateContact(now)
	if n.JudgeDone {
		if now >= n.Last.Get().Time {
			n.Ended = true
			n.stopFeedback()
		}
		return
	}
	if n.OnHead {
		n.headTimeout(now)
	} else {
		n.nodeTimeout(now)
	}
	if n.JudgeDone || n.Ended {
		return
	}
	n.register(now)
}

// advanceSegment moves the live segment's front to the last node reached.
func (n *PlaySlideBase) advanceSegment(now float64) {
	for k := 0; k < 64; k++ {
		next := n.First
		if !n.SegHead {
			next = n.Seg.Get().Next
		}
		if next.Index <= 0 || now < next.Get().Time {
			return
		}
		n.Seg = next
		n.SegHead = false
	}
}

// updateContact places the contact point on the judge line: the stopped
// node slides toward the next node, linearly in time.
func (n *PlaySlideBase) updateContact(now float64) {
	ft, fl, fw := n.Time, n.Lane+n.Exit, n.Width
	next := n.First
	if !n.SegHead {
		f := n.Seg.Get()
		ft, fl, fw = f.Time, f.Lane+f.Exit, f.Width
		next = f.Next
	}
	if now < ft {
		n.HasContact = false
		return
	}
	n.HasContact = true
	n.ContactWidth = fw
	if next.Index <= 0 {
		n.ContactLane = fl
		return
	}
	a := next.Get()
	k := 1.0
	if a.Time > ft {
		k = sonolus.Clamp((now-ft)/(a.Time-ft), 0, 1)
	}
	n.ContactLane = fl + (a.Lane-fl)*k
}

// headArrived: the head waits once it has crossed the line and reached the
// virtual perfect line (stopSlideHeadAtJudgeLine: progress > 1 — visual under
// SV — and τ ≤ 0); only then can it time out.
func (n *PlaySlideBase) headArrived(now float64) bool {
	return n.progress(now) > 1 && now >= n.Time+Layout.VPLDelay
}

func (n *PlaySlideBase) headTimeout(now float64) {
	if n.HeadDone || !n.headArrived(now) {
		return
	}
	elapsed := now - n.Time
	remaining := 1e9
	if fv := firstVisible(n.First); fv.Index > 0 {
		remaining = fv.Get().Time - now
	}
	if elapsed > missSeconds || remaining <= 0 || elapsed > remaining {
		n.Grade, n.Error = gradeMiss, missSeconds
		n.HeadDone = true
		n.State = stDone
		lane := n.headLane()
		applyJudgement(now, judgement{Width: n.Width, Grade: gradeMiss, Type: jtSlide, X: laneX(lane), Lane: lane,
			Key: n.Time, Count: gestureCount(n.Kind, n.KindWidth), Direction: n.Direction, Divisor: 1, Mid: true, Slide: true})
		n.stopFeedback()
		n.Broken = 0
		n.leaveHead()
	}
}

func (n *PlaySlideBase) leaveHead() {
	n.OnHead = false
	n.GCached = -1
	n.Cursor = firstVisible(n.First)
	if n.Cursor.Index <= 0 {
		n.JudgeDone = true
		n.stopFeedback()
	}
}

// nodeTimeout is NoteSlideAfter.StopState for the current node.
func (n *PlaySlideBase) nodeTimeout(now float64) {
	c := n.Cursor.Get()
	stop := c.Time
	if c.Terminal {
		stop += Layout.VPLDelay // the terminal stops at the virtual perfect line
	}
	if c.Judged || now < stop {
		return
	}
	elapsed := now - c.Time
	miss := false
	divisor := 1.0
	switch {
	case !c.Terminal:
		divisor = 5
		vis := firstVisible(c.Next)
		if vis.Index <= 0 {
			// only Hidden nodes follow (extension): 13/60 s like an end
			miss = elapsed > missSeconds
		} else {
			// slideAfterTimeoutDue: the next visible node has stopped (the
			// terminal stops at the virtual perfect line), or more time has
			// passed since this node than remains to the next
			v := vis.Get()
			stopAt := v.Time
			if v.Terminal {
				stopAt += Layout.VPLDelay
			}
			remaining := v.Time - now
			miss = elapsed > missSeconds || now >= stopAt || remaining > 0 && elapsed > remaining
		}
	case c.Kind == NodeFlick:
		c.Frames += frameStep()
		miss = c.Frames >= flickWaitFrames
	default:
		miss = elapsed > missSeconds
	}
	if miss {
		n.stopFeedback()
		n.Broken = c.Index
		n.judgeNode(n.Cursor, gradeMiss, timingNone, missSeconds, jtSlide, divisor, now)
	}
}

// register offers the current node to this frame's touch arbitration.
func (n *PlaySlideBase) register(now float64) {
	t, start, width, y := n.Time, n.SpanStart, n.Width, Layout.G
	if n.OnHead {
		if now < n.Time {
			y = notePos(n.Lane, math.Min(curveOf(n.progress(now)), 1)).Y
		}
	} else {
		c := n.Cursor.Get()
		t, start, width = c.Time, c.SpanStart, c.Width
		if now < c.Time {
			y = notePos(c.Lane, math.Min(curveOf(c.progress(now)), 1)).Y
		}
	}
	if n.awaitingRegistration() {
		return // registered from any held touch instead (registerHidden)
	}
	if slideTau(t, now) > 9 {
		return
	}
	registerCandidate(play.Entity.Info().Index, start, width, t, y, true, n.Finger >= 0)
}

// judgeNode records a node's judgement and advances the cursor past it.
func (n *PlaySlideBase) judgeNode(ref sonolus.EntityRef[PlaySlideNodeBase], grade, timing, err, jt, divisor, now float64) {
	c := ref.Get()
	c.Judged = true
	c.Grade, c.Error = grade, err
	c.State = stDone
	n.GCached = -1
	if !c.Terminal && grade != gradeMiss {
		n.Passed = ref
	}
	lane := receiverLane(c.SpanStart, c.Width)
	if c.Kind == NodeDirectional || c.Kind == NodeMultiDir {
		lane = dirBodyLane(c.SpanStart, c.Width, c.Direction)
	}
	applyJudgement(now, judgement{Width: c.Width, Grade: grade, Type: jt, Timing: timing, Skill: c.Skill, X: laneX(lane), Lane: lane,
		Count: gestureCount(c.Kind, c.KindWidth), Direction: c.Direction, Key: c.Time, Divisor: divisor, Mid: !c.Terminal, Slide: true})
	if c.Terminal {
		n.JudgeDone = true
		n.Ended = true
		n.Finger = -1
		n.stopFeedback()
		return
	}
	n.Cursor = firstVisible(c.Next)
	if n.Cursor.Index <= 0 {
		n.JudgeDone = true
		n.stopFeedback()
	}
}

func (n *PlaySlideBase) startFeedback(now float64) {
	n.Feedback = true
	n.FlashStart = now
	if Config.SFX && !Config.AutoSFX && n.Loop.ID <= 0 {
		n.Loop = startHoldLoop()
	}
	if Config.NoteEffects && !fxAlive(n.Keep, n.KeepPlane, n.KeepPlane2) {
		// a Slide's TapKeep is always the base one
		n.Keep, n.KeepPlane, n.KeepPlane2 = spawnFx(fxKeep, laneX(n.contactOr(n.Lane)), durKeep, durKeepPlane, true)
	}
}

func (n *PlaySlideBase) stopFeedback() {
	n.Feedback = false
	if n.Loop.ID > 0 {
		releaseHoldLoop(n.Loop)
		n.Loop.ID = 0
	}
	destroyFx(n.Keep, n.KeepPlane, n.KeepPlane2)
	n.Keep.ID, n.KeepPlane.ID, n.KeepPlane2.ID = 0, 0, 0
}

func (n *PlaySlideBase) contactOr(lane float64) float64 {
	if n.HasContact {
		return n.ContactLane
	}
	return lane
}

// ----------------------------------------------------------------- touch ----

func (n *PlaySlideBase) Touch() {
	if n.Ended {
		return
	}
	now := play.Time.Now()
	if n.Assigned >= 0 {
		id := n.Assigned
		n.Assigned = -1
		if n.Finger < 0 && !n.JudgeDone && n.AGrade != gradeNone {
			n.pickUp(id, now)
		}
	}
	if n.Finger < 0 && n.awaitingRegistration() {
		n.registerHidden(now)
	}
	if n.Finger < 0 {
		return
	}
	for touch := range play.Touches.Values() {
		// a touch that began this frame is only processed when it also ended
		if touch.ID != n.Finger || (touch.Started && !touch.Ended) {
			continue
		}
		if n.JudgeDone {
			if touch.Ended {
				n.Finger = -1
			}
			return
		}
		// held reports a release forced by the tail block (D15)
		if touch.Ended || n.held(touch.Position, now) {
			n.release(touch.Position, touch.Time, now)
		}
		return
	}
}

// pickUp is a touch-down on the current node (ExecTouchBegan).
func (n *PlaySlideBase) pickUp(id, now float64) {
	n.Finger = id
	n.Touched = true
	if n.OnHead {
		if n.Kind != NodePlain {
			// gesture head: the flick tail's touch-down caches the grade; the
			// gesture shows until it is judged
			n.cacheGesture(n.AGrade, n.ATiming, n.ATime-n.Time)
			n.startFeedback(now)
			return
		}
		n.Grade, n.Error = n.AGrade, n.ATime-n.Time
		n.HeadDone = true
		n.State = stDone
		lane := receiverLane(n.SpanStart, n.Width)
		applyJudgement(now, judgement{Width: n.Width, Grade: n.AGrade, Type: jtSlide, Timing: n.ATiming, Skill: n.Skill,
			X: laneX(lane), Lane: lane, Key: n.Time, Count: 1, Divisor: 1, Mid: true, Slide: true})
		n.startFeedback(now)
		n.leaveHead()
		return
	}
	c := n.Cursor.Get()
	switch {
	case c.Terminal && c.Kind != NodePlain:
		// flick-family end picked up: cache the grade, judge on the gesture
		n.Cached, n.CachedTiming, n.CachedError = n.AGrade, n.ATiming, n.ATime-c.Time
		n.OX, n.OY = n.AX, n.AY
		c.State = stWait
		n.startFeedback(now)
	case c.Kind != NodePlain:
		// intermediate gesture: the flick tail's touch-down as well
		n.cacheGesture(n.AGrade, n.ATiming, n.ATime-c.Time)
		n.GestureAt = float64(n.Cursor.Index)
		n.startFeedback(now)
	default:
		n.judgeNode(n.Cursor, n.AGrade, n.ATiming, n.ATime-c.Time, jtSlide, 1, now)
		if !n.JudgeDone {
			n.startFeedback(now)
		}
	}
}

// awaitingRegistration: the first visible node of a headless Slide is never a
// button candidate; any held touch registers it (Hidden-head extension rule).
func (n *PlaySlideBase) awaitingRegistration() bool {
	return n.Hidden && !n.Touched && !n.JudgeDone && !n.OnHead && n.Cursor.Index > 0 && !n.Cursor.Get().Judged
}

// registerHidden checks every touch that has not ended against the first
// visible node of a headless Slide: a plain node judges Perfect at τ ≤ 1, a
// gesture node binds within its window.
func (n *PlaySlideBase) registerHidden(now float64) {
	c := n.Cursor.Get()
	start, width := c.SpanStart, c.Width
	tau := slideTau(c.Time, now)
	if c.Kind == NodePlain && tau > 1 || c.Kind != NodePlain && tau > 8 {
		return
	}
	for touch := range play.Touches.Values() {
		if touch.Ended || !insideSpan(touch.Position, start, width, false) {
			continue
		}
		n.Finger, n.Touched = touch.ID, true
		if !touch.Started {
			n.moveLane(c.SpanStart, c.Width, now)
		}
		if c.Kind == NodePlain {
			n.judgeNode(n.Cursor, gradePerfect, timingNone, now-c.Time, jtSlide, 1, now)
		} else {
			n.GX, n.GY, n.GGrace, n.GestureAt = touch.Position.X, touch.Position.Y, 0, float64(n.Cursor.Index)
		}
		if !n.JudgeDone {
			n.startFeedback(now)
		}
		return
	}
}

// cacheGesture is the flick tail's touch-down (ExecTouchBegan cachedResult)
// on a gesture head or middle node: the touch's grade and timing wait for
// the flick distance from the touch-down position.
func (n *PlaySlideBase) cacheGesture(grade, timing, err float64) {
	n.GCached, n.GCachedTime, n.GCachedError = grade, timing, err
	n.GX, n.GY, n.GGrace = n.AX, n.AY, 0
}

// gestureJudged is a gesture node's own judgement, the original flick tail
// path on the node's own state: a cached touch-down completes on the flick
// distance alone (no direction, grace or VPL gate); otherwise the contact
// grace refreshes inside the node, the origin follows the finger until the
// node reaches the VPL, then a flick from the origin completes it (Perfect).
func (n *PlaySlideBase) gestureJudged(p sonolus.Vec2, now, t, kind, kindWidth, direction, start, width float64) (bool, float64, float64, float64) {
	if n.GCached >= 0 {
		return cachedFlickDone(p.X-n.GX, p.Y-n.GY, kind, kindWidth), n.GCached, n.GCachedTime, n.GCachedError
	}
	if insideSpan(p, start, width, false) {
		n.GGrace = leaveGrace
	} else {
		n.GGrace -= frameStep()
	}
	if slideTau(t, now) > 0 {
		n.GX, n.GY = p.X, p.Y
		return false, 0, 0, 0
	}
	done := n.GGrace > 0 && gestureDone(p.X-n.GX, p.Y-n.GY, gestureDir(kind, direction), gestureCount(kind, kindWidth))
	return done, gradePerfect, timingNone, now - t
}

// cachedFlickDone is the cached tail's distance test from the touch-down.
func cachedFlickDone(dx, dy, kind, kindWidth float64) bool {
	switch kind {
	case NodeFlick:
		return sq(dx)+sq(dy) > sq(Layout.FlickWorld)
	case NodeDirectional:
		return math.Abs(dx) > Layout.FlickWorld
	}
	return math.Abs(dx) > Layout.DirWorld && sq(dx)+sq(dy) > sq(((kindWidth-1)*0.01+0.01)*uiWidth/750)
}

// held handles a Moved / Stationary frame of the owning finger; true means
// the Slide must release (a terminal flick completed with a node pending).
func (n *PlaySlideBase) held(p sonolus.Vec2, now float64) bool {
	if n.OnHead {
		// gesture head (a plain head is judged on touch-down): its own
		// judgement as its tail family
		// contact span: the node's own span, as the matching tail uses
		if ok, g, timing, err := n.gestureJudged(p, now, n.Time, n.Kind, n.KindWidth, n.Direction, n.SpanStart, n.Width); ok {
			n.Grade, n.Error = g, err
			n.HeadDone = true
			n.State = stDone
			lane := n.headLane()
			applyJudgement(now, judgement{Width: n.Width, Grade: g, Type: endType(n.Kind), Timing: timing, Skill: n.Skill, X: laneX(lane), Lane: lane,
				Key: n.Time, Count: gestureCount(n.Kind, n.KindWidth), Direction: n.Direction, Divisor: 1, Mid: true, Slide: true})
			n.leaveHead()
			n.OX, n.OY = p.X, p.Y
		}
		return false
	}
	c := n.Cursor.Get()
	if c.Kind != NodePlain && !c.Terminal && n.GestureAt != float64(n.Cursor.Index) {
		// reaching a gesture node while held: its own state starts here
		n.GX, n.GY, n.GGrace, n.GestureAt = p.X, p.Y, 0, float64(n.Cursor.Index)
	}
	if !c.Terminal {
		if c.Kind == NodePlain {
			if slideTau(c.Time, now) <= 1 && insideSpan(p, c.SpanStart, c.Width, false) {
				n.moveLane(c.SpanStart, c.Width, now)
				n.judgeNode(n.Cursor, gradePerfect, timingNone, now-c.Time, jtSlide, 1, now)
				n.OX, n.OY = p.X, p.Y
			}
		} else {
			if ok, g, timing, err := n.gestureJudged(p, now, c.Time, c.Kind, c.KindWidth, c.Direction, c.SpanStart, c.Width); ok {
				// a successful middle node moves the lit lane, gesture or not
				n.moveLane(c.SpanStart, c.Width, now)
				n.judgeNode(n.Cursor, g, timing, err, endType(c.Kind), 1, now)
				n.OX, n.OY = p.X, p.Y
			}
		}
		switch n.tailBlock(p, now) {
		case tailRelease:
			return true
		case tailNone:
			return false
		}
		// tailTerminal: the terminal is current now, checked in this frame
	}
	n.terminalMoved(p, now)
	return false
}

// tailBlock outcomes.
const (
	tailNone     = 0
	tailTerminal = 1 // the terminal is now current: check its gesture
	tailRelease  = 2 // a terminal flick completed with a node still pending
)

// tailBlock is the original's tail-gesture block, which every Moved on an
// intermediate node also reaches (ExecTouchMoved 0x321CB48..0x321CB78 with
// judgement adjust B = 0), gesture node or not: when the
// node just passed was the last before the terminal, the terminal gesture is
// checked in the same frame; otherwise the flick-family terminal's contact
// and origin advance, and a completed flick releases the Slide with a Miss
// on the node still pending (÷5 when intermediate). A Directional terminal
// resets its origin here, so only a Flick terminal can complete this way.
func (n *PlaySlideBase) tailBlock(p sonolus.Vec2, now float64) float64 {
	if n.JudgeDone || n.Cursor.Index <= 0 {
		return tailNone
	}
	last := n.Last.Get()
	if last.Hidden || last.Kind == NodePlain || last.Judged {
		return tailNone
	}
	if n.Cursor.Index == n.Last.Index {
		return tailTerminal
	}
	if last.Kind != NodeFlick {
		n.OX, n.OY = p.X, p.Y
	}
	tau := slideTau(last.Time, now)
	if tau > 0 {
		n.OX, n.OY = p.X, p.Y
	}
	if insideSpan(p, last.SpanStart, last.Width, false) {
		n.Grace = leaveGrace
	} else {
		n.Grace -= frameStep()
	}
	if tau <= 0 && n.Grace > 0 && gestureDone(p.X-n.OX, p.Y-n.OY, gestureDir(last.Kind, last.Direction), gestureCount(last.Kind, last.KindWidth)) {
		return tailRelease
	}
	return tailNone
}

// terminalMoved is a Moved / Stationary frame with the cursor on the
// terminal node.
func (n *PlaySlideBase) terminalMoved(p sonolus.Vec2, now float64) {
	c := n.Cursor.Get()
	if c.Kind == NodePlain {
		return // a plain end is judged on release only
	}
	if n.Cached >= 0 {
		// picked-up end: no VPL or grace requirement
		if cachedFlickDone(p.X-n.OX, p.Y-n.OY, c.Kind, c.KindWidth) {
			n.judgeNode(n.Cursor, n.Cached, n.CachedTiming, n.CachedError, endType(c.Kind), 1, now)
		}
		return
	}
	tau := slideTau(c.Time, now)
	if tau > 0 {
		n.OX, n.OY = p.X, p.Y
	}
	if insideSpan(p, c.SpanStart, c.Width, false) {
		n.Grace = leaveGrace
	} else {
		n.Grace -= frameStep()
	}
	if tau <= 0 && n.Grace > 0 && gestureDone(p.X-n.OX, p.Y-n.OY, gestureDir(c.Kind, c.Direction), gestureCount(c.Kind, c.KindWidth)) {
		n.judgeNode(n.Cursor, gradePerfect, timingNone, now-c.Time, endType(c.Kind), 1, now)
	}
}

// moveLane moves the lit lane of the owning finger to a node's receiver.
func (n *PlaySlideBase) moveLane(spanStart, width, now float64) {
	lane := math.Round(receiverLane(spanStart, width))
	old, ok := Input.TouchLane.GetOK(n.Finger)
	if ok && old != lane {
		laneOff(old)
	}
	laneOn(lane)
	if ok || !Input.TouchLane.IsFull() {
		Input.TouchLane.Set(n.Finger, lane)
	}
}

// release is ExecTouchEnded of the owning finger.
func (n *PlaySlideBase) release(p sonolus.Vec2, t, now float64) {
	n.Finger = -1
	n.stopFeedback()
	if n.OnHead {
		// gesture head released before its gesture: a non-final node's
		// release is onMiss (type 8), as for a middle node
		n.Grade, n.Error = gradeMiss, missSeconds
		n.HeadDone = true
		n.State = stDone
		lane := n.headLane()
		applyJudgement(now, judgement{Width: n.Width, Grade: gradeMiss, Type: jtSlide, X: laneX(lane), Lane: lane,
			Key: n.Time, Count: gestureCount(n.Kind, n.KindWidth), Direction: n.Direction, Divisor: 1, Mid: true, Slide: true})
		n.Broken = 0
		n.leaveHead()
		return
	}
	c := n.Cursor.Get()
	if !c.Terminal {
		n.Broken = c.Index // everything up to the pending node and out of it
	}
	switch {
	case !c.Terminal:
		// onMiss (type 8) on any non-final node
		n.judgeNode(n.Cursor, gradeMiss, timingNone, missSeconds, jtSlide, 5, now)
	case c.Kind != NodePlain:
		n.judgeNode(n.Cursor, gradeMiss, timingNone, missSeconds, endType(c.Kind), 1, now)
	case insideSpan(p, c.SpanStart, c.Width, true):
		tau := slideTau(c.Time, t)
		g := slideGrade(tau)
		timing := timingNone
		if g == gradeNone {
			g = gradeMiss
		} else if g != gradePerfect && tau > 0 {
			timing = timingFast
		}
		err := t - c.Time
		if g == gradeMiss {
			err = missSeconds
		}
		n.judgeNode(n.Cursor, g, timing, err, jtSlide, 1, now)
	default:
		n.judgeNode(n.Cursor, gradeMiss, timingNone, missSeconds, jtSlide, 1, now)
	}
}

// -------------------------------------------------------------- drawing ----

func (n *PlaySlideBase) UpdateParallel() {
	if n.Ended {
		if !n.Hidden {
			writeResult(n.Grade, n.Error, Buckets.SlideHead)
		}
		play.Entity.SetDespawn(true)
		return
	}
	now := play.Time.Now()
	n.drawBody(now)
	if !n.Hidden && !n.HeadDone {
		pr := n.progress(now)
		if n.Feedback {
			pr = math.Max(pr, 1) // a touched-down gesture head is placed on the line
		}
		if pr >= 0 {
			// an unheld head snaps onto the line at the target-button scale
			// and, like a held one, rides the contact point (slidingMove runs
			// in Wait and Stop alike) until its Miss hides it
			l := n.look()
			l.Stopped = pr >= 1
			lane := n.Lane
			if l.Stopped {
				lane = n.anchorLane()
			}
			drawTerminal(l, n.Kind, n.KindWidth, lane, math.Min(curveOf(pr), 1), now-n.SpawnTime, n.z())
		}
	}
	if n.Feedback {
		lane := n.anchorLane()
		p := heldPos(lane)
		chain := n.chainSprite(now)
		body, flash, px := slideHoldSprites(n.Width, lane)
		// the held body replaces a gesture head only once it is judged
		drawHoldFlash(body, flash, p, px, !n.Hidden && n.HeadDone && chain == chainHead, now-n.FlashStart, n.z())
		if chain == chainPassed {
			// a passed middle node rides as a plain one: note_slide_among only
			drawAmong(lane, 1, n.Passed.Get().Width, NodePlain, 0, true, n.z())
		}
		moveFx(fxKeep, p.X, n.Keep, n.KeepPlane, n.KeepPlane2)
	}
}

// anchorLane is where the front's sprites sit: the contact point, less the
// front's exit offset while its gesture sprite is still shown, so a wide
// Directional's arrow does not jump when the line goes on from its far end.
func (n *PlaySlideBase) anchorLane() float64 {
	if !n.HasContact {
		return n.Lane
	}
	lane := n.ContactLane
	if n.SegHead {
		if !n.HeadDone {
			lane -= n.Exit
		}
	} else if f := n.Seg.Get(); !f.Judged {
		lane -= f.Exit
	}
	return lane
}

// chainSprite follows the original's hide-before: passing a node hides the
// one before it (an invisible node too, NoteSlide.killFromInvisibleNotesToVisibleNote),
// so the held head body shows until the first middle node is passed, then
// the last passed node rides until the next visible node stops.
func (n *PlaySlideBase) chainSprite(now float64) float64 {
	from, chain := n.First, chainHead
	if n.Passed.Index > 0 {
		from, chain = n.Passed.Get().Next, chainPassed
	}
	for k := 0; k < 4096 && from.Index > 0; k++ {
		d := from.Get()
		if !d.Hidden {
			if chain == chainPassed && now >= d.stopTime() {
				return chainNone
			}
			return chain
		}
		if now >= d.Time {
			return chainNone // an invisible node passed
		}
		from = d.Next
	}
	return chain
}

// stopTime is when a node stops on the line: a middle node on the real
// line, the terminal on the virtual perfect line.
func (n *PlaySlideNodeBase) stopTime() float64 {
	if n.Terminal {
		return n.Time + Layout.VPLDelay
	}
	return n.Time
}

// drawBody draws the live segments: from the contact point (or the head) to
// each following node until the first one that has not appeared yet. The
// segments out of a missed node and every one before it are gone (Broken).
func (n *PlaySlideBase) drawBody(now float64) {
	alpha := bodyAlpha()
	if n.AllHidden {
		alpha *= 0.5
	}
	z := lineZ(zBelt, n.Time)
	var fp sonolus.Vec2
	var fh float64
	next := n.First
	front := 0.0 // order of the segment's front node
	if n.HasContact {
		fp, fh = heldPos(n.ContactLane), heldHalf(n.ContactWidth)
		if !n.SegHead {
			next = n.Seg.Get().Next
			front = n.Seg.Get().Index
		}
	} else {
		pr := n.progress(now)
		if pr < 0 {
			return
		}
		fp = notePos(n.Lane+n.Exit, math.Min(curveOf(pr), 1))
		fh = bodyHalf(fp.Y, n.Width)
		if pr >= 1 {
			fh = heldHalf(n.Width) // snapped on the line
		}
	}
	for k := 0; k < 64 && next.Index > 0; k++ {
		a := next.Get()
		pr := a.progress(now)
		c := 0.0
		if pr >= 0 {
			c = math.Min(curveOf(pr), 1)
		}
		ap := notePos(a.Lane, c)
		ah := bodyHalf(ap.Y, a.Width)
		switch {
		case pr < 0:
			ah = waitHalf(a.Width) // not appeared yet: pinned at the spawn row
		case pr >= 1:
			ah = heldHalf(a.Width)
		}
		if front > n.Broken {
			drawBelt(fp, fh, ap, ah, n.Curve, z, alpha)
		}
		if pr < 0 {
			return
		}
		fp, fh = notePos(a.Lane+a.Exit, c), ah
		front = a.Index
		next = a.Next
	}
}

// ------------------------------------------------------------- sync line ----

// PlaySimLine joins two simultaneous notes while both are approaching.
type PlaySimLine struct {
	play.Archetype `archetype:"name=SimLine"`
	A              sonolus.EntityRef[PlayNoteBase] `archetype:"imported,name=a"`
	B              sonolus.EntityRef[PlayNoteBase] `archetype:"imported,name=b"`
	ALane          float64                         `archetype:"imported,name=aLane,default=-1000"` // drawing lane of each end
	BLane          float64                         `archetype:"imported,name=bLane,default=-1000"`
	SpeedMin       float64                         `archetype:"imported,name=speedMin,default=0"` // note speeds (tenths) the line exists at
	SpeedMax       float64                         `archetype:"imported,name=speedMax,default=1000"`
	ANoMargin      bool                            `archetype:"imported,name=aNoMargin"` // directional-note end: no edge margin
	BNoMargin      bool                            `archetype:"imported,name=bNoMargin"`
	SpawnTime      float64                         `archetype:"data"`
	EndTime        float64                         `archetype:"data"`
}

func (l *PlaySimLine) Preprocess() {
	a, b := l.A.Get(), l.B.Get()
	l.ALane, l.BLane = simLane(l.ALane, a.Lane), simLane(l.BLane, b.Lane)
	l.SpawnTime = math.Max(spawnTimeOf(a.Beat, a.Group, a.AxisHit), spawnTimeOf(b.Beat, b.Group, b.AxisHit))
	l.EndTime = math.Min(play.Time.BeatToTime(a.Beat), play.Time.BeatToTime(b.Beat))
}
func (l *PlaySimLine) SpawnOrder() float64 {
	if !Config.SyncLine || !syncLineAtSpeed(l.SpeedMin, l.SpeedMax) {
		return 1e9 // never spawns: keep it out of the way of the queue
	}
	return l.SpawnTime
}
func (l *PlaySimLine) ShouldSpawn() bool {
	return Config.SyncLine && syncLineAtSpeed(l.SpeedMin, l.SpeedMax) && play.Time.Now() >= l.SpawnTime
}
func (l *PlaySimLine) UpdateParallel() {
	now := play.Time.Now()
	a, b := l.A.Get(), l.B.Get()
	if now > l.EndTime+missSeconds || a.State == stDone || b.State == stDone {
		play.Entity.SetDespawn(true)
		return
	}
	if a.State != stMove || b.State != stMove {
		return
	}
	pa, pb := a.progress(now), b.progress(now)
	if pa < 0 || pb < 0 || pa > 1 || pb > 1 {
		return
	}
	drawSimLine(notePos(l.ALane, curveOf(pa)), notePos(l.BLane, curveOf(pb)), math.Min(a.Time, b.Time), l.ANoMargin, l.BNoMargin)
}

// headLane is where the head's judgement plays: a directional gesture head at
// its flick-start end, otherwise the receiver.
func (n *PlaySlideBase) headLane() float64 {
	if n.Kind == NodeDirectional || n.Kind == NodeMultiDir {
		return dirBodyLane(n.SpanStart, n.Width, n.Direction)
	}
	return receiverLane(n.SpanStart, n.Width)
}

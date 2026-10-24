//go:build tutorial

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus/tutorial"
)

// The tutorial loops one demonstration per phase: Tap, Flick, Directional
// flick, Long and Slide, each arriving at tutHit seconds into the phase.

type TutorialState struct {
	sonolus.LevelMemoryResource
	Phase    float64
	Start    float64
	Previous float64
	Loop     float64 // hold sound main / intro handle IDs
	Intro    float64
	Keep     sonolus.ParticleHandle
	KeepP    sonolus.ParticleHandle // the lane-plane TapKeep parts
	KeepQ    sonolus.ParticleHandle
}

var Tut = TutorialState{}

type TutorialGlobals struct{ tutorial.GlobalCallbacks }

var Global TutorialGlobals

const tutPhases = 5

func tutHit() float64    { return Layout.Arrival + 0.3 }
func tutPeriod() float64 { return tutHit() + 2.2 }

func Preprocess() {
	initLayout(tutorial.Screen.Rect(), Config.NoteSpeed, Config.NoteSize, Config.Mirror, false, -1)
	initHoldSE()
	initFx()
	tutorial.Background.Set(tutBackgroundQuad())
	Tut.Phase = 0
	Tut.Start = tutorial.Time.Now()
	Tut.Previous = -1
	screen := tutorial.Screen.Rect()
	c := uiTopRight(-42, 54)
	tutorial.UI.SetMenu(sonolus.RuntimeUIBasicLayout{
		Anchor: c, Pivot: sonolus.Vec2{X: 0.5, Y: 0.5},
		Size: sonolus.Vec2{X: 68 * Layout.U, Y: 70 * Layout.U}, Alpha: 1, Background: true,
	})
	nav := tutorial.UI.NavigationConfiguration()
	tutorial.UI.SetPrevious(sonolus.RuntimeUIBasicLayout{
		Anchor: sonolus.Vec2{X: screen.L + 0.05, Y: 0}, Pivot: sonolus.Vec2{X: 0, Y: 0.5},
		Size: sonolus.Vec2{X: 0.15 * nav.Scale, Y: 0.15 * nav.Scale}, Alpha: nav.Alpha, Background: true,
	})
	tutorial.UI.SetNext(sonolus.RuntimeUIBasicLayout{
		Anchor: sonolus.Vec2{X: screen.R - 0.05, Y: 0}, Pivot: sonolus.Vec2{X: 1, Y: 0.5},
		Size: sonolus.Vec2{X: 0.15 * nav.Scale, Y: 0.15 * nav.Scale}, Alpha: nav.Alpha, Background: true,
	})
	ins := tutorial.UI.InstructionConfiguration()
	tutorial.UI.SetInstruction(sonolus.RuntimeUIBasicLayout{
		Anchor: sonolus.Vec2{X: 0, Y: 0.25}, Pivot: sonolus.Vec2{X: 0.5, Y: 0.5},
		Size: sonolus.Vec2{X: 1.2 * ins.Scale, Y: 0.15 * ins.Scale}, Alpha: ins.Alpha, Background: true,
	})
}

func Navigate() {
	p := Tut.Phase
	if tutorial.Navigation.Direction() > 0 {
		p++
	} else {
		p--
	}
	tutStopHold()
	Tut.Phase = math.Mod(p+tutPhases, tutPhases)
	Tut.Start = tutorial.Time.Now()
	Tut.Previous = -1
}

func tutStopHold() {
	if Tut.Loop > 0 {
		stopHoldSound(Tut.Loop, Tut.Intro, tutorial.Time.Now(), true)
		Tut.Loop, Tut.Intro = 0, 0
	}
	destroyFx(Tut.Keep, Tut.KeepP, Tut.KeepQ)
	Tut.Keep.ID, Tut.KeepP.ID, Tut.KeepQ.ID = 0, 0, 0
}

// crossed reports whether phase time passed at since the previous update.
func crossed(prev, now, at float64) bool { return prev < at && now >= at }

// tutCurve is the motion curve of a note hitting at h, or −1 off screen.
func tutCurve(t, h float64, stop bool) float64 {
	pr := 1 - (h-t)/Layout.Arrival
	if pr < 0 {
		return -1
	}
	c := curveOf(pr)
	if stop {
		c = math.Min(c, 1)
	}
	return c
}

func tutHand(x, y, t float64) {
	a := 1.0
	if t < 0 {
		a = 0
	}
	tutorial.Instruction.Paint(InstructionIcons.Hand, sonolus.Vec2{X: x, Y: y - 0.12}, 0.25, 0, 100, a)
}

func Update() {
	now := tutorial.Time.Now()
	t := math.Mod(now-Tut.Start, tutPeriod())
	prev := Tut.Previous
	if t < prev {
		tutStopHold()
		prev = -1
		for i := 0; i < 7; i++ {
			laneOff(float64(i))
		}
	}
	Tut.Previous = t
	updateLaneReserve(now)
	tutorial.Background.Set(tutBackgroundQuad())
	drawField(now, 1)
	drawLaneEffects(now, zLaneEffect)
	h := tutHit()
	z := noteZ(zNote, 0, 0)
	g := Layout.G
	switch Tut.Phase {
	case 0: // Tap
		if c := tutCurve(t, h, false); c >= 0 && t < h {
			drawTap(noteLook{SpriteLane: 3, Lane: 3, Width: 1, SpanStart: 3}, false, false, c, z)
		}
		if crossed(prev, t, h) {
			tutJudge(laneX(3), jtNormal)
		}
		tutText(t, h, Instructions.Tap)
		if math.Abs(t-h) < 0.4 {
			tutHand(laneX(3), g, t)
		}
	case 1: // Flick
		if c := tutCurve(t, h, false); c >= 0 && t < h {
			drawFlick(noteLook{SpriteLane: 3, Lane: 3, Width: 1, SpanStart: 3}, c, t, z)
		}
		if crossed(prev, t, h) {
			tutJudge(laneX(3), jtFlick)
		}
		tutText(t, h, Instructions.Flick)
		if math.Abs(t-h) < 0.4 {
			tutHand(laneX(3), g+math.Max(t-h, 0)*1.2, t)
		}
	case 2: // Directional flick
		l := noteLook{SpriteLane: 3, Lane: 3, Width: 1, Direction: 1, SpanStart: 3}
		if c := tutCurve(t, h, false); c >= 0 && t < h {
			drawDirectionalBody(l, c, t, true, false, l.Lane, z)
		}
		if crossed(prev, t, h) {
			tutJudge(laneX(3), jtDirectional)
		}
		tutText(t, h, Instructions.Flick)
		if math.Abs(t-h) < 0.4 {
			tutHand(laneX(3)+math.Max(t-h, 0)*1.2, g, t)
		}
	case 3: // Long
		end := h + 1.2
		held := t >= h && t < end
		if c := tutCurve(t, h, true); c >= 0 && t < end {
			drawLong(noteLook{SpriteLane: 3, Lane: 3, Width: 1, SpanStart: 3}, c, 3, 1, math.Max(tutCurve(t, end, true), 0), held, t-h, z, lineZ(zBelt, 0))
		}
		if c := tutCurve(t, end, true); c >= 0 && t < end {
			drawTerminal(noteLook{SpriteLane: 3, Lane: 3, Width: 1, SpanStart: 3}, NodePlain, 1, 3, c, t, subZ(z, 1))
		}
		if crossed(prev, t, h) {
			tutJudge(laneX(3), jtLongFront)
			tutStartHold(laneX(3))
			laneOn(3) // the held touch lights its lane
		}
		if crossed(prev, t, end) {
			tutStopHold()
			tutJudge(laneX(3), jtLongEnd)
			laneOnReserve(3, now) // the Long end pulses, the release fades it
			playHoldRelease(tutorial.Time.Now(), true)
		}
		if t < h+0.5 {
			tutText(t, h, Instructions.Hold)
		} else {
			tutText(t, end, Instructions.Release)
		}
		if t > h-0.4 && t < end+0.3 {
			tutHand(laneX(3), g, t)
		}
	default: // Slide: lane 2 → 4 → 5
		t1, t2 := h+0.7, h+1.4
		lanes := [3]float64{2, 4, 5}
		times := [3]float64{h, t1, t2}
		contact := 2.0
		if t >= h {
			k := sonolus.Clamp((t-h)/(t1-h), 0, 1)
			contact = 2 + 2*k
			if t >= t1 {
				contact = 4 + sonolus.Clamp((t-t1)/(t2-t1), 0, 1)
			}
		}
		held := t >= h && t < t2
		tutSlideBody(t, h, contact, lanes, times)
		if c := tutCurve(t, h, true); c >= 0 && t < h {
			drawNoteSprite(Skin.NoteLong[2], 2, c, 1, z, 1, false)
		}
		if c := tutCurve(t, t1, true); c >= 0 && t < t1 {
			drawAmong(4, c, 1, NodePlain, 0, false, z)
		}
		if c := tutCurve(t, t2, true); c >= 0 && t < t2 {
			drawTerminal(noteLook{SpriteLane: 5, Lane: 5, Width: 1, SpanStart: 5}, NodePlain, 1, 5, c, t, z)
		}
		if held {
			// Play's hide-before (chainSprite): the held head body shows
			// until the middle node is passed; then that node rides the
			// front as note_slide_among until the hold ends (the terminal
			// stops on the virtual perfect line after t2)
			if t >= t1 {
				drawAmong(contact, cRide, 1, NodePlain, 0, true, z)
			}
			p := heldPos(contact)
			body, flash, px := slideHoldSprites(1, contact)
			drawHoldFlash(body, flash, p, px, t < t1, t-h, subZ(z, 2))
			moveFx(fxKeep, p.X, Tut.Keep, Tut.KeepP, Tut.KeepQ)
		}
		if crossed(prev, t, h) {
			tutJudge(laneX(2), jtSlide)
			tutStartHold(laneX(2))
			laneOn(2) // the touch lights its lane, the Slide's target
		}
		if crossed(prev, t, t1) {
			tutJudge(laneX(4), jtSlide)
			// a passed middle node: Off(old target), On, OffReserve
			laneOff(2)
			laneOnReserve(4, now)
		}
		if crossed(prev, t, t2) {
			tutStopHold()
			tutJudge(laneX(5), jtSlide)
			// the release on the end: Off(old target), On, OffReserve
			laneOff(4)
			laneOnReserve(5, now)
			playHoldRelease(tutorial.Time.Now(), true)
		}
		tutText(t, h, Instructions.Follow)
		if t > h-0.4 && t < t2+0.3 {
			tutHand(laneX(contact), g, t)
		}
	}
}

// tutSlideBody draws the Slide's live segments like Play's drawBody: from the
// held contact point (heldPos / heldHalf) or the approaching head to each
// following node until the first one that has not appeared yet.
func tutSlideBody(t, h, contact float64, lanes, times [3]float64) {
	if t >= times[2] {
		return
	}
	var fp sonolus.Vec2
	var fh float64
	next := 1
	if t >= h {
		fp, fh = heldPos(contact), heldHalf(1)
		if t >= times[1] {
			next = 2
		}
	} else {
		c := tutCurve(t, h, true)
		if c < 0 {
			return
		}
		fp = notePos(lanes[0], c)
		fh = bodyHalf(fp.Y, 1)
	}
	for i := next; i < 3; i++ {
		c := tutCurve(t, times[i], true)
		ap := notePos(lanes[i], math.Max(c, 0))
		ah := bodyHalf(ap.Y, 1)
		switch {
		case c < 0:
			ah = waitHalf(1) // not appeared yet: pinned at the spawn row
		case c >= 1:
			ah = heldHalf(1)
		}
		drawBelt(fp, fh, ap, ah, false, lineZ(zBelt, 0), bodyAlpha())
		if c < 0 {
			return
		}
		fp, fh = ap, ah
	}
}

// tutBackgroundQuad is Play's backgroundQuad (stage_fx.go, play/watch only)
// at the opening's end pose (Stage scale 0.92, y 0), which the tutorial
// keeps throughout.
func tutBackgroundQuad() sonolus.Quad {
	k := (1 - 0.23*Layout.HAR) * Layout.UB
	cy := 14 * Layout.HAR * Layout.UB
	return sonolus.Rect{T: cy + 820*k, R: 883.2 * k, B: cy - 820*k, L: -883.2 * k}.ToQuad()
}

func tutText(t, h float64, text sonolus.Text) {
	if t > h-0.8 && t < h+0.8 {
		tutorial.Instruction.Show(text)
	} else {
		tutorial.Instruction.Clear()
	}
}

func tutStartHold(x float64) {
	if Config.SFX && Tut.Loop <= 0 {
		Tut.Loop, Tut.Intro = startHoldSound(tutorial.Time.Now(), true)
	}
	if Config.NoteEffects && !fxAlive(Tut.Keep, Tut.KeepP, Tut.KeepQ) {
		Tut.Keep, Tut.KeepP, Tut.KeepQ = spawnFx(fxKeep, x, durKeep, durKeepPlane, true)
	}
}

// tutJudge plays a Perfect judgement's SE and particles.
func tutJudge(x, jt float64) {
	switch jt {
	case jtNormal, jtFlick, jtDirectional:
		// a quick tap lights its lane and fades (as Auto does)
		laneOnReserve(math.Round(x/Layout.W+3), tutorial.Time.Now())
	}
	if Config.SFX {
		switch jt {
		case jtFlick:
			Effects.Flick.Play(0)
		case jtDirectional:
			Effects.Directional[0].Play(0)
		default:
			Effects.Perfect.Play(0)
		}
	}
	if !Config.NoteEffects {
		return
	}
	switch jt {
	case jtFlick:
		spawnFx(fxSwipe, x, durSwipe, durSwipe, false)
	case jtDirectional:
		dirEffect(true, 0).Spawn(particleQuad(x, 1), durDir1, false)
		spawnFx(dirFinger(true), x, durFinger, durFinger, false)
	default:
		spawnFx(fxPerfect, x, durPerfect, durPerfect, false)
	}
}

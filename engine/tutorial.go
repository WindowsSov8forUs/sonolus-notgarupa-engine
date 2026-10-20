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
	}
	Tut.Previous = t
	drawField(now, 1)
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
		}
		if crossed(prev, t, end) {
			tutStopHold()
			tutJudge(laneX(3), jtLongEnd)
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
		var fp sonolus.Vec2
		if t >= h {
			fp = notePos(contact, 1)
		} else if c := tutCurve(t, h, true); c >= 0 {
			fp = notePos(2, c)
		}
		for i := 1; i < 3; i++ {
			if t >= times[i] || tutCurve(t, h, true) < 0 {
				continue
			}
			c := math.Max(tutCurve(t, times[i], true), 0)
			ap := notePos(lanes[i], c)
			front := fp
			if t < times[i-1] {
				front = notePos(lanes[i-1], math.Max(tutCurve(t, times[i-1], true), 0))
			}
			drawBelt(front, bodyHalf(front.Y, 1), ap, bodyHalf(ap.Y, 1), false, lineZ(zBelt, 0), bodyAlpha())
		}
		if c := tutCurve(t, h, true); c >= 0 && t < h {
			drawNoteSprite(Skin.NoteLong[2], 2, c, 1, z, 1, false)
		}
		if c := tutCurve(t, t1, true); c >= 0 && t < t1 {
			p := notePos(4, c)
			Skin.NoteSlideAmong.DrawZ(noteQuad(p, noteScale(p.Y, 1)), z, 1)
		}
		if c := tutCurve(t, t2, true); c >= 0 && t < t2 {
			drawTerminal(noteLook{SpriteLane: 5, Lane: 5, Width: 1, SpanStart: 5}, NodePlain, 1, 5, c, t, z)
		}
		if held {
			drawHoldFlash(Skin.NoteLong[3], Skin.NoteLongFlash[3], heldPos(contact), noteSpritePx, true, t-h, subZ(z, 2))
			moveFx(fxKeep, laneX(contact), Tut.Keep, Tut.KeepP, Tut.KeepQ)
		}
		if crossed(prev, t, h) {
			tutJudge(laneX(2), jtSlide)
			tutStartHold(laneX(2))
		}
		if crossed(prev, t, t1) {
			tutJudge(laneX(4), jtSlide)
		}
		if crossed(prev, t, t2) {
			tutStopHold()
			tutJudge(laneX(5), jtSlide)
			playHoldRelease(tutorial.Time.Now(), true)
		}
		tutText(t, h, Instructions.Follow)
		if t > h-0.4 && t < t2+0.3 {
			tutHand(laneX(contact), g, t)
		}
	}
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
		Particles.DirR[0].Spawn(particleQuad(x, 1), durDir1, false)
		spawnFx(fxFingerR, x, durFinger, durFinger, false)
	default:
		spawnFx(fxPerfect, x, durPerfect, durPerfect, false)
	}
}

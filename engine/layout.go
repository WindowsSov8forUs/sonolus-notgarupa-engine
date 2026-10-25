package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
)

// Original layout constants (GameCamera world units = Sonolus screen units).
const (
	referenceScreenX = 9.578571319580078
	aspectBase       = 1.778666615486145
	highAspectMax    = 2.0
	androidSafe      = 0.8999999761581421
	buttonSpacing    = 2.2
	buttonY          = -3.45
	launcherY        = 5.42
	launchRate       = 0.05
	curveBase        = 1.100000023841858
	curveScale       = 50.0
	uiWidth          = 1334.0

	noteSpritePx  = 308.0 // note body sprites, PPU 100
	noteSpritePy  = 120.0
	notePPU       = 100.0
	judgeLinePPU  = 69.0
	laneEffectPPU = 69.0
)

// The generated background is the bgImage widget laid out at stage scale
// 0.92 on a canvas 1766.4 UI units wide from bgCanvasTop down to −820 (1537 of
// the composed 1901 px rows over 1640 units; the black rows above the widget
// are cut, the background colour shows there).
const bgCanvasTop = 820 - 364*1640/1901.0

var scaleMinRatio = [7]float64{0.98, 0.988, 0.9898, 0.9899, 0.991, 0.9915, 0.9917}

// LayoutData holds the field geometry computed once in preprocess.
type LayoutData struct {
	sonolus.LevelDataResource
	Aspect  float64
	Rate    float64 // s = aspect / 9.5786
	HAR     float64 // high-aspect ratio 0..1
	Safe    float64
	M       float64 // MusicObjects scale
	MLines  float64 // RhythmGameLines scale
	P       float64 // NoteParent scale
	S       float64 // s·m
	G       float64 // judge line y
	W       float64 // lane spacing at the judge line
	V       float64 // vanishing y
	Y0      float64 // spawn row y
	Ly      float64 // launcher y for the scale formula
	N       float64 // note size factor
	Arrival float64 // seconds from launch to the judge line
	U       float64 // world units per UI unit (incl. safe-area factor)
	UK      float64 // UI safe factor k
	Screen  sonolus.Rect
	SafeL   float64
	SafeR   float64
	Mirror  bool
	// Slide virtual perfect line delay: first 1/60 s sample at/after arrival.
	VPLDelay float64
	// Collision radius² of a lane button and its "sweet" variant.
	Collision2 float64
	FlickWorld float64 // 0.04 rate in world units
	DirWorld   float64 // 0.01 rate in world units
	// TargetLocal is the target-button local scale rate·N: held, stopped and
	// snapped sprites on the line use it instead of the moving-note scale.
	TargetLocal float64
	// WaitLocal is the virtual local scale of a body end whose node has not
	// appeared yet (getOrdinaryNoteMeshAfterScale, phase wait).
	WaitLocal float64
	// HeldK maps a held Slide's contact point from button-local space (P/M).
	HeldK float64
	// UB is the background UI unit 2/activeHeight (integer activeHeight).
	UB float64
	// Sudden (original live setting): ratio, clip line of bands and lines,
	// mask of note sprites [MaskB, MaskT].
	SuddenR     float64
	SuddenClipY float64
	SuddenMaskB float64
	SuddenMaskT float64
	// HABAHIRO: the level is a #HABAHIRO chart; HabaSkin: the skin has its
	// sprites; LaneChange: time of the lane change (−1: none).
	Habahiro   bool
	HabaSkin   bool
	HabaFx     bool // the particle item has the HABAHIRO width variants
	LaneChange float64
}

// roundHalfEvenInt is Mathf.RoundToInt (ties to even).
func roundHalfEvenInt(v float64) float64 {
	f := math.Floor(v)
	switch d := v - f; {
	case d > 0.5:
		return f + 1
	case d < 0.5:
		return f
	}
	return f + math.Mod(f, 2)
}

// rhythmLinesY converts a RhythmGameLines-local y to world y.
func rhythmLinesY(l float64) float64 { return (13*Layout.HAR + l*Layout.M) * Layout.UB }

// suddenHides reports whether the Sudden cover hides a note sprite centred at
// height y (Sonolus has no masks: the whole sprite hides).
func suddenHides(y float64) bool { return Layout.SuddenR > 0 && y > Layout.SuddenMaskB }

var Layout = LayoutData{}

func arrivalSeconds(speed float64) float64 {
	if speed <= 11.01 {
		return (speed-1)*-0.5 + 5.5
	}
	return (speed-11)/-10 + 0.5
}

func initLayout(screen sonolus.Rect, noteSpeed, noteSize float64, mirror, habahiro bool, laneChange float64) {
	if habahiro {
		noteSize = sonolus.Clamp(noteSize, 0.8, 1.5) // multi-range charts clamp the note size
	}
	Layout.Habahiro = habahiro
	Layout.HabaSkin = habahiro && Skin.HabaNormal[0].Exists()
	Layout.HabaFx = habahiro && Particles.PerfectW[0].Exists()
	Layout.LaneChange = laneChange
	aspect := (screen.R - screen.L) / (screen.T - screen.B)
	s := aspect / referenceScreenX
	har := sonolus.Clamp((aspect-aspectBase)/(highAspectMax-aspectBase), 0, 1)
	safe := 1.0
	uk := 1.0
	if aspect > aspectBase {
		safe = androidSafe * aspectBase / aspect
		uk = safe
	}
	m := safe * (1 + 0.054*har)
	S := s * m
	yb := buttonY + 0.25*har
	g := yb * S
	v := (yb + 8.870001) * S
	Layout.Aspect = aspect
	Layout.Rate = s
	Layout.HAR = har
	Layout.Safe = safe
	Layout.M = m
	Layout.MLines = m
	Layout.P = safe * (1 + 0.0700000524520874*har)
	Layout.S = S
	Layout.G = g
	Layout.W = buttonSpacing * S
	Layout.V = v
	Layout.Y0 = g + (1-launchRate)*(v-g)
	Layout.Ly = launcherY * S
	Layout.N = noteSize
	Layout.Arrival = arrivalSeconds(noteSpeed)
	// the original UIRoot: FitWidth with an integer activeHeight
	Layout.U = 2 / roundHalfEvenInt(uiWidth/aspect) * uk
	Layout.UK = uk
	Layout.Screen = screen
	halfSafe := aspect * safe / uk
	if aspect > aspectBase {
		halfSafe = aspect * androidSafe
	}
	Layout.SafeL = -halfSafe
	Layout.SafeR = halfSafe
	Layout.Mirror = mirror
	frames := math.Ceil(Layout.Arrival*60 - 1e-9)
	Layout.VPLDelay = frames/60 - Layout.Arrival
	if aspect > aspectBase {
		w := Layout.W * 1.1679999828338623
		Layout.Collision2 = w * w
	} else {
		Layout.Collision2 = 0.23136097192764282
	}
	Layout.FlickWorld = 0.04 * uiWidth / 750
	Layout.DirWorld = 0.01 * uiWidth / 750
	Layout.TargetLocal = s * noteSize
	Layout.WaitLocal = s * noteSize * math.Abs(Layout.Ly/safe-Layout.Y0/Layout.P) / math.Abs((Layout.Ly-g)/safe)
	Layout.HeldK = Layout.P / m
	Layout.UB = 2 / roundHalfEvenInt(uiWidth/aspect)
	Layout.SuddenR = 0
	if Config.SuddenRate > 0 {
		Layout.SuddenR = 0.05 + 0.95*Config.SuddenRate/100
	}
	r := Layout.SuddenR
	Layout.SuddenClipY = rhythmLinesY(370 - 610*r)
	Layout.SuddenMaskT = rhythmLinesY(368.39 - 0.1286*r)
	Layout.SuddenMaskB = rhythmLinesY(368.39 - 608.375*r)
}

// mirrorLane maps a lane through the mirror option.
func mirrorLane(lane float64) float64 {
	if Layout.Mirror {
		return 6 - lane
	}
	return lane
}

// curveOf converts approach progress (0 = launch, 1 = judge line) to the
// original exponential motion curve.
func curveOf(progress float64) float64 {
	return math.Pow(curveBase, (progress-1)*curveScale)
}

// notePos returns the screen position of a lane centre at motion curve c.
func notePos(lane, c float64) sonolus.Vec2 {
	k := launchRate + (1-launchRate)*c
	return sonolus.Vec2{X: (lane - 3) * Layout.W * k, Y: Layout.Y0 - c*(Layout.Y0-Layout.G)}
}

// laneX returns the judge-line x of a lane centre.
func laneX(lane float64) float64 { return (lane - 3) * Layout.W }

// noteScale returns the world scale of a note at height y.
func noteScale(y float64, buttons float64) float64 {
	vr := Layout.Rate * Layout.N * math.Abs(Layout.Ly-y) / math.Abs(Layout.Ly-Layout.G)
	idx := int(sonolus.Clamp(buttons, 1, 7)) - 1
	a := Layout.HAR*(scaleMinRatio[idx]-0.996) + 0.996
	return (vr*a + (1 - a)) * Layout.P
}

// noteLocalScale returns the note's local scale (without NoteParent), used
// for the body mesh half width and line thickness.
func noteLocalScale(y float64, buttons float64) float64 {
	return noteScale(y, buttons) / Layout.P
}

// targetScale is the world scale of a sprite held, stopped or snapped on the
// line (target-button scale).
func targetScale() float64 { return Layout.TargetLocal * Layout.P }

// heldPos is a held Slide's contact point: the head moves in button-local
// space under NoteParent, i.e. (x, G)·P/M.
func heldPos(lane float64) sonolus.Vec2 {
	return sonolus.Vec2{X: laneX(lane) * Layout.HeldK, Y: Layout.G * Layout.HeldK}
}

// Body mesh half widths: moving, held / stopped on the line, waiting at the
// spawn row (calculateNoteMeshHalfWidth: local·width·safe).
// HABAHIRO widens wide bodies (the original grid width rate from HighAspectRatio).
func bodyHalf(y, width float64) float64 {
	return noteLocalScale(y, width) * width * Layout.Safe * widthRate(width)
}
func heldHalf(width float64) float64 {
	return Layout.TargetLocal * width * Layout.Safe * widthRate(width)
}
func waitHalf(width float64) float64 {
	return Layout.WaitLocal * width * Layout.Safe * widthRate(width)
}

func widthRate(width float64) float64 {
	switch {
	case !Layout.Habahiro || width <= 1:
		return 1
	case width == 2:
		return 1.0499999523162842
	}
	return 1.05 + 0.03000009059906006*Layout.HAR
}

// spriteQuad returns an axis-aligned quad of a pixel sprite centred on p.
func spriteQuad(p sonolus.Vec2, px, py, ppu, scale float64) sonolus.Quad {
	hw, hh := px/ppu*scale/2, py/ppu*scale/2
	return sonolus.Rect{T: p.Y + hh, R: p.X + hw, B: p.Y - hh, L: p.X - hw}.ToQuad()
}

// noteQuad is the quad of a note body sprite at position p with world scale.
func noteQuad(p sonolus.Vec2, scale float64) sonolus.Quad {
	return spriteQuad(p, noteSpritePx, noteSpritePy, notePPU, scale)
}

// beltQuad builds the straight body segment between two endpoints with the
// given half widths (front = earlier node, lower on screen).
func beltQuad(front, after sonolus.Vec2, frontHalf, afterHalf float64) sonolus.Quad {
	return sonolus.Quad{
		BL: sonolus.Vec2{X: front.X - frontHalf, Y: front.Y},
		TL: sonolus.Vec2{X: after.X - afterHalf, Y: after.Y},
		TR: sonolus.Vec2{X: after.X + afterHalf, Y: after.Y},
		BR: sonolus.Vec2{X: front.X + frontHalf, Y: front.Y},
	}
}

// lineQuad builds a constant-width line between two points (LineRenderer).
func lineQuad(a, b sonolus.Vec2, half float64) sonolus.Quad {
	d := b.Sub(a)
	l := d.Magnitude()
	n := sonolus.Vec2{X: 0, Y: half}
	if l > 1e-9 {
		n = sonolus.Vec2{X: -d.Y / l * half, Y: d.X / l * half}
	}
	return sonolus.Quad{BL: a.Sub(n), TL: a.Add(n), TR: b.Add(n), BR: b.Sub(n)}
}

// uiPoint converts centred UI units (y up) to screen coordinates.
func uiPoint(x, y float64) sonolus.Vec2 {
	return sonolus.Vec2{X: x * Layout.U, Y: y * Layout.U}
}

// uiTopLeft converts root-local (x, y_down) UI units of an element anchored at
// the safe-area top-left corner.
func uiTopLeft(x, yDown float64) sonolus.Vec2 {
	return sonolus.Vec2{X: Layout.SafeL + x*Layout.U, Y: Layout.Screen.T - yDown*Layout.U}
}

// uiTopRight converts root-local (x ≤ 0, y_down) UI units anchored at the
// safe-area top-right corner.
func uiTopRight(x, yDown float64) sonolus.Vec2 {
	return sonolus.Vec2{X: Layout.SafeR + x*Layout.U, Y: Layout.Screen.T - yDown*Layout.U}
}

// uiRect returns a quad of w×h UI units centred on p.
func uiRect(p sonolus.Vec2, w, h float64) sonolus.Quad {
	hw, hh := w*Layout.U/2, h*Layout.U/2
	return sonolus.Rect{T: p.Y + hh, R: p.X + hw, B: p.Y - hh, L: p.X - hw}.ToQuad()
}

func sq(x float64) float64 { return x * x }

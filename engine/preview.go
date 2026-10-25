//go:build preview

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus/preview"
)

// Preview draws the chart as left-to-right columns of pvColumnSeconds each,
// lanes vertical and time upward, in real time (SV is not applied). As in
// the Sonolus Bandori engine a column fills the screen height and a lane is
// 1/20 of it. Each column shows the skin's lane image unprojected into a flat
// strip; notes look as they do on the judge line in Play (target-button
// scale). Left of the strip: whole seconds; right: every 50th combo and the
// BPM changes.

const (
	pvColumnSeconds = 2.0
	pvColumnLanes   = 12.0 // column width in lanes: the strip and two label margins
	pvHeightLanes   = 20.0 // column height in lanes
	// pvStripHalf is the half width of the preview lane strip in lanes: the
	// field image is 1160 UI units wide, a lane 2.2·S = 153.2 units.
	pvStripHalf = 580 / (2.2 * 1334 / (2 * 9.578571319580078))
	pvComboStep = 50.0
)

type PreviewLayout struct {
	sonolus.LevelDataResource
	ColW       float64 // column width
	W          float64 // lane spacing
	Left       float64 // x of the first column's left edge
	Bottom     float64
	Top        float64
	LastTime   float64
	Columns    float64
	LaneChange float64 // time the HABAHIRO lane image is swapped in (−1: none)
}

var PL = PreviewLayout{}

func pvColumn(t float64) float64 { return math.Max(math.Floor(t/pvColumnSeconds), 0) }

// pvCenter is the x of lane 3 in column col.
func pvCenter(col float64) float64 { return PL.Left + (col+0.5)*PL.ColW }

// pvY is the height of time t inside column col.
func pvY(t, col float64) float64 {
	return PL.Bottom + (t-col*pvColumnSeconds)/pvColumnSeconds*(PL.Top-PL.Bottom)
}

// pvPos maps (lane, time) to the canvas.
func pvPos(lane, t float64) sonolus.Vec2 {
	col := pvColumn(t)
	return sonolus.Vec2{X: pvCenter(col) + (lane-3)*PL.W, Y: pvY(t, col)}
}

// pvScale is the world scale of a note sprite on the preview: Play's
// target-button scale rate·N·P relative to the lane spacing 2.2·rate·M.
func pvScale() float64 { return PL.W / buttonSpacing * Layout.N * Layout.P / Layout.M }

// pvLocal is the matching local scale (without NoteParent): line widths.
func pvLocal() float64 { return pvScale() / Layout.P }

// pvHalf is a body's half width: Play's held / stopped mesh half width
// (heldHalf) relative to the lane spacing.
func pvHalf(width float64) float64 {
	return PL.W / buttonSpacing * width * Layout.N * Layout.Safe / Layout.M * widthRate(width)
}

// The original's per-lane note sprites (note_*_<lane>, HABAHIRO runs) are
// drawn for the perspective field: only the centre lane's sprite faces the
// viewer. The flat preview draws every note with the frontal sprite.
const pvSpriteLane = 3

// pvRun is the HABAHIRO run of a note's width centred on lane 3 (an even
// width has no centred run: the one nearer the note, as Play's held Slide
// head in slideHoldSprites).
func pvRun(spanStart, width float64) int {
	w := sonolus.Clamp(width, 1, 7)
	start := math.Floor((7 - w) / 2)
	if math.Mod(w, 2) == 0 && spanStart+(w-1)/2 > 3 {
		start++
	}
	return habaRun(start, w)
}

// Icon offsets: the middle of Play's 1/3 s icon animations.
func pvFlickTopOffset() float64  { return flickTopOffset(1.0 / 6) }
func pvSideArrowOffset() float64 { return sideArrowOffset(1.0 / 6) }

// pvTextH is the label height, one lane (the Bandori engine's screen.h/20).
func pvTextH() float64 { return PL.W }

// pvLabel prints a label beside the strip at time t: left of it (side −1,
// right-aligned) or right of it (side 1, left-aligned), vertically centred on
// t's height and pushed inside the column only at its top and bottom edges.
func pvLabel(value float64, format sonolus.PrintFormat, decimals int, t, side float64, color sonolus.PrintColor) {
	col := pvColumn(t)
	h := pvTextH()
	margin := (pvColumnLanes/2 - pvStripHalf) * PL.W
	align, pivot := sonolus.HorizontalAlignLeft, 0.0
	if side < 0 {
		align, pivot = sonolus.HorizontalAlignRight, 1
	}
	preview.Canvas.Print(preview.PrintOptions{
		Value: value, Format: format, DecimalPlaces: decimals,
		Anchor: sonolus.Vec2{X: pvCenter(col) + side*(pvStripHalf+0.1)*PL.W, Y: sonolus.Clamp(pvY(t, col), PL.Bottom+h/2, PL.Top-h/2)},
		Pivot:  sonolus.Vec2{X: pivot, Y: 0.5},
		Size:   sonolus.Vec2{X: margin - 0.2*PL.W, Y: h},
		Color:  color, Alpha: 1, HorizontalAlign: align,
	})
}

// pvLineHalf is the half thickness of the beat and BPM lines in screen units
// (Next-RUSH's PREVIEW_BAR_LINE_H).
func pvLineHalf() float64 { return 0.0025 }

// pvLine draws a horizontal line at time t from lane offset l to r (lanes
// from the column centre, lane 3).
func pvLine(sp sonolus.Sprite, t, l, r, half float64, z sonolus.ZIndex, a float64) {
	col := pvColumn(t)
	c, y := pvCenter(col), pvY(t, col)
	sp.DrawZ(sonolus.Rect{T: y + half, R: c + r*PL.W, B: y - half, L: c + l*PL.W}.ToQuad(), z, a)
}

func pvNote(sp sonolus.Sprite, lane, t float64, z sonolus.ZIndex) sonolus.Vec2 {
	p := pvPos(lane, t)
	sp.DrawZ(noteQuad(p, pvScale()), z, 1)
	return p
}

// pvWide draws a HABAHIRO wide-note sprite of the given pixel width.
func pvWide(sp sonolus.Sprite, px, lane, t float64, z sonolus.ZIndex) sonolus.Vec2 {
	p := pvPos(lane, t)
	sp.DrawZ(spriteQuad(p, px, noteSpritePy, notePPU, pvScale()), z, 1)
	return p
}

// pvLong draws a Long / Slide head or plain end (HABAHIRO: its lane run).
// A skill head is note_skill like a skill tap; ends never are.
func (n *PreviewNoteBase) pvLong(skill bool, z sonolus.ZIndex) {
	if Layout.HabaSkin {
		run := pvRun(n.SpanStart, n.Width)
		sp := Skin.HabaLong[run]
		if skill {
			sp = Skin.HabaSkill[run]
		}
		pvWide(sp, habaWidthPx(n.Width), n.Lane, n.Time, z)
		return
	}
	sp := Skin.NoteLong[pvSpriteLane]
	if skill {
		sp = Skin.NoteSkill[pvSpriteLane]
	}
	pvNote(sp, n.Lane, n.Time, z)
}

// pvFlick draws a Flick body with its up-flick icon (a skill Flick note is
// note_skill, as Play's drawFlick).
func (n *PreviewNoteBase) pvFlick(skill bool, z sonolus.ZIndex) {
	var p sonolus.Vec2
	if Layout.HabaSkin {
		run := pvRun(n.SpanStart, n.Width)
		sp := Skin.HabaFlick[run]
		if skill {
			sp = Skin.HabaSkill[run]
		}
		p = pvWide(sp, habaWidthPx(n.Width), n.Lane, n.Time, z)
	} else {
		sp := Skin.NoteFlick[pvSpriteLane]
		if skill {
			sp = Skin.NoteSkill[pvSpriteLane]
		}
		p = pvNote(sp, n.Lane, n.Time, z)
	}
	pvFlickTop(p, n.Width, z)
}

// pvFlickTop draws an up-flick icon above p (Play's drawFlickTop).
func pvFlickTop(p sonolus.Vec2, width float64, z sonolus.ZIndex) {
	s := pvScale()
	top := sonolus.Vec2{X: p.X, Y: p.Y + pvFlickTopOffset()*s}
	if Layout.HabaSkin {
		i := int(sonolus.Clamp(width, 1, 3)) - 1
		Skin.HabaFlickTop[i].DrawZ(spriteQuad(top, habaTopPx[i], 138, notePPU, s), subZ(z, 1), 1)
		return
	}
	Skin.NoteFlickTop.DrawZ(spriteQuad(top, 171, 138, notePPU, s), subZ(z, 1), 1)
}

// pvBelt draws a body from (laneA, tA) to (laneB, tB), split at columns.
func pvBelt(laneA, tA, laneB, tB, widthA, widthB float64, curve bool, z sonolus.ZIndex, alpha float64) {
	sp := Skin.LongLine
	if Layout.HabaSkin {
		sp = Skin.HabaLongLine
		if curve {
			sp = Skin.HabaLongLine2
		}
	} else if curve {
		sp = Skin.LongLine2
	}
	if tB <= tA {
		return
	}
	t0 := tA
	for k := 0; k < 256 && t0 < tB; k++ {
		col := pvColumn(t0)
		t1 := math.Min(tB, (col+1)*pvColumnSeconds)
		r0 := (t0 - tA) / (tB - tA)
		r1 := (t1 - tA) / (tB - tA)
		c := pvCenter(col)
		// both ends in this column (a segment end on the column top
		// belongs to it, not to the next column's bottom)
		p0 := sonolus.Vec2{X: c + (laneA+(laneB-laneA)*r0-3)*PL.W, Y: pvY(t0, col)}
		p1 := sonolus.Vec2{X: c + (laneA+(laneB-laneA)*r1-3)*PL.W, Y: pvY(t1, col)}
		h0 := pvHalf(widthA + (widthB-widthA)*r0)
		h1 := pvHalf(widthA + (widthB-widthA)*r1)
		sp.DrawZ(beltQuad(p0, p1, h0, h1), z, alpha)
		t0 = t1
	}
}

type PreviewInitialization struct {
	preview.Archetype      `archetype:"name=Initialization"`
	preview.CallbackOrders `archetype:"preprocess=-10"`
	NoteCount              float64                            `archetype:"imported,name=noteCount"`
	Habahiro               bool                               `archetype:"imported,name=habahiro"`
	LaneChangeBeat         float64                            `archetype:"imported,name=laneChangeBeat,default=-1"`
	FirstScored            sonolus.EntityRef[PreviewNoteBase] `archetype:"imported,name=firstScored"`
}

func (n *PreviewInitialization) Preprocess() {
	r := preview.Screen.Rect()
	laneChange := -1.0
	if n.LaneChangeBeat >= 0 {
		laneChange = preview.Time.BeatToTime(n.LaneChangeBeat)
	}
	initLayout(r, Config.NoteSpeed, Config.NoteSize, Config.Mirror, n.Habahiro, laneChange)
	PL.LaneChange = -1
	if Layout.HabaSkin && laneChange >= 0 {
		PL.LaneChange = laneChange + laneChangeSwap
	}

	// The menu sits where Play's pause button is (as in Tutorial); the
	// progress bar runs along the bottom as in Watch. Both lie over the
	// columns, which fill the screen height.
	menu := preview.UI.MenuConfiguration()
	preview.UI.SetMenu(sonolus.RuntimeUIBasicLayout{
		Anchor: uiTopRight(-42, 54), Pivot: sonolus.Vec2{X: 0.5, Y: 0.5},
		Size:  sonolus.Vec2{X: 68 * Layout.U * menu.Scale, Y: 70 * Layout.U * menu.Scale},
		Alpha: menu.Alpha, Background: true,
	})
	progress := preview.UI.ProgressConfiguration()
	preview.UI.SetProgress(sonolus.RuntimeUIBasicLayout{
		Anchor: sonolus.Vec2{X: 0, Y: r.B + 0.05}, Pivot: sonolus.Vec2{X: 0.5, Y: 0},
		Size:  sonolus.Vec2{X: 2 * Layout.SafeR * 0.9, Y: 0.15 * progress.Scale},
		Alpha: progress.Alpha, Background: true,
	})
	PL.Bottom, PL.Top = r.B, r.T
	PL.W = (PL.Top - PL.Bottom) / pvHeightLanes
	PL.ColW = PL.W * pvColumnLanes
	PL.Left = r.L
}

type PreviewStage struct {
	preview.Archetype      `archetype:"name=Stage"`
	preview.CallbackOrders `archetype:"preprocess=100"`
}

func (*PreviewStage) Preprocess() {
	PL.Columns = pvColumn(PL.LastTime) + 1
	r := preview.Screen.Rect()
	preview.Canvas.Set(preview.CanvasOptions{
		Scroll: preview.ScrollLeftToRight,
		Size:   math.Max(PL.Columns*PL.ColW, r.R-r.L),
	})
}

const (
	zPvGrid   = 15.0
	zPvMarker = 16.0
)

func (*PreviewStage) Render() {
	end := PL.Columns * pvColumnSeconds
	for c := 0.0; c < PL.Columns && c < 1000; c++ {
		pvDrawStrip(c)
	}
	// a beat line on every quarter note
	for b := 0.0; b < 16384; b++ {
		t := preview.Time.BeatToTime(b)
		if t >= end {
			break
		}
		pvLine(Skin.PreviewLine, t, -3.5, 3.5, pvLineHalf(), sonolus.ZIndex{Z1: zPvGrid}, 0.5)
	}
	// whole seconds on the left
	for s := 1.0; s < end && s < 4096; s++ {
		pvLabel(s, sonolus.PrintFormatTime, 0, s, -1, sonolus.PrintColorNeutral)
	}
	pvPrintCombo()
}

// pvDrawStrip draws a column's lane strip; on a HABAHIRO chart the lane
// image changes at the lane change.
func pvDrawStrip(col float64) {
	c := pvCenter(col)
	t0, t1 := col*pvColumnSeconds, (col+1)*pvColumnSeconds
	quad := func(a, b float64) sonolus.Quad {
		return sonolus.Rect{T: pvY(b, col), R: c + pvStripHalf*PL.W, B: pvY(a, col), L: c - pvStripHalf*PL.W}.ToQuad()
	}
	z := sonolus.ZIndex{Z1: zField}
	switch {
	case PL.LaneChange < 0 || PL.LaneChange >= t1:
		Skin.PreviewLane.DrawZ(quad(t0, t1), z, 1)
	case PL.LaneChange <= t0:
		Skin.HabaPreviewLane.DrawZ(quad(t0, t1), z, 1)
	default:
		Skin.PreviewLane.DrawZ(quad(t0, PL.LaneChange), z, 1)
		Skin.HabaPreviewLane.DrawZ(quad(PL.LaneChange, t1), z, 1)
	}
}

// pvPrintCombo prints every pvComboStep-th combo on the right at the time
// of that scoring unit, walking the scoring chain (sorted by judgement
// time); level data written before the chain existed has no counts.
func pvPrintCombo() {
	unit := sonolus.EntityRef[PreviewInitialization]{Index: 0}.Get().FirstScored
	for combo := 1.0; combo < 100000 && unit.Index > 0; combo++ {
		n := unit.Get()
		if math.Mod(combo, pvComboStep) == 0 {
			pvLabel(combo, sonolus.PrintFormatEntityCount, 0, n.Time, 1, sonolus.PrintColorNeutral)
		}
		unit = n.ScoreNext
	}
}

type PreviewInputManager struct {
	preview.Archetype `archetype:"name=InputManager"`
}

type PreviewBPMChange struct {
	preview.Archetype `archetype:"name=#BPM_CHANGE"`
	Beat              float64 `archetype:"imported,name=#BEAT"`
	BPM               float64 `archetype:"imported,name=#BPM"`
}

// Render marks the change with a dark red line across the lanes (as thick
// as a beat line) and its BPM on the right in red (Print has no dark red).
func (n *PreviewBPMChange) Render() {
	t := preview.Time.BeatToTime(n.Beat)
	pvLine(Skin.PreviewLineBPM, t, -3.5, 3.5, pvLineHalf(), sonolus.ZIndex{Z1: zPvMarker}, 1)
	pvLabel(n.BPM, sonolus.PrintFormatBPM, -1, t, 1, sonolus.PrintColorRed)
}

type PreviewTimingChange struct {
	preview.Archetype `archetype:"name=TimingChange"`
	Time              float64 `archetype:"imported,name=time"`
	Speed             float64 `archetype:"imported,name=speed"`
	Axis              float64 `archetype:"imported,name=axis"`
	Max               float64 `archetype:"imported,name=max"`
}

type PreviewTimingGroup struct {
	preview.Archetype `archetype:"name=TimingGroup"`
	First             sonolus.EntityRef[PreviewTimingChange] `archetype:"imported,name=first"`
	Count             float64                                `archetype:"imported,name=count"`
}

// ----------------------------------------------------------------- notes ----

type PreviewNoteBase struct {
	preview.Archetype `archetype:"abstract"`
	Beat              float64                            `archetype:"imported,name=#BEAT"`
	Lane              float64                            `archetype:"imported,name=lane"`
	Width             float64                            `archetype:"imported,name=width"`
	SpanStart         float64                            `archetype:"imported,name=spanStart"`
	Skill             bool                               `archetype:"imported,name=skill"`
	Direction         float64                            `archetype:"imported,name=direction"`
	ScoreNext         sonolus.EntityRef[PreviewNoteBase] `archetype:"imported,name=scoreNext"`
	Time              float64                            `archetype:"data"`
}

func (n *PreviewNoteBase) prepare() {
	if Config.Mirror {
		n.Lane = 6 - n.Lane
		n.SpanStart = 6 - (n.SpanStart + n.Width - 1)
		n.Direction = -n.Direction
	}
	n.Time = preview.Time.BeatToTime(n.Beat)
	PL.LastTime = math.Max(PL.LastTime, n.Time)
}

func (n *PreviewNoteBase) z() sonolus.ZIndex { return noteZ(zNote, n.Time, n.Lane) }

func (n *PreviewNoteBase) Preprocess() { n.prepare() }

// pvTerminal draws a Slide head, a Long / Slide end or a gesture node; skill
// (a skill Slide head) applies to a plain body only, as Play's drawTerminal.
func (n *PreviewNoteBase) pvTerminal(kind, kindWidth float64, skill bool, z sonolus.ZIndex) {
	switch kind {
	case NodeFlick:
		n.pvFlick(false, z)
	case NodeDirectional, NodeMultiDir:
		members := 1.0
		if kind == NodeMultiDir {
			members = math.Max(kindWidth, 1)
		}
		var p sonolus.Vec2
		line := Skin.FlickLineR
		if n.Direction < 0 {
			line = Skin.FlickLineL
		}
		for k := 0.0; k < members && k < 7; k++ {
			ml := n.Lane + n.Direction*k
			body := Skin.NoteFlickR[pvSpriteLane]
			if n.Direction < 0 {
				body = Skin.NoteFlickL[pvSpriteLane]
			}
			q := pvNote(body, ml, n.Time, subZ(z, k))
			if k > 0 {
				// the members' back lines, as a moving wide directional end
				line.DrawZ(lineQuad(p, q, pvLocal()*0.75/2), lineZ(zDirLine, n.Time), 1)
			}
			p = q
		}
		pvArrow(p, n.Direction, iconZ(z))
	default:
		n.pvLong(skill, z)
	}
}

// pvArrow draws a directional arrow beside a note body at p (Play's
// drawSideArrow).
func pvArrow(p sonolus.Vec2, direction float64, z sonolus.ZIndex) {
	s := pvScale()
	off := pvSideArrowOffset()
	if direction < 0 {
		Skin.NoteFlickTopL.DrawZ(spriteQuad(sonolus.Vec2{X: p.X - off*s, Y: p.Y}, 138, 171, notePPU, s), z, 1)
	} else {
		Skin.NoteFlickTopR.DrawZ(spriteQuad(sonolus.Vec2{X: p.X + off*s, Y: p.Y}, 138, 170, notePPU, s), z, 1)
	}
}

type PreviewTapNote struct {
	PreviewNoteBase   `archetype:"base"`
	preview.Archetype `archetype:"name=TapNote"`
	Offbeat           bool `archetype:"imported,name=offbeat"`
}

func (n *PreviewTapNote) Render() {
	if Layout.HabaSkin {
		run := pvRun(n.SpanStart, n.Width)
		sp := Skin.HabaNormal[run]
		if n.Skill {
			sp = Skin.HabaSkill[run]
		} else if n.Offbeat && Config.NoteColor {
			sp = Skin.HabaNormal16[run]
		}
		pvWide(sp, habaWidthPx(n.Width), n.Lane, n.Time, n.z())
		return
	}
	sp := Skin.NoteNormal[pvSpriteLane]
	if n.Skill {
		sp = Skin.NoteSkill[pvSpriteLane]
	} else if n.Offbeat && Config.NoteColor {
		sp = Skin.NoteNormal16[pvSpriteLane]
	}
	pvNote(sp, n.Lane, n.Time, n.z())
}

type PreviewFlickNote struct {
	PreviewNoteBase   `archetype:"base"`
	preview.Archetype `archetype:"name=FlickNote"`
}

func (n *PreviewFlickNote) Render() { n.pvFlick(n.Skill, n.z()) }

type PreviewDirectionalBase struct {
	PreviewNoteBase   `archetype:"base"`
	preview.Archetype `archetype:"abstract"`
	Leader            sonolus.EntityRef[PreviewDirectionalBase] `archetype:"imported,name=leader"`
	Left              sonolus.EntityRef[PreviewDirectionalBase] `archetype:"imported,name=left"`
	Right             sonolus.EntityRef[PreviewDirectionalBase] `archetype:"imported,name=right"`
	Count             float64                                   `archetype:"imported,name=count"`
	Arrow             bool                                      `archetype:"imported,name=arrow"`
}

type PreviewDirectionalNote struct {
	PreviewDirectionalBase `archetype:"base"`
	preview.Archetype      `archetype:"name=DirectionalNote"`
}

type PreviewDirectionalMember struct {
	PreviewDirectionalBase `archetype:"base"`
	preview.Archetype      `archetype:"name=DirectionalMember"`
}

func (n *PreviewDirectionalBase) Render() {
	body := Skin.NoteFlickR[pvSpriteLane]
	if n.Direction < 0 {
		body = Skin.NoteFlickL[pvSpriteLane]
	}
	p := pvNote(body, n.Lane, n.Time, memberZ(n.z(), n.Direction, n.Lane, groupFirst(n.Leader.Get().Lane, n.Count)))
	if n.Arrow {
		pvArrow(p, n.Direction, iconZ(n.z()))
	}
	if n.Right.Index > 0 {
		// the back line to the next member (Play's drawBackLine)
		r := n.Right.Get()
		line := Skin.FlickLineR
		if n.Direction < 0 {
			line = Skin.FlickLineL
		}
		line.DrawZ(lineQuad(p, pvPos(r.Lane, r.Time), pvLocal()*0.75/2), lineZ(zDirLine, n.Time), 1)
	}
}

type PreviewLongHead struct {
	PreviewNoteBase   `archetype:"base"`
	preview.Archetype `archetype:"name=LongHead"`
	Tail              sonolus.EntityRef[PreviewLongTail] `archetype:"imported,name=tail"`
}

type PreviewLongTail struct {
	PreviewNoteBase   `archetype:"base"`
	preview.Archetype `archetype:"name=LongTail"`
	Head              sonolus.EntityRef[PreviewLongHead] `archetype:"imported,name=head"`
	Kind              float64                            `archetype:"imported,name=kind"`
}

func (n *PreviewLongHead) Render() {
	t := n.Tail.Get()
	pvBelt(n.Lane, n.Time, t.Lane, t.Time, n.Width, t.Width, false, lineZ(zBelt, n.Time), bodyAlpha())
	n.pvLong(n.Skill, n.z())
}

func (n *PreviewLongTail) Render() { n.pvTerminal(n.Kind, n.Width, false, n.z()) }

type PreviewSlideNodeBase struct {
	PreviewNoteBase   `archetype:"base"`
	preview.Archetype `archetype:"abstract"`
	Index             float64                                 `archetype:"imported,name=index"`
	Kind              float64                                 `archetype:"imported,name=kind"`
	KindWidth         float64                                 `archetype:"imported,name=kindWidth"`
	Exit              float64                                 `archetype:"imported,name=exit"`
	Hidden            bool                                    `archetype:"imported,name=hidden"`
	Terminal          bool                                    `archetype:"imported,name=terminal"`
	Head              sonolus.EntityRef[PreviewSlideBase]     `archetype:"imported,name=head"`
	Prev              sonolus.EntityRef[PreviewSlideNodeBase] `archetype:"imported,name=prev"`
	Next              sonolus.EntityRef[PreviewSlideNodeBase] `archetype:"imported,name=next"`
}

type PreviewSlideNode struct {
	PreviewSlideNodeBase `archetype:"base"`
	preview.Archetype    `archetype:"name=SlideNode"`
}

type PreviewSlideHidden struct {
	PreviewSlideNodeBase `archetype:"base"`
	preview.Archetype    `archetype:"name=SlideHidden"`
}

func (n *PreviewSlideNodeBase) Preprocess() {
	n.prepare()
	if Config.Mirror {
		n.Exit = -n.Exit
	}
}

func (n *PreviewSlideNodeBase) Render() {
	h := n.Head.Get()
	fl, ft, fw := h.Lane+h.Exit, h.Time, h.Width
	if n.Prev.Index > 0 {
		p := n.Prev.Get()
		fl, ft, fw = p.Lane+p.Exit, p.Time, p.Width
	}
	alpha := bodyAlpha()
	if h.AllHidden {
		alpha *= 0.5
	}
	pvBelt(fl, ft, n.Lane, n.Time, fw, n.Width, h.Curve, lineZ(zBelt, h.Time), alpha)
	if n.Hidden {
		return
	}
	if !n.Terminal && (n.Kind == NodePlain || n.Kind == NodeFlick) {
		// intermediate node: note_slide_among (+ the up-flick icon)
		var p sonolus.Vec2
		if Layout.HabaSkin {
			w := sonolus.Clamp(n.Width, 1, 7)
			p = pvWide(Skin.HabaAmong[int(w)-1], habaWidthPx(w), n.Lane, n.Time, n.z())
		} else {
			p = pvNote(Skin.NoteSlideAmong, n.Lane, n.Time, n.z())
		}
		if n.Kind == NodeFlick {
			pvFlickTop(p, n.Width, n.z())
		}
		return
	}
	n.pvTerminal(n.Kind, n.KindWidth, false, n.z())
}

type PreviewSlideBase struct {
	PreviewNoteBase   `archetype:"base"`
	preview.Archetype `archetype:"abstract"`
	Kind              float64                                 `archetype:"imported,name=kind"`
	KindWidth         float64                                 `archetype:"imported,name=kindWidth"`
	Exit              float64                                 `archetype:"imported,name=exit"`
	Hidden            bool                                    `archetype:"imported,name=hidden"`
	Curve             bool                                    `archetype:"imported,name=curve"`
	AllHidden         bool                                    `archetype:"imported,name=allHidden"`
	First             sonolus.EntityRef[PreviewSlideNodeBase] `archetype:"imported,name=first"`
	Last              sonolus.EntityRef[PreviewSlideNodeBase] `archetype:"imported,name=last"`
}

type PreviewSlideHead struct {
	PreviewSlideBase  `archetype:"base"`
	preview.Archetype `archetype:"name=SlideHead"`
}

type PreviewSlideStart struct {
	PreviewSlideBase  `archetype:"base"`
	preview.Archetype `archetype:"name=SlideStart"`
}

func (n *PreviewSlideBase) Preprocess() {
	n.prepare()
	if Config.Mirror {
		n.Exit = -n.Exit
	}
}

func (n *PreviewSlideBase) Render() {
	if n.Hidden {
		return
	}
	n.pvTerminal(n.Kind, n.KindWidth, n.Skill, n.z())
}

type PreviewSimLine struct {
	preview.Archetype `archetype:"name=SimLine"`
	A                 sonolus.EntityRef[PreviewNoteBase] `archetype:"imported,name=a"`
	B                 sonolus.EntityRef[PreviewNoteBase] `archetype:"imported,name=b"`
	ALane             float64                            `archetype:"imported,name=aLane,default=-1000"`
	BLane             float64                            `archetype:"imported,name=bLane,default=-1000"`
	SpeedMin          float64                            `archetype:"imported,name=speedMin,default=0"`
	SpeedMax          float64                            `archetype:"imported,name=speedMax,default=1000"`
	ANoMargin         bool                               `archetype:"imported,name=aNoMargin"`
	BNoMargin         bool                               `archetype:"imported,name=bNoMargin"`
}

// pvSyncAtSpeed is Play's syncLineAtSpeed (feedback.go is play/watch only):
// a line the converter paired only for some note speeds (option tenths).
func pvSyncAtSpeed(min, max float64) bool {
	speed := math.Round(Config.NoteSpeed * 10)
	return speed >= min && speed <= max
}

// Render draws the sync line as Play's drawSimLine: ends pulled in by the
// skin's edge margin except at directional ends, width 0.28 × note scale.
func (l *PreviewSimLine) Render() {
	if !Config.SyncLine || !pvSyncAtSpeed(l.SpeedMin, l.SpeedMax) {
		return
	}
	a, b := l.A.Get(), l.B.Get()
	la, lb := simLane(l.ALane, a.Lane), simLane(l.BLane, b.Lane)
	p, q := pvPos(la, a.Time), pvPos(lb, b.Time)
	if math.Abs(p.Y-q.Y) > 1e-6 {
		return // different columns or times
	}
	s := pvScale()
	if Skin.SyncEdgeMarginOpt.Exists() {
		dir := 1.0
		if p.X > q.X {
			dir = -1
		}
		if !l.ANoMargin {
			p.X += syncEdgeMargin * s * dir
		}
		if !l.BNoMargin {
			q.X -= syncEdgeMargin * s * dir
		}
	}
	z := lineZ(zSyncLine, math.Min(a.Time, b.Time))
	if Layout.HabaSkin {
		Skin.HabaSimLine.DrawZ(lineQuad(p, q, pvLocal()*0.28/2), z, 1)
		return
	}
	Skin.SimLine.DrawZ(lineQuad(p, q, pvLocal()*0.28/2), z, 1)
}

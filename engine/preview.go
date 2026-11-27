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
// scale). As in GarupaEditor's static preview: faint measure lines, one event
// line per instant (BPM, SV and lane change merged) with its labels left of
// the track, whole seconds on the left, every 50th combo on the right, and a
// column holding a note outside lanes 0–6 widened by as many lanes each side.

const (
	pvColumnSeconds = 2.0
	pvColumnLanes   = 12.0 // column width in lanes: the strip and two label margins
	pvHeightLanes   = 20.0 // column height in lanes
	// pvStripHalf is the half width of the preview lane strip in lanes: the
	// field image is 1160 UI units wide, a lane 2.2·S = 153.2 units.
	pvStripHalf = 580 / (2.2 * 1334 / (2 * 9.578571319580078))
	pvComboStep = 50.0
	// The level data has no time signature: measures of 4 beats from beat 0.
	pvBeatsPerMeasure = 4.0
	// Capacities in PreviewData (4096 values): column edges (2000 s of
	// chart), the BPM change table and the timing groups.
	pvMaxColumns = 1000
	pvMaxBPM     = 512
	pvMaxGroups  = 32
	pvMaxExtra   = 64.0 // extra lanes a column may get each side
	// pvEventEps (seconds) merges events at the same instant: BPM change
	// times come from the BPM timeline, SV times from the converter.
	pvEventEps = 1e-3
)

type PreviewLayout struct {
	sonolus.LevelDataResource
	ColW         float64 // width of a column without extra lanes
	W            float64 // lane spacing
	Left         float64 // x of the first column's left edge
	Bottom       float64
	Top          float64
	LastTime     float64
	Columns      float64
	LaneChange   float64 // time the HABAHIRO lane image is swapped in (−1: none)
	LaneChangeAt float64 // time of the lane change event line (−1: none)
	// ColX: during the notes' preprocess, column c's extra lanes each side;
	// from the Stage's preprocess on, the x of its left edge from Left
	// (ColX[Columns] is the right edge of the last column).
	ColX [pvMaxColumns + 1]float64
	// BPM changes sorted by time (entity index of each), and the timing
	// groups (entity indices) in preprocess order.
	BPMCount   float64
	BPMTime    [pvMaxBPM]float64
	BPMEntity  [pvMaxBPM]float64
	GroupCount float64
	Group      [pvMaxGroups]float64
}

var PL = PreviewLayout{}

func pvColumn(t float64) float64 { return math.Max(math.Floor(t/pvColumnSeconds), 0) }

// pvColLeft is the x of column col's left edge from PL.Left; columns past the
// last one (events after the last note) follow at the plain width.
func pvColLeft(col float64) float64 {
	if col < PL.Columns {
		return PL.ColX[int(col)]
	}
	return PL.ColX[int(PL.Columns)] + (col-PL.Columns)*PL.ColW
}

// pvColExtra is the number of lanes column col is widened by each side.
func pvColExtra(col float64) float64 {
	if col >= PL.Columns {
		return 0
	}
	return math.Round((PL.ColX[int(col)+1] - PL.ColX[int(col)] - PL.ColW) / (2 * PL.W))
}

// pvCenter is the x of lane 3 in column col.
func pvCenter(col float64) float64 {
	return PL.Left + pvColLeft(col) + PL.ColW/2 + pvColExtra(col)*PL.W
}

// pvPx is the editor's 1 px line width: the screen height over its design
// height of 750.
func pvPx() float64 { return (PL.Top - PL.Bottom) / 750 }

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
	pvLabelAt(value, format, decimals, col, sonolus.Clamp(pvY(t, col), PL.Bottom+h/2, PL.Top-h/2), side, color)
}

// pvLabelEdge is the lane offset from the column centre that labels keep
// clear of: the lane art or, in a widened column, the lane range.
func pvLabelEdge(col float64) float64 { return math.Max(pvStripHalf, 3.5+pvColExtra(col)) }

// pvLabelAt prints a label beside column col's track, centred on height y.
func pvLabelAt(value float64, format sonolus.PrintFormat, decimals int, col, y, side float64, color sonolus.PrintColor) {
	edge := pvLabelEdge(col)
	align, pivot := sonolus.HorizontalAlignLeft, 0.0
	if side < 0 {
		align, pivot = sonolus.HorizontalAlignRight, 1
	}
	preview.Canvas.Print(preview.PrintOptions{
		Value: value, Format: format, DecimalPlaces: decimals,
		Anchor: sonolus.Vec2{X: pvCenter(col) + side*(edge+0.1)*PL.W, Y: y},
		Pivot:  sonolus.Vec2{X: pivot, Y: 0.5},
		Size:   sonolus.Vec2{X: (pvColumnLanes/2 + pvColExtra(col) - edge - 0.2) * PL.W, Y: pvTextH()},
		Color:  color, Alpha: 1, HorizontalAlign: align,
	})
}

// pvLine draws a horizontal line at time t from lane offset l to r (lanes
// from the column centre, lane 3).
func pvLine(sp sonolus.Sprite, t, l, r, half float64, z sonolus.ZIndex, a float64) {
	col := pvColumn(t)
	c, y := pvCenter(col), pvY(t, col)
	sp.DrawZ(sonolus.Rect{T: y + half, R: c + r*PL.W, B: y - half, L: c + l*PL.W}.ToQuad(), z, a)
}

// Measure lines and the widened columns' lane range guides: the editor's
// rgba(220, 230, 255, 0.3), 1 px (the white line sprite at that alpha).
const pvGuideAlpha = 0.3

// pvBeatAlpha is the beat lines' alpha (1 px, the same sprite), fainter than
// the measure lines.
const pvBeatAlpha = 0.15

// pvDashedGuide draws a vertical dashed line (4 px dashes, 4 px gaps) at x
// over the column's height.
func pvDashedGuide(x float64) {
	px := pvPx()
	z := sonolus.ZIndex{Z1: zPvGrid}
	for y := PL.Bottom; y < PL.Top; y += 8 * px {
		Skin.PreviewLine.DrawZ(sonolus.Rect{T: math.Min(y+4*px, PL.Top), R: x + px/2, B: y, L: x - px/2}.ToQuad(), z, pvGuideAlpha)
	}
}

// ---------------------------------------------------------------- events ----

// Event lines take the editor's event colour #FF3B72 at its alpha 0.8, 1 px;
// the skin's only uniform #FF3B72 fill is the middle of the Auto Live pill.
const pvEventAlpha = 0.8

func pvEventSprite() sonolus.Sprite { return Skin.AutoPill[1] }

// pvBPMFrom is the first entry of the BPM table at or after time t.
func pvBPMFrom(t float64) float64 {
	lo, hi := 0.0, PL.BPMCount
	for lo < hi {
		mid := math.Floor((lo + hi) / 2)
		if PL.BPMTime[int(mid)] < t {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// pvBPMAt is the first entry of the BPM table at time t, or −1.
func pvBPMAt(t float64) float64 {
	b := pvBPMFrom(t - pvEventEps)
	if b < PL.BPMCount && PL.BPMTime[int(b)] <= t+pvEventEps {
		return b
	}
	return -1
}

func pvGroup(i float64) *PreviewTimingGroup {
	return sonolus.EntityRef[PreviewTimingGroup]{Index: PL.Group[int(i)]}.Get()
}

// pvGroupFrom is the first change of the i-th timing group at or after time
// t. A group's changes are sorted by time, one per instant.
func pvGroupFrom(i, t float64) float64 {
	g := pvGroup(i)
	lo, hi := 0.0, g.Count
	for lo < hi {
		mid := math.Floor((lo + hi) / 2)
		if g.change(mid).Time < t {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	return lo
}

// pvGroupChange is the change of the i-th timing group at time t, or −1.
func pvGroupChange(i, t float64) float64 {
	k := pvGroupFrom(i, t-pvEventEps)
	if k < pvGroup(i).Count && pvGroup(i).change(k).Time <= t+pvEventEps {
		return k
	}
	return -1
}

// pvSVShown reports whether the i-th group's change at t prints its value:
// every group carries the #Global changes, so a value an earlier group
// already printed at t is not repeated.
func pvSVShown(i, t float64) bool {
	k := pvGroupChange(i, t)
	if k < 0 {
		return false
	}
	v := pvGroup(i).change(k).Speed
	for j := 0.0; j < i; j++ {
		if q := pvGroupChange(j, t); q >= 0 && math.Abs(pvGroup(j).change(q).Speed-v) < 1e-6 {
			return false
		}
	}
	return true
}

// pvSVBefore reports whether a timing group registered before the i-th one
// has a change at t (i = GroupCount: any group).
func pvSVBefore(i, t float64) bool {
	for j := 0.0; j < i && j < PL.GroupCount; j++ {
		if pvGroupChange(j, t) >= 0 {
			return true
		}
	}
	return false
}

// pvEventLines is the number of label lines of the event at t: its BPM and
// its distinct SV values (the lane change has no printable label).
func pvEventLines(t float64) float64 {
	n := 0.0
	if pvBPMAt(t) >= 0 {
		n++
	}
	for i := 0.0; i < PL.GroupCount; i++ {
		if pvSVShown(i, t) {
			n++
		}
	}
	return n
}

// pvEventTop is the centre height of the first of n label lines of the event
// at t: the stack centred on t's height, kept inside the column.
func pvEventTop(t, n float64) float64 {
	h := pvTextH()
	y := pvY(t, pvColumn(t)) + h*(n-1)/2
	return math.Max(math.Min(y, PL.Top-h/2), PL.Bottom+h*(n-0.5))
}

// pvEvent draws the event at t once for every event at that instant: one
// line across the column's lane range and, left of the track, one label line
// per event (BPM, then the SV values, in Print's red: it has no #FF3B72).
func pvEvent(t float64) {
	col := pvColumn(t)
	r := 3.5 + pvColExtra(col)
	pvLine(pvEventSprite(), t, -r, r, pvPx()/2, sonolus.ZIndex{Z1: zPvMarker}, pvEventAlpha)
	n := pvEventLines(t)
	if n == 0 {
		return
	}
	h := pvTextH()
	y := pvEventTop(t, n)
	if b := pvBPMAt(t); b >= 0 {
		bpm := sonolus.EntityRef[PreviewBPMChange]{Index: PL.BPMEntity[int(b)]}.Get().BPM
		pvLabelAt(bpm, sonolus.PrintFormatBPM, -1, col, y, -1, sonolus.PrintColorRed)
		y -= h
	}
	for i := 0.0; i < PL.GroupCount; i++ {
		if pvSVShown(i, t) {
			v := pvGroup(i).change(pvGroupChange(i, t)).Speed
			pvLabelAt(v, sonolus.PrintFormatTimeScale, -1, col, y, -1, sonolus.PrintColorRed)
			y -= h
		}
	}
}

// pvCoversLabel reports whether the labels of the event at te cover a label
// centred on height y in column col.
func pvCoversLabel(te, col, y float64) bool {
	if pvColumn(te) != col {
		return false
	}
	n := pvEventLines(te)
	if n == 0 {
		return false
	}
	h := pvTextH()
	top := pvEventTop(te, n) + h/2
	return y-h/2 < top && y+h/2 > top-n*h
}

// pvTimeLabelCovered reports whether an event label takes the place of the
// time label of second s (the editor's rule: an event label wins).
func pvTimeLabelCovered(s float64) bool {
	col := pvColumn(s)
	h := pvTextH()
	y := sonolus.Clamp(pvY(s, col), PL.Bottom+h/2, PL.Top-h/2)
	// an event stack reaches at most this far (seconds) from its time
	reach := (PL.GroupCount + 2) * pvColumnSeconds / pvHeightLanes
	for b := pvBPMFrom(s - reach); b < PL.BPMCount && PL.BPMTime[int(b)] <= s+reach; b++ {
		if pvCoversLabel(PL.BPMTime[int(b)], col, y) {
			return true
		}
	}
	for i := 0.0; i < PL.GroupCount; i++ {
		g := pvGroup(i)
		for k := pvGroupFrom(i, s-reach); k < g.Count && g.change(k).Time <= s+reach; k++ {
			if pvCoversLabel(g.change(k).Time, col, y) {
				return true
			}
		}
	}
	return false
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
	PL.LaneChangeAt = laneChange

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

// Preprocess runs after every note has recorded its column's extra lanes in
// ColX and turns them into the columns' left edges (a prefix sum of the
// column widths).
func (*PreviewStage) Preprocess() {
	PL.Columns = math.Min(pvColumn(PL.LastTime)+1, pvMaxColumns)
	x := 0.0
	for c := 0.0; c < PL.Columns; c++ {
		extra := PL.ColX[int(c)]
		PL.ColX[int(c)] = x
		x += PL.ColW + 2*extra*PL.W
	}
	PL.ColX[int(PL.Columns)] = x
	r := preview.Screen.Rect()
	preview.Canvas.Set(preview.CanvasOptions{
		Scroll: preview.ScrollLeftToRight,
		Size:   math.Max(x, r.R-r.L),
	})
}

const (
	zPvGrid   = 15.0
	zPvMarker = 16.0
)

func (*PreviewStage) Render() {
	end := PL.Columns * pvColumnSeconds
	for c := 0.0; c < PL.Columns; c++ {
		pvDrawStrip(c)
		if extra := pvColExtra(c); extra > 0 {
			// the widened column's lane range
			pvDashedGuide(pvCenter(c) - (3.5+extra)*PL.W)
			pvDashedGuide(pvCenter(c) + (3.5+extra)*PL.W)
		}
	}
	// across the seven lanes: a measure line every pvBeatsPerMeasure beats,
	// a fainter beat line on the other beats
	for b := 0.0; b < 65536; b++ {
		t := preview.Time.BeatToTime(b)
		if t >= end {
			break
		}
		alpha := pvBeatAlpha
		if math.Mod(b, pvBeatsPerMeasure) == 0 {
			alpha = pvGuideAlpha
		}
		pvLine(Skin.PreviewLine, t, -3.5, 3.5, pvPx()/2, sonolus.ZIndex{Z1: zPvGrid}, alpha)
	}
	// the lane change, unless a BPM change or an SV change at the same
	// instant draws its line
	if t := PL.LaneChangeAt; t >= 0 && pvBPMAt(t) < 0 && !pvSVBefore(PL.GroupCount, t) {
		pvEvent(t)
	}
	// whole seconds on the left, except where an event label is
	for s := 1.0; s < end && s < 4096; s++ {
		if !pvTimeLabelCovered(s) {
			pvLabel(s, sonolus.PrintFormatTime, 0, s, -1, sonolus.PrintColorNeutral)
		}
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
	Listed            bool    `archetype:"data"` // in the BPM table
}

// Preprocess inserts the change into the BPM table, sorted by time (equal
// times in preprocess order).
func (n *PreviewBPMChange) Preprocess() {
	if PL.BPMCount >= pvMaxBPM {
		return
	}
	t := preview.Time.BeatToTime(n.Beat)
	i := PL.BPMCount
	for ; i > 0 && PL.BPMTime[int(i-1)] > t; i-- {
		PL.BPMTime[int(i)] = PL.BPMTime[int(i-1)]
		PL.BPMEntity[int(i)] = PL.BPMEntity[int(i-1)]
	}
	PL.BPMTime[int(i)] = t
	PL.BPMEntity[int(i)] = preview.Entity.Info().Index
	PL.BPMCount++
	n.Listed = true
}

// Render draws the event at the change's time if it is the first BPM change
// there (a BPM change draws the events it shares its instant with); one
// past the table's capacity draws its own line and label.
func (n *PreviewBPMChange) Render() {
	t := preview.Time.BeatToTime(n.Beat)
	if !n.Listed {
		pvLine(pvEventSprite(), t, -3.5, 3.5, pvPx()/2, sonolus.ZIndex{Z1: zPvMarker}, pvEventAlpha)
		pvLabel(n.BPM, sonolus.PrintFormatBPM, -1, t, -1, sonolus.PrintColorRed)
		return
	}
	if PL.BPMEntity[int(pvBPMAt(t))] == preview.Entity.Info().Index {
		pvEvent(t)
	}
}

// PreviewTimingChange is one SV step; its time is in seconds (the
// converter's tempo.Time of the SV beat, compared with Time.Now in Play).
type PreviewTimingChange struct {
	preview.Archetype `archetype:"name=TimingChange"`
	Time              float64 `archetype:"imported,name=time"`
	Speed             float64 `archetype:"imported,name=speed"`
	Axis              float64 `archetype:"imported,name=axis"`
	Max               float64 `archetype:"imported,name=max"`
}

// Render draws the event at the change's time unless a BPM change or an
// earlier timing group's change at the same instant draws it: each group
// carries the #Global changes as well as its own.
func (c *PreviewTimingChange) Render() {
	if pvBPMAt(c.Time) >= 0 {
		return
	}
	self := preview.Entity.Info().Index
	for i := 0.0; i < PL.GroupCount; i++ {
		g := pvGroup(i)
		if self >= g.First.Index && self < g.First.Index+g.Count {
			break // its own group: no earlier group has this instant
		}
		if pvGroupChange(i, c.Time) >= 0 {
			return
		}
	}
	pvEvent(c.Time)
}

type PreviewTimingGroup struct {
	preview.Archetype `archetype:"name=TimingGroup"`
	First             sonolus.EntityRef[PreviewTimingChange] `archetype:"imported,name=first"`
	Count             float64                                `archetype:"imported,name=count"`
}

func (g *PreviewTimingGroup) change(k float64) *PreviewTimingChange {
	return sonolus.EntityRef[PreviewTimingChange]{Index: g.First.Index + k}.Get()
}

// Preprocess registers the group for the event lookups.
func (*PreviewTimingGroup) Preprocess() {
	if PL.GroupCount < pvMaxGroups {
		PL.Group[int(PL.GroupCount)] = preview.Entity.Info().Index
		PL.GroupCount++
	}
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
	// a lane outside 0–6 (to the nearest lane) widens the note's column by
	// as many lanes each side (ColX holds the extra lanes until the Stage's
	// preprocess)
	if outside := math.Min(math.Round(math.Max(0, math.Max(-n.Lane, n.Lane-6))), pvMaxExtra); outside > 0 {
		c := int(math.Min(pvColumn(n.Time), pvMaxColumns-1))
		PL.ColX[c] = math.Max(PL.ColX[c], outside)
	}
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

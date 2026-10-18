//go:build preview

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus/preview"
)

// Preview draws the chart as left-to-right columns of pvColumnSeconds each,
// lanes vertical and time upward, in real time (SV is not applied).

const pvColumnSeconds = 2.0

type PreviewLayout struct {
	sonolus.LevelDataResource
	ColW     float64 // column width
	W        float64 // lane spacing
	Left     float64 // x of the first column's left edge
	Bottom   float64
	Top      float64
	LastTime float64
	LastBeat float64
	Columns  float64
}

var PL = PreviewLayout{}

// pvPos maps (lane, time) to the canvas.
func pvPos(lane, t float64) sonolus.Vec2 {
	col := math.Max(math.Floor(t/pvColumnSeconds), 0)
	local := (t - col*pvColumnSeconds) / pvColumnSeconds
	return sonolus.Vec2{
		X: PL.Left + col*PL.ColW + PL.ColW/2 + (lane-3)*PL.W,
		Y: PL.Bottom + local*(PL.Top-PL.Bottom),
	}
}

// pvScale is the world scale of a note sprite (308 px = 1.4 lanes).
func pvScale() float64 { return PL.W / buttonSpacing }

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
func (n *PreviewNoteBase) pvLong(z sonolus.ZIndex) {
	if Layout.HabaSkin {
		pvWide(Skin.HabaLong[habaRun(n.SpanStart, n.Width)], habaWidthPx(n.Width), n.Lane, n.Time, z)
		return
	}
	pvNote(lanePick(Skin.NoteLong, n.SpriteLane), n.Lane, n.Time, z)
}

// pvFlickTop draws an up-flick icon above p.
func pvFlickTop(p sonolus.Vec2, width float64, z sonolus.ZIndex) {
	s := pvScale()
	top := sonolus.Vec2{X: p.X, Y: p.Y + 0.9*s}
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
		colEnd := (math.Floor(t0/pvColumnSeconds) + 1) * pvColumnSeconds
		t1 := math.Min(tB, colEnd)
		r0 := (t0 - tA) / (tB - tA)
		r1 := (t1 - tA) / (tB - tA)
		p0 := pvPos(laneA+(laneB-laneA)*r0, t0)
		p1 := pvPos(laneA+(laneB-laneA)*r1, t1)
		if t1 == colEnd {
			p1.Y = PL.Top
			p1.X -= PL.ColW // the column end belongs to the previous column
		}
		h0 := (widthA + (widthB-widthA)*r0) * PL.W * 0.45 * widthRate(widthA)
		h1 := (widthA + (widthB-widthA)*r1) * PL.W * 0.45 * widthRate(widthB)
		sp.DrawZ(beltQuad(p0, p1, h0, h1), z, alpha)
		t0 = t1
	}
}

type PreviewInitialization struct {
	preview.Archetype      `archetype:"name=Initialization"`
	preview.CallbackOrders `archetype:"preprocess=-10"`
	NoteCount              float64 `archetype:"imported,name=noteCount"`
	Habahiro               bool    `archetype:"imported,name=habahiro"`
}

func (n *PreviewInitialization) Preprocess() {
	r := preview.Screen.Rect()
	initLayout(r, Config.NoteSpeed, Config.NoteSize, Config.Mirror, n.Habahiro, -1)
	h := r.T - r.B
	PL.Bottom = r.B + h*0.06
	PL.Top = r.T - h*0.06
	PL.ColW = (PL.Top - PL.Bottom) * 0.55
	PL.W = PL.ColW / 9
	PL.Left = r.L + h*0.05
}

type PreviewStage struct {
	preview.Archetype      `archetype:"name=Stage"`
	preview.CallbackOrders `archetype:"preprocess=100"`
}

func (*PreviewStage) Preprocess() {
	PL.Columns = math.Floor(PL.LastTime/pvColumnSeconds) + 1
	r := preview.Screen.Rect()
	preview.Canvas.Set(preview.CanvasOptions{
		Scroll: preview.ScrollLeftToRight,
		Size:   math.Max(PL.Left-r.L+PL.Columns*PL.ColW+(r.T-r.B)*0.05, r.R-r.L),
	})
}

func (*PreviewStage) Render() {
	for c := 0.0; c < PL.Columns && c < 1000; c++ {
		x := PL.Left + c*PL.ColW + PL.ColW/2
		preview.Canvas.Print(preview.PrintOptions{
			Value: c * pvColumnSeconds, Format: sonolus.PrintFormatTime, DecimalPlaces: 0,
			Anchor: sonolus.Vec2{X: x, Y: PL.Bottom - 0.01}, Pivot: sonolus.Vec2{X: 0.5, Y: 1},
			Size: sonolus.Vec2{X: PL.ColW * 0.8, Y: 0.05}, Alpha: 1, Color: sonolus.PrintColorNeutral,
			HorizontalAlign: sonolus.HorizontalAlignCenter,
		})
	}
	for c := 0.0; c < PL.Columns && c < 1000; c++ {
		x0 := PL.Left + c*PL.ColW + PL.ColW/2
		for l := 0.0; l <= 7; l++ {
			x := x0 + (l-3.5)*PL.W
			a := 0.25
			if l == 0 || l == 7 {
				a = 0.6
			}
			Skin.SimLine.Draw(lineQuad(sonolus.Vec2{X: x, Y: PL.Bottom}, sonolus.Vec2{X: x, Y: PL.Top}, 0.003), 1, a)
		}
	}
	// beat lines (measure every 4 beats)
	for b := 0.0; b <= PL.LastBeat && b < 4096; b++ {
		t := preview.Time.BeatToTime(b)
		p := pvPos(3, t)
		half := 0.0015
		a := 0.25
		if math.Mod(b, 4) == 0 {
			half, a = 0.003, 0.6
		}
		Skin.SimLine.Draw(lineQuad(sonolus.Vec2{X: p.X - 3.5*PL.W, Y: p.Y}, sonolus.Vec2{X: p.X + 3.5*PL.W, Y: p.Y}, half), 2, a)
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

func (n *PreviewBPMChange) Render() {
	p := pvPos(3, preview.Time.BeatToTime(n.Beat))
	preview.Canvas.Print(preview.PrintOptions{
		Value: n.BPM, Format: sonolus.PrintFormatBPM, DecimalPlaces: 0,
		Anchor: sonolus.Vec2{X: p.X + 3.6*PL.W, Y: p.Y}, Pivot: sonolus.Vec2{X: 0, Y: 0.5},
		Size: sonolus.Vec2{X: PL.W * 1.2, Y: 0.04}, Alpha: 1, Color: sonolus.PrintColorPurple,
		HorizontalAlign: sonolus.HorizontalAlignLeft,
	})
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
	Beat              float64 `archetype:"imported,name=#BEAT"`
	Lane              float64 `archetype:"imported,name=lane"`
	SpriteLane        float64 `archetype:"imported,name=spriteLane"`
	Width             float64 `archetype:"imported,name=width"`
	SpanStart         float64 `archetype:"imported,name=spanStart"`
	Skill             bool    `archetype:"imported,name=skill"`
	Direction         float64 `archetype:"imported,name=direction"`
	Time              float64 `archetype:"data"`
}

func (n *PreviewNoteBase) prepare() {
	if Config.Mirror {
		n.Lane = 6 - n.Lane
		n.SpanStart = 6 - (n.SpanStart + n.Width - 1)
		n.SpriteLane = 6 - n.SpriteLane
		n.Direction = -n.Direction
	}
	n.Time = preview.Time.BeatToTime(n.Beat)
	PL.LastTime = math.Max(PL.LastTime, n.Time)
	PL.LastBeat = math.Max(PL.LastBeat, math.Ceil(n.Beat))
}

func (n *PreviewNoteBase) z() sonolus.ZIndex { return noteZ(zNote, n.Time, n.Lane) }

func (n *PreviewNoteBase) Preprocess() { n.prepare() }

// pvTerminal draws a Long / Slide end or gesture node.
func (n *PreviewNoteBase) pvTerminal(kind, kindWidth float64, z sonolus.ZIndex) {
	switch kind {
	case NodeFlick:
		var p sonolus.Vec2
		if Layout.HabaSkin {
			p = pvWide(Skin.HabaFlick[habaRun(n.SpanStart, n.Width)], habaWidthPx(n.Width), n.Lane, n.Time, z)
		} else {
			p = pvNote(lanePick(Skin.NoteFlick, n.SpriteLane), n.Lane, n.Time, z)
		}
		pvFlickTop(p, n.Width, z)
	case NodeDirectional, NodeMultiDir:
		members := 1.0
		if kind == NodeMultiDir {
			members = math.Max(kindWidth, 1)
		}
		var p sonolus.Vec2
		for k := 0.0; k < members && k < 7; k++ {
			ml := n.Lane + n.Direction*k
			body := lanePick(Skin.NoteFlickR, n.SpriteLane+ml-n.Lane)
			if n.Direction < 0 {
				body = lanePick(Skin.NoteFlickL, n.SpriteLane+ml-n.Lane)
			}
			p = pvNote(body, ml, n.Time, subZ(z, k))
		}
		pvArrow(p, n.Direction, iconZ(z))
	default:
		n.pvLong(z)
	}
}

func pvArrow(p sonolus.Vec2, direction float64, z sonolus.ZIndex) {
	s := pvScale()
	if direction < 0 {
		Skin.NoteFlickTopL.DrawZ(spriteQuad(sonolus.Vec2{X: p.X - 1.9*s, Y: p.Y}, 138, 171, notePPU, s), z, 1)
	} else {
		Skin.NoteFlickTopR.DrawZ(spriteQuad(sonolus.Vec2{X: p.X + 1.9*s, Y: p.Y}, 138, 170, notePPU, s), z, 1)
	}
}

type PreviewTapNote struct {
	PreviewNoteBase   `archetype:"base"`
	preview.Archetype `archetype:"name=TapNote"`
	Offbeat           bool `archetype:"imported,name=offbeat"`
}

func (n *PreviewTapNote) Render() {
	if Layout.HabaSkin {
		run := habaRun(n.SpanStart, n.Width)
		sp := Skin.HabaNormal[run]
		if n.Skill {
			sp = Skin.HabaSkill[run]
		} else if n.Offbeat && Config.NoteColor {
			sp = Skin.HabaNormal16[run]
		}
		pvWide(sp, habaWidthPx(n.Width), n.Lane, n.Time, n.z())
		return
	}
	sp := lanePick(Skin.NoteNormal, n.SpriteLane)
	if n.Skill {
		sp = lanePick(Skin.NoteSkill, n.SpriteLane)
	} else if n.Offbeat && Config.NoteColor {
		sp = lanePick(Skin.NoteNormal16, n.SpriteLane)
	}
	pvNote(sp, n.Lane, n.Time, n.z())
}

type PreviewFlickNote struct {
	PreviewNoteBase   `archetype:"base"`
	preview.Archetype `archetype:"name=FlickNote"`
}

func (n *PreviewFlickNote) Render() { n.pvTerminal(NodeFlick, 1, n.z()) }

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
	body := lanePick(Skin.NoteFlickR, n.SpriteLane)
	if n.Direction < 0 {
		body = lanePick(Skin.NoteFlickL, n.SpriteLane)
	}
	p := pvNote(body, n.Lane, n.Time, memberZ(n.z(), n.Direction, n.Lane, groupFirst(n.Leader.Get().Lane, n.Count)))
	if n.Arrow {
		pvArrow(p, n.Direction, iconZ(n.z()))
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
	n.pvLong(n.z())
}

func (n *PreviewLongTail) Render() { n.pvTerminal(n.Kind, n.Width, n.z()) }

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
	n.pvTerminal(n.Kind, n.KindWidth, n.z())
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
	n.pvTerminal(n.Kind, n.KindWidth, n.z())
}

type PreviewSimLine struct {
	preview.Archetype `archetype:"name=SimLine"`
	A                 sonolus.EntityRef[PreviewNoteBase] `archetype:"imported,name=a"`
	B                 sonolus.EntityRef[PreviewNoteBase] `archetype:"imported,name=b"`
	ALane             float64                            `archetype:"imported,name=aLane,default=-1000"`
	BLane             float64                            `archetype:"imported,name=bLane,default=-1000"`
}

func (l *PreviewSimLine) Render() {
	if !Config.SyncLine {
		return
	}
	a, b := l.A.Get(), l.B.Get()
	la, lb := simLane(l.ALane, a.Lane), simLane(l.BLane, b.Lane)
	p, q := pvPos(la, a.Time), pvPos(lb, b.Time)
	if math.Abs(p.Y-q.Y) > 1e-6 {
		return // different columns or times
	}
	Skin.SimLine.Draw(lineQuad(p, q, PL.W*0.06), zSyncLine, 1)
}

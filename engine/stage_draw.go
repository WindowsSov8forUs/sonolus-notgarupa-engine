package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
)

// Layers (z). Larger z draws on top.
const (
	zField      = 10.0
	zInfo       = 20.0 // combo, judge text, AddScore (sorting order 0 panel)
	zLaneEffect = 30.0
	zDirLine    = 31.0
	zJudgeLine  = 40.0
	zBelt       = 60.0
	zSyncLine   = 69.0
	zNote       = 70.0
	zIcon       = 71.0
	zHUD        = 100.0
)

// noteZ orders a note-layer sprite like the original (C §1.10): layer, then
// earlier chart time on top, then lower lane on top. Each key is its own
// Sonolus z (z1..z3) so none is lost to rounding; z4 orders sprites inside one
// note (subZ).
func noteZ(layer, time, lane float64) sonolus.ZIndex {
	return sonolus.ZIndex{Z1: layer, Z2: -time, Z3: -lane}
}

// lineZ orders bodies and connector lines within a layer: earlier on top.
func lineZ(layer, time float64) sonolus.ZIndex { return sonolus.ZIndex{Z1: layer, Z2: -time} }

// iconZ moves a note's z to the icon / flash layer (sorting order 71).
func iconZ(z sonolus.ZIndex) sonolus.ZIndex {
	z.Z1 += zIcon - zNote
	return z
}

// subZ draws a sprite of the same note above the earlier ones (larger on top).
func subZ(z sonolus.ZIndex, sub float64) sonolus.ZIndex {
	z.Z4 += sub
	return z
}

// simLane is a sync line end's drawing lane: the converter's aLane / bLane
// (mirrored like the notes), or the note's own (already mirrored) lane in
// level data written before they existed.
func simLane(imported, noteLane float64) float64 {
	if imported <= -1000 {
		return noteLane
	}
	if Config.Mirror {
		return 6 - imported
	}
	return imported
}

// groupFirst is the leftmost lane of a directional group from its leader
// (the first member in original order; mirrored, it is the rightmost).
func groupFirst(leaderLane, count float64) float64 {
	if Layout.Mirror {
		return leaderLane - (count - 1)
	}
	return leaderLane
}

// memberZ orders linked directional members like the original's chained
// depth nudge (±0.00001 per link from the leftmost member, twice the 0.000005
// lane step): z3 = −first + 2·direction·(lane − first).
func memberZ(z sonolus.ZIndex, direction, lane, first float64) sonolus.ZIndex {
	z.Z3 = -first + 2*direction*(lane-first)
	return z
}

// drawField draws the lane image and the judge line at alpha la (opening fade).
func drawField(now, la float64) {
	field, judge, judgePx := Skin.Field, Skin.JudgeLine, 38.0
	if Layout.HabaSkin && Layout.LaneChange >= 0 && now >= Layout.LaneChange+laneChangeSwap {
		// HABAHIRO lane change: only the sprites change (habahiro field and
		// its 43 px judge line); before it the field is skin00's
		field, judge, judgePx = Skin.HabaField, Skin.HabaJudgeLine, 43
	}
	ml := Layout.MLines
	u := Layout.UB // the lane image is a UITexture under RhythmGameLines (2/activeHeight)
	bottom := (13*Layout.HAR - 240*ml) * u
	half := 580 * ml * u
	height := 610 * ml * u
	field.Draw(sonolus.Rect{T: bottom + height, R: half, B: bottom, L: -half}.ToQuad(), zField, la)
	// the judge line is a Button4 child: it follows the note size
	w := 1800.0 / judgeLinePPU * 0.99 * Layout.S * Layout.N
	h := judgePx / judgeLinePPU * 0.99 * Layout.S * Layout.N
	judge.Draw(sonolus.Rect{T: Layout.G + h/2, R: w / 2, B: Layout.G - h/2, L: -w / 2}.ToQuad(), zJudgeLine, la)
	drawLaneChangeFlash(now)
	drawSuddenBar(la)
}

// laneChangeSwap is the ChangeLane event of the lane-change animation (frame 25).
const laneChangeSwap = 0.4166666567325592

// smoothstep interpolates between keyframes a and b with zero tangents.
func smoothstep(a, b, r float64) float64 { return a + (b-a)*(3*r*r-2*r*r*r) }

// Lane-change flash alpha keys: A (0, 0.001) (0.0833, 1) (0.75, 1) (1, 0.001);
// B (0, 0.001) (0.3833, 1) (0.4167, 1) (0.45, 1) (1, 0.001).
func flashAlphaA(t float64) float64 {
	switch {
	case t < 0.0833333:
		return smoothstep(0.001, 1, t/0.0833333)
	case t < 0.75:
		return 1
	}
	return smoothstep(1, 0.001, math.Min((t-0.75)/0.25, 1))
}

func flashAlphaB(t float64) float64 {
	switch {
	case t < 0.3833333:
		return smoothstep(0.001, 1, t/0.3833333)
	case t < 0.45:
		return 1
	}
	return smoothstep(1, 0.001, math.Min((t-0.45)/0.55, 1))
}

// drawLaneChangeFlash draws the HABAHIRO lane-change flash (1 s): the lane and
// judge-line flash sprites over RhythmGameLines-local rects, above the judge
// line and below the notes.
func drawLaneChangeFlash(now float64) {
	e := now - Layout.LaneChange
	if !Layout.HabaSkin || Layout.LaneChange < 0 || e < 0 || e > 1 {
		return
	}
	k := Layout.MLines * Layout.UB
	rect := func(x0, y0, x1, y1 float64) sonolus.Quad {
		return sonolus.Rect{T: rhythmLinesY(y1), R: x1 * k, B: rhythmLinesY(y0), L: x0 * k}.ToQuad()
	}
	a, b := flashAlphaA(e), flashAlphaB(e)
	Skin.HabaFlash[0].Draw(rect(-580, -242, 580, 368), zJudgeLine+0.5, a)
	Skin.HabaFlash[1].Draw(rect(-580, -242, 580, 368), zJudgeLine+0.5, b)
	// obj_scale y −242 ×100, obj2_pos +3.05: game_play_line_flash at −3.0
	// (centre −237), game_play_line_flash2 at −3.05 (centre −242)
	Skin.HabaFlash[2].Draw(rect(-940, -297, 940, -177), zJudgeLine+0.6, a)
	Skin.HabaFlash[3].Draw(rect(-940, -302, 940, -182), zJudgeLine+0.6, b)
}

// drawSuddenBar draws the Sudden appear bar (sudden_line, 9-slice 26/26 px
// borders) at the lower edge of the cover, below the notes.
func drawSuddenBar(alpha float64) {
	r := Layout.SuddenR
	if r <= 0 {
		return
	}
	k := Layout.M * Layout.UB
	y := rhythmLinesY(368.39 - 608.391*r)
	half, hh, cap := (1.98+105.67*r)*10*k/2, 3*k/2, 2.6*k
	a := 0.6980392336845398 * alpha
	z := zSyncLine + 0.5
	quad := func(l, r float64) sonolus.Quad {
		return sonolus.Rect{T: y + hh, R: r, B: y - hh, L: l}.ToQuad()
	}
	Skin.SuddenLine[0].Draw(quad(-half, -half+cap), z, a)
	Skin.SuddenLine[1].Draw(quad(-half+cap, half-cap), z, a)
	Skin.SuddenLine[2].Draw(quad(half-cap, half), z, a)
}

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
)

// Original result grades.
const (
	gradeNone    float64 = -1
	gradeMiss    float64 = 0
	gradeBad     float64 = 1
	gradeGood    float64 = 2
	gradeGreat   float64 = 3
	gradePerfect float64 = 4
)

// Node kinds of a Long / Slide end or gesture node (converter NodeKind).
const (
	NodePlain       float64 = 0
	NodeFlick       float64 = 1
	NodeDirectional float64 = 2 // width 1
	NodeMultiDir    float64 = 3 // width > 1
)

// Judge timing (FAST / SLOW display).
const (
	timingNone float64 = 0
	timingFast float64 = 1
	timingSlow float64 = 2
)

// Judge note types (original JudgeNoteType): select SE, particles, damage.
const (
	jtNormal        float64 = 0
	jtLongMiss      float64 = 1
	jtLongEnd       float64 = 2
	jtFlick         float64 = 3
	jtLongFront     float64 = 4
	jtLongEndFlick  float64 = 5
	jtLongEndDir    float64 = 6
	jtLongEndMulti  float64 = 7
	jtSlide         float64 = 8
	jtDirectional   float64 = 9
	jtMultiDir      float64 = 10
	missSeconds             = 0.21666666865348816 // MissSecondInterval, 13/60 s
	frameSeconds            = 1.0 / 60
	flickWaitFrames         = 7.0
	leaveGrace              = 8.0
)

// roundHalfEven rounds to the nearest integer, ties to even (original GetResult).
func roundHalfEven(v float64) float64 {
	f := math.Floor(v)
	d := v - f
	if d > 0.5 {
		return f + 1
	}
	if d < 0.5 {
		return f
	}
	if math.Mod(f, 2) == 0 {
		return f
	}
	return f + 1
}

// timeGrade grades |Δt| in seconds with the original frame windows; sweet is
// 0 for heads and singles, 1 for Long tails.
func timeGrade(dt, sweet float64) float64 {
	r := roundHalfEven(math.Abs(dt) / frameSeconds)
	if r < sweet+3 {
		return gradePerfect
	}
	if r < sweet+6 {
		return gradeGreat
	}
	if r < sweet+7 {
		return gradeGood
	}
	if r < sweet+8 {
		return gradeBad
	}
	return gradeNone
}

// timeWindowEnd is the last late offset (s) that can still grade.
func timeWindowEnd(sweet float64) float64 { return (sweet + 7.5) * frameSeconds }

func timeTiming(grade, dt float64) float64 {
	if grade == gradePerfect || grade == gradeMiss {
		return timingNone
	}
	if dt <= 0 {
		return timingSlow
	}
	return timingFast
}

// slideGrade grades a slide node by its time before the virtual perfect line
// (tau, frames, positive = early). Late side is always Perfect.
func slideGrade(tau float64) float64 {
	if tau <= 3 {
		return gradePerfect
	}
	if tau <= 6 {
		return gradeGreat
	}
	if tau <= 7 {
		return gradeGood
	}
	if tau <= 8 {
		return gradeBad
	}
	return gradeNone
}

// slideTau returns frames by which time t precedes the node's VPL sample.
func slideTau(nodeTime, t float64) float64 {
	return (nodeTime + Layout.VPLDelay - t) * 60
}

// sonolusJudgment maps an original grade to the Sonolus judgment.
func sonolusJudgment(grade float64) sonolus.Judgment {
	switch grade {
	case gradePerfect:
		return sonolus.JudgmentPerfect
	case gradeGreat:
		return sonolus.JudgmentGreat
	case gradeGood, gradeBad:
		return sonolus.JudgmentGood
	}
	return sonolus.JudgmentMiss
}

// Arcade score: the Sonolus Level Score shown on the result screen, which the
// HUD shows as well (as Next-RUSH does). Each judgement scores
// judgmentMultiplier × (1 + consecutive bonus) and the sum is normalised to
// 1 000 000. The weights keep the original's structure: result rates 1.1 /
// 0.8 / 0.5 (Bad counts as Sonolus Good) and a combo bonus up to +0.11 (the
// original's combo correction 1.00…1.11, approximated as +0.01 per 50
// consecutive Great-or-better).
const (
	arcadePerfect, arcadeGreat, arcadeGood = 1.1, 0.8, 0.5
	arcadeStep, arcadeBonus, arcadeCap     = 50.0, 0.01, 550.0
)

// arcadeMultiplier is the judgment multiplier of a grade.
func arcadeMultiplier(grade float64) float64 {
	switch grade {
	case gradePerfect:
		return arcadePerfect
	case gradeGreat:
		return arcadeGreat
	case gradeGood, gradeBad:
		return arcadeGood
	}
	return 0
}

// arcadeWeight is a judgement's score after a streak of streak consecutive
// Great-or-better judgements (counting itself): the bonus accumulates per
// step reached, up to the cap (Level Score consecutive Great).
func arcadeWeight(grade, streak float64) float64 {
	return arcadeMultiplier(grade) * (1 + math.Floor(math.Min(streak, arcadeCap)/arcadeStep)*arcadeBonus)
}

// lifeDamage returns the life change of a grade for a judge note type.
func lifeDamage(grade, judgeType float64, fullLength bool) float64 {
	miss, bad := -100.0, -50.0
	if fullLength {
		miss, bad = -50, -25
	}
	switch grade {
	case gradeMiss:
		return miss
	case gradeBad:
		return bad
	}
	return 0
}

// The extra particle anchor scale: one particle unit in world units.
func particleUnit() float64 { return Layout.Rate * Layout.N * Layout.Safe }

// particleQuad returns the spawn quad of a judgement particle at x on the judge
// line (scaleY squashes lane-plane effects).
func particleQuad(x, scaleY float64) sonolus.Quad {
	g := particleUnit()
	return sonolus.Rect{T: Layout.G + g*scaleY, R: x + g, B: Layout.G - g*scaleY, L: x - g}.ToQuad()
}

func gestureDir(kind, direction float64) float64 {
	if kind == NodeDirectional || kind == NodeMultiDir {
		return direction
	}
	return 0
}

func gestureCount(kind, width float64) float64 {
	if kind == NodeMultiDir {
		return width
	}
	return 1
}

// endType is the judge note type of a flick-family end; a gesture head or
// intermediate node judges its gesture as the same original Slide-tail
// family (chart extension).
func endType(kind float64) float64 {
	switch kind {
	case NodeFlick:
		return jtLongEndFlick
	case NodeDirectional:
		return jtLongEndDir
	case NodeMultiDir:
		return jtLongEndMulti
	}
	return jtSlide
}

// dirBodyLane is where a directional judgement's body particle plays: the
// end of the span the flick starts from (Left: the right end).
func dirBodyLane(spanStart, width, direction float64) float64 {
	if direction < 0 {
		return spanStart + width - 1
	}
	return spanStart
}

// receiverX is where a note's judgement particles play: the middle button for
// odd widths, the half button between the two middle ones for even widths.
func receiverLane(spanStart, width float64) float64 {
	return spanStart + (width-1)/2
}

//go:build play || watch

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
)

// HudState is the live HUD model, written where judgements are applied
// (play: touch / updateSequential; watch: updateSequential) and drawn by the
// Stage in updateParallel.
type HudState struct {
	sonolus.LevelMemoryResource
	Combo        float64
	MaxCombo     float64
	ComboStart   float64 // time the combo number last changed
	Score        float64
	ScoreRaw     float64 // arcade weight earned so far
	Life         float64
	JudgeSprite  float64 // index into Skin.JudgeText, −1 hidden
	JudgeTiming  float64
	JudgeStart   float64
	JudgeGrade   float64    // raw result of the shown judge text (frame representative)
	AddValue     [4]float64 // AddScore pool: 4 roots, round robin
	AddStart     [4]float64
	AddSeq       [4]float64
	AddNext      float64
	LastAdded    float64 // score added by the last judgement
	AllPerfect   bool
	PerfectCount float64
	GreatCount   float64
	Judged       float64
	GameOver     bool
	// GameOverCover (life vignette): logical level 0–3, the running scale
	// tween (from, to, start) and whether the texture is enabled
	CoverLevel float64
	CoverFrom  float64
	CoverTo    float64
	CoverStart float64
	CoverOn    bool
	SSStart    float64
	// life warning (life ≤ 25 %): its TweenAlpha starts at 0.6 when the
	// warning switches on; the AP combo pulse starts when it first shows
	Warn      bool
	WarnStart float64
	APShown   bool
	APStart   float64
	StageS    float64 // Stage TRS scale and y of the opening (updateStagePose)
	StageY    float64
	HudAlpha  float64 // opening fade of the gameplay HUD layer (updateStagePose)
	// judgement particles per owner (receiver half-lane × prefab family),
	// restarted by the next judgement of the same owner
	Fx0           [224]sonolus.ParticleHandle // fxKey: 14 half-lanes × fxOwnerFamilies
	Fx1           [224]sonolus.ParticleHandle
	Fx2           [224]sonolus.ParticleHandle
	SilencePos    float64
	SilenceType   float64
	SilenceGrade  float64
	SilenceCount  float64
	SilenceTime   float64
	SpeakerKind   [2]float64 // 0 idle, 1 normal, 2 exciting, 3 continuous
	SpeakerStart  [2]float64
	LightPale     [12]float64 // colour-table entry each light shows pale (random once)
	ScoreTotal    float64     // arcade weight of an all-Perfect full combo
	ScoreMax      float64
	RankThreshold [5]float64
	NoteCount     float64
	FullLength    bool // level flag: halved life damage (original musicDataType "full")
}

var Hud = HudState{}

// Score-rank thresholds on the 1 000 000 scale: the original's ratios
// 1 : 6 : 12 : 18 : 24 (300·(E+95)·…) with SS at 90 %, as GarupaEditor's
// normalised rule; the gauge spans SS × 1.111111, i.e. the whole million.
func initRanks() {
	f := [5]float64{1, 6, 12, 18, 24}
	for i := 0; i < 5; i++ {
		Hud.RankThreshold[i] = 37500 * f[i]
	}
	Hud.ScoreMax = 1000000
}

// ---------------------------------------------------------------- drawing ----

// sliced3 draws a horizontally three-sliced sprite from left x over width w
// (UI units), between y bottom/top (screen units already).
func sliced3(parts [3]sonolus.Sprite, left, right, bottom, top, capL, capR, z, a float64) {
	u := Layout.U
	l2 := left + capL*u
	r2 := right - capR*u
	if r2 < l2 {
		m := (l2 + r2) / 2
		l2, r2 = m, m
	}
	a *= Hud.HudAlpha
	parts[0].Draw(sonolus.Rect{T: top, R: l2, B: bottom, L: left}.ToQuad(), z, a)
	parts[1].Draw(sonolus.Rect{T: top, R: r2, B: bottom, L: l2}.ToQuad(), z, a)
	parts[2].Draw(sonolus.Rect{T: top, R: right, B: bottom, L: r2}.ToQuad(), z, a)
}

// sliced9 draws a nine-sliced sprite with uniform borders (UI units).
func sliced9(parts [9]sonolus.Sprite, left, right, bottom, top, bl, br, bt, bb, z, a float64) {
	u := Layout.U
	xs := [4]float64{left, left + bl*u, right - br*u, right}
	ys := [4]float64{top, top - bt*u, bottom + bb*u, bottom}
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			parts[row*3+col].Draw(sonolus.Rect{T: ys[row], R: xs[col+1], B: ys[row+1], L: xs[col]}.ToQuad(), z, a*Hud.HudAlpha)
		}
	}
}

// glyph draws a pre-rendered glyph with its pen origin at (penX, baseline).
func glyph(s sonolus.Sprite, penX, baseline, w, h, originX, descent, scale, z, a float64) {
	u := Layout.U * scale
	l := penX - originX*u
	b := baseline - descent*u
	s.Draw(sonolus.Rect{T: b + h*u, R: l + w*u, B: b, L: l}.ToQuad(), z, a*Hud.HudAlpha)
}

const (
	scoreDigitW, scoreDigitH, scoreDigitAdvance = 23.0, 29.0, 21.25
	scoreAscent, scoreDescent                   = 24.640625, 3.3671875
	lifeDigitW, lifeDigitH, lifeDigitAdvance    = 15.0, 18.0, 13.65625
	lifeSlashW, lifeSlashAdvance                = 10.0, 9.0
	lifeAscent, lifeDescent                     = 15.8359375, 2.1640625
)

func drawScore(z float64) {
	// Panel (gauge_base_score, L216 R16) at x 25..495, y_down 27..109.
	p0 := uiTopLeft(25, 109)
	p1 := uiTopLeft(495, 27)
	sliced3(Skin.ScorePanel, p0.X, p1.X, p0.Y, p1.Y, 216, 16, z, 1)
	// Fill by rank: x 66..487, y_down 34..58.
	ratio := sonolus.Clamp(Hud.Score/math.Max(Hud.ScoreMax, 1), 0, 1)
	rank := 0
	for i := 0; i < 5; i++ {
		if Hud.Score >= Hud.RankThreshold[i] {
			rank = i + 1
		}
	}
	f0 := uiTopLeft(66, 58)
	f1 := uiTopLeft(66, 34)
	if ratio >= 0.001 {
		switch rank {
		case 0:
			w := 8 + 413*ratio
			sliced9(Skin.MeterBlue, f0.X, f0.X+w*Layout.U, f0.Y, f1.Y, 4, 4, 3, 3, z+0.1, 1)
		case 1, 2, 3:
			w := 10 + 411*ratio
			parts := Skin.MeterGreen
			if rank == 2 {
				parts = Skin.MeterOrange
			} else if rank == 3 {
				parts = Skin.MeterPink
			}
			sliced3(parts, f0.X, f0.X+w*Layout.U, f0.Y, f1.Y, 5, 5, z+0.1, 1)
		default:
			Skin.MeterS.Draw(sonolus.Rect{T: f1.Y, R: f0.X + 421*ratio*Layout.U, B: f0.Y, L: f0.X}.ToQuad(), z+0.1, Hud.HudAlpha)
		}
	}
	// Cover ring: x 63..490, y_down 29.5..62.5, 9-slice 8.
	c0 := uiTopLeft(63, 62.5)
	c1 := uiTopLeft(490, 29.5)
	sliced9(Skin.ScoreCover, c0.X, c1.X, c0.Y, c1.Y, 8, 8, 8, 8, z+0.2, 1)
	// Rank marks C, B, A, S, SS.
	for i := 0; i < 5; i++ {
		x := 25 + 41 + Hud.RankThreshold[i]*421/math.Max(Hud.ScoreMax, 1)
		m := uiTopLeft(x, 53)
		Skin.LevelMark.Draw(uiRect(m, 8, 6), z+0.3, Hud.HudAlpha)
		c := uiTopLeft(x, 45)
		baseline := c.Y - (10.5625+1.4453125)/2*Layout.U + 1.4453125*Layout.U
		if i < 4 {
			glyph(Skin.RankLetter[i], c.X-9.4/2*Layout.U, baseline, 11, 13, 0.5, 1.4453125, 1, z+0.4, 1)
		} else {
			glyph(Skin.RankLetter[3], c.X-9.109375*Layout.U, baseline, 11, 13, 0.5, 1.4453125, 1, z+0.4, 1)
			glyph(Skin.RankLetter[3], c.X, baseline, 11, 13, 0.5, 1.4453125, 1, z+0.4, 1)
		}
	}
	// Total score: right edge x = 212, centre y_down 84, min 8 digits, pitch advance+1.
	digits := 1
	v := math.Floor(Hud.Score)
	for t := v; t >= 10; t = math.Floor(t / 10) {
		digits++
	}
	// the original shrinks by integer font size until the run fits 188:
	// 8 digits size 28 pitch 22, 9 digits 27 / 21, 10 digits 23 / 18
	count := max(digits, 8)
	size, pitch := 1.0, 22.0
	if count == 9 {
		size, pitch = 27.0/28, 21
	} else if count >= 10 {
		size, pitch = 23.0/28, 18
	}
	right := uiTopLeft(212, 84)
	baseline := right.Y - (scoreAscent+scoreDescent)/2*size*Layout.U + scoreDescent*size*Layout.U
	for i := 0; i < count && i < 10; i++ {
		d := math.Mod(math.Floor(v/math.Pow(10, float64(i))), 10)
		pen := right.X - (float64(i+1)*pitch-1)*Layout.U
		idx := int(d)
		if i < digits {
			glyph(Skin.ScoreDigitPink[idx], pen, baseline, scoreDigitW, scoreDigitH, 0.5, scoreDescent, size, z+0.4, 1)
		} else {
			glyph(Skin.ScoreDigitGray[0], pen, baseline, scoreDigitW, scoreDigitH, 0.5, scoreDescent, size, z+0.4, 1)
		}
	}
}

// AddScore pop: number root at x 282 (safe top-left), moving along its local
// X; three phases of 9 updates (≥ 0.14 s each), 4 roots round robin.
var addInner = [10]float64{43, 30, 45, 45, 47, 46, 46, 45, 46, 44}

// addPop starts a pop on the next root of the pool.
func addPop(value, start float64) {
	i := int(math.Mod(Hud.AddNext, 4))
	Hud.AddValue[i], Hud.AddStart[i], Hud.AddSeq[i] = value, start, Hud.AddNext
	Hud.AddNext++
}

func clearAddPops() {
	for i := 0; i < 4; i++ {
		Hud.AddStart[i] = 0
	}
	Hud.AddNext = 0
}

func drawAddScore(now, z float64) {
	for i := 0; i < 4; i++ {
		// later pops draw on top (depth cycles over 8)
		drawAddPop(Hud.AddValue[i], now-Hud.AddStart[i], Hud.AddStart[i] > 0, z+math.Mod(Hud.AddSeq[i], 8)*0.001)
	}
}

func drawAddPop(value, t float64, live bool, z float64) {
	f := math.Floor(t * 60)
	if !live || t < 0 || f >= 27 {
		return
	}
	phase := math.Floor(f / 9)
	within := f - 9*phase
	// AddScoreObject.MoveNext resumes on the creation frame: the first frame
	// is already one step in
	e := (within + 1) / 60
	off, alpha := 0.0, 1.0
	switch phase {
	case 0:
		off = 8 * (within + 1)
		alpha = math.Min(0.2+0.8*e/0.14, 1)
	case 1:
		off = 72 + within + 1
	default:
		off = 81 + within + 1
		alpha = math.Max(1-e/0.14, 0)
	}
	value = math.Floor(value)
	digits := 1
	for v := value; v >= 10; v = math.Floor(v / 10) {
		digits++
	}
	// AddScoreObject moves along the root's local X: authored (282, 85 down),
	// Play shifts x by −50, then every coroutine resume adds +8 / +1 / +1
	origin := uiTopLeft(282-50+off, 85)
	scale := 0.6
	Skin.AddPlus.Draw(uiRect(origin, 47*scale, 70*scale), z, alpha*Hud.HudAlpha)
	// centres: the plus at 0, then previous centre + inner width + 1
	c, inner := 0.0, 47.0
	for i := digits - 1; i >= 0; i-- {
		d := int(math.Mod(math.Floor(value/math.Pow(10, float64(i))), 10))
		c += inner + 1
		inner = addInner[d]
		p := sonolus.Vec2{X: origin.X + c*scale*Layout.U, Y: origin.Y}
		Skin.AddDigit[d].Draw(uiRect(p, 47*scale, 70*scale), z, alpha*Hud.HudAlpha)
	}
}

// Combo: root (default right centre 434.7, 82.8), digits 82×116 step 70.
var comboRoots = [6][2]float64{{338, 194}, {434.7, 82.8}, {530, -57}, {-338, 194}, {-434.7, 82.8}, {-505, -57}}

func comboScale(t float64) float64 {
	if t < 1.0/12 {
		return -1036.8*t*t*t + 129.6*t*t + 0.8
	}
	if t < 1.0/6 {
		d := t - 1.0/12
		return 1.1 - 1.2*d
	}
	return 1
}

func drawCombo(now, z float64) {
	// stays at its latest value until Combo returns to zero (the one-second
	// hide coroutine has no caller); a change replays the scale clip
	if Hud.Combo <= 0 {
		return
	}
	root := comboRoots[Config.ComboPos]
	rx, ry := root[0], root[1]
	if Config.ComboPos == 2 {
		rx += 60 * Layout.HAR
	} else if Config.ComboPos == 5 {
		rx -= 70 * Layout.HAR
	}
	s := comboScale(now - Hud.ComboStart)
	n := 1
	for v := Hud.Combo; v >= 10; v = math.Floor(v / 10) {
		n++
	}
	n = min(n, 4)
	ap := Config.ComboAP && Hud.AllPerfect
	apAlpha := 1.0
	if ap {
		ph := math.Mod(math.Max(now-Hud.APStart, 0), 0.8333333134651184)
		if ph < 0.41666666 {
			apAlpha = 1 - 0.5*sonolus.Smoothstep(ph/0.41666666)
		} else {
			apAlpha = 0.5 + 0.5*sonolus.Smoothstep((ph-0.41666666)/0.41666666)
		}
	}
	for i := 0; i < n; i++ {
		d := int(math.Mod(math.Floor(Hud.Combo/math.Pow(10, float64(i))), 10))
		x := 22 + (-70*float64(i+1) + 35*float64(n))
		c := uiPoint(rx+x*s, ry)
		Skin.ComboDigit[d].Draw(uiRect(c, 82*s, 116*s), z, Hud.HudAlpha)
		if ap {
			Skin.ComboDigitAP[d].Draw(uiRect(c, 82*s, 116*s), z+0.01, apAlpha*Hud.HudAlpha)
		}
	}
	c := uiPoint(rx-6*s, ry-72*s)
	// the COMBO unit (depth 6) sits above the digits (depth 5)
	Skin.ComboText.Draw(uiRect(c, 150*s, 42*s), z+0.02, Hud.HudAlpha)
	if ap {
		Skin.ComboTextAP.Draw(uiRect(c, 150*s, 42*s), z+0.03, apAlpha*Hud.HudAlpha)
	}
}

// Judge text: Result at (0, −160) scale 0.8, clip GameJudge, hidden after 1 s.
var judgeSize = [6][2]float64{{286, 78}, {224, 70}, {218, 70}, {148, 70}, {166, 70}, {204, 78}}

func drawJudge(now, z float64) {
	if Hud.JudgeSprite < 0 {
		return
	}
	t := now - Hud.JudgeStart
	if t < 0 || t >= 1 {
		return
	}
	s, a := 1.0, 1.0
	if t < 0.04 {
		s = -3125*t*t*t + 125*t*t + 7.5*t + 0.8
		a = -3125*t*t*t + 125*t*t + 10*t + 0.6
	} else if t < 0.08 {
		d := t - 0.04
		s = 3905.3*d*d*d - 281.2*d*d + 2.5*d + 1.1
	}
	a = math.Min(a, 1)
	idx := int(Hud.JudgeSprite)
	size := judgeSize[idx]
	c := uiPoint(0, -160)
	Skin.JudgeText[idx].Draw(uiRect(c, size[0]*0.8*s, size[1]*0.8*s), z, a*Hud.HudAlpha)
	if Hud.JudgeTiming != timingNone && Config.FastSlow {
		ft := uiPoint(4*0.8*s, -160-38*0.8*s)
		sp := Skin.JudgeText[6]
		if Hud.JudgeTiming == timingSlow {
			sp = Skin.JudgeText[7]
		}
		sp.Draw(uiRect(ft, 144*s, 30*s), z+0.01, a*Hud.HudAlpha)
	}
}

// Life gauge, anchored at the safe-area top-right corner.
func drawLife(now, z float64) {
	life := Hud.Life
	b0 := uiTopRight(-358, 88)
	b1 := uiTopRight(-92, 26)
	sliced3(Skin.LifeBG, b0.X, b1.X, b0.Y, b1.Y, 39, 143, z, 1)
	f0 := uiTopRight(-323, 57)
	f1 := uiTopRight(-99, 31)
	primary := math.Min(life/1000, 1)
	if primary > 0 {
		parts := Skin.LifeFill
		if life/1000 <= 0.20000000298023224 {
			parts = Skin.LifeFillLow
		}
		sliced3(parts, f0.X, f0.X+(8+216*primary)*Layout.U, f0.Y, f1.Y, 4, 4, z+0.1, 1)
	}
	second := math.Max(life/1000-1, 0)
	if second > 0 {
		sliced3(Skin.LifeFill2, f0.X, f0.X+(16+208*second)*Layout.U, f0.Y, f1.Y, 8, 8, z+0.2, 1)
	}
	if life/1000 <= 0.25 { // also at game over
		ph := math.Mod(math.Max(now-Hud.WarnStart, 0), 2)
		tw := ph
		if ph > 1 {
			tw = 2 - ph
		}
		wa := 0.6000000238418579 + 0.4*tw
		w := uiTopRight(-224, 58)
		// the warning (depth 44) sits above the life label (depth 42)
		sliced3(Skin.LifeWarning, w.X-287/2*Layout.U, w.X+287/2*Layout.U, w.Y-83/2*Layout.U, w.Y+83/2*Layout.U, 15, 158, z+0.45, wa)
		wb := uiTopRight(-213, 46)
		sliced9(Skin.LifeWarningBody, wb.X-114*Layout.U, wb.X+114*Layout.U, wb.Y-13.5*Layout.U, wb.Y+13.5*Layout.U, 6, 6, 5, 5, z+0.45, 0.703*wa)
	}
	// "current/1000": right edge x = −104, centre y_down 74, size 18.
	right := uiTopRight(-104, 74)
	baseline := right.Y - (lifeAscent+lifeDescent)/2*Layout.U + lifeDescent*Layout.U
	pen := right.X
	// maximum "1000"
	for i := 0; i < 4; i++ {
		d := 0
		if i == 3 {
			d = 1
		}
		pen -= lifeDigitAdvance * Layout.U
		glyph(Skin.LifeDigitGreen[d], pen, baseline, lifeDigitW, lifeDigitH, 0.5, lifeDescent, 1, z+0.4, 1)
	}
	pen -= lifeSlashAdvance * Layout.U
	glyph(Skin.LifeSlash, pen, baseline, lifeSlashW, lifeDigitH, 0.5, lifeDescent, 1, z+0.4, 1)
	v := math.Max(math.Floor(life), 0)
	n := 1
	for t := v; t >= 10; t = math.Floor(t / 10) {
		n++
	}
	for i := 0; i < n; i++ {
		d := int(math.Mod(math.Floor(v/math.Pow(10, float64(i))), 10))
		pen -= lifeDigitAdvance * Layout.U
		sp := Skin.LifeDigitGreen[d]
		if v <= 0 {
			sp = Skin.LifeDigitRed[d]
		}
		glyph(sp, pen, baseline, lifeDigitW, lifeDigitH, 0.5, lifeDescent, 1, z+0.4, 1)
	}
}

// Auto Live caption (watch auto): pill 206×38 at (130, 134 down).
func drawAutoCaption(z float64) {
	c := uiTopLeft(130, 134)
	sliced3(Skin.AutoPill, c.X-103*Layout.U, c.X+103*Layout.U, c.Y-19*Layout.U, c.Y+19*Layout.U, 25, 25, z, 1)
	baseline := c.Y + Layout.U - (21.1171875+2.8828125)/2*Layout.U + 2.8828125*Layout.U
	glyph(Skin.AutoText, c.X-138/2*Layout.U, baseline, 140, 24, 0.5, 2.8828125, 1, z+0.1, 1)
}

// initHud resets the HUD model for a level of noteCount scoring units.
func initHud(noteCount float64) {
	Hud.NoteCount = noteCount
	Hud.StageS, Hud.StageY = 0.7, 111 // the opening's start pose
	Hud.ScoreTotal = 0
	for i := 1.0; i <= noteCount; i++ {
		Hud.ScoreTotal += arcadeWeight(gradePerfect, i)
	}
	initRanks()
	resetHudState()
}

// resetHudState clears the running HUD model (start of the level).
func resetHudState() {
	Hud.Combo, Hud.MaxCombo, Hud.ComboStart = 0, 0, 0
	Hud.Score, Hud.ScoreRaw, Hud.LastAdded = 0, 0, 0
	clearAddPops()
	Hud.Life = 1000
	Hud.AllPerfect = true
	Hud.Warn, Hud.WarnStart, Hud.APShown, Hud.APStart = false, 0, false, 0
	Hud.PerfectCount, Hud.GreatCount, Hud.Judged = 0, 0, 0
	Hud.GameOver = false
	resetCover()
	Hud.JudgeSprite, Hud.JudgeTiming, Hud.JudgeStart, Hud.JudgeGrade = -1, timingNone, 0, -1
	Hud.SilencePos, Hud.SilenceTime = -1e9, -1e9
	for i := 0; i < 7; i++ {
		LaneFx.On[i] = 0
		LaneFx.Reserve[i] = 0
	}
}

// coverScales are GameOverCoverManager's localScale per level (x = y; the
// level-1/2 values are float32 1.7 and 1.35).
var coverScales = [4]float64{3, 1.7000000476837158, 1.350000023841858, 1}

// coverLevel is GameOverCoverControllerBase.calcCoverLevel of
// min(life/1000, 1): > 0.3 level 0, ≤ 0 level 3, ≤ 0.15 level 2, else 1.
func coverLevel(life float64) float64 {
	r := math.Min(life/1000, 1)
	switch {
	case r > 0.30000001192092896:
		return 0
	case r <= 0:
		return 3
	case r <= 0.15000000596046448:
		return 2
	}
	return 1
}

func coverScale(now float64) float64 {
	k := sonolus.Clamp((now-Hud.CoverStart)/0.5, 0, 1)
	return Hud.CoverFrom + (Hud.CoverTo-Hud.CoverFrom)*k
}

func resetCover() {
	Hud.CoverLevel, Hud.CoverFrom, Hud.CoverTo, Hud.CoverStart, Hud.CoverOn = 0, coverScales[0], coverScales[0], -1e9, false
}

// updateCover is UpdateCover, run every frame with the current life: only a
// level change enables / disables the texture (non-zero low life follows
// isEnableFearOfDeath, on at the default light and graphics settings) and
// starts a 0.5 s linear DOScale from the current scale (no DOKill: the newer
// tween wins). snap jumps to the level's scale (watch seek).
func updateCover(now float64, snap bool) {
	level := coverLevel(Hud.Life)
	if snap {
		Hud.CoverLevel, Hud.CoverStart = level, -1e9
		Hud.CoverFrom, Hud.CoverTo = coverScales[int(level)], coverScales[int(level)]
		Hud.CoverOn = level > 0
		return
	}
	if level == Hud.CoverLevel {
		return
	}
	Hud.CoverFrom = coverScale(now)
	Hud.CoverTo = coverScales[int(level)]
	Hud.CoverStart = now
	Hud.CoverLevel = level
	Hud.CoverOn = level > 0
}

// drawCover draws GameOverCover: the UITexture sub_glow, 1334×1500 at (0, 190)
// under UI_Root_Back/Display (behind every UI_Root element). The original
// subtracts the texture (ReverseSubtract), so its white rim darkens the
// background; the skin carries it as black with that strength as alpha.
func drawCover(now, z float64) {
	if !Hud.CoverOn {
		return
	}
	s := coverScale(now) * Layout.UB
	y := 190 * Layout.UB
	Skin.GameOverCover.Draw(sonolus.Rect{T: y + 750*s, R: 667 * s, B: y - 750*s, L: -667 * s}.ToQuad(), z, 1)
}

// pauseCenter is the original PauseButton (68×70 widget, button_pause) under
// the safe-area top-right; the Sonolus menu button lies on it, transparent.
func pauseCenter() sonolus.Vec2 { return uiTopRight(-42, 54) }

func drawPauseButton(z float64) {
	Skin.PauseButton.Draw(uiRect(pauseCenter(), 68, 70), z, Hud.HudAlpha)
}

// drawHud draws every HUD element in its original layer.
func drawHud(now float64) {
	drawScore(zHUD)
	drawSSEffect(now, zHUD+0.5) // its own panel above the score panel's rank marks
	drawLife(now, zHUD+1)
	drawPauseButton(zHUD + 1)
	drawCombo(now, zInfo)
	drawJudge(now, zInfo+1)
	drawAddScore(now, zInfo+2)
}

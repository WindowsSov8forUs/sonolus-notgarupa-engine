//go:build play || watch

package main

import (
	"math"
	"math/rand"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
)

// Stage dressing of the original Standard live: stage lights, speakers and
// the SS gauge effect.

// stageRoot is the back UI root scale 1 − 0.23·HAR (no safe-area factor,
// 3c89bdf), in UI units.
func stageRoot() float64 { return 1 - 0.23*Layout.HAR }

// openingClock maps level time to the original opening (StartupDirection
// after the information view, GarupaEditor startupDirectionController): there
// BgCover fades out from 0.92 by 0.0075 a frame; from 0.5 s the Stage TRS
// tweens from (scale 0.7, y 111) to (0.92, 0) with OutQuad over 3 s while the
// stage backdrop brightens 0.5 → 1 over 0.75 s; from 2.5 s the lanes fade in
// over 1 s; the music starts at 3 s. Sonolus starts a level 2 s before the
// music, so the opening runs 1.5× faster with its 3 s at level time 0. The
// information view (jacket, title) and the SD characters are not drawn.
func openingClock(now float64) float64 { return (now + 2) * 1.5 }

// updateStagePose sets the Stage TRS of the opening and the 0.5 s fade-in
// of the gameplay HUD layer (score, life, combo, judge, AddScore, Auto label,
// pause button) that starts with it.
func updateStagePose(now float64) {
	Hud.HudAlpha = sonolus.Clamp(openingClock(now)/0.5, 0, 1)
	p := sonolus.Clamp((openingClock(now)-0.5)/3, 0, 1)
	p = p * (2 - p)
	Hud.StageS, Hud.StageY = 0.7+0.22*p, 111*(1-p)
}

// lineAlpha is the opening fade of the lanes, judge line and Sudden bar.
func lineAlpha(now float64) float64 { return sonolus.Clamp(openingClock(now)-2.5, 0, 1) }

// drawOpeningCovers draws BgCover (point_charactor_shadow tinted black,
// 2000² at scale 2.5, x 80 under UI_Root_Back) over the stage, and the
// backdrop's dimming (multiply by 0.5 + 0.5·progress) under the lights.
func drawOpeningCovers(now float64) {
	o := openingClock(now)
	k := Layout.UB
	q := sonolus.Rect{T: 2500 * k, R: 2580 * k, B: -2500 * k, L: -2420 * k}.ToQuad()
	if c := sonolus.Clamp((o-0.5)/0.75, 0, 1); c < 1 {
		Skin.BgCover.Draw(q, 1, 0.5*(1-c))
	}
	// 0.0075 per 60 Hz frame of the opening clock
	if a := 0.92 - 0.45*o; a > 0 {
		Skin.BgCover.Draw(q, 3.5, a)
	}
}

// stageK is the Stage scale under the back UI root.
func stageK() float64 { return Hud.StageS * stageRoot() }

// stageScale converts stage-space sizes to the UI units drawTrimmed expects.
func stageScale() float64 { return stageK() * Layout.UB / Layout.U }

// stagePoint converts stage-space UI units (centred, y up) to the screen;
// the stage root sits 14·HAR units higher on wide screens.
func stagePoint(x, y float64) sonolus.Vec2 {
	return sonolus.Vec2{X: x * stageK() * Layout.UB, Y: (y*stageK() + 14*Layout.HAR + Hud.StageY*stageRoot()) * Layout.UB}
}

// backgroundQuad places the generated background (1766.4 × 1640 UI units:
// the bgImage widget laid out at stage scale 0.92) under the same stage
// transform. Away from 0.92 the backdrop scales about its own centre, which
// sits 170·scale below the root and follows the pose's y.
func backgroundQuad() sonolus.Quad {
	k := stageRoot() * Layout.UB
	cy := 14 * Layout.HAR * Layout.UB
	f := Hud.StageS / 0.92
	c0, c := cy-170*0.92*k, cy+(Hud.StageY-170*Hud.StageS)*k
	return sonolus.Rect{T: c + (cy+820*k-c0)*f, R: 883.2 * k * f, B: c + (cy-820*k-c0)*f, L: -883.2 * k * f}.ToQuad()
}

// drawTrimmed draws a trimmed sprite so that its untrimmed full rect of
// w×h UI units is centred on c (trim: px removed left, top, right, bottom).
func drawTrimmed(sp sonolus.Sprite, c sonolus.Vec2, w, h, fullW, fullH float64, trim [4]float64, flipX bool, z, a float64) {
	kx, ky := w*Layout.U/fullW, h*Layout.U/fullH
	l := c.X - w*Layout.U/2 + trim[0]*kx
	r := c.X + w*Layout.U/2 - trim[2]*kx
	t := c.Y + h*Layout.U/2 - trim[1]*ky
	b := c.Y - h*Layout.U/2 + trim[3]*ky
	if flipX {
		l, r = 2*c.X-l, 2*c.X-r
	}
	sp.Draw(sonolus.Quad{BL: sonolus.Vec2{X: l, Y: b}, TL: sonolus.Vec2{X: l, Y: t}, TR: sonolus.Vec2{X: r, Y: t}, BR: sonolus.Vec2{X: r, Y: b}}, z, a)
}

// ---------------------------------------------------------- stage lights ----

var lightX = [12]float64{-540, -451, -354, -253, -152, -46, 60, 155, 266, 354, 451, 540}

// initStageLights picks each light's pale entry (0..6) at random, once.
func initStageLights() {
	for i := 0; i < 12; i++ {
		Hud.LightPale[i] = math.Floor(rand.Float64() * 7)
	}
}

// light sizes: lights 1, 2, 11 are 181.22×179.87, light 12 242.96×241.62;
// the white cores of lights 1–3 and 10–12 are 98.73×99.72.
var lightW = [12]float64{181.22, 181.22, 179.87, 179.87, 179.87, 179.87, 179.87, 179.87, 179.87, 179.87, 181.22, 242.96}
var lightH = [12]float64{179.87, 179.87, 179.87, 179.87, 179.87, 179.87, 179.87, 179.87, 179.87, 179.87, 179.87, 241.62}
var whiteW = [12]float64{98.73, 98.73, 98.73, 99.72, 99.72, 99.72, 99.72, 99.72, 99.72, 98.73, 98.73, 98.73}

var miniShineTrim = [4]float64{6, 4, 0, 3}

func drawStageLights(now, z float64) {
	if !Config.StageEffects {
		return
	}
	t := math.Max(now-0.3, 0)
	for i := 0; i < 12; i++ {
		// lerp entry k → k+1 over each 0.2 s step of the 7-entry table
		step := math.Floor(t / 0.2)
		f := (t - step*0.2) / 0.2
		k := math.Mod(step, 7)
		from, to := 0.0, 0.0
		if k == Hud.LightPale[i] {
			from = 1
		}
		if math.Mod(k+1, 7) == Hud.LightPale[i] {
			to = 1
		}
		w := from + (to-from)*f
		c := stagePoint(lightX[i], -99.9000015258789)
		sw, sh := lightW[i]*stageScale(), lightH[i]*stageScale()
		drawTrimmed(Skin.LightBlue, c, sw, sh, 128, 128, miniShineTrim, false, z, 1-w)
		drawTrimmed(Skin.LightPale, c, sw, sh, 128, 128, miniShineTrim, false, z, w)
		drawTrimmed(Skin.LightWhite, c, whiteW[i]*stageScale(), 99.72*stageScale(), 128, 128, miniShineTrim, false, z+0.01, 1)
	}
}

// -------------------------------------------------------------- speakers ----

// Animation table: speaker clips at 60 Hz (stageSpeakers.json, Root-local
// units), concatenated; cv* are their offsets.
const (
	cvSpkNTopA   = 0
	cvSpkNBotA   = 28
	cvSpkNGlowS  = 56
	cvSpkNTopX   = 84
	cvSpkNTopY   = 112
	cvSpkNBotX   = 140
	cvSpkNBotY   = 168
	cvSpkNRootS  = 196
	cvSpkNOvA    = 224
	cvSpkNOvS    = 252
	cvSpkERootY  = 280
	cvSpkERootSX = 297
	cvSpkERootSY = 314
	cvSpkEOvA    = 331
	cvSpkEOvS    = 348
	cvSpkEGlowA  = 365
	cvSpkETopS   = 382
	cvSpkEBotS   = 399
	cvSpkCRootS  = 416
	cvSpkCOvA    = 427
	cvSpkCOvS    = 438
	cvSpkCGlowA  = 449
	cvSpkCTopS   = 460
	cvSpkCBotS   = 471
	cvSpkCTopX   = 482
	cvSpkCTopY   = 493
)

var curveTable = [504]float64{0, 0, 0, 0.5, 1, 1, 1, 1, 0.82, 0.562, 0.273, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.5, 1, 1, 1, 1, 0.891, 0.626, 0.297, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.9, 0.89, 0.9, 0.94, 1, 1.07, 1.15, 1.23, 1.31, 1.39, 1.45, 1.5, 1.53, 1.55, 1.57, 1.58, 1.58, 1.58, 1.58, 1.57, 1.56, 1.55, 1.54, 1.53, 1.52, 1.51, 1.5, 1.5, -11, -10.75, -10.5, -10.25, -10, -9.63, -9.26, -9.008, -8.756, -8.504, -8.252, -8, -7.932, -7.864, -7.796, -7.728, -7.66, -7.6909, -7.7218, -7.7527, -7.7836, -7.8145, -7.8455, -7.8764, -7.9073, -7.9382, -7.9691, -8, 55, 55.75, 56.5, 57.25, 58, 58.25, 58.5, 58.2, 57.9, 57.6, 57.3, 57, 56.966, 56.932, 56.898, 56.864, 56.83, 56.8455, 56.8609, 56.8764, 56.8918, 56.9073, 56.9227, 56.9382, 56.9536, 56.9691, 56.9845, 57, -34.9, -35.425, -35.95, -36.475, -37, -37.2857, -37.5714, -37.8571, -38.1429, -38.4286, -38.7143, -39, -39.068, -39.136, -39.204, -39.272, -39.34, -39.3091, -39.2782, -39.2473, -39.2164, -39.1855, -39.1545, -39.1236, -39.0927, -39.0618, -39.0309, -39, -54.1, -54.575, -55.05, -55.525, -56, -57, -58, -59, -60, -61, -62, -63, -63.236, -63.472, -63.708, -63.944, -64.18, -64.0727, -63.9655, -63.8582, -63.7509, -63.6436, -63.5364, -63.4291, -63.3218, -63.2145, -63.1073, -63, 1, 1.005, 1.01, 1.015, 1.02, 1.02, 1.02, 1.02, 1.0133, 1.0067, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0.03, 0.03, 0.03, 0.418, 0.8, 0.931, 0.976, 0.955, 0.89, 0.801, 0.707, 0.63, 0.556, 0.464, 0.358, 0.247, 0.135, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 1, 1.011, 1.022, 1.033, 1.044, 1.055, 1.066, 1.077, 1.088, 1.099, 1.11, 1.11, 1.11, 1.11, 1.1015, 1.0931, 1.0846, 1.0762, 1.0677, 1.0592, 1.0508, 1.0423, 1.0338, 1.0254, 1.0169, 1.0085, 1, 1, 0, 25.9, 49, 65.1, 77.4, 88, 96, 93.5, 88.1, 80.1, 70.3, 59, 46.9, 34.3, 22, 10.4, 0, 1, 1.11, 1.2, 1.25, 1.28, 1.3, 1.3, 1.28, 1.26, 1.23, 1.2, 1.17, 1.13, 1.1, 1.06, 1.03, 1, 1, 1.16, 1.3, 1.36, 1.39, 1.4, 1.4, 1.38, 1.35, 1.31, 1.27, 1.23, 1.18, 1.13, 1.09, 1.04, 1, 0, 0.177, 0.309, 0.363, 0.395, 0.406, 0.401, 0.381, 0.35, 0.309, 0.262, 0.212, 0.16, 0.111, 0.066, 0.028, 0, 1, 1.065, 1.13, 1.195, 1.26, 1.26, 1.26, 1.234, 1.208, 1.182, 1.156, 1.13, 1.104, 1.078, 1.052, 1.026, 1, 1, 1, 1, 0.959, 0.91, 0.852, 0.787, 0.717, 0.641, 0.563, 0.481, 0.398, 0.315, 0.232, 0.152, 0.074, 0, 0.88, 0.88, 0.88, 0.9314, 0.9829, 1.0343, 1.0857, 1.1371, 1.1886, 1.24, 1.2914, 1.3429, 1.3943, 1.4457, 1.4971, 1.5486, 1.6, 0.9, 0.9, 0.9, 0.95, 1, 1.05, 1.1, 1.15, 1.2, 1.25, 1.3, 1.35, 1.4, 1.45, 1.5, 1.55, 1.6, 1, 1.03, 1.06, 1.09, 1.1, 1.1, 1.09, 1.07, 1.04, 1.02, 1, 0.03, 0.108, 0.193, 0.264, 0.3, 0.294, 0.26, 0.207, 0.145, 0.083, 0.03, 1, 1.022, 1.044, 1.066, 1.088, 1.11, 1.088, 1.066, 1.044, 1.022, 1, 1, 1, 1, 1, 1, 0.946, 0.87, 0.781, 0.685, 0.589, 0.5, 0.89, 0.89, 0.88, 0.88, 0.9, 0.93, 0.98, 1.03, 1.09, 1.15, 1.2, 0.82, 0.83, 0.85, 0.87, 0.9, 0.94, 0.99, 1.04, 1.09, 1.15, 1.2, -9.9, -9.425, -8.95, -8.475, -8, -7.895, -7.79, -7.8425, -7.895, -7.9475, -8, 55.1, 55.575, 56.05, 56.525, 57, 57.105, 57.21, 57.1575, 57.105, 57.0525, 57}

// CurveData holds curveTable in level data: written once in preprocess, it
// is read in place instead of being copied for every dynamic index.
type CurveData struct {
	sonolus.LevelDataResource
	Flat [504]float64
}

var Curves = CurveData{}

func initCurves() { Curves.Flat = curveTable }

// curve samples the n-sample table at off at fractional frame f (linear,
// clamped to the table).
func curve(off, n, f float64) float64 {
	f = sonolus.Clamp(f, 0, n-1)
	i := math.Min(math.Floor(f), n-2)
	a := Curves.Flat[int(off+i)]
	return a + (Curves.Flat[int(off+i)+1]-a)*(f-i)
}

// drawSpeakers draws both speakers: parent (0, 75) scale 1.3, node x −488
// (mirrored) / 484, an animated Root holding the body, the obj_speakerAlpha
// overlay (rest alpha 0.03) and the two glows, driven by the original clips
// speaker_normal / _exciting / _continuous.
func drawSpeakers(now, z float64) {
	if !Config.StageEffects {
		return
	}
	k := 1.3 * stageScale()
	for side := 0; side < 2; side++ {
		nodeX, m := 484.0, 1.0
		if side == 0 {
			nodeX, m = -488, -1
		}
		kind, t := Hud.SpeakerKind[side], now-Hud.SpeakerStart[side]
		if Hud.SpeakerStart[side] <= 0 || t < 0 {
			kind = 0
		}
		rootY, sx, sy := 0.0, 1.0, 1.0
		ovA, ovS := 0.03, 1.0
		topA, botA, topS, botS := 0.0, 0.0, 0.88437, 0.9
		topX, topY, botX, botY := -8.0, 57.0, -34.9, -54.1
		switch kind {
		case 1:
			f := t * 60
			sx = curve(cvSpkNRootS, 28, f)
			sy = sx
			ovA, ovS = curve(cvSpkNOvA, 28, f), curve(cvSpkNOvS, 28, f)
			topA, botA = curve(cvSpkNTopA, 28, f), curve(cvSpkNBotA, 28, f)
			topS = curve(cvSpkNGlowS, 28, f)
			botS = topS
			topX, topY, botX, botY = curve(cvSpkNTopX, 28, f), curve(cvSpkNTopY, 28, f), curve(cvSpkNBotX, 28, f), curve(cvSpkNBotY, 28, f)
		case 2:
			f := t * 60
			rootY, sx, sy = curve(cvSpkERootY, 17, f), curve(cvSpkERootSX, 17, f), curve(cvSpkERootSY, 17, f)
			ovA, ovS = curve(cvSpkEOvA, 17, f), curve(cvSpkEOvS, 17, f)
			topA = curve(cvSpkEGlowA, 17, f)
			botA = topA
			topS, botS = curve(cvSpkETopS, 17, f), curve(cvSpkEBotS, 17, f)
		case 3:
			f := math.Mod(t, 0.1666666716337204) * 60
			sx = curve(cvSpkCRootS, 11, f)
			sy = sx
			ovA, ovS = curve(cvSpkCOvA, 11, f), curve(cvSpkCOvS, 11, f)
			topA = curve(cvSpkCGlowA, 11, f)
			botA = topA
			topS, botS = curve(cvSpkCTopS, 11, f), curve(cvSpkCBotS, 11, f)
			topX, topY = curve(cvSpkCTopX, 11, f), curve(cvSpkCTopY, 11, f)
		}
		pos := func(ex, ey float64) sonolus.Vec2 {
			return stagePoint(1.3*(nodeX+m*sx*ex), 75+1.3*(rootY+sy*ey))
		}
		body := pos(0, 0)
		trim := [4]float64{0, 1, 0, 0}
		drawTrimmed(Skin.Speaker, body, 188*k*sx, 256*k*sy, 187, 256, trim, side == 0, z, 1)
		if ovA > 0 {
			drawTrimmed(Skin.Speaker, body, 189*k*sx*ovS, 256*k*sy*ovS, 187, 256, trim, side == 0, z+0.005, ovA)
		}
		glowTrim := [4]float64{4, 6, 8, 2}
		if topA > 0 {
			drawTrimmed(Skin.SpeakerGlow, pos(topX, topY), 118*k*sx*topS, 114*k*sy*topS, 118, 114, glowTrim, side == 0, z+0.01, topA)
		}
		if botA > 0 {
			drawTrimmed(Skin.SpeakerGlow, pos(botX, botY), 118*k*sx*botS, 114*k*sy*botS, 118, 114, glowTrim, side == 0, z+0.01, botA)
		}
	}
}

// ------------------------------------------------------------ SS effect ----

var kiraSize = [8]float64{20, 12, 8, 28, 12, 15, 22, 12}
var kiraY = [8]float64{2.4, -5, 2.4, 0, -5, 3.5, -5, 5.3}
var kiraX0 = [8]float64{416.0995, 378.3228, 340.5461, 295.4568, 252.6993, 205.9903, 125.1584, 59.9356}
var kiraWrap = [8]float64{0.05, 0.3167, 0.5833, 0.8833, 1.2, 1.4833, 2.1, 2.6667}
var kiraPulse = [8]bool{true, true, false, true, false, false, true, false}

// kiraRotation is a kira's z rotation (degrees) at loop phase lt (ScoreGaugeSS).
func kiraRotation(i int, lt float64) float64 {
	switch i {
	case 0:
		return 5 + 35*sonolus.Smoothstep(lt/3)
	case 1:
		return -20
	case 3:
		if lt < 0.8833 {
			return 10 + 50*lt/0.8833
		}
		return -50 + 60*sonolus.Clamp((lt-0.9)/2.1, 0, 1)
	case 4:
		return -5
	case 6:
		if lt < 2.1 {
			return 15 - 55*sonolus.Smoothstep(lt/2.1)
		}
		switch {
		case lt < 2.3:
			return -21.9 + 8.4*(lt-2.1667)/0.1333
		case lt < 2.5:
			return -13.5 + 12.1*(lt-2.3)/0.2
		case lt < 2.7:
			return -1.4 + 9.8*(lt-2.5)/0.2
		case lt < 2.9:
			return 8.4 + 5.8*(lt-2.7)/0.2
		}
		return 14.2 + 0.8*(lt-2.9)/0.1
	case 7:
		return -10
	}
	return 0
}

// drawSSEffect draws HighRankEffect (Flash + 8 kira) inside the >SS panel
// clip, from Hud.SSStart on.
func drawSSEffect(now, z float64) {
	if Hud.SSStart <= 0 || now < Hud.SSStart {
		return
	}
	t := now - Hud.SSStart
	ratio := sonolus.Clamp(Hud.Score/math.Max(Hud.ScoreMax, 1), 0, 1)
	indicator := math.Trunc(ratio * 422)
	clipR := 63 + math.Max(indicator, 4) // root-local x
	// Flash: 467×24 from x 67.01, alpha ping-pong 0 → 0.3 over 1 s
	ph := math.Mod(t, 2)
	if ph > 1 {
		ph = 2 - ph
	}
	fa := 0.30000001192092896 * ph * ph
	left := 25 + 41.05 + 0.96
	right := math.Min(left+467, clipR)
	if right > left {
		p0 := uiTopLeft(left, 46+12)
		p1 := uiTopLeft(right, 46-12)
		Skin.SSOverlay.Draw(sonolus.Rect{T: p1.Y, R: p1.X, B: p0.Y, L: p0.X}.ToQuad(), z, fa*Hud.HudAlpha)
	}
	lt := math.Mod(t, 3)
	for i := 0; i < 8; i++ {
		w := kiraWrap[i]
		var x float64
		if lt < w {
			x = kiraX0[i] + (423.65-kiraX0[i])*lt/w
		} else {
			if lt < w+0.05 {
				continue // hidden at the wrap
			}
			x = -7 + (kiraX0[i]+7)*(lt-w)/(3-w)
		}
		rx := 25 + 41.05 + x
		// SoftClip (softness 20): fade over the 20 units inside each x edge
		a := sonolus.Clamp(math.Min(rx-67, clipR-rx)/20, 0, 1)
		if a <= 0 {
			continue
		}
		s := 1.0
		if kiraPulse[i] && lt < 2.6667 {
			s = 0.9 + 0.1*math.Cos(2*math.Pi*lt/0.6666667)
		}
		c := uiTopLeft(rx, 45-kiraY[i])
		Skin.SSKira[i].Draw(rotQuad(c, kiraSize[i]*s, kiraSize[i]*s, kiraRotation(i, lt)), z+0.01, a*Hud.HudAlpha)
	}
}

// rotQuad is a w×h UI-unit quad centred on c, rotated deg counter-clockwise.
func rotQuad(c sonolus.Vec2, w, h, deg float64) sonolus.Quad {
	r := deg * math.Pi / 180
	cs, sn := math.Cos(r), math.Sin(r)
	hw, hh := w*Layout.U/2, h*Layout.U/2
	at := func(x, y float64) sonolus.Vec2 { return sonolus.Vec2{X: c.X + x*cs - y*sn, Y: c.Y + x*sn + y*cs} }
	return sonolus.Quad{BL: at(-hw, -hh), TL: at(-hw, hh), TR: at(hw, hh), BR: at(hw, -hh)}
}

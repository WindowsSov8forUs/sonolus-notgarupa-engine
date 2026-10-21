//go:build play || watch

package main

import (
	"math"
	"math/rand"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
)

// Stage dressing of the original Standard live: stage lights, speakers and
// the SS gauge effect.

// stageRoot is the original back UI root scale 1 − 0.23·HAR (integer
// activeHeight, no safe-area factor), in UI units.
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
		// a uniform tint of the backdrop (BgCover's sprite is a soft ellipse)
		Skin.Black.Draw(q, 1, 0.5*(1-c))
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
	// ColorFader.PlayColorTable cycles from the stage's creation (the level
	// starts with the stage, 2 s before the music), not from chart time
	t := now + 2
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

// Animation table: the original speaker clips (streamed cubic keys) sampled
// at 60 Hz, Root-local units, concatenated; cv* are their offsets.
const (
	cvSpkNTopA   = 0
	cvSpkNBotA   = 31
	cvSpkNGlowS  = 62
	cvSpkNTopX   = 93
	cvSpkNTopY   = 124
	cvSpkNBotX   = 155
	cvSpkNBotY   = 186
	cvSpkNRootS  = 217
	cvSpkNOvA    = 248
	cvSpkNOvS    = 279
	cvSpkERootY  = 310
	cvSpkERootSX = 327
	cvSpkERootSY = 344
	cvSpkEOvA    = 361
	cvSpkEOvS    = 378
	cvSpkEGlowA  = 395
	cvSpkETopS   = 412
	cvSpkEBotS   = 429
	cvSpkCRootS  = 446
	cvSpkCOvA    = 457
	cvSpkCOvS    = 468
	cvSpkCGlowA  = 479
	cvSpkCTopS   = 490
	cvSpkCBotS   = 501
	cvSpkCTopX   = 512
	cvSpkCTopY   = 523
)

var curveTable = [534]float64{0, 0, 0, 0.50002, 1, 1, 1, 0.99945, 0.81981, 0.56214, 0.27327, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.49996, 1, 1, 1, 0.99945, 0.8911, 0.62551, 0.29703, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0.9, 0.89375, 0.9, 0.94107, 1, 1.06793, 1.1465, 1.23003, 1.31283, 1.38921, 1.4535, 1.5, 1.53139, 1.55469, 1.57073, 1.58036, 1.5844, 1.58371, 1.5791, 1.57143, 1.56152, 1.55022, 1.53836, 1.52679, 1.51632, 1.50781, 1.50209, 1.5, 1.5, 1.5, 1.5, -11, -11.0625, -11, -10.53571, -10, -9.61808, -9.26093, -8.93294, -8.63848, -8.38192, -8.16764, -8, -7.87444, -7.78125, -7.71708, -7.67857, -7.66239, -7.66518, -7.68359, -7.71429, -7.75391, -7.79911, -7.84654, -7.89286, -7.93471, -7.96875, -7.99163, -8, -8, -8, -8, 55, 54.8125, 55, 56.51786, 58, 58.45189, 58.52332, 58.32362, 57.9621, 57.54811, 57.19096, 57, 56.93722, 56.89063, 56.85854, 56.83929, 56.83119, 56.83259, 56.8418, 56.85714, 56.87695, 56.89955, 56.92327, 56.94643, 56.96735, 56.98438, 56.99581, 57, 57, 57, 57, -34.9, -34.76875, -34.9, -35.91429, -37, -37.58396, -38.01968, -38.33644, -38.56356, -38.73032, -38.86603, -39, -39.12556, -39.21875, -39.28292, -39.32143, -39.33761, -39.33482, -39.31641, -39.28571, -39.24609, -39.20089, -39.15346, -39.10714, -39.06529, -39.03125, -39.00837, -39, -39, -39, -39, -54.1, -53.98125, -54.1, -54.925, -56, -57.04286, -58.17857, -59.34286, -60.47143, -61.5, -62.36429, -63, -63.43945, -63.76562, -63.99023, -64.125, -64.18164, -64.17187, -64.10742, -64, -63.86133, -63.70312, -63.53711, -63.375, -63.22852, -63.10937, -63.0293, -63, -63, -63, -63, 1, 0.99875, 1, 1.01036, 1.02, 1.02169, 1.02026, 1.01668, 1.0119, 1.00688, 1.00259, 1, 0.99874, 0.99781, 0.99717, 0.99679, 0.99662, 0.99665, 0.99684, 0.99714, 0.99754, 0.99799, 0.99847, 0.99893, 0.99935, 0.99969, 0.99992, 1, 1, 1, 1, 0.03, -0.01813, 0.03, 0.41804, 0.8, 0.9307, 0.9757, 0.95542, 0.89029, 0.80073, 0.70716, 0.63, 0.55629, 0.46365, 0.35839, 0.24683, 0.13526, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 0.03, 1, 0.99688, 1, 1.02393, 1.05, 1.06555, 1.07868, 1.08944, 1.09789, 1.10409, 1.10811, 1.11, 1.10971, 1.10732, 1.1031, 1.0973, 1.09018, 1.08202, 1.07306, 1.06357, 1.05382, 1.04405, 1.03455, 1.02556, 1.01735, 1.01018, 1.00431, 1, 0.99847, 0.99924, 1, 0, 25.9375, 48.99999, 65.11111, 77.38889, 88, 96, 93.528, 88.064, 80.136, 70.27201, 59.00001, 46.84801, 34.34401, 22.01601, 10.39201, 0.00001, 1, 1.10833, 1.2, 1.25185, 1.28148, 1.3, 1.3, 1.28215, 1.2592, 1.23205, 1.2016, 1.16875, 1.1344, 1.09945, 1.0648, 1.03135, 1, 1, 1.16458, 1.3, 1.36296, 1.38704, 1.4, 1.4, 1.3762, 1.3456, 1.3094, 1.2688, 1.225, 1.1792, 1.1326, 1.0864, 1.0418, 1, 0, 0.17657, 0.309, 0.36305, 0.39458, 0.40629, 0.40089, 0.38107, 0.34954, 0.309, 0.26215, 0.21171, 0.16036, 0.11081, 0.06576, 0.02793, 0, 1, 1.11429, 1.2, 1.23499, 1.25539, 1.26297, 1.25948, 1.24665, 1.22624, 1.2, 1.16968, 1.13703, 1.10379, 1.07172, 1.04257, 1.01808, 1, 1, 1, 1, 0.95937, 0.90962, 0.85186, 0.78717, 0.71665, 0.6414, 0.5625, 0.48105, 0.39814, 0.31487, 0.23233, 0.1516, 0.0738, 0, 0.88437, 0.87798, 0.88437, 0.91345, 0.94905, 0.99039, 1.03668, 1.08714, 1.141, 1.19746, 1.25575, 1.31508, 1.37467, 1.43374, 1.49151, 1.54719, 1.6, 0.9022, 0.89597, 0.9022, 0.93055, 0.96526, 1.00557, 1.05071, 1.09992, 1.15243, 1.20749, 1.26432, 1.32218, 1.38028, 1.43788, 1.49421, 1.5485, 1.6, 1, 1.02891, 1.06042, 1.08672, 1.1, 1.0978, 1.08519, 1.06563, 1.04259, 1.01956, 1, 0.03, 0.10805, 0.19313, 0.26414, 0.3, 0.29406, 0.26, 0.20719, 0.145, 0.08281, 0.03, 1, 1.02734, 1.05625, 1.08203, 1.1, 1.10868, 1.11111, 1.10938, 1.10556, 1.10174, 1.1, 1, 1, 1, 1, 1, 0.9456, 0.87037, 0.78125, 0.68519, 0.58912, 0.5, 0.89238, 0.88977, 0.88416, 0.88457, 0.9, 0.9333, 0.97862, 1.03196, 1.08931, 1.14666, 1.2, 0.8159, 0.83421, 0.85071, 0.87083, 0.9, 0.93994, 0.98712, 1.03913, 1.09356, 1.14799, 1.2, -9.9, -9.38047, -8.83125, -8.34141, -8, -7.83507, -7.78889, -7.82188, -7.89444, -7.96701, -8, 55.1, 55.61953, 56.16875, 56.65859, 57, 57.16493, 57.21111, 57.17813, 57.10556, 57.03299, 57}

// CurveData holds curveTable in level data: written once in preprocess, it
// is read in place instead of being copied for every dynamic index.
type CurveData struct {
	sonolus.LevelDataResource
	Flat [534]float64
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
			sx = curve(cvSpkNRootS, 31, f)
			sy = sx
			ovA, ovS = curve(cvSpkNOvA, 31, f), curve(cvSpkNOvS, 31, f)
			topA, botA = curve(cvSpkNTopA, 31, f), curve(cvSpkNBotA, 31, f)
			topS = curve(cvSpkNGlowS, 31, f)
			botS = topS
			topX, topY, botX, botY = curve(cvSpkNTopX, 31, f), curve(cvSpkNTopY, 31, f), curve(cvSpkNBotX, 31, f), curve(cvSpkNBotY, 31, f)
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

// -------------------------------------------------------------- psyllium ----

// Psyllium: the 73 PsylliumPos instances under Stage/TRSRoot/StageFront
// (stage units; scale x negative when mirrored). PsylliumManager.ExecStart
// shows them with the stage; Stage.ShowIntroAnimation fades them out 1.5 s
// after the entrance starts (FadeOut: RGB → 0 in 45 per-frame steps, the
// first at once). Each has a penlight (texture right half) and a core (left
// half), 64×128 widgets with a Bottom pivot, Additive Cheap, animated by the
// 4 s StartIdle loop.
var psyX = [73]float64{-635.0002, -346, -481.0001, 251, -226, -98, 644.0002, 13, 144, 403.0001, 523, -311.0001, -560, -172, -37, -670, 752.0002, 106, 242, 402, 498.0001, 609, -713.0001, -87, -521, 155, -245, 574, -419, 29, 437, 317, 714.0001, 828.0001, -830.0001, -474, -149, -629, -30, 137, 940.0001, -336, 300, 756.0001, 484, 640.0001, 814, 787.0001, 712, 555, 899.0001, 203, 35, 370, -246, -376, -529.0001, -939.0001, -721, -112, -916.0001, -769.0001, -482, -308, -591, -168, 144, 325, -2, 500, 740, 636.0001, 907.0001}
var psyY = [73]float64{-172, -181, -180, -174, -177, -179, -178, -174, -174, -173, -167, -267.0001, -219, -248, -236, -234, -238, -246, -240, -242, -236, -235, -313, -331, -292, -303, -348, -309, -294, -324, -304, -295, -295, -306, -410, -389, -394, -395, -402, -405, -406, -383, -404, -421, -403, -413, -408, -567.0001, -529.0001, -514, -517, -508, -502, -509, -495, -493, -494, -520.0001, -499, -521, -649.0001, -618, -675, -634, -602.0001, -632, -656, -628, -643, -663.0001, -664.0001, -624, -678.0001}
var psyS = [73]float64{-0.6, -0.6, -0.6, 0.6, -0.6, -0.6, 0.6, -0.6, 0.6, 0.6, 0.6, -0.8, -0.8, -0.8, -0.8, -0.8, 0.8, -0.8, 0.8, -0.8, -0.8, -0.8, 1, -1, 1, -1, -1, 1, -1, 1, 1, 1, 1, 1, -1.2, -1.2, -1.2, -1.2, 1.2, 1.2, 1.2, -1.2, 1.2, 1.2, -1.2, 1.2, -1.2, 1.4, 1.4, -1.4, 1.4, -1.4, 1.4, 1.4, -1.4, -1.4, -1.4, -1.4, -1.4, 1.4, -1.6, -1.6, -1.6, 1.6, -1.6, -1.6, 1.6, 1.6, -1.6, 1.6, 1.6, 1.6, 1.6}

// psylliumPose is StartIdle at τ (s, looped): both widgets at (x, 20),
// rotated by z degrees (the clip's two cubic segments).
func psylliumPose(tau float64) (float64, float64) {
	t := math.Mod(tau, 4)
	if t < 2 {
		return ((-7.5*t+22.5)*t)*t - 15, ((5*t-15)*t)*t + 10
	}
	d := t - 2
	return ((7.5*d-22.5)*d)*d + 15, ((-5*d+15)*d)*d - 10
}

func drawPsyllium(now, z float64) {
	if !Config.StageEffects {
		return
	}
	o := openingClock(now)
	k := 1.0
	if fade := o - 2.0; fade >= 0 { // the entrance starts at O = 0.5
		steps := math.Floor(fade*60) + 1
		if steps >= 45 {
			return
		}
		k = 1 - steps/45
	}
	px, deg := psylliumPose(o)
	a := deg * math.Pi / 180
	cs, sn := math.Cos(a), math.Sin(a)
	for i := 0; i < len(psyX); i++ {
		sx, sy := psyS[i], math.Abs(psyS[i])
		corner := func(x, y float64) sonolus.Vec2 {
			return stagePoint(psyX[i]+(px+cs*x-sn*y)*sx, psyY[i]+(20+sn*x+cs*y)*sy)
		}
		q := sonolus.Quad{BL: corner(-32, 0), TL: corner(-32, 128), TR: corner(32, 128), BR: corner(32, 0)}
		Skin.PsyPen.Draw(q, z, k)
		Skin.PsyCore.Draw(q, z+0.01, k)
	}
}

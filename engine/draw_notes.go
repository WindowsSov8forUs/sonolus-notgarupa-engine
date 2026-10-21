package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
)

// noteLook is what note drawing needs from an entity.
type noteLook struct {
	SpriteLane, Lane, Width, Direction, SpanStart float64
	// Stopped: held, stopped or snapped on the judge line, drawn at the
	// target-button scale instead of the motion scale.
	Stopped bool
}

// spriteScale is a note sprite's world scale: the motion scale at height y,
// or the target-button scale when stopped on the line.
func spriteScale(y, buttons float64, stopped bool) float64 {
	if stopped {
		return targetScale()
	}
	return noteScale(y, buttons)
}

// HABAHIRO wide-note sprites: pixel width by lane count (120 px high, PPU 100).
var habaPx = [7]float64{308, 524, 736, 952, 1164, 1380, 1596}
var habaTopPx = [3]float64{171, 288, 400}

// habaRun is the run index of the HABAHIRO sprite covering lanes
// s..s+w−1, s = clamp(floor(spanStart), 0, 7−w): (w−1)(16−w)/2 + s.
func habaRun(spanStart, width float64) int {
	w := sonolus.Clamp(width, 1, 7)
	s := sonolus.Clamp(math.Floor(spanStart), 0, 7-w)
	return int((w-1)*(16-w)/2 + s)
}

func habaWidthPx(width float64) float64 { return habaPx[int(sonolus.Clamp(width, 1, 7))-1] }

// drawWide draws a note body sprite of the given pixel width.
// cRide is the curve value of a sprite riding a held Slide's line: the head
// (slidingMove) and stopped middle nodes (MoveToNextAfterNoteButton) move in
// button-local space, so they sit at heldPos, with the flash and body front.
const cRide = -2.0

// spritePos is a note sprite's position for curve c (cRide: riding).
func spritePos(lane, c float64) sonolus.Vec2 {
	if c == cRide {
		return heldPos(lane)
	}
	return notePos(lane, c)
}

func drawWide(sp sonolus.Sprite, px, lane, c, buttons float64, z sonolus.ZIndex, a float64, stopped bool) sonolus.Vec2 {
	p := spritePos(lane, c)
	if !suddenHides(p.Y) {
		sp.DrawZ(spriteQuad(p, px, noteSpritePy, notePPU, spriteScale(p.Y, buttons, stopped)), z, a)
	}
	return p
}

// drawFlickTop draws the up-flick icon above a flick body at p.
func drawFlickTop(p sonolus.Vec2, width, iconT float64, z sonolus.ZIndex, stopped bool) {
	if suddenHides(p.Y) {
		return
	}
	s := spriteScale(p.Y, width, stopped)
	top := sonolus.Vec2{X: p.X, Y: p.Y + flickTopOffset(iconT)*s}
	if Layout.HabaSkin {
		i := int(sonolus.Clamp(width, 1, 3)) - 1
		Skin.HabaFlickTop[i].DrawZ(spriteQuad(top, habaTopPx[i], 138, notePPU, s), subZ(z, 1), 1)
		return
	}
	Skin.NoteFlickTop.DrawZ(spriteQuad(top, 171, 138, notePPU, s), subZ(z, 1), 1)
}

// slideHoldSprites are a held Slide head's body and flash: note_long_3
// normally; HABAHIRO uses a centred run of the Slide's width.
func slideHoldSprites(width, centre float64) (sonolus.Sprite, sonolus.Sprite, float64) {
	if !Layout.HabaSkin {
		return Skin.NoteLong[3], Skin.NoteLongFlash[3], noteSpritePx
	}
	w := sonolus.Clamp(width, 1, 7)
	start := math.Floor((7 - w) / 2)
	if math.Mod(w, 2) == 0 && centre > 3 {
		start++
	}
	run := habaRun(start, w)
	return Skin.HabaLong[run], Skin.HabaLongFlash[run], habaWidthPx(w)
}

// drawNoteSprite draws a note body sprite at lane position for curve c.
func drawNoteSprite(sp sonolus.Sprite, lane, c, buttons float64, z sonolus.ZIndex, a float64, stopped bool) sonolus.Vec2 {
	p := spritePos(lane, c)
	if !suddenHides(p.Y) {
		sp.DrawZ(noteQuad(p, spriteScale(p.Y, buttons, stopped)), z, a)
	}
	return p
}

// Sprite riding at the front of a held Slide (Play and Watch chainSprite).
const (
	chainNone   = -1.0
	chainHead   = 0.0 // the held head body
	chainPassed = 1.0 // the last passed middle node
)

// drawAmong draws an intermediate Slide node: note_slide_among, with the
// up-flick icon on a Flick gesture node whose gesture still shows (chart
// extension); a node stopped on the line takes the target-button scale.
func drawAmong(lane, c, width, kind, iconT float64, stopped bool, z sonolus.ZIndex) {
	p := spritePos(lane, c)
	if suddenHides(p.Y) {
		return
	}
	s := spriteScale(p.Y, width, stopped)
	if Layout.HabaSkin {
		w := sonolus.Clamp(width, 1, 7)
		Skin.HabaAmong[int(w)-1].DrawZ(spriteQuad(p, habaWidthPx(w), noteSpritePy, notePPU, s), z, 1)
	} else {
		Skin.NoteSlideAmong.DrawZ(noteQuad(p, s), z, 1)
	}
	if kind == NodeFlick {
		drawFlickTop(p, width, iconT, z, stopped)
	}
}

// bodyAlpha is the Long / Slide body opacity (live setting, default 80 %).
func bodyAlpha() float64 { return Config.LineBright / 100 }

func lanePick(arr [7]sonolus.Sprite, lane float64) sonolus.Sprite {
	return arr[int(sonolus.Clamp(lane, 0, 6))]
}

// flickTopOffset is the up-flick arrow's animated local y (0.7 → 1.3 per 1/3 s).
func flickTopOffset(t float64) float64 {
	return 0.699999988079071 + 1.7999998331069946*math.Mod(math.Max(t, 0), 0.3333333432674408)
}

// sideArrowOffset is the directional arrow's animated local x magnitude.
func sideArrowOffset(t float64) float64 {
	return 1.600000023841858 + 2.0999996662139893*math.Mod(math.Max(t, 0), 0.3333333432674408)
}

// drawTap draws a Single / Skill note.
func drawTap(l noteLook, skill, offbeat bool, c float64, z sonolus.ZIndex) {
	if Layout.HabaSkin {
		run := habaRun(l.SpanStart, l.Width)
		sp := Skin.HabaNormal[run]
		if skill {
			sp = Skin.HabaSkill[run]
		} else if offbeat && Config.NoteColor {
			sp = Skin.HabaNormal16[run]
		}
		drawWide(sp, habaWidthPx(l.Width), l.Lane, c, l.Width, z, 1, false)
		return
	}
	sp := lanePick(Skin.NoteNormal, l.SpriteLane)
	if skill {
		sp = lanePick(Skin.NoteSkill, l.SpriteLane)
	} else if offbeat && Config.NoteColor {
		sp = lanePick(Skin.NoteNormal16, l.SpriteLane)
	}
	drawNoteSprite(sp, l.Lane, c, l.Width, z, 1, false)
}

// drawFlick draws a Flick note with its animated up arrow.
func drawFlick(l noteLook, c, iconT float64, z sonolus.ZIndex) {
	var p sonolus.Vec2
	if Layout.HabaSkin {
		p = drawWide(Skin.HabaFlick[habaRun(l.SpanStart, l.Width)], habaWidthPx(l.Width), l.Lane, c, l.Width, z, 1, l.Stopped)
	} else {
		p = drawNoteSprite(lanePick(Skin.NoteFlick, l.SpriteLane), l.Lane, c, l.Width, z, 1, l.Stopped)
	}
	drawFlickTop(p, l.Width, iconT, z, l.Stopped)
}

// drawSideArrow draws a directional arrow beside a note body at p.
func drawSideArrow(p sonolus.Vec2, direction, iconT float64, z sonolus.ZIndex, stopped bool) {
	if suddenHides(p.Y) {
		return
	}
	s := spriteScale(p.Y, 1, stopped)
	off := sideArrowOffset(iconT)
	if direction < 0 {
		Skin.NoteFlickTopL.DrawZ(spriteQuad(sonolus.Vec2{X: p.X - off*s, Y: p.Y}, 138, 171, notePPU, s), z, 1)
	} else {
		Skin.NoteFlickTopR.DrawZ(spriteQuad(sonolus.Vec2{X: p.X + off*s, Y: p.Y}, 138, 170, notePPU, s), z, 1)
	}
}

// drawDirectionalBody draws one directional member (and its arrow).
// first is the leftmost lane of a linked group.
func drawDirectionalBody(l noteLook, c, iconT float64, arrow, grouped bool, first float64, z sonolus.ZIndex) sonolus.Vec2 {
	body := lanePick(Skin.NoteFlickR, l.SpriteLane)
	if l.Direction < 0 {
		body = lanePick(Skin.NoteFlickL, l.SpriteLane)
	}
	zb := z
	if grouped {
		zb = memberZ(z, l.Direction, l.Lane, first)
	}
	p := drawNoteSprite(body, l.Lane, c, 1, zb, 1, l.Stopped)
	if arrow {
		drawSideArrow(p, l.Direction, iconT, iconZ(z), l.Stopped)
	}
	return p
}

// drawBackLine joins two directional members (thickness 0.75 × the note's
// local scale; Sudden clips it).
func drawBackLine(p, q sonolus.Vec2, direction, t float64) {
	if Layout.SuddenR > 0 && (p.Y+q.Y)/2 > Layout.SuddenClipY {
		return
	}
	line := Skin.FlickLineR
	if direction < 0 {
		line = Skin.FlickLineL
	}
	line.DrawZ(lineQuad(p, q, noteLocalScale(p.Y, 1)*0.75/2), lineZ(zDirLine, t), 1)
}

// drawTerminal draws a Long / Slide end (or gesture node) sprite with its
// flick icon at lane for motion curve c. A multiple directional node draws
// one body per member lane and the arrow on the far member.
func drawTerminal(l noteLook, kind, kindWidth, lane, c, iconT float64, z sonolus.ZIndex) {
	shift := lane - l.Lane
	switch kind {
	case NodeFlick:
		drawFlick(noteLook{SpriteLane: l.SpriteLane + shift, Lane: lane, Width: l.Width, SpanStart: l.SpanStart + shift, Stopped: l.Stopped}, c, iconT, z)
	case NodeDirectional, NodeMultiDir:
		members := 1.0
		if kind == NodeMultiDir {
			members = math.Max(kindWidth, 1)
		}
		var p sonolus.Vec2
		for k := 0.0; k < members && k < 7; k++ {
			ml := lane + l.Direction*k
			body := lanePick(Skin.NoteFlickR, l.SpriteLane+ml-l.Lane)
			if l.Direction < 0 {
				body = lanePick(Skin.NoteFlickL, l.SpriteLane+ml-l.Lane)
			}
			// k grows toward the flick direction: the nearer member on top
			p = drawNoteSprite(body, ml, c, 1, subZ(z, k), 1, l.Stopped)
		}
		drawSideArrow(p, l.Direction, iconT, iconZ(z), l.Stopped)
	default:
		if Layout.HabaSkin {
			drawWide(Skin.HabaLong[habaRun(l.SpanStart+shift, l.Width)], habaWidthPx(l.Width), lane, c, l.Width, z, 1, l.Stopped)
			return
		}
		drawNoteSprite(lanePick(Skin.NoteLong, l.SpriteLane+shift), lane, c, l.Width, z, 1, l.Stopped)
	}
}

// drawBelt draws one Long / Slide body segment between two endpoints with
// their mesh half widths (bodyHalf / heldHalf / waitHalf); Sudden cuts it at
// the clip line.
func drawBelt(front sonolus.Vec2, fh float64, after sonolus.Vec2, ah float64, curve bool, z sonolus.ZIndex, alpha float64) {
	if Layout.SuddenR > 0 && after.Y > Layout.SuddenClipY {
		if front.Y >= Layout.SuddenClipY {
			return
		}
		k := (Layout.SuddenClipY - front.Y) / (after.Y - front.Y)
		after = sonolus.Vec2{X: front.X + (after.X-front.X)*k, Y: Layout.SuddenClipY}
		ah = fh + (ah-fh)*k
	}
	sp := Skin.LongLine
	if Layout.HabaSkin {
		sp = Skin.HabaLongLine
		if curve {
			sp = Skin.HabaLongLine2
		}
	} else if curve {
		sp = Skin.LongLine2
	}
	sp.DrawZ(beltQuad(front, after, fh, ah), z, alpha)
}

// drawHoldFlash draws the held Long / Slide root: optional body and the
// additive LongNoteFlash.
// p is the held point on the line and px the sprites' pixel width; both use
// the target-button scale, and the flash is not covered by Sudden.
func drawHoldFlash(body, flash sonolus.Sprite, p sonolus.Vec2, px float64, withBody bool, t float64, z sonolus.ZIndex) {
	s := targetScale()
	q := spriteQuad(p, px, noteSpritePy, notePPU, s)
	if withBody && !suddenHides(p.Y) {
		body.DrawZ(q, z, 1)
	}
	flash.DrawZ(q, iconZ(z), flashValue(t))
}

// flashValue is the LongNoteFlash clip: 0.2 → 0.6 → 0.2 over 0.8333 s.
func flashValue(t float64) float64 {
	ph := math.Mod(math.Max(t, 0), 0.8333333134651184)
	if ph < 0.4166666567325592 {
		return ((-5.529601097*ph+2.303999901)*ph+0.960000098)*ph + 0.2
	}
	d := ph - 0.4166666567325592
	return ((5.529601097*d-4.6080002785)*d)*d + 0.6
}

// syncEdgeMargin is the master noteSyncEdgeMargin of note skin TYPE7, the
// only non-zero value (its items carry Skin.SyncEdgeMarginOpt).
const syncEdgeMargin = 1.100000023841858

// drawSimLine joins two simultaneous notes (width 0.28 × note local scale).
// NoteSyncLine pulls each end in by the note skin's noteSyncEdgeMargin ×
// the note's world scale, except at directional-note ends (aNo / bNo).
func drawSimLine(p, q sonolus.Vec2, t float64, aNo, bNo bool) {
	if Layout.SuddenR > 0 && (p.Y+q.Y)/2 > Layout.SuddenClipY {
		return
	}
	if Skin.SyncEdgeMarginOpt.Exists() {
		dir := 1.0
		if p.X > q.X {
			dir = -1
		}
		if !aNo {
			p.X += syncEdgeMargin * noteScale(p.Y, 1) * dir
		}
		if !bNo {
			q.X -= syncEdgeMargin * noteScale(q.Y, 1) * dir
		}
	}
	half := noteLocalScale(p.Y, 1) * 0.28 / 2
	if Layout.HabaSkin {
		Skin.HabaSimLine.DrawZ(lineQuad(p, q, half), lineZ(zSyncLine, t), 1)
		return
	}
	Skin.SimLine.DrawZ(lineQuad(p, q, half), lineZ(zSyncLine, t), 1)
}

// drawLong draws a Long's head (held: with flash) and its straight body to
// the tail at curve tc (0 = pinned at the spawn row).
func drawLong(l noteLook, c, tailLane, tailWidth, tc float64, held bool, flashT float64, z, zb sonolus.ZIndex) {
	head := notePos(l.Lane, c)
	tp := notePos(tailLane, tc)
	// the body exists only while the tail is above the line (an unheld head
	// that moved on past the line leaves no body once the tail arrives)
	if tp.Y > head.Y && tc < 1 {
		fh, ah := bodyHalf(head.Y, l.Width), bodyHalf(tp.Y, tailWidth)
		if held {
			fh = heldHalf(l.Width)
		}
		if tc <= 0 {
			ah = waitHalf(tailWidth) // the tail has not appeared yet
		}
		drawBelt(head, fh, tp, ah, false, zb, bodyAlpha())
	}
	if Layout.HabaSkin {
		// a Long's flash uses its real lanes
		run := habaRun(l.SpanStart, l.Width)
		if held {
			drawHoldFlash(Skin.HabaLong[run], Skin.HabaLongFlash[run], head, habaWidthPx(l.Width), true, flashT, z)
			return
		}
		drawWide(Skin.HabaLong[run], habaWidthPx(l.Width), l.Lane, c, l.Width, z, 1, l.Stopped)
		return
	}
	if held {
		drawHoldFlash(lanePick(Skin.NoteLong, l.SpriteLane), lanePick(Skin.NoteLongFlash, l.SpriteLane), head, noteSpritePx, true, flashT, z)
		return
	}
	drawNoteSprite(lanePick(Skin.NoteLong, l.SpriteLane), l.Lane, c, l.Width, z, 1, l.Stopped)
}

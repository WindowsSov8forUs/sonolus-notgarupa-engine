//go:build play || watch

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
)

// judgement is one original OneFrameData entry.
type judgement struct {
	Grade     float64
	Type      float64 // judge note type
	Timing    float64
	Skill     bool
	X         float64 // receiver x on the judge line
	Lane      float64 // receiver lane (for lane effects / speakers), real
	Count     float64 // multiple-directional count
	Direction float64 // −1 left, +1 right (directional particles)
	Key       float64 // chart time of the note (same-position SE silence)
	Divisor   float64 // life damage divisor (2 for Long head timeout, 5 for slide among)
	Auto      bool
	Mid       bool // a Slide head / intermediate node: no end lane light
	FingerX   float64
	HasFinger bool    // a manual directional group: finger particle on the touched member
	Width     float64 // note width (HABAHIRO width-variant particles)
	Slide     bool    // a Slide judgement: never skill particles
	// Frame: now is the judging update's own time (Play), so one OneFrame
	// batch is exactly the judgements at that time; Watch preprocess groups
	// replay times by 1/60 s instead
	Frame bool
}

// hudApply updates combo, score, life and the judge text of a judgement and
// returns the life change it caused (original OneFrameData reflection).
func hudApply(now float64, j judgement) float64 {
	g := j.Grade
	Hud.Judged++
	if g >= gradeGreat {
		Hud.Combo++
		Hud.MaxCombo = math.Max(Hud.MaxCombo, Hud.Combo)
		Hud.ComboStart = now
	} else {
		Hud.Combo = 0
	}
	if g == gradePerfect {
		Hud.PerfectCount++
	} else {
		Hud.AllPerfect = false
		if g == gradeGreat {
			Hud.GreatCount++
		}
	}
	// arcade score (the result screen's): the combo is the streak of
	// consecutive Great-or-better judgements
	Hud.ScoreRaw += arcadeWeight(g, Hud.Combo)
	score := math.Floor(Hud.ScoreRaw/math.Max(Hud.ScoreTotal, 1e-9)*1000000 + 1e-6)
	added := score - Hud.Score
	Hud.Score = score
	Hud.LastAdded = added
	if added > 0 {
		addPop(added, now) // every judgement with score starts its own pop
	}
	// Life 0 does not end the play (Sonolus cannot stop the live; the level
	// is failed, as Sonolus' own life rules it). As in the original multi
	// live, life stays at 0 and everything else goes on, the clear included.
	dmg := 0.0
	if !Hud.GameOver {
		dmg = lifeDamage(g, j.Type, Config.FullLength)
		if j.Divisor > 1 {
			dmg = math.Trunc(dmg / j.Divisor)
		}
		if dmg != 0 {
			Hud.Life = sonolus.Clamp(Hud.Life+dmg, 0, 2000)
			if Hud.Life <= 0 {
				Hud.GameOver = true
			}
		}
	}
	// one representative result per frame: the highest raw result, the first
	// on ties (scoreLifeStateManager)
	sameFrame := math.Floor(now*60) == math.Floor(Hud.JudgeStart*60)
	if j.Frame {
		sameFrame = now == Hud.JudgeStart
	}
	if sameFrame && Hud.JudgeSprite >= 0 && g <= Hud.JudgeGrade && !j.Auto {
		return dmg
	}
	Hud.JudgeGrade = g
	switch {
	case j.Auto:
		Hud.JudgeSprite = 5
	case g == gradePerfect:
		Hud.JudgeSprite = 0
	case g == gradeGreat:
		Hud.JudgeSprite = 1
	case g == gradeGood:
		Hud.JudgeSprite = 2
	case g == gradeBad:
		Hud.JudgeSprite = 3
	default:
		Hud.JudgeSprite = 4
	}
	Hud.JudgeTiming = j.Timing
	if g == gradePerfect || g == gradeMiss || j.Auto {
		Hud.JudgeTiming = timingNone
	}
	Hud.JudgeStart = now
	return dmg
}

// judgementLights lights the receiver lane of end judgements and drives the
// stage speakers.
func judgementLights(now float64, j judgement) {
	if j.Auto || !j.Mid && (j.Type == jtLongEnd || j.Type == jtLongEndFlick || j.Type == jtLongEndDir || j.Type == jtLongEndMulti) {
		laneOnReserve(j.Lane, now)
	}
	triggerSpeaker(j, now)
}

// SE clip codes chosen by judgementClip.
const (
	clipNone    = 0.0
	clipPerfect = 1.0
	clipGreat   = 2.0
	clipGood    = 3.0
	clipFlick   = 4.0
	clipDir1    = 5.0
	clipDir2    = 6.0
	clipDir3    = 7.0
)

// kindClip is the SE code of a Perfect on a node of the given kind.
func kindClip(kind, width float64) float64 {
	switch kind {
	case NodeFlick:
		return clipFlick
	case NodeDirectional:
		return clipDir1
	case NodeMultiDir:
		if width >= 3 {
			return clipDir3
		} else if width == 2 {
			return clipDir2
		}
		return clipDir1
	}
	return clipPerfect
}

func clipOf(code float64) sonolus.Clip {
	switch code {
	case clipPerfect:
		return Effects.Perfect
	case clipGreat:
		return Effects.Great
	case clipGood:
		return Effects.Good
	case clipFlick:
		return Effects.Flick
	case clipDir1:
		return Effects.Directional[0]
	case clipDir2:
		return Effects.Directional[1]
	}
	return Effects.Directional[2]
}

func isFlickType(t float64) bool {
	return t == jtFlick || t == jtLongEndFlick || t == jtLongEndDir || t == jtLongEndMulti || t == jtDirectional || t == jtMultiDir
}

// judgementClip follows InGameManager.onJudgeNote with the TapSEStatus
// same-position silence (3-frame window) and returns the SE clip code.
// skillSE: judgeFrontNote / NoteLong.judgeAfterNote request SE_RHYTHM_TAP_SKILL
// themselves for a skill note whose adjusted result is Great or Perfect. It
// is its own request beside the normal cue: never silenced by, and never
// recorded for, the same-position silence (the note's normal cue still is).
func skillSE(j judgement) bool {
	return j.Skill && (j.Grade == gradePerfect || j.Grade == gradeGreat)
}

func judgementClip(now float64, j judgement) float64 {
	g := j.Grade
	silent := false
	if g != gradeMiss && now-Hud.SilenceTime < 3*frameSeconds && Hud.SilencePos == j.Key {
		p := Hud.SilenceType
		switch {
		case p == jtFlick || p == jtLongEndFlick:
			silent = j.Type == jtFlick || j.Type == jtLongEndFlick
		case p == jtLongEndDir || p == jtDirectional:
			silent = j.Type == jtLongEndDir || j.Type == jtDirectional
		case p == jtLongEndMulti || p == jtMultiDir:
			silent = (j.Type == jtLongEndMulti || j.Type == jtMultiDir) && Hud.SilenceCount == j.Count
		default:
			silent = Hud.SilenceGrade == g && !isFlickType(j.Type)
		}
	}
	Hud.SilencePos, Hud.SilenceType, Hud.SilenceGrade, Hud.SilenceCount, Hud.SilenceTime = j.Key, j.Type, g, j.Count, now
	if silent || g < gradeGood {
		return clipNone
	}
	switch {
	case j.Type == jtFlick || j.Type == jtLongEndFlick:
		return clipFlick
	case j.Type == jtLongEndDir || j.Type == jtDirectional:
		return clipDir1
	case j.Type == jtLongEndMulti || j.Type == jtMultiDir:
		if j.Count >= 3 {
			return clipDir3
		} else if j.Count == 2 {
			return clipDir2
		}
		return clipDir1
	case g == gradePerfect:
		return clipPerfect
	case g == gradeGreat:
		return clipGreat
	}
	return clipGood
}

// fxOwnerFamilies is the number of particle owner families per half-lane.
const fxOwnerFamilies = 16

// fxKey is a particle owner (GamePlayButton.PlayAnimation: button and effect
// root): receiver half-lane and prefab family (0 perfect, 1 great, 2 good,
// 3..5 skill, 6 swipe, 7..9 / 10..12 directional L / R for counts 1, 2, 3+,
// 13 / 14 directional finger L / R, 15 empty tap).
func fxKey(lane float64, family int) int {
	return int(sonolus.Clamp(math.Round(lane*2), 0, 13))*fxOwnerFamilies + family
}

// fxRestart stops an owner's previous particles: a new hit restarts the
// prefab (Stop + Clear + Play, restartIfActive).
func fxRestart(k int) {
	if Hud.Fx0[k].ID > 0 {
		Hud.Fx0[k].Destroy()
	}
	if Hud.Fx1[k].ID > 0 {
		Hud.Fx1[k].Destroy()
	}
	if Hud.Fx2[k].ID > 0 {
		Hud.Fx2[k].Destroy()
	}
	Hud.Fx0[k].ID, Hud.Fx1[k].ID, Hud.Fx2[k].ID = 0, 0, 0
}

// habaW: a note of width w plays the HABAHIRO width variant of its effect.
func habaW(w int) bool { return Layout.HabaFx && w >= 2 }

// playFx spawns the one-shot family f into feedback slot k.
func playFx(k, f int, x, dur float64) {
	Hud.Fx0[k], Hud.Fx1[k], Hud.Fx2[k] = spawnFx(f, x, dur, dur, false)
}

// playJudgementParticles follows GamePlayButton.PlayAnimation: Bad and Miss
// play nothing; skill notes first, then directional, flick, normal routes.
// A HABAHIRO note of width w ≥ 2 plays the width variants (view part and
// their fixed-tilt plane parts) of the HABAHIRO tap effect.
func playJudgementParticles(j judgement) {
	g := j.Grade
	if g < gradeGood {
		return
	}
	x := j.X
	w := int(sonolus.Clamp(j.Width, 1, 7))
	q1 := particleQuad(x, 1)
	switch {
	case j.Skill && !j.Slide:
		switch g {
		case gradePerfect:
			k := fxKey(j.Lane, 3)
			fxRestart(k)
			if habaW(w) {
				Hud.Fx0[k] = Particles.SkillPerfectW[w-2].Spawn(q1, durPerfect, false)
				Hud.Fx1[k] = Particles.SkillPerfectPlaneW[w-2].Spawn(particleQuad(x, planeCos46), durPerfect, false)
				Hud.Fx2[k] = Particles.SkillPerfectPlane2W[w-2].Spawn(particleQuad(x, planeCos50), durPerfect, false)
			} else {
				playFx(k, fxSkillPerfect, x, durPerfect)
			}
		case gradeGreat:
			k := fxKey(j.Lane, 4)
			fxRestart(k)
			if habaW(w) {
				Hud.Fx0[k] = Particles.SkillGreatW[w-2].Spawn(q1, durGreat, false)
				Hud.Fx1[k] = Particles.SkillGreatPlaneW[w-2].Spawn(particleQuad(x, planeCos46), durGreat, false)
			} else {
				playFx(k, fxSkillGreat, x, durGreat)
			}
		default:
			k := fxKey(j.Lane, 5)
			fxRestart(k)
			if habaW(w) {
				Hud.Fx0[k] = Particles.SkillGoodW[w-2].Spawn(q1, durGood, false)
			} else {
				playFx(k, fxSkillGood, x, durGood)
			}
		}
	case j.Type == jtLongEndDir || j.Type == jtLongEndMulti || j.Type == jtDirectional || j.Type == jtMultiDir:
		idx := 0
		if j.Count >= 3 {
			idx = 2
		} else if j.Count == 2 {
			idx = 1
		}
		dur := durDir23
		if idx == 0 {
			dur = durDir1
		}
		fx := x
		if j.HasFinger {
			fx = j.FingerX
		}
		// the finger family has at most one plane part
		// the body and the finger are separate owners: the body's root
		// carries the count variant, the finger sits on its own button
		fingerLane := j.Lane
		if j.HasFinger {
			fingerLane = j.FingerX/Layout.W + 3
		}
		if j.Direction < 0 {
			k := fxKey(j.Lane, 7+idx)
			fxRestart(k)
			Hud.Fx0[k] = Particles.DirL[idx].Spawn(q1, dur, false)
			kf := fxKey(fingerLane, 13)
			fxRestart(kf)
			Hud.Fx0[kf], Hud.Fx1[kf], _ = spawnFx(fxFingerL, fx, durFinger, durFinger, false)
		} else {
			k := fxKey(j.Lane, 10+idx)
			fxRestart(k)
			Hud.Fx0[k] = Particles.DirR[idx].Spawn(q1, dur, false)
			kf := fxKey(fingerLane, 14)
			fxRestart(kf)
			Hud.Fx0[kf], Hud.Fx1[kf], _ = spawnFx(fxFingerR, fx, durFinger, durFinger, false)
		}
	case j.Type == jtFlick || j.Type == jtLongEndFlick:
		k := fxKey(j.Lane, 6)
		fxRestart(k)
		if habaW(w) {
			Hud.Fx0[k] = Particles.SwipeW[w-2].Spawn(q1, durSwipe, false)
			Hud.Fx1[k] = Particles.SwipePlaneW[w-2].Spawn(particleQuad(x, planeCos46), durSwipe, false)
		} else {
			playFx(k, fxSwipe, x, durSwipe)
		}
	default:
		switch g {
		case gradePerfect:
			k := fxKey(j.Lane, 0)
			fxRestart(k)
			if habaW(w) {
				Hud.Fx0[k] = Particles.PerfectW[w-2].Spawn(q1, durPerfect, false)
				Hud.Fx1[k] = Particles.PerfectPlaneW[w-2].Spawn(particleQuad(x, planeCos46), durPerfect, false)
			} else {
				playFx(k, fxPerfect, x, durPerfect)
			}
		case gradeGreat:
			k := fxKey(j.Lane, 1)
			fxRestart(k)
			if habaW(w) {
				Hud.Fx0[k] = Particles.GreatW[w-2].Spawn(q1, durGreat, false)
				Hud.Fx1[k] = Particles.GreatPlaneW[w-2].Spawn(particleQuad(x, planeCos50), durGreat, false)
			} else {
				playFx(k, fxGreat, x, durGreat)
			}
		default:
			k := fxKey(j.Lane, 2)
			fxRestart(k)
			if habaW(w) {
				Hud.Fx0[k] = Particles.GoodW[w-2].Spawn(q1, durGood, false)
			} else {
				playFx(k, fxGood, x, durGood)
			}
		}
	}
}

// laneIndex returns the integer lane 0..6 of a receiver lane, or −1.
func laneIndex(lane float64) int {
	l := math.Round(lane)
	if l != lane || l < 0 || l > 6 {
		return -1
	}
	return int(l)
}

// laneOn lights a lane (NoteLaneEffectOn).
func laneOn(lane float64) {
	i := laneIndex(lane)
	if i < 0 {
		return
	}
	Hud.LaneOn[i] = 1
}

// laneOnReserve lights a lane for one update, then fades (On + OffReserve);
// events apply before the countdown of the same update (tapLaneEffectOwner).
func laneOnReserve(lane, now float64) {
	i := laneIndex(lane)
	if i < 0 {
		return
	}
	Hud.LaneOn[i] = 1
	Hud.LaneReserve[i] = 1
}

// laneAnimatedOff starts the 1/6 s fade (restarts it while fading); it does
// nothing on a switched-off lane.
func laneAnimatedOff(lane, now float64) {
	i := laneIndex(lane)
	if i < 0 || Hud.LaneOn[i] == 0 {
		return
	}
	Hud.LaneOn[i] = 2
	Hud.LaneStart[i] = now
}

// laneOff switches a lane off immediately.
func laneOff(lane float64) {
	i := laneIndex(lane)
	if i < 0 {
		return
	}
	Hud.LaneOn[i] = 0
}

// updateLaneReserve counts the OffReserve counters down once per update.
func updateLaneReserve(now float64) {
	for i := 0; i < 7; i++ {
		if Hud.LaneReserve[i] > 0 {
			Hud.LaneReserve[i]--
			if Hud.LaneReserve[i] == 0 && Hud.LaneOn[i] != 0 {
				// the fade started with the previous update's countdown
				Hud.LaneOn[i] = 2
				Hud.LaneStart[i] = now - frameSeconds
			}
		}
	}
}

// triggerSpeaker drives the stage speakers from outer-lane judgements.
func triggerSpeaker(j judgement, now float64) {
	// any judgement whose lane span contains lane 0 / 6 (both for a note
	// covering both)
	half := (math.Max(j.Width, 1) - 1) / 2
	if j.Lane-half <= 0 {
		speakerClip(0, j, now)
	}
	if j.Lane+half >= 6 {
		speakerClip(1, j, now)
	}
}

func speakerClip(side int, j judgement, now float64) {
	switch {
	case j.Type == jtLongMiss:
		Hud.SpeakerKind[side] = 0
	case j.Grade <= gradeMiss:
		return
	case j.Type == jtLongFront:
		Hud.SpeakerKind[side] = 3
	case isFlickType(j.Type):
		Hud.SpeakerKind[side] = 2
	default:
		Hud.SpeakerKind[side] = 1
	}
	Hud.SpeakerStart[side] = now
}

// holdLoopSchedule applies the cue limit to scheduled hold sounds, fed in
// start order (the scoring chain). A sound that ends normally stops and
// plays its release; a sound the limit steals stops at the new start,
// silently.
type holdLoopSchedule struct {
	Loops [holdSlots]float64 // main handle IDs
	Intro [holdSlots]float64 // intro handle IDs
	End   [holdSlots]float64
	Order [holdSlots]float64
	Live  [holdSlots]bool
	Seq   float64
}

func (s *holdLoopSchedule) retire(i int) {
	releaseHoldSound(s.Loops[i], s.Intro[i], s.End[i], false)
	s.Live[i] = false
}

func (s *holdLoopSchedule) start(t, end float64) {
	limit := int(HoldSE.Limit)
	slot := -1
	for i := 0; i < limit; i++ {
		if s.Live[i] && s.End[i] <= t {
			s.retire(i)
		}
		if !s.Live[i] && slot < 0 {
			slot = i
		}
	}
	if slot < 0 {
		slot = 0
		for i := 1; i < limit; i++ {
			if s.Order[i] < s.Order[slot] {
				slot = i
			}
		}
		stopHoldSound(s.Loops[slot], s.Intro[slot], t, false)
	}
	s.Loops[slot], s.Intro[slot] = startHoldSound(t, false)
	s.Seq++
	s.End[slot], s.Order[slot], s.Live[slot] = end, s.Seq, true
}

func (s *holdLoopSchedule) finish() {
	for i := 0; i < int(HoldSE.Limit); i++ {
		if s.Live[i] {
			s.retire(i)
		}
	}
}

// syncLineAtSpeed: a sync line the converter paired only for some note speeds
// (min..max in option tenths) exists at the current one.
func syncLineAtSpeed(min, max float64) bool {
	speed := math.Round(Config.NoteSpeed * 10)
	return speed >= min && speed <= max
}

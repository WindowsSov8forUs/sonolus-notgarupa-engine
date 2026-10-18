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
	if added > 0 && !Hud.GameOver {
		addPop(added, now) // every judgement with score starts its own pop
	}
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
				Hud.GameOverStart = now
				for i := 0.0; i < 7; i++ {
					laneOff(i) // game over switches every lane effect off
				}
			}
		}
	}
	// one representative result per frame: the highest raw result, the first
	// on ties (scoreLifeStateManager)
	if math.Floor(now*60) == math.Floor(Hud.JudgeStart*60) && Hud.JudgeSprite >= 0 && g <= Hud.JudgeGrade && !j.Auto {
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
	if Hud.GameOver {
		return
	}
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
	clipSkill   = 8.0
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
	case clipSkill:
		return Effects.Skill
	}
	return Effects.Directional[2]
}

func isFlickType(t float64) bool {
	return t == jtFlick || t == jtLongEndFlick || t == jtLongEndDir || t == jtLongEndMulti || t == jtDirectional || t == jtMultiDir
}

// judgementClip follows InGameManager.onJudgeNote with the TapSEStatus
// same-position silence (3-frame window) and returns the SE clip code.
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
	case j.Skill:
		return clipSkill // SE_RHYTHM_TAP_SKILL replaces the cue (Good and above)
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

// fxKey is a judgement particle owner: receiver half-lane and prefab family
// (0 perfect, 1 great, 2 good, 3..5 skill, 6 swipe, 7 / 8 directional L / R).
func fxKey(lane float64, family int) int {
	return int(sonolus.Clamp(math.Round(lane*2), 0, 13))*9 + family
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

// pickW selects a HABAHIRO width variant for a note of width w.
func pickW(base sonolus.Effect, variants [6]sonolus.Effect, w int) sonolus.Effect {
	if Layout.HabaFx && w >= 2 {
		return variants[w-2]
	}
	return base
}

// playJudgementParticles follows GamePlayButton.PlayAnimation: Bad and Miss
// play nothing; skill notes first, then directional, flick, normal routes.
func playJudgementParticles(j judgement) {
	g := j.Grade
	if g < gradeGood || Hud.GameOver {
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
			Hud.Fx0[k] = pickW(Particles.SkillPerfect, Particles.SkillPerfectW, w).Spawn(q1, durPerfect, false)
			Hud.Fx1[k] = pickW(Particles.SkillPerfectPlane, Particles.SkillPerfectPlaneW, w).Spawn(particleQuad(x, planeCos46), durPerfect, false)
			Hud.Fx2[k] = pickW(Particles.SkillPerfectPlane2, Particles.SkillPerfectPlane2W, w).Spawn(particleQuad(x, planeCos50), durPerfect, false)
		case gradeGreat:
			k := fxKey(j.Lane, 4)
			fxRestart(k)
			Hud.Fx0[k] = pickW(Particles.SkillGreat, Particles.SkillGreatW, w).Spawn(q1, durGreat, false)
			Hud.Fx1[k] = pickW(Particles.SkillGreatPlane, Particles.SkillGreatPlaneW, w).Spawn(particleQuad(x, planeCos46), durGreat, false)
		default:
			k := fxKey(j.Lane, 5)
			fxRestart(k)
			Hud.Fx0[k] = pickW(Particles.SkillGood, Particles.SkillGoodW, w).Spawn(q1, durGood, false)
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
		if j.Direction < 0 {
			k := fxKey(j.Lane, 7)
			fxRestart(k)
			Hud.Fx0[k] = Particles.DirL[idx].Spawn(q1, dur, false)
			Hud.Fx1[k] = Particles.FingerL.Spawn(particleQuad(fx, 1), durFinger, false)
			Hud.Fx2[k] = Particles.FingerLPlane.Spawn(particleQuad(fx, planeCos46), durFinger, false)
		} else {
			k := fxKey(j.Lane, 8)
			fxRestart(k)
			Hud.Fx0[k] = Particles.DirR[idx].Spawn(q1, dur, false)
			Hud.Fx1[k] = Particles.FingerR.Spawn(particleQuad(fx, 1), durFinger, false)
			Hud.Fx2[k] = Particles.FingerRPlane.Spawn(particleQuad(fx, planeCos46), durFinger, false)
		}
	case j.Type == jtFlick || j.Type == jtLongEndFlick:
		k := fxKey(j.Lane, 6)
		fxRestart(k)
		Hud.Fx0[k] = pickW(Particles.Swipe, Particles.SwipeW, w).Spawn(q1, durSwipe, false)
		Hud.Fx1[k] = pickW(Particles.SwipePlane, Particles.SwipePlaneW, w).Spawn(particleQuad(x, planeCos46), durSwipe, false)
	default:
		switch g {
		case gradePerfect:
			k := fxKey(j.Lane, 0)
			fxRestart(k)
			Hud.Fx0[k] = pickW(Particles.Perfect, Particles.PerfectW, w).Spawn(q1, durPerfect, false)
			Hud.Fx1[k] = pickW(Particles.PerfectPlane, Particles.PerfectPlaneW, w).Spawn(particleQuad(x, planeCos46), durPerfect, false)
		case gradeGreat:
			k := fxKey(j.Lane, 1)
			fxRestart(k)
			Hud.Fx0[k] = pickW(Particles.Great, Particles.GreatW, w).Spawn(q1, durGreat, false)
			Hud.Fx1[k] = pickW(Particles.GreatPlane, Particles.GreatPlaneW, w).Spawn(particleQuad(x, planeCos50), durGreat, false)
		default:
			k := fxKey(j.Lane, 2)
			fxRestart(k)
			Hud.Fx0[k] = pickW(Particles.Good, Particles.GoodW, w).Spawn(q1, durGood, false)
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

// holdLoopLimit is the SE_RHYTHM_TAP_LONG cue limit: the original keeps at
// most 4 playbacks of the cue, a new start stopping the oldest
// (GarupaEditor originalCueLimitProfile, newest-replaces-oldest).
const holdLoopLimit = 4

// holdLoopSchedule applies the cue limit to scheduled hold loops, fed in
// start order (the scoring chain). A loop that ends normally stops and plays
// its release; a loop the limit steals stops at the new start, silently.
type holdLoopSchedule struct {
	Loops [holdLoopLimit]sonolus.ScheduledLoopedEffectHandle
	End   [holdLoopLimit]float64
	Order [holdLoopLimit]float64
	Live  [holdLoopLimit]bool
	Seq   float64
}

func (s *holdLoopSchedule) retire(i int) {
	s.Loops[i].Stop(s.End[i])
	Effects.HoldRelease.PlayScheduled(s.End[i], 0)
	s.Live[i] = false
}

func (s *holdLoopSchedule) start(t, end float64) {
	slot := -1
	for i := 0; i < holdLoopLimit; i++ {
		if s.Live[i] && s.End[i] <= t {
			s.retire(i)
		}
		if !s.Live[i] && slot < 0 {
			slot = i
		}
	}
	if slot < 0 {
		slot = 0
		for i := 1; i < holdLoopLimit; i++ {
			if s.Order[i] < s.Order[slot] {
				slot = i
			}
		}
		s.Loops[slot].Stop(t)
	}
	s.Loops[slot] = Effects.HoldLoop.PlayLoopedScheduled(t)
	s.Seq++
	s.End[slot], s.Order[slot], s.Live[slot] = end, s.Seq, true
}

func (s *holdLoopSchedule) finish() {
	for i := 0; i < holdLoopLimit; i++ {
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

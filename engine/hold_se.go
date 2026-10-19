package main

import "github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"

// Hold sound (SE_RHYTHM_TAP_LONG). The original cue plays from its start: a
// looping cue repeats [LoopStart, LoopEnd) after its intro [0, LoopStart), a
// non-looping cue plays once; a release fades it out over 0.3 s. The effect
// item describes its cue through optional marker clips, read once here:
// notgarupa/hold/introUsBit<k> and onceUsBit<k> are the bits of the intro
// and of the non-looping cue length in µs, noCueLimit removes the cue limit.

// holdMarkerBits is the number of µs bits (up to 16.7 s).
const holdMarkerBits = 24

// holdLoopLimit is the usual SE_RHYTHM_TAP_LONG cue limit: the original
// keeps at most 4 playbacks of the cue, a new start stopping the oldest
// (newest-replaces-oldest). The effect item can lift it (HoldSE.Limit).
const holdLoopLimit = 4

// holdSlots is the size of the hold-sound tables; a cue without a limit may
// use all of them.
const holdSlots = 16

type HoldSEData struct {
	sonolus.LevelDataResource
	Intro float64 // seconds of #intro before #loop starts (0: none)
	Once  float64 // seconds of a non-looping cue (0: the cue loops)
	Limit float64 // cue limit: hold sounds alive at once
}

var HoldSE = HoldSEData{}

func initHoldSE() {
	HoldSE.Intro, HoldSE.Once = 0, 0
	us := 1.0
	for k := 0; k < holdMarkerBits; k++ {
		if Effects.IntroUsBitOpt[k].Exists() {
			HoldSE.Intro += us
		}
		if Effects.OnceUsBitOpt[k].Exists() {
			HoldSE.Once += us
		}
		us *= 2
	}
	HoldSE.Intro /= 1e6
	HoldSE.Once /= 1e6
	if !Effects.HoldLoopOpt.Exists() && HoldSE.Once <= 0 {
		HoldSE.Once = 1e-3 // no loop and no length: never sound longer than a frame
	}
	HoldSE.Limit = holdLoopLimit
	if Effects.NoCueLimitOpt.Exists() {
		HoldSE.Limit = holdSlots
	}
}

// holdScheduled: a live hold sound needs scheduled handles (an intro or a
// non-looping cue); a plain loop plays as a looped effect.
func holdScheduled() bool { return HoldSE.Intro > 0 || HoldSE.Once > 0 }

// startHoldSound starts a hold sound at t and returns its main handle ID and
// its intro handle ID (0: none). live starts a plain loop as a looped effect
// at once; otherwise every handle is scheduled.
func startHoldSound(t float64, live bool) (float64, float64) {
	if HoldSE.Once > 0 {
		h := Effects.HoldWhole.PlayLoopedScheduled(t)
		h.Stop(t + HoldSE.Once)
		return h.ID, 0
	}
	if HoldSE.Intro > 0 {
		intro := Effects.HoldIntroOpt.PlayLoopedScheduled(t)
		intro.Stop(t + HoldSE.Intro)
		return Effects.HoldLoopOpt.PlayLoopedScheduled(t + HoldSE.Intro).ID, intro.ID
	}
	if live {
		return Effects.HoldLoopOpt.PlayLooped().ID, 0
	}
	return Effects.HoldLoopOpt.PlayLoopedScheduled(t).ID, 0
}

// stopHoldSound stops a hold sound at t, silently (the cue limit stealing
// it). live: it was started by startHoldSound(…, true).
func stopHoldSound(main, intro, t float64, live bool) {
	if live && !holdScheduled() {
		sonolus.LoopedEffectHandle{ID: main}.Stop()
		return
	}
	sonolus.ScheduledLoopedEffectHandle{ID: main}.Stop(t)
	if intro > 0 {
		sonolus.ScheduledLoopedEffectHandle{ID: intro}.Stop(t)
	}
}

// releaseHoldSound stops a hold sound at t and plays the release fade of a
// loop. A non-looping cue just stops: its release would fade the cue's
// current position, which no clip holds, and after the cue's end there is
// nothing to fade.
func releaseHoldSound(main, intro, t float64, live bool) {
	stopHoldSound(main, intro, t, live)
	playHoldRelease(t, live)
}

// playHoldRelease plays the release fade of a stopped hold loop at t (live:
// now); a non-looping cue has none.
func playHoldRelease(t float64, live bool) {
	switch {
	case HoldSE.Once > 0:
	case live:
		Effects.HoldRelease.Play(0)
	default:
		Effects.HoldRelease.PlayScheduled(t, 0)
	}
}

// scheduleHoldSound plays a whole hold, start to end, without the cue limit.
func scheduleHoldSound(start, end float64) {
	main, intro := startHoldSound(start, false)
	releaseHoldSound(main, intro, end, false)
}

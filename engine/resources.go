package main

import "github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"

// Effect clips (original cue names, RESOURCE_CONTRACT.md).
type EffectsData struct {
	sonolus.EffectResource
	Perfect, Great, Good, Flick sonolus.Clip
	Directional                 [3]sonolus.Clip
	EmptyTap                    sonolus.Clip
	// Hold sound: the whole cue, its release fade, and the optional parts
	// (see hold_se.go): #loop when the cue loops, #intro when the loop does
	// not start at 0, and the marker clips of the item's hold profile.
	HoldWhole, HoldRelease sonolus.Clip
	HoldLoopOpt            sonolus.Clip
	HoldIntroOpt           sonolus.Clip
	IntroUsBitOpt          [holdMarkerBits]sonolus.Clip
	OnceUsBitOpt           [holdMarkerBits]sonolus.Clip
	NoCueLimitOpt          sonolus.Clip
	Skill                  sonolus.Clip
}

var Effects = &EffectsData{
	Perfect: sonolus.EffectClip("perfect"), Great: sonolus.EffectClip("great"), Good: sonolus.EffectClip("good"),
	Flick:       sonolus.EffectClip("flick"),
	Directional: [3]sonolus.Clip{sonolus.EffectClip("directional_fl"), sonolus.EffectClip("directional_fl_2"), sonolus.EffectClip("directional_fl_3")},
	EmptyTap:    sonolus.EffectClip("game_button"),
	HoldWhole:   sonolus.EffectClip("SE_RHYTHM_TAP_LONG"), HoldRelease: sonolus.EffectClip("SE_RHYTHM_TAP_LONG#release"),
	HoldLoopOpt: sonolus.EffectClip("SE_RHYTHM_TAP_LONG#loop"), HoldIntroOpt: sonolus.EffectClip("SE_RHYTHM_TAP_LONG#intro"),
	IntroUsBitOpt: [holdMarkerBits]sonolus.Clip{sonolus.EffectClip("notgarupa/hold/introUsBit0"), sonolus.EffectClip("notgarupa/hold/introUsBit1"), sonolus.EffectClip("notgarupa/hold/introUsBit2"), sonolus.EffectClip("notgarupa/hold/introUsBit3"), sonolus.EffectClip("notgarupa/hold/introUsBit4"), sonolus.EffectClip("notgarupa/hold/introUsBit5"), sonolus.EffectClip("notgarupa/hold/introUsBit6"), sonolus.EffectClip("notgarupa/hold/introUsBit7"), sonolus.EffectClip("notgarupa/hold/introUsBit8"), sonolus.EffectClip("notgarupa/hold/introUsBit9"), sonolus.EffectClip("notgarupa/hold/introUsBit10"), sonolus.EffectClip("notgarupa/hold/introUsBit11"), sonolus.EffectClip("notgarupa/hold/introUsBit12"), sonolus.EffectClip("notgarupa/hold/introUsBit13"), sonolus.EffectClip("notgarupa/hold/introUsBit14"), sonolus.EffectClip("notgarupa/hold/introUsBit15"), sonolus.EffectClip("notgarupa/hold/introUsBit16"), sonolus.EffectClip("notgarupa/hold/introUsBit17"), sonolus.EffectClip("notgarupa/hold/introUsBit18"), sonolus.EffectClip("notgarupa/hold/introUsBit19"), sonolus.EffectClip("notgarupa/hold/introUsBit20"), sonolus.EffectClip("notgarupa/hold/introUsBit21"), sonolus.EffectClip("notgarupa/hold/introUsBit22"), sonolus.EffectClip("notgarupa/hold/introUsBit23")},
	OnceUsBitOpt:  [holdMarkerBits]sonolus.Clip{sonolus.EffectClip("notgarupa/hold/onceUsBit0"), sonolus.EffectClip("notgarupa/hold/onceUsBit1"), sonolus.EffectClip("notgarupa/hold/onceUsBit2"), sonolus.EffectClip("notgarupa/hold/onceUsBit3"), sonolus.EffectClip("notgarupa/hold/onceUsBit4"), sonolus.EffectClip("notgarupa/hold/onceUsBit5"), sonolus.EffectClip("notgarupa/hold/onceUsBit6"), sonolus.EffectClip("notgarupa/hold/onceUsBit7"), sonolus.EffectClip("notgarupa/hold/onceUsBit8"), sonolus.EffectClip("notgarupa/hold/onceUsBit9"), sonolus.EffectClip("notgarupa/hold/onceUsBit10"), sonolus.EffectClip("notgarupa/hold/onceUsBit11"), sonolus.EffectClip("notgarupa/hold/onceUsBit12"), sonolus.EffectClip("notgarupa/hold/onceUsBit13"), sonolus.EffectClip("notgarupa/hold/onceUsBit14"), sonolus.EffectClip("notgarupa/hold/onceUsBit15"), sonolus.EffectClip("notgarupa/hold/onceUsBit16"), sonolus.EffectClip("notgarupa/hold/onceUsBit17"), sonolus.EffectClip("notgarupa/hold/onceUsBit18"), sonolus.EffectClip("notgarupa/hold/onceUsBit19"), sonolus.EffectClip("notgarupa/hold/onceUsBit20"), sonolus.EffectClip("notgarupa/hold/onceUsBit21"), sonolus.EffectClip("notgarupa/hold/onceUsBit22"), sonolus.EffectClip("notgarupa/hold/onceUsBit23")},
	NoCueLimitOpt: sonolus.EffectClip("notgarupa/hold/noCueLimit"),
	Skill:         sonolus.EffectClip("SE_RHYTHM_TAP_SKILL"),
}

// Particle effects: one per original prefab, plus its lane-plane part.
type ParticlesData struct {
	sonolus.ParticleResource
	// The prefab families of fx.go: the view part and the lane-plane parts at
	// every tilt any tap effect skin uses; an item has only some of them.
	BaseOpt    [fxFamilies]sonolus.Effect
	PlaneOpt   [fxFamilies * fxPlanes]sonolus.Effect
	DirL, DirR [3]sonolus.Effect
	// Clear fireworks: GameClearParticle roots in spawn order, then the
	// Full Combo / All Perfect systems (fireworkStart / fireworkAnchor).
	// HABAHIRO width variants _2.._7 (separate HABAHIRO particle item).
	PerfectW            [6]sonolus.Effect
	PerfectPlaneW       [6]sonolus.Effect
	GreatW              [6]sonolus.Effect
	GreatPlaneW         [6]sonolus.Effect
	GoodW               [6]sonolus.Effect
	SkillPerfectW       [6]sonolus.Effect
	SkillPerfectPlaneW  [6]sonolus.Effect
	SkillPerfectPlane2W [6]sonolus.Effect
	SkillGreatW         [6]sonolus.Effect
	SkillGreatPlaneW    [6]sonolus.Effect
	SkillGoodW          [6]sonolus.Effect
	SwipeW              [6]sonolus.Effect
	SwipePlaneW         [6]sonolus.Effect
	KeepW               [6]sonolus.Effect
	KeepPlaneW          [6]sonolus.Effect
}

var Particles = &ParticlesData{
	BaseOpt: [fxFamilies]sonolus.Effect{sonolus.ParticleEffect("effect_tap"), sonolus.ParticleEffect("effect_tap_good"), sonolus.ParticleEffect("effect_tap_great"), sonolus.ParticleEffect("effect_tap_perfect"), sonolus.ParticleEffect("effect_tap_skill_good"), sonolus.ParticleEffect("effect_tap_skill_great"), sonolus.ParticleEffect("effect_tap_skill_perfect"), sonolus.ParticleEffect("effect_tap_swipe"), sonolus.ParticleEffect("effect_TapKeep"), sonolus.ParticleEffect("effect_tap_directional_flick_l_finger"), sonolus.ParticleEffect("effect_tap_directional_flick_r_finger")},
	PlaneOpt: [fxFamilies * fxPlanes]sonolus.Effect{
		sonolus.ParticleEffect("effect_tap#plane-65"), sonolus.ParticleEffect("effect_tap#plane44"), sonolus.ParticleEffect("effect_tap#plane46"), sonolus.ParticleEffect("effect_tap#plane48"), sonolus.ParticleEffect("effect_tap#plane50"), sonolus.ParticleEffect("effect_tap#plane52"), sonolus.ParticleEffect("effect_tap#plane53"), sonolus.ParticleEffect("effect_tap#plane55"), sonolus.ParticleEffect("effect_tap#plane58"), sonolus.ParticleEffect("effect_tap#plane60"), sonolus.ParticleEffect("effect_tap#plane62"), sonolus.ParticleEffect("effect_tap#plane64"),
		sonolus.ParticleEffect("effect_tap_good#plane-65"), sonolus.ParticleEffect("effect_tap_good#plane44"), sonolus.ParticleEffect("effect_tap_good#plane46"), sonolus.ParticleEffect("effect_tap_good#plane48"), sonolus.ParticleEffect("effect_tap_good#plane50"), sonolus.ParticleEffect("effect_tap_good#plane52"), sonolus.ParticleEffect("effect_tap_good#plane53"), sonolus.ParticleEffect("effect_tap_good#plane55"), sonolus.ParticleEffect("effect_tap_good#plane58"), sonolus.ParticleEffect("effect_tap_good#plane60"), sonolus.ParticleEffect("effect_tap_good#plane62"), sonolus.ParticleEffect("effect_tap_good#plane64"),
		sonolus.ParticleEffect("effect_tap_great#plane-65"), sonolus.ParticleEffect("effect_tap_great#plane44"), sonolus.ParticleEffect("effect_tap_great#plane46"), sonolus.ParticleEffect("effect_tap_great#plane48"), sonolus.ParticleEffect("effect_tap_great#plane50"), sonolus.ParticleEffect("effect_tap_great#plane52"), sonolus.ParticleEffect("effect_tap_great#plane53"), sonolus.ParticleEffect("effect_tap_great#plane55"), sonolus.ParticleEffect("effect_tap_great#plane58"), sonolus.ParticleEffect("effect_tap_great#plane60"), sonolus.ParticleEffect("effect_tap_great#plane62"), sonolus.ParticleEffect("effect_tap_great#plane64"),
		sonolus.ParticleEffect("effect_tap_perfect#plane-65"), sonolus.ParticleEffect("effect_tap_perfect#plane44"), sonolus.ParticleEffect("effect_tap_perfect#plane46"), sonolus.ParticleEffect("effect_tap_perfect#plane48"), sonolus.ParticleEffect("effect_tap_perfect#plane50"), sonolus.ParticleEffect("effect_tap_perfect#plane52"), sonolus.ParticleEffect("effect_tap_perfect#plane53"), sonolus.ParticleEffect("effect_tap_perfect#plane55"), sonolus.ParticleEffect("effect_tap_perfect#plane58"), sonolus.ParticleEffect("effect_tap_perfect#plane60"), sonolus.ParticleEffect("effect_tap_perfect#plane62"), sonolus.ParticleEffect("effect_tap_perfect#plane64"),
		sonolus.ParticleEffect("effect_tap_skill_good#plane-65"), sonolus.ParticleEffect("effect_tap_skill_good#plane44"), sonolus.ParticleEffect("effect_tap_skill_good#plane46"), sonolus.ParticleEffect("effect_tap_skill_good#plane48"), sonolus.ParticleEffect("effect_tap_skill_good#plane50"), sonolus.ParticleEffect("effect_tap_skill_good#plane52"), sonolus.ParticleEffect("effect_tap_skill_good#plane53"), sonolus.ParticleEffect("effect_tap_skill_good#plane55"), sonolus.ParticleEffect("effect_tap_skill_good#plane58"), sonolus.ParticleEffect("effect_tap_skill_good#plane60"), sonolus.ParticleEffect("effect_tap_skill_good#plane62"), sonolus.ParticleEffect("effect_tap_skill_good#plane64"),
		sonolus.ParticleEffect("effect_tap_skill_great#plane-65"), sonolus.ParticleEffect("effect_tap_skill_great#plane44"), sonolus.ParticleEffect("effect_tap_skill_great#plane46"), sonolus.ParticleEffect("effect_tap_skill_great#plane48"), sonolus.ParticleEffect("effect_tap_skill_great#plane50"), sonolus.ParticleEffect("effect_tap_skill_great#plane52"), sonolus.ParticleEffect("effect_tap_skill_great#plane53"), sonolus.ParticleEffect("effect_tap_skill_great#plane55"), sonolus.ParticleEffect("effect_tap_skill_great#plane58"), sonolus.ParticleEffect("effect_tap_skill_great#plane60"), sonolus.ParticleEffect("effect_tap_skill_great#plane62"), sonolus.ParticleEffect("effect_tap_skill_great#plane64"),
		sonolus.ParticleEffect("effect_tap_skill_perfect#plane-65"), sonolus.ParticleEffect("effect_tap_skill_perfect#plane44"), sonolus.ParticleEffect("effect_tap_skill_perfect#plane46"), sonolus.ParticleEffect("effect_tap_skill_perfect#plane48"), sonolus.ParticleEffect("effect_tap_skill_perfect#plane50"), sonolus.ParticleEffect("effect_tap_skill_perfect#plane52"), sonolus.ParticleEffect("effect_tap_skill_perfect#plane53"), sonolus.ParticleEffect("effect_tap_skill_perfect#plane55"), sonolus.ParticleEffect("effect_tap_skill_perfect#plane58"), sonolus.ParticleEffect("effect_tap_skill_perfect#plane60"), sonolus.ParticleEffect("effect_tap_skill_perfect#plane62"), sonolus.ParticleEffect("effect_tap_skill_perfect#plane64"),
		sonolus.ParticleEffect("effect_tap_swipe#plane-65"), sonolus.ParticleEffect("effect_tap_swipe#plane44"), sonolus.ParticleEffect("effect_tap_swipe#plane46"), sonolus.ParticleEffect("effect_tap_swipe#plane48"), sonolus.ParticleEffect("effect_tap_swipe#plane50"), sonolus.ParticleEffect("effect_tap_swipe#plane52"), sonolus.ParticleEffect("effect_tap_swipe#plane53"), sonolus.ParticleEffect("effect_tap_swipe#plane55"), sonolus.ParticleEffect("effect_tap_swipe#plane58"), sonolus.ParticleEffect("effect_tap_swipe#plane60"), sonolus.ParticleEffect("effect_tap_swipe#plane62"), sonolus.ParticleEffect("effect_tap_swipe#plane64"),
		sonolus.ParticleEffect("effect_TapKeep#plane-65"), sonolus.ParticleEffect("effect_TapKeep#plane44"), sonolus.ParticleEffect("effect_TapKeep#plane46"), sonolus.ParticleEffect("effect_TapKeep#plane48"), sonolus.ParticleEffect("effect_TapKeep#plane50"), sonolus.ParticleEffect("effect_TapKeep#plane52"), sonolus.ParticleEffect("effect_TapKeep#plane53"), sonolus.ParticleEffect("effect_TapKeep#plane55"), sonolus.ParticleEffect("effect_TapKeep#plane58"), sonolus.ParticleEffect("effect_TapKeep#plane60"), sonolus.ParticleEffect("effect_TapKeep#plane62"), sonolus.ParticleEffect("effect_TapKeep#plane64"),
		sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane-65"), sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane44"), sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane46"), sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane48"), sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane50"), sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane52"), sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane53"), sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane55"), sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane58"), sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane60"), sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane62"), sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane64"),
		sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane-65"), sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane44"), sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane46"), sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane48"), sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane50"), sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane52"), sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane53"), sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane55"), sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane58"), sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane60"), sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane62"), sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane64"),
	},
	DirL:                [3]sonolus.Effect{sonolus.ParticleEffect("effect_tap_directional_flick_l"), sonolus.ParticleEffect("effect_tap_directional_flick_l_2"), sonolus.ParticleEffect("effect_tap_directional_flick_l_3")},
	DirR:                [3]sonolus.Effect{sonolus.ParticleEffect("effect_tap_directional_flick_r"), sonolus.ParticleEffect("effect_tap_directional_flick_r_2"), sonolus.ParticleEffect("effect_tap_directional_flick_r_3")},
	PerfectW:            [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_perfect_2"), sonolus.ParticleEffect("effect_tap_perfect_3"), sonolus.ParticleEffect("effect_tap_perfect_4"), sonolus.ParticleEffect("effect_tap_perfect_5"), sonolus.ParticleEffect("effect_tap_perfect_6"), sonolus.ParticleEffect("effect_tap_perfect_7")},
	PerfectPlaneW:       [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_perfect_2#plane46"), sonolus.ParticleEffect("effect_tap_perfect_3#plane46"), sonolus.ParticleEffect("effect_tap_perfect_4#plane46"), sonolus.ParticleEffect("effect_tap_perfect_5#plane46"), sonolus.ParticleEffect("effect_tap_perfect_6#plane46"), sonolus.ParticleEffect("effect_tap_perfect_7#plane46")},
	GreatW:              [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_great_2"), sonolus.ParticleEffect("effect_tap_great_3"), sonolus.ParticleEffect("effect_tap_great_4"), sonolus.ParticleEffect("effect_tap_great_5"), sonolus.ParticleEffect("effect_tap_great_6"), sonolus.ParticleEffect("effect_tap_great_7")},
	GreatPlaneW:         [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_great_2#plane50"), sonolus.ParticleEffect("effect_tap_great_3#plane50"), sonolus.ParticleEffect("effect_tap_great_4#plane50"), sonolus.ParticleEffect("effect_tap_great_5#plane50"), sonolus.ParticleEffect("effect_tap_great_6#plane50"), sonolus.ParticleEffect("effect_tap_great_7#plane50")},
	GoodW:               [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_good_2"), sonolus.ParticleEffect("effect_tap_good_3"), sonolus.ParticleEffect("effect_tap_good_4"), sonolus.ParticleEffect("effect_tap_good_5"), sonolus.ParticleEffect("effect_tap_good_6"), sonolus.ParticleEffect("effect_tap_good_7")},
	SkillPerfectW:       [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_skill_perfect_2"), sonolus.ParticleEffect("effect_tap_skill_perfect_3"), sonolus.ParticleEffect("effect_tap_skill_perfect_4"), sonolus.ParticleEffect("effect_tap_skill_perfect_5"), sonolus.ParticleEffect("effect_tap_skill_perfect_6"), sonolus.ParticleEffect("effect_tap_skill_perfect_7")},
	SkillPerfectPlaneW:  [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_skill_perfect_2#plane46"), sonolus.ParticleEffect("effect_tap_skill_perfect_3#plane46"), sonolus.ParticleEffect("effect_tap_skill_perfect_4#plane46"), sonolus.ParticleEffect("effect_tap_skill_perfect_5#plane46"), sonolus.ParticleEffect("effect_tap_skill_perfect_6#plane46"), sonolus.ParticleEffect("effect_tap_skill_perfect_7#plane46")},
	SkillPerfectPlane2W: [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_skill_perfect_2#plane50"), sonolus.ParticleEffect("effect_tap_skill_perfect_3#plane50"), sonolus.ParticleEffect("effect_tap_skill_perfect_4#plane50"), sonolus.ParticleEffect("effect_tap_skill_perfect_5#plane50"), sonolus.ParticleEffect("effect_tap_skill_perfect_6#plane50"), sonolus.ParticleEffect("effect_tap_skill_perfect_7#plane50")},
	SkillGreatW:         [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_skill_great_2"), sonolus.ParticleEffect("effect_tap_skill_great_3"), sonolus.ParticleEffect("effect_tap_skill_great_4"), sonolus.ParticleEffect("effect_tap_skill_great_5"), sonolus.ParticleEffect("effect_tap_skill_great_6"), sonolus.ParticleEffect("effect_tap_skill_great_7")},
	SkillGreatPlaneW:    [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_skill_great_2#plane46"), sonolus.ParticleEffect("effect_tap_skill_great_3#plane46"), sonolus.ParticleEffect("effect_tap_skill_great_4#plane46"), sonolus.ParticleEffect("effect_tap_skill_great_5#plane46"), sonolus.ParticleEffect("effect_tap_skill_great_6#plane46"), sonolus.ParticleEffect("effect_tap_skill_great_7#plane46")},
	SkillGoodW:          [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_skill_good_2"), sonolus.ParticleEffect("effect_tap_skill_good_3"), sonolus.ParticleEffect("effect_tap_skill_good_4"), sonolus.ParticleEffect("effect_tap_skill_good_5"), sonolus.ParticleEffect("effect_tap_skill_good_6"), sonolus.ParticleEffect("effect_tap_skill_good_7")},
	SwipeW:              [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_swipe_2"), sonolus.ParticleEffect("effect_tap_swipe_3"), sonolus.ParticleEffect("effect_tap_swipe_4"), sonolus.ParticleEffect("effect_tap_swipe_5"), sonolus.ParticleEffect("effect_tap_swipe_6"), sonolus.ParticleEffect("effect_tap_swipe_7")},
	SwipePlaneW:         [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_swipe_2#plane46"), sonolus.ParticleEffect("effect_tap_swipe_3#plane46"), sonolus.ParticleEffect("effect_tap_swipe_4#plane46"), sonolus.ParticleEffect("effect_tap_swipe_5#plane46"), sonolus.ParticleEffect("effect_tap_swipe_6#plane46"), sonolus.ParticleEffect("effect_tap_swipe_7#plane46")},
	KeepW:               [6]sonolus.Effect{sonolus.ParticleEffect("effect_TapKeep_2"), sonolus.ParticleEffect("effect_TapKeep_3"), sonolus.ParticleEffect("effect_TapKeep_4"), sonolus.ParticleEffect("effect_TapKeep_5"), sonolus.ParticleEffect("effect_TapKeep_6"), sonolus.ParticleEffect("effect_TapKeep_7")},
	KeepPlaneW:          [6]sonolus.Effect{sonolus.ParticleEffect("effect_TapKeep_2#plane48"), sonolus.ParticleEffect("effect_TapKeep_3#plane48"), sonolus.ParticleEffect("effect_TapKeep_4#plane48"), sonolus.ParticleEffect("effect_TapKeep_5#plane48"), sonolus.ParticleEffect("effect_TapKeep_6#plane48"), sonolus.ParticleEffect("effect_TapKeep_7#plane48")},
}

// Particle effect durations (seconds): the canonical duration of each
// family, the same in every particle item (the generator puts every tap
// effect skin on it; good / great / perfect also cover the skill and width
// variants).
const (
	durTap, durGood, durGreat, durPerfect = 0.8, 0.8, 0.85, 1
	durSwipe, durKeep, durKeepPlane       = 1, 1, 18
	durDir1, durDir23, durFinger          = 0.7, 1.1, 0.35
	planeCos46, planeCos48, planeCos50    = 0.694658, 0.669131, 0.642788
)

type BucketsData struct {
	sonolus.BucketsResource
	Tap, Flick, Directional, LongHead, LongTail, SlideHead, SlideNode, SlideTail sonolus.Bucket
}

var Buckets = &BucketsData{
	Tap:         sonolus.JudgmentBucket("#MILLISECONDS", sonolus.JudgmentBucketSprite(Skin.NoteNormal[3], 0, 0, 2, 2, 0)),
	Flick:       sonolus.JudgmentBucket("#MILLISECONDS", sonolus.JudgmentBucketSprite(Skin.NoteFlick[3], 0, 0, 2, 2, 0), sonolus.JudgmentBucketSprite(Skin.NoteFlickTop, 0, 0.8, 1, 1, 0)),
	Directional: sonolus.JudgmentBucket("#MILLISECONDS", sonolus.JudgmentBucketSprite(Skin.NoteFlickR[3], 0, 0, 2, 2, 0), sonolus.JudgmentBucketSprite(Skin.NoteFlickTopR, 1.4, 0, 1, 1, 0)),
	LongHead:    sonolus.JudgmentBucket("#MILLISECONDS", sonolus.JudgmentBucketSprite(Skin.LongLine, 0, 1.5, 2, 3, 0), sonolus.JudgmentBucketSprite(Skin.NoteLong[3], 0, 0, 2, 2, 0)),
	LongTail:    sonolus.JudgmentBucket("#MILLISECONDS", sonolus.JudgmentBucketSprite(Skin.LongLine, 0, -1.5, 2, 3, 0), sonolus.JudgmentBucketSprite(Skin.NoteLong[3], 0, 0, 2, 2, 0)),
	SlideHead:   sonolus.JudgmentBucket("#MILLISECONDS", sonolus.JudgmentBucketSprite(Skin.LongLine2, 0, 1.5, 2, 3, 0), sonolus.JudgmentBucketSprite(Skin.NoteLong[3], 0, 0, 2, 2, 0)),
	SlideNode:   sonolus.JudgmentBucket("#MILLISECONDS", sonolus.JudgmentBucketSprite(Skin.LongLine2, 0, 0, 2, 4, 0), sonolus.JudgmentBucketSprite(Skin.NoteSlideAmong, 0, 0, 2, 2, 0)),
	SlideTail:   sonolus.JudgmentBucket("#MILLISECONDS", sonolus.JudgmentBucketSprite(Skin.LongLine2, 0, -1.5, 2, 3, 0), sonolus.JudgmentBucketSprite(Skin.NoteLong[3], 0, 0, 2, 2, 0)),
}

type InstructionsData struct {
	sonolus.InstructionResource
	Tap, Flick, Hold, Release, Follow sonolus.Text
}

var Instructions = &InstructionsData{
	Tap: sonolus.InstructionText("#TAP"), Flick: sonolus.InstructionText("#TAP_FLICK"),
	Hold: sonolus.InstructionText("#TAP_HOLD"), Release: sonolus.InstructionText("#RELEASE"),
	Follow: sonolus.InstructionText("#HOLD_FOLLOW"),
}

type InstructionIconsData struct {
	sonolus.InstructionIconResource
	Hand sonolus.Icon
}

var InstructionIcons = &InstructionIconsData{Hand: sonolus.InstructionIcon("#HAND")}

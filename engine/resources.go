package main

import "github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"

// Effect clips (original cue names, RESOURCE_CONTRACT.md).
type EffectsData struct {
	sonolus.EffectResource
	Perfect, Great, Good, Flick  sonolus.Clip
	Directional                  [3]sonolus.Clip
	EmptyTap                     sonolus.Clip
	HoldLoop, HoldRelease        sonolus.Clip
	Skill                        sonolus.Clip
	FullCombo, Clear, ClearVoice sonolus.Clip
}

var Effects = &EffectsData{
	Perfect: sonolus.EffectClip("perfect"), Great: sonolus.EffectClip("great"), Good: sonolus.EffectClip("good"),
	Flick:       sonolus.EffectClip("flick"),
	Directional: [3]sonolus.Clip{sonolus.EffectClip("directional_fl"), sonolus.EffectClip("directional_fl_2"), sonolus.EffectClip("directional_fl_3")},
	EmptyTap:    sonolus.EffectClip("game_button"),
	HoldLoop:    sonolus.EffectClip("SE_RHYTHM_TAP_LONG#loop"), HoldRelease: sonolus.EffectClip("SE_RHYTHM_TAP_LONG#release"),
	Skill:     sonolus.EffectClip("SE_RHYTHM_TAP_SKILL"),
	FullCombo: sonolus.EffectClip("SE_RHYTHM_FULLCOMBO"), Clear: sonolus.EffectClip("SE_RHYTHM_CLEAR"), ClearVoice: sonolus.EffectClip("SE_RHYTHM_CLEAR_VO"),
}

// Particle effects: one per original prefab, plus its lane-plane part.
type ParticlesData struct {
	sonolus.ParticleResource
	Tap                                                 sonolus.Effect
	Good, Great, GreatPlane, Perfect, PerfectPlane      sonolus.Effect
	SkillGood, SkillGreat, SkillGreatPlane              sonolus.Effect
	SkillPerfect, SkillPerfectPlane, SkillPerfectPlane2 sonolus.Effect
	Swipe, SwipePlane                                   sonolus.Effect
	Keep, KeepPlane                                     sonolus.Effect
	DirL, DirR                                          [3]sonolus.Effect
	FingerL, FingerLPlane, FingerR, FingerRPlane        sonolus.Effect
	// Clear fireworks: GameClearParticle roots in spawn order, then the
	// Full Combo / All Perfect systems (fireworkStart / fireworkAnchor).
	Firework   [11]sonolus.Effect
	FireworkFC sonolus.Effect
	FireworkAP sonolus.Effect
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
	Tap:  sonolus.ParticleEffect("effect_tap"),
	Good: sonolus.ParticleEffect("effect_tap_good"), Great: sonolus.ParticleEffect("effect_tap_great"), GreatPlane: sonolus.ParticleEffect("effect_tap_great#plane50"),
	Perfect: sonolus.ParticleEffect("effect_tap_perfect"), PerfectPlane: sonolus.ParticleEffect("effect_tap_perfect#plane46"),
	SkillGood: sonolus.ParticleEffect("effect_tap_skill_good"), SkillGreat: sonolus.ParticleEffect("effect_tap_skill_great"),
	SkillGreatPlane: sonolus.ParticleEffect("effect_tap_skill_great#plane46"),
	SkillPerfect:    sonolus.ParticleEffect("effect_tap_skill_perfect"), SkillPerfectPlane: sonolus.ParticleEffect("effect_tap_skill_perfect#plane46"),
	SkillPerfectPlane2: sonolus.ParticleEffect("effect_tap_skill_perfect#plane50"),
	Swipe:              sonolus.ParticleEffect("effect_tap_swipe"), SwipePlane: sonolus.ParticleEffect("effect_tap_swipe#plane46"),
	Keep: sonolus.ParticleEffect("effect_TapKeep"), KeepPlane: sonolus.ParticleEffect("effect_TapKeep#plane48"),
	DirL:    [3]sonolus.Effect{sonolus.ParticleEffect("effect_tap_directional_flick_l"), sonolus.ParticleEffect("effect_tap_directional_flick_l_2"), sonolus.ParticleEffect("effect_tap_directional_flick_l_3")},
	DirR:    [3]sonolus.Effect{sonolus.ParticleEffect("effect_tap_directional_flick_r"), sonolus.ParticleEffect("effect_tap_directional_flick_r_2"), sonolus.ParticleEffect("effect_tap_directional_flick_r_3")},
	FingerL: sonolus.ParticleEffect("effect_tap_directional_flick_l_finger"), FingerLPlane: sonolus.ParticleEffect("effect_tap_directional_flick_l_finger#plane46"),
	FingerR: sonolus.ParticleEffect("effect_tap_directional_flick_r_finger"), FingerRPlane: sonolus.ParticleEffect("effect_tap_directional_flick_r_finger#plane46"), Firework: [11]sonolus.Effect{sonolus.ParticleEffect("GameClearParticle/effect_par_firework_star_pink1"), sonolus.ParticleEffect("GameClearParticle/effect_par_firework_star_green1 (1)"), sonolus.ParticleEffect("GameClearParticle/effect_par_firework_star_yellow1 (1)"), sonolus.ParticleEffect("GameClearParticle/effect_par_firework_star_blue1 (1)"), sonolus.ParticleEffect("GameClearParticle/effect_par_firework_star_purple1 (1)"), sonolus.ParticleEffect("GameClearParticle/effect_par_firework_star_green1#effect_gameclear_firework_star_a"), sonolus.ParticleEffect("GameClearParticle/effect_par_firework_star_green1#effect_gameclear_firework_star_b"), sonolus.ParticleEffect("GameClearParticle/effect_par_firework_star_pink1 (1)"), sonolus.ParticleEffect("GameClearParticle/effect_par_firework_star_blue1"), sonolus.ParticleEffect("GameClearParticle/effect_par_firework_star_yellow1"), sonolus.ParticleEffect("GameClearParticle/effect_par_firework_star_purple1")},
	FireworkFC: sonolus.ParticleEffect("FullComboAnimation/effect_par_firework_star"),
	FireworkAP: sonolus.ParticleEffect("AllPerfectAnimation/effect_par_firework_star"), PerfectW: [6]sonolus.Effect{sonolus.ParticleEffect("effect_tap_perfect_2"), sonolus.ParticleEffect("effect_tap_perfect_3"), sonolus.ParticleEffect("effect_tap_perfect_4"), sonolus.ParticleEffect("effect_tap_perfect_5"), sonolus.ParticleEffect("effect_tap_perfect_6"), sonolus.ParticleEffect("effect_tap_perfect_7")},
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

// Particle effect durations (seconds), from the particle pack contract.
const (
	durTap, durGood, durGreat, durPerfect = 0.6, 0.6, 0.7, 0.9
	durSwipe, durKeep, durKeepPlane       = 0.75, 1.0, 6.0
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

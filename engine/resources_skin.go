package main

import "github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"

// Skin sprites, named as in RESOURCE_CONTRACT.md (original sprite names).
type SkinData struct {
	sonolus.SkinResource
	NoteNormal     [7]sonolus.Sprite
	NoteNormal16   [7]sonolus.Sprite
	NoteSkill      [7]sonolus.Sprite
	NoteFlick      [7]sonolus.Sprite
	NoteLong       [7]sonolus.Sprite
	NoteLongFlash  [7]sonolus.Sprite
	NoteFlickL     [7]sonolus.Sprite
	NoteFlickR     [7]sonolus.Sprite
	NoteSlideAmong sonolus.Sprite
	NoteFlickTop   sonolus.Sprite
	NoteFlickTopL  sonolus.Sprite
	NoteFlickTopR  sonolus.Sprite
	SimLine        sonolus.Sprite
	// SyncEdgeMarginOpt exists in the items whose note skin has the master
	// noteSyncEdgeMargin 1.1 (TYPE7); optional, absent means 0.
	SyncEdgeMarginOpt sonolus.Sprite
	LongLine          sonolus.Sprite
	LongLine2         sonolus.Sprite
	FlickLineL        sonolus.Sprite
	FlickLineR        sonolus.Sprite
	Field             sonolus.Sprite
	JudgeLine         sonolus.Sprite
	LaneEffect        [4]sonolus.Sprite
	// JudgeText: the item's judge atlas, except judge_auto, which comes from
	// the default atlas (Auto Live never uses a special judge skin).
	JudgeText       [8]sonolus.Sprite
	ComboDigit      [10]sonolus.Sprite
	ComboDigitAP    [10]sonolus.Sprite
	ComboText       sonolus.Sprite
	ComboTextAP     sonolus.Sprite
	AddDigit        [10]sonolus.Sprite
	AddPlus         sonolus.Sprite
	ScorePanel      [3]sonolus.Sprite
	ScoreCover      [9]sonolus.Sprite
	MeterBlue       [9]sonolus.Sprite
	MeterGreen      [3]sonolus.Sprite
	MeterOrange     [3]sonolus.Sprite
	MeterPink       [3]sonolus.Sprite
	MeterS          sonolus.Sprite
	LevelMark       sonolus.Sprite
	ScoreDigitGray  [10]sonolus.Sprite
	ScoreDigitPink  [10]sonolus.Sprite
	RankLetter      [4]sonolus.Sprite
	LifeBG          [3]sonolus.Sprite
	LifeFill        [3]sonolus.Sprite
	LifeFillLow     [3]sonolus.Sprite
	LifeFill2       [3]sonolus.Sprite
	LifeWarning     [3]sonolus.Sprite
	LifeWarningBody [9]sonolus.Sprite
	NoHealth        [3]sonolus.Sprite
	LifeDigitGreen  [10]sonolus.Sprite
	LifeDigitRed    [10]sonolus.Sprite
	LifeSlash       sonolus.Sprite
	AutoPill        [3]sonolus.Sprite
	AutoText        sonolus.Sprite
	GameOverText1   sonolus.Sprite
	GameOverText2   sonolus.Sprite
	SSOverlay       sonolus.Sprite
	SSKira          [8]sonolus.Sprite
	LightBlue       sonolus.Sprite
	LightPale       sonolus.Sprite
	LightWhite      sonolus.Sprite
	Speaker         sonolus.Sprite
	SpeakerGlow     sonolus.Sprite
	FullCombo       [14]sonolus.Sprite
	AllPerfect      [18]sonolus.Sprite
	SuddenLine      [3]sonolus.Sprite
	// HABAHIRO (wide-note) sprites of the separate HABAHIRO skin item; a
	// family array is indexed by run (start lane s, width w):
	// (w−1)(16−w)/2 + s. See the skin generator's HABAHIRO.md.
	HabaNormal    [28]sonolus.Sprite
	HabaNormal16  [28]sonolus.Sprite
	HabaSkill     [28]sonolus.Sprite
	HabaFlick     [28]sonolus.Sprite
	HabaLong      [28]sonolus.Sprite
	HabaLongFlash [28]sonolus.Sprite
	HabaAmong     [7]sonolus.Sprite
	HabaFlickTop  [3]sonolus.Sprite
	HabaSimLine   sonolus.Sprite
	HabaLongLine  sonolus.Sprite
	HabaLongLine2 sonolus.Sprite
	HabaField     sonolus.Sprite
	HabaJudgeLine sonolus.Sprite
	HabaFlash     [4]sonolus.Sprite
}

var Skin = &SkinData{
	SkinResource:      sonolus.SkinResource{RenderMode: sonolus.RenderModeLightweight},
	NoteNormal:        [7]sonolus.Sprite{sonolus.SkinSprite("note_normal_0"), sonolus.SkinSprite("note_normal_1"), sonolus.SkinSprite("note_normal_2"), sonolus.SkinSprite("note_normal_3"), sonolus.SkinSprite("note_normal_4"), sonolus.SkinSprite("note_normal_5"), sonolus.SkinSprite("note_normal_6")},
	NoteNormal16:      [7]sonolus.Sprite{sonolus.SkinSprite("note_normal_16_0"), sonolus.SkinSprite("note_normal_16_1"), sonolus.SkinSprite("note_normal_16_2"), sonolus.SkinSprite("note_normal_16_3"), sonolus.SkinSprite("note_normal_16_4"), sonolus.SkinSprite("note_normal_16_5"), sonolus.SkinSprite("note_normal_16_6")},
	NoteSkill:         [7]sonolus.Sprite{sonolus.SkinSprite("note_skill_0"), sonolus.SkinSprite("note_skill_1"), sonolus.SkinSprite("note_skill_2"), sonolus.SkinSprite("note_skill_3"), sonolus.SkinSprite("note_skill_4"), sonolus.SkinSprite("note_skill_5"), sonolus.SkinSprite("note_skill_6")},
	NoteFlick:         [7]sonolus.Sprite{sonolus.SkinSprite("note_flick_0"), sonolus.SkinSprite("note_flick_1"), sonolus.SkinSprite("note_flick_2"), sonolus.SkinSprite("note_flick_3"), sonolus.SkinSprite("note_flick_4"), sonolus.SkinSprite("note_flick_5"), sonolus.SkinSprite("note_flick_6")},
	NoteLong:          [7]sonolus.Sprite{sonolus.SkinSprite("note_long_0"), sonolus.SkinSprite("note_long_1"), sonolus.SkinSprite("note_long_2"), sonolus.SkinSprite("note_long_3"), sonolus.SkinSprite("note_long_4"), sonolus.SkinSprite("note_long_5"), sonolus.SkinSprite("note_long_6")},
	NoteLongFlash:     [7]sonolus.Sprite{sonolus.SkinSprite("note_long_flash_0"), sonolus.SkinSprite("note_long_flash_1"), sonolus.SkinSprite("note_long_flash_2"), sonolus.SkinSprite("note_long_flash_3"), sonolus.SkinSprite("note_long_flash_4"), sonolus.SkinSprite("note_long_flash_5"), sonolus.SkinSprite("note_long_flash_6")},
	NoteFlickL:        [7]sonolus.Sprite{sonolus.SkinSprite("note_flick_l_0"), sonolus.SkinSprite("note_flick_l_1"), sonolus.SkinSprite("note_flick_l_2"), sonolus.SkinSprite("note_flick_l_3"), sonolus.SkinSprite("note_flick_l_4"), sonolus.SkinSprite("note_flick_l_5"), sonolus.SkinSprite("note_flick_l_6")},
	NoteFlickR:        [7]sonolus.Sprite{sonolus.SkinSprite("note_flick_r_0"), sonolus.SkinSprite("note_flick_r_1"), sonolus.SkinSprite("note_flick_r_2"), sonolus.SkinSprite("note_flick_r_3"), sonolus.SkinSprite("note_flick_r_4"), sonolus.SkinSprite("note_flick_r_5"), sonolus.SkinSprite("note_flick_r_6")},
	NoteSlideAmong:    sonolus.SkinSprite("note_slide_among"),
	NoteFlickTop:      sonolus.SkinSprite("note_flick_top"),
	NoteFlickTopL:     sonolus.SkinSprite("note_flick_top_l"),
	NoteFlickTopR:     sonolus.SkinSprite("note_flick_top_r"),
	SimLine:           sonolus.SkinSprite("simultaneous_line"),
	SyncEdgeMarginOpt: sonolus.SkinSprite("notgarupa/noteSyncEdgeMargin_1.1"),
	LongLine:          sonolus.SkinSprite("longNoteLine@E6E6E6"),
	LongLine2:         sonolus.SkinSprite("longNoteLine2@E6E6E6"),
	FlickLineL:        sonolus.SkinSprite("FlickNoteLine_l"),
	FlickLineR:        sonolus.SkinSprite("FlickNoteLine_r"),
	Field:             sonolus.SkinSprite("bg_line_rhythm"),
	JudgeLine:         sonolus.SkinSprite("game_play_line"),
	LaneEffect:        [4]sonolus.Sprite{sonolus.SkinSprite("NoteLaneEffect_1"), sonolus.SkinSprite("NoteLaneEffect_2"), sonolus.SkinSprite("NoteLaneEffect_3"), sonolus.SkinSprite("NoteLaneEffect_4")},
	JudgeText:         [8]sonolus.Sprite{sonolus.SkinSprite("judge_perfect"), sonolus.SkinSprite("judge_great"), sonolus.SkinSprite("judge_good"), sonolus.SkinSprite("judge_bad"), sonolus.SkinSprite("judge_miss"), sonolus.SkinSprite("default/judge_auto"), sonolus.SkinSprite("judge_fast"), sonolus.SkinSprite("judge_slow")},
	ComboDigit:        [10]sonolus.Sprite{sonolus.SkinSprite("icon_number_big_0"), sonolus.SkinSprite("icon_number_big_1"), sonolus.SkinSprite("icon_number_big_2"), sonolus.SkinSprite("icon_number_big_3"), sonolus.SkinSprite("icon_number_big_4"), sonolus.SkinSprite("icon_number_big_5"), sonolus.SkinSprite("icon_number_big_6"), sonolus.SkinSprite("icon_number_big_7"), sonolus.SkinSprite("icon_number_big_8"), sonolus.SkinSprite("icon_number_big_9")},
	ComboDigitAP:      [10]sonolus.Sprite{sonolus.SkinSprite("icon_number_big_AP_0"), sonolus.SkinSprite("icon_number_big_AP_1"), sonolus.SkinSprite("icon_number_big_AP_2"), sonolus.SkinSprite("icon_number_big_AP_3"), sonolus.SkinSprite("icon_number_big_AP_4"), sonolus.SkinSprite("icon_number_big_AP_5"), sonolus.SkinSprite("icon_number_big_AP_6"), sonolus.SkinSprite("icon_number_big_AP_7"), sonolus.SkinSprite("icon_number_big_AP_8"), sonolus.SkinSprite("icon_number_big_AP_9")},
	ComboText:         sonolus.SkinSprite("combo"),
	ComboTextAP:       sonolus.SkinSprite("combo_AP"),
	AddDigit:          [10]sonolus.Sprite{sonolus.SkinSprite("icon_number_0"), sonolus.SkinSprite("icon_number_1"), sonolus.SkinSprite("icon_number_2"), sonolus.SkinSprite("icon_number_3"), sonolus.SkinSprite("icon_number_4"), sonolus.SkinSprite("icon_number_5"), sonolus.SkinSprite("icon_number_6"), sonolus.SkinSprite("icon_number_7"), sonolus.SkinSprite("icon_number_8"), sonolus.SkinSprite("icon_number_9")},
	AddPlus:           sonolus.SkinSprite("icon_number_plus"),
	ScorePanel:        [3]sonolus.Sprite{sonolus.SkinSprite("gauge_base_score#L"), sonolus.SkinSprite("gauge_base_score#M"), sonolus.SkinSprite("gauge_base_score#R")},
	ScoreCover:        [9]sonolus.Sprite{sonolus.SkinSprite("bg_gauge_score_multi#TL"), sonolus.SkinSprite("bg_gauge_score_multi#T"), sonolus.SkinSprite("bg_gauge_score_multi#TR"), sonolus.SkinSprite("bg_gauge_score_multi#L"), sonolus.SkinSprite("bg_gauge_score_multi#M"), sonolus.SkinSprite("bg_gauge_score_multi#R"), sonolus.SkinSprite("bg_gauge_score_multi#BL"), sonolus.SkinSprite("bg_gauge_score_multi#B"), sonolus.SkinSprite("bg_gauge_score_multi#BR")},
	MeterBlue:         [9]sonolus.Sprite{sonolus.SkinSprite("score_meter_blue#TL"), sonolus.SkinSprite("score_meter_blue#T"), sonolus.SkinSprite("score_meter_blue#TR"), sonolus.SkinSprite("score_meter_blue#L"), sonolus.SkinSprite("score_meter_blue#M"), sonolus.SkinSprite("score_meter_blue#R"), sonolus.SkinSprite("score_meter_blue#BL"), sonolus.SkinSprite("score_meter_blue#B"), sonolus.SkinSprite("score_meter_blue#BR")},
	MeterGreen:        [3]sonolus.Sprite{sonolus.SkinSprite("score_meter_green#L"), sonolus.SkinSprite("score_meter_green#M"), sonolus.SkinSprite("score_meter_green#R")},
	MeterOrange:       [3]sonolus.Sprite{sonolus.SkinSprite("score_meter_orange#L"), sonolus.SkinSprite("score_meter_orange#M"), sonolus.SkinSprite("score_meter_orange#R")},
	MeterPink:         [3]sonolus.Sprite{sonolus.SkinSprite("score_meter_pink#L"), sonolus.SkinSprite("score_meter_pink#M"), sonolus.SkinSprite("score_meter_pink#R")},
	MeterS:            sonolus.SkinSprite("score_meter_s"),
	LevelMark:         sonolus.SkinSprite("level_mark"),
	ScoreDigitGray:    [10]sonolus.Sprite{sonolus.SkinSprite("sgm_28_BEBEBE_0"), sonolus.SkinSprite("sgm_28_BEBEBE_1"), sonolus.SkinSprite("sgm_28_BEBEBE_2"), sonolus.SkinSprite("sgm_28_BEBEBE_3"), sonolus.SkinSprite("sgm_28_BEBEBE_4"), sonolus.SkinSprite("sgm_28_BEBEBE_5"), sonolus.SkinSprite("sgm_28_BEBEBE_6"), sonolus.SkinSprite("sgm_28_BEBEBE_7"), sonolus.SkinSprite("sgm_28_BEBEBE_8"), sonolus.SkinSprite("sgm_28_BEBEBE_9")},
	ScoreDigitPink:    [10]sonolus.Sprite{sonolus.SkinSprite("sgm_28_FF3B72_0"), sonolus.SkinSprite("sgm_28_FF3B72_1"), sonolus.SkinSprite("sgm_28_FF3B72_2"), sonolus.SkinSprite("sgm_28_FF3B72_3"), sonolus.SkinSprite("sgm_28_FF3B72_4"), sonolus.SkinSprite("sgm_28_FF3B72_5"), sonolus.SkinSprite("sgm_28_FF3B72_6"), sonolus.SkinSprite("sgm_28_FF3B72_7"), sonolus.SkinSprite("sgm_28_FF3B72_8"), sonolus.SkinSprite("sgm_28_FF3B72_9")},
	RankLetter:        [4]sonolus.Sprite{sonolus.SkinSprite("sgm_12_FFFFFF_C"), sonolus.SkinSprite("sgm_12_FFFFFF_B"), sonolus.SkinSprite("sgm_12_FFFFFF_A"), sonolus.SkinSprite("sgm_12_FFFFFF_S")},
	LifeBG:            [3]sonolus.Sprite{sonolus.SkinSprite("bg_health#L"), sonolus.SkinSprite("bg_health#M"), sonolus.SkinSprite("bg_health#R")},
	LifeFill:          [3]sonolus.Sprite{sonolus.SkinSprite("hp_meter@6EFF69#L"), sonolus.SkinSprite("hp_meter@6EFF69#M"), sonolus.SkinSprite("hp_meter@6EFF69#R")},
	LifeFillLow:       [3]sonolus.Sprite{sonolus.SkinSprite("hp_meter@FF9999#L"), sonolus.SkinSprite("hp_meter@FF9999#M"), sonolus.SkinSprite("hp_meter@FF9999#R")},
	LifeFill2:         [3]sonolus.Sprite{sonolus.SkinSprite("additive/hp_meter@5DFF0064#L"), sonolus.SkinSprite("additive/hp_meter@5DFF0064#M"), sonolus.SkinSprite("additive/hp_meter@5DFF0064#R")},
	LifeWarning:       [3]sonolus.Sprite{sonolus.SkinSprite("effect_health_caution_outline#L"), sonolus.SkinSprite("effect_health_caution_outline#M"), sonolus.SkinSprite("effect_health_caution_outline#R")},
	LifeWarningBody:   [9]sonolus.Sprite{sonolus.SkinSprite("effect_health_caution_inside#TL"), sonolus.SkinSprite("effect_health_caution_inside#T"), sonolus.SkinSprite("effect_health_caution_inside#TR"), sonolus.SkinSprite("effect_health_caution_inside#L"), sonolus.SkinSprite("effect_health_caution_inside#M"), sonolus.SkinSprite("effect_health_caution_inside#R"), sonolus.SkinSprite("effect_health_caution_inside#BL"), sonolus.SkinSprite("effect_health_caution_inside#B"), sonolus.SkinSprite("effect_health_caution_inside#BR")},
	NoHealth:          [3]sonolus.Sprite{sonolus.SkinSprite("bg_no_health#L"), sonolus.SkinSprite("bg_no_health#M"), sonolus.SkinSprite("bg_no_health#R")},
	LifeDigitGreen:    [10]sonolus.Sprite{sonolus.SkinSprite("sgm_18_00C000_0"), sonolus.SkinSprite("sgm_18_00C000_1"), sonolus.SkinSprite("sgm_18_00C000_2"), sonolus.SkinSprite("sgm_18_00C000_3"), sonolus.SkinSprite("sgm_18_00C000_4"), sonolus.SkinSprite("sgm_18_00C000_5"), sonolus.SkinSprite("sgm_18_00C000_6"), sonolus.SkinSprite("sgm_18_00C000_7"), sonolus.SkinSprite("sgm_18_00C000_8"), sonolus.SkinSprite("sgm_18_00C000_9")},
	LifeDigitRed:      [10]sonolus.Sprite{sonolus.SkinSprite("sgm_18_FE2349_0"), sonolus.SkinSprite("sgm_18_FE2349_1"), sonolus.SkinSprite("sgm_18_FE2349_2"), sonolus.SkinSprite("sgm_18_FE2349_3"), sonolus.SkinSprite("sgm_18_FE2349_4"), sonolus.SkinSprite("sgm_18_FE2349_5"), sonolus.SkinSprite("sgm_18_FE2349_6"), sonolus.SkinSprite("sgm_18_FE2349_7"), sonolus.SkinSprite("sgm_18_FE2349_8"), sonolus.SkinSprite("sgm_18_FE2349_9")},
	LifeSlash:         sonolus.SkinSprite("sgm_18_505050_/"),
	AutoPill:          [3]sonolus.Sprite{sonolus.SkinSprite("label_round_white@FF3B72#L"), sonolus.SkinSprite("label_round_white@FF3B72#M"), sonolus.SkinSprite("label_round_white@FF3B72#R")},
	AutoText:          sonolus.SkinSprite("sgmtext_24_FFFFFF_autolive"),
	GameOverText1:     sonolus.SkinSprite("sgmtext_22_FF0000_gameover_1"),
	GameOverText2:     sonolus.SkinSprite("sgmtext_22_FF0000_gameover_2"),
	SSOverlay:         sonolus.SkinSprite("ss_overlay@FFDDED"),
	SSKira:            [8]sonolus.Sprite{sonolus.SkinSprite("ss_kira@FFE8FBDC"), sonolus.SkinSprite("ss_kira@D7FBFF"), sonolus.SkinSprite("ss_kira@ECFFD3"), sonolus.SkinSprite("ss_kira@FFFFFFDC"), sonolus.SkinSprite("ss_kira@FDE3FF"), sonolus.SkinSprite("ss_kira@E6F6FF"), sonolus.SkinSprite("ss_kira@F1FFE6DC"), sonolus.SkinSprite("ss_kira@FFEED7")},
	LightBlue:         sonolus.SkinSprite("Tex_miniShine@0037FF"),
	LightPale:         sonolus.SkinSprite("Tex_miniShine@D5FFFF"),
	LightWhite:        sonolus.SkinSprite("Tex_miniShine"),
	Speaker:           sonolus.SkinSprite("Tex_speaker"),
	SpeakerGlow:       sonolus.SkinSprite("speaker_effect"),
	FullCombo:         [14]sonolus.Sprite{sonolus.SkinSprite("FullCombo_F"), sonolus.SkinSprite("FullCombo_U"), sonolus.SkinSprite("FullCombo_L"), sonolus.SkinSprite("FullCombo_C"), sonolus.SkinSprite("FullCombo_O"), sonolus.SkinSprite("FullCombo_M"), sonolus.SkinSprite("FullCombo_B"), sonolus.SkinSprite("FullCombo_fu"), sonolus.SkinSprite("FullCombo_ru"), sonolus.SkinSprite("FullCombo_ko"), sonolus.SkinSprite("FullCombo_un"), sonolus.SkinSprite("FullCombo_bo"), sonolus.SkinSprite("FullCombo_star"), sonolus.SkinSprite("FullCombo_thunder")},
	AllPerfect:        [18]sonolus.Sprite{sonolus.SkinSprite("AllPerfect_A"), sonolus.SkinSprite("AllPerfect_L"), sonolus.SkinSprite("AllPerfect_P"), sonolus.SkinSprite("AllPerfect_E_1"), sonolus.SkinSprite("AllPerfect_R"), sonolus.SkinSprite("AllPerfect_F"), sonolus.SkinSprite("AllPerfect_C"), sonolus.SkinSprite("AllPerfect_T"), sonolus.SkinSprite("AllPerfect_o"), sonolus.SkinSprite("AllPerfect_-"), sonolus.SkinSprite("AllPerfect_ru"), sonolus.SkinSprite("AllPerfect_pa"), sonolus.SkinSprite("AllPerfect_fu"), sonolus.SkinSprite("AllPerfect_e_2"), sonolus.SkinSprite("AllPerfect_ku"), sonolus.SkinSprite("AllPerfect_to"), sonolus.SkinSprite("AllPerfect_star"), sonolus.SkinSprite("AllPerfect_Thunder")},
	SuddenLine:        [3]sonolus.Sprite{sonolus.SkinSprite("sudden_line#L"), sonolus.SkinSprite("sudden_line#M"), sonolus.SkinSprite("sudden_line#R")}, HabaNormal: [28]sonolus.Sprite{sonolus.SkinSprite("habahiro/note_normal_0"), sonolus.SkinSprite("habahiro/note_normal_1"), sonolus.SkinSprite("habahiro/note_normal_2"), sonolus.SkinSprite("habahiro/note_normal_3"), sonolus.SkinSprite("habahiro/note_normal_4"), sonolus.SkinSprite("habahiro/note_normal_5"), sonolus.SkinSprite("habahiro/note_normal_6"), sonolus.SkinSprite("habahiro/note_normal_0_1"), sonolus.SkinSprite("habahiro/note_normal_1_2"), sonolus.SkinSprite("habahiro/note_normal_2_3"), sonolus.SkinSprite("habahiro/note_normal_3_4"), sonolus.SkinSprite("habahiro/note_normal_4_5"), sonolus.SkinSprite("habahiro/note_normal_5_6"), sonolus.SkinSprite("habahiro/note_normal_0_1_2"), sonolus.SkinSprite("habahiro/note_normal_1_2_3"), sonolus.SkinSprite("habahiro/note_normal_2_3_4"), sonolus.SkinSprite("habahiro/note_normal_3_4_5"), sonolus.SkinSprite("habahiro/note_normal_4_5_6"), sonolus.SkinSprite("habahiro/note_normal_0_1_2_3"), sonolus.SkinSprite("habahiro/note_normal_1_2_3_4"), sonolus.SkinSprite("habahiro/note_normal_2_3_4_5"), sonolus.SkinSprite("habahiro/note_normal_3_4_5_6"), sonolus.SkinSprite("habahiro/note_normal_0_1_2_3_4"), sonolus.SkinSprite("habahiro/note_normal_1_2_3_4_5"), sonolus.SkinSprite("habahiro/note_normal_2_3_4_5_6"), sonolus.SkinSprite("habahiro/note_normal_0_1_2_3_4_5"), sonolus.SkinSprite("habahiro/note_normal_1_2_3_4_5_6"), sonolus.SkinSprite("habahiro/note_normal_0_1_2_3_4_5_6")},
	HabaNormal16:  [28]sonolus.Sprite{sonolus.SkinSprite("habahiro/note_normal_16_0"), sonolus.SkinSprite("habahiro/note_normal_16_1"), sonolus.SkinSprite("habahiro/note_normal_16_2"), sonolus.SkinSprite("habahiro/note_normal_16_3"), sonolus.SkinSprite("habahiro/note_normal_16_4"), sonolus.SkinSprite("habahiro/note_normal_16_5"), sonolus.SkinSprite("habahiro/note_normal_16_6"), sonolus.SkinSprite("habahiro/note_normal_16_0_1"), sonolus.SkinSprite("habahiro/note_normal_16_1_2"), sonolus.SkinSprite("habahiro/note_normal_16_2_3"), sonolus.SkinSprite("habahiro/note_normal_16_3_4"), sonolus.SkinSprite("habahiro/note_normal_16_4_5"), sonolus.SkinSprite("habahiro/note_normal_16_5_6"), sonolus.SkinSprite("habahiro/note_normal_16_0_1_2"), sonolus.SkinSprite("habahiro/note_normal_16_1_2_3"), sonolus.SkinSprite("habahiro/note_normal_16_2_3_4"), sonolus.SkinSprite("habahiro/note_normal_16_3_4_5"), sonolus.SkinSprite("habahiro/note_normal_16_4_5_6"), sonolus.SkinSprite("habahiro/note_normal_16_0_1_2_3"), sonolus.SkinSprite("habahiro/note_normal_16_1_2_3_4"), sonolus.SkinSprite("habahiro/note_normal_16_2_3_4_5"), sonolus.SkinSprite("habahiro/note_normal_16_3_4_5_6"), sonolus.SkinSprite("habahiro/note_normal_16_0_1_2_3_4"), sonolus.SkinSprite("habahiro/note_normal_16_1_2_3_4_5"), sonolus.SkinSprite("habahiro/note_normal_16_2_3_4_5_6"), sonolus.SkinSprite("habahiro/note_normal_16_0_1_2_3_4_5"), sonolus.SkinSprite("habahiro/note_normal_16_1_2_3_4_5_6"), sonolus.SkinSprite("habahiro/note_normal_16_0_1_2_3_4_5_6")},
	HabaSkill:     [28]sonolus.Sprite{sonolus.SkinSprite("habahiro/note_skill_0"), sonolus.SkinSprite("habahiro/note_skill_1"), sonolus.SkinSprite("habahiro/note_skill_2"), sonolus.SkinSprite("habahiro/note_skill_3"), sonolus.SkinSprite("habahiro/note_skill_4"), sonolus.SkinSprite("habahiro/note_skill_5"), sonolus.SkinSprite("habahiro/note_skill_6"), sonolus.SkinSprite("habahiro/note_skill_0_1"), sonolus.SkinSprite("habahiro/note_skill_1_2"), sonolus.SkinSprite("habahiro/note_skill_2_3"), sonolus.SkinSprite("habahiro/note_skill_3_4"), sonolus.SkinSprite("habahiro/note_skill_4_5"), sonolus.SkinSprite("habahiro/note_skill_5_6"), sonolus.SkinSprite("habahiro/note_skill_0_1_2"), sonolus.SkinSprite("habahiro/note_skill_1_2_3"), sonolus.SkinSprite("habahiro/note_skill_2_3_4"), sonolus.SkinSprite("habahiro/note_skill_3_4_5"), sonolus.SkinSprite("habahiro/note_skill_4_5_6"), sonolus.SkinSprite("habahiro/note_skill_0_1_2_3"), sonolus.SkinSprite("habahiro/note_skill_1_2_3_4"), sonolus.SkinSprite("habahiro/note_skill_2_3_4_5"), sonolus.SkinSprite("habahiro/note_skill_3_4_5_6"), sonolus.SkinSprite("habahiro/note_skill_0_1_2_3_4"), sonolus.SkinSprite("habahiro/note_skill_1_2_3_4_5"), sonolus.SkinSprite("habahiro/note_skill_2_3_4_5_6"), sonolus.SkinSprite("habahiro/note_skill_0_1_2_3_4_5"), sonolus.SkinSprite("habahiro/note_skill_1_2_3_4_5_6"), sonolus.SkinSprite("habahiro/note_skill_0_1_2_3_4_5_6")},
	HabaFlick:     [28]sonolus.Sprite{sonolus.SkinSprite("habahiro/note_flick_0"), sonolus.SkinSprite("habahiro/note_flick_1"), sonolus.SkinSprite("habahiro/note_flick_2"), sonolus.SkinSprite("habahiro/note_flick_3"), sonolus.SkinSprite("habahiro/note_flick_4"), sonolus.SkinSprite("habahiro/note_flick_5"), sonolus.SkinSprite("habahiro/note_flick_6"), sonolus.SkinSprite("habahiro/note_flick_0_1"), sonolus.SkinSprite("habahiro/note_flick_1_2"), sonolus.SkinSprite("habahiro/note_flick_2_3"), sonolus.SkinSprite("habahiro/note_flick_3_4"), sonolus.SkinSprite("habahiro/note_flick_4_5"), sonolus.SkinSprite("habahiro/note_flick_5_6"), sonolus.SkinSprite("habahiro/note_flick_0_1_2"), sonolus.SkinSprite("habahiro/note_flick_1_2_3"), sonolus.SkinSprite("habahiro/note_flick_2_3_4"), sonolus.SkinSprite("habahiro/note_flick_3_4_5"), sonolus.SkinSprite("habahiro/note_flick_4_5_6"), sonolus.SkinSprite("habahiro/note_flick_0_1_2_3"), sonolus.SkinSprite("habahiro/note_flick_1_2_3_4"), sonolus.SkinSprite("habahiro/note_flick_2_3_4_5"), sonolus.SkinSprite("habahiro/note_flick_3_4_5_6"), sonolus.SkinSprite("habahiro/note_flick_0_1_2_3_4"), sonolus.SkinSprite("habahiro/note_flick_1_2_3_4_5"), sonolus.SkinSprite("habahiro/note_flick_2_3_4_5_6"), sonolus.SkinSprite("habahiro/note_flick_0_1_2_3_4_5"), sonolus.SkinSprite("habahiro/note_flick_1_2_3_4_5_6"), sonolus.SkinSprite("habahiro/note_flick_0_1_2_3_4_5_6")},
	HabaLong:      [28]sonolus.Sprite{sonolus.SkinSprite("habahiro/note_long_0"), sonolus.SkinSprite("habahiro/note_long_1"), sonolus.SkinSprite("habahiro/note_long_2"), sonolus.SkinSprite("habahiro/note_long_3"), sonolus.SkinSprite("habahiro/note_long_4"), sonolus.SkinSprite("habahiro/note_long_5"), sonolus.SkinSprite("habahiro/note_long_6"), sonolus.SkinSprite("habahiro/note_long_0_1"), sonolus.SkinSprite("habahiro/note_long_1_2"), sonolus.SkinSprite("habahiro/note_long_2_3"), sonolus.SkinSprite("habahiro/note_long_3_4"), sonolus.SkinSprite("habahiro/note_long_4_5"), sonolus.SkinSprite("habahiro/note_long_5_6"), sonolus.SkinSprite("habahiro/note_long_0_1_2"), sonolus.SkinSprite("habahiro/note_long_1_2_3"), sonolus.SkinSprite("habahiro/note_long_2_3_4"), sonolus.SkinSprite("habahiro/note_long_3_4_5"), sonolus.SkinSprite("habahiro/note_long_4_5_6"), sonolus.SkinSprite("habahiro/note_long_0_1_2_3"), sonolus.SkinSprite("habahiro/note_long_1_2_3_4"), sonolus.SkinSprite("habahiro/note_long_2_3_4_5"), sonolus.SkinSprite("habahiro/note_long_3_4_5_6"), sonolus.SkinSprite("habahiro/note_long_0_1_2_3_4"), sonolus.SkinSprite("habahiro/note_long_1_2_3_4_5"), sonolus.SkinSprite("habahiro/note_long_2_3_4_5_6"), sonolus.SkinSprite("habahiro/note_long_0_1_2_3_4_5"), sonolus.SkinSprite("habahiro/note_long_1_2_3_4_5_6"), sonolus.SkinSprite("habahiro/note_long_0_1_2_3_4_5_6")},
	HabaLongFlash: [28]sonolus.Sprite{sonolus.SkinSprite("habahiro/note_long_flash_0"), sonolus.SkinSprite("habahiro/note_long_flash_1"), sonolus.SkinSprite("habahiro/note_long_flash_2"), sonolus.SkinSprite("habahiro/note_long_flash_3"), sonolus.SkinSprite("habahiro/note_long_flash_4"), sonolus.SkinSprite("habahiro/note_long_flash_5"), sonolus.SkinSprite("habahiro/note_long_flash_6"), sonolus.SkinSprite("habahiro/note_long_flash_0_1"), sonolus.SkinSprite("habahiro/note_long_flash_1_2"), sonolus.SkinSprite("habahiro/note_long_flash_2_3"), sonolus.SkinSprite("habahiro/note_long_flash_3_4"), sonolus.SkinSprite("habahiro/note_long_flash_4_5"), sonolus.SkinSprite("habahiro/note_long_flash_5_6"), sonolus.SkinSprite("habahiro/note_long_flash_0_1_2"), sonolus.SkinSprite("habahiro/note_long_flash_1_2_3"), sonolus.SkinSprite("habahiro/note_long_flash_2_3_4"), sonolus.SkinSprite("habahiro/note_long_flash_3_4_5"), sonolus.SkinSprite("habahiro/note_long_flash_4_5_6"), sonolus.SkinSprite("habahiro/note_long_flash_0_1_2_3"), sonolus.SkinSprite("habahiro/note_long_flash_1_2_3_4"), sonolus.SkinSprite("habahiro/note_long_flash_2_3_4_5"), sonolus.SkinSprite("habahiro/note_long_flash_3_4_5_6"), sonolus.SkinSprite("habahiro/note_long_flash_0_1_2_3_4"), sonolus.SkinSprite("habahiro/note_long_flash_1_2_3_4_5"), sonolus.SkinSprite("habahiro/note_long_flash_2_3_4_5_6"), sonolus.SkinSprite("habahiro/note_long_flash_0_1_2_3_4_5"), sonolus.SkinSprite("habahiro/note_long_flash_1_2_3_4_5_6"), sonolus.SkinSprite("habahiro/note_long_flash_0_1_2_3_4_5_6")},
	HabaAmong:     [7]sonolus.Sprite{sonolus.SkinSprite("habahiro/note_slide_among"), sonolus.SkinSprite("habahiro/note_slide_among_2"), sonolus.SkinSprite("habahiro/note_slide_among_3"), sonolus.SkinSprite("habahiro/note_slide_among_4"), sonolus.SkinSprite("habahiro/note_slide_among_5"), sonolus.SkinSprite("habahiro/note_slide_among_6"), sonolus.SkinSprite("habahiro/note_slide_among_7")},
	HabaFlickTop:  [3]sonolus.Sprite{sonolus.SkinSprite("habahiro/note_flick_top"), sonolus.SkinSprite("habahiro/note_flick_top_2"), sonolus.SkinSprite("habahiro/note_flick_top_3")},
	HabaSimLine:   sonolus.SkinSprite("habahiro/simultaneous_line"),
	HabaLongLine:  sonolus.SkinSprite("habahiro/longNoteLine@E6E6E6"),
	HabaLongLine2: sonolus.SkinSprite("habahiro/longNoteLine2@E6E6E6"),
	HabaField:     sonolus.SkinSprite("habahiro/bg_line_rhythm"),
	HabaJudgeLine: sonolus.SkinSprite("habahiro/game_play_line"),
	HabaFlash:     [4]sonolus.Sprite{sonolus.SkinSprite("habahiro/bg_line_rhythm_flash"), sonolus.SkinSprite("habahiro/bg_line_rhythm_flash2"), sonolus.SkinSprite("habahiro/game_play_line_flash"), sonolus.SkinSprite("habahiro/game_play_line_flash2")},
}

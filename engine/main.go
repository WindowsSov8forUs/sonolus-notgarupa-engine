// Package main is the Ex-Garupa Sonolus engine: a reproduction of BanG
// Dream! Girls Band Party!'s live gameplay with GarupaEditor's chart
// extensions. See docs/ENGINE_DESIGN.md.
package main

import (
	_ "embed"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
)

type EngineConfiguration struct {
	sonolus.Configuration

	Speed        float64
	NoteSpeed    float64
	NoteSize     float64
	SyncLine     bool
	NoteColor    bool
	LaneEffect   bool
	FastSlow     bool
	ComboAP      bool
	ComboPos     int
	Mirror       bool
	SFX          bool
	AutoSFX      bool
	NoteEffects  bool
	DirEffect    int
	LineBright   float64
	SuddenRate   float64
	IllustBright float64
	UI           sonolus.UIConfig
	Replay       []string
}

const scope = "ExGarupa"

// Option titles, descriptions and select values are ##LOCALIZE texts (en
// first, the fallback for other languages); the names stay as they were, as
// Sonolus keys saved option values and replays by scope and name. Options
// that are original live settings use the original's wording of each server
// (dialog_liveSettings_* of each server's master, or of its wording
// collection bundle when the master has none: en EN, zhs CN, zht TW, ja JP);
// standard #names are localized by Sonolus.
const (
	titleNoteColor = `##LOCALIZE:{"en":"Off-Beat Coloring","zhs":"节奏的色觉辅助","zht":"節奏支援","ja":"リズムサポート"}`
	descNoteColor  = `##LOCALIZE:{"en":"Taps off the eighth-note grid use the 16th-note colour.","zhs":"不在八分音符位置上的单键使用十六分音符的颜色。","zht":"不在八分音符位置上的單鍵使用十六分音符的顏色。","ja":"8分音符の位置にないタップノーツを16分音符の色で表示します。"}`
	titleFastSlow  = `##LOCALIZE:{"en":"Show Fast/Slow","zhs":"FAST/SLOW表示","zht":"顯示FAST/SLOW","ja":"FAST/SLOW表示"}`
	titleComboAP   = `##LOCALIZE:{"en":"Display Combo Status","zhs":"Perfect状态显示","zht":"COMBO狀態表示","ja":"コンボ状態表示"}`
	descComboAP    = `##LOCALIZE:{"en":"Shows the All Perfect colour on the combo counter while every note is PERFECT.","zhs":"全部 PERFECT 时连击数以 All Perfect 配色显示。","zht":"全部 PERFECT 時 COMBO 數以 All Perfect 配色顯示。","ja":"すべて PERFECT の間、コンボ数をオールパーフェクトの色で表示します。"}`
	titleComboPos  = `##LOCALIZE:{"en":"Combo Counter Position","zhs":"连击数显示位置","zht":"COMBO數顯示位置","ja":"コンボ数表示位置"}`
	comboRightTop  = `##LOCALIZE:{"en":"Right / Top","zhs":"右・上","zht":"右・上","ja":"右・上"}`
	comboRightMid  = `##LOCALIZE:{"en":"Right / Default","zhs":"右・标准","zht":"右・標準","ja":"右・標準"}`
	comboRightBot  = `##LOCALIZE:{"en":"Right / Bottom","zhs":"右・下","zht":"右・下","ja":"右・下"}`
	comboLeftTop   = `##LOCALIZE:{"en":"Left / Top","zhs":"左・上","zht":"左・上","ja":"左・上"}`
	comboLeftMid   = `##LOCALIZE:{"en":"Left / Default","zhs":"左・标准","zht":"左・標準","ja":"左・標準"}`
	comboLeftBot   = `##LOCALIZE:{"en":"Left / Bottom","zhs":"左・下","zht":"左・下","ja":"左・下"}`
	titleDirEffect = `##LOCALIZE:{"en":"Directional Flick Note Effect","zhs":"左右flick标志 效果","zht":"左右滑動圖案特效","ja":"左右フリックアイコンエフェクト"}`
	dirEffectLarge = `##LOCALIZE:{"en":"Increase","zhs":"大","zht":"大","ja":"大"}`
	dirEffectSmall = `##LOCALIZE:{"en":"Decrease","zhs":"小","zht":"小","ja":"小"}`
	titleLineAlpha = `##LOCALIZE:{"en":"Long Note Transparency","zhs":"长按按键条不透明度","zht":"長壓線的深淺","ja":"長押しラインの濃さ"}`
	descLineAlpha  = `##LOCALIZE:{"en":"Opacity of Long and Slide bodies.","zhs":"长按与滑条主体的不透明度。","zht":"長壓與滑動主體的不透明度。","ja":"ロング・スライドのラインの不透明度。"}`
	titleSudden    = `##LOCALIZE:{"en":"Note Position","zhs":"节奏图标的出现位置","zht":"節奏圖示的表示位置","ja":"リズムアイコンの表示位置"}`
	descSudden     = `##LOCALIZE:{"en":"Covers the upper part of the lanes.","zhs":"遮住轨道的上方部分。","zht":"遮住軌跡的上方部分。","ja":"レーンの上側を隠します。"}`
	// dialog_liveSetting_LightModeBrightness of each server's master
	titleIllustBright = `##LOCALIZE:{"en":"Background Brightness","zhs":"背景亮度","zht":"佈景亮度","ja":"背景の明るさ"}`
	descIllustBright  = `##LOCALIZE:{"en":"Brightness of a level's custom background.","zhs":"关卡自定义背景的亮度。","zht":"關卡自訂背景的亮度。","ja":"譜面のカスタム背景の明るさ。"}`
)

var Config = EngineConfiguration{
	Speed: sonolus.SliderOption(sonolus.SliderOptionConfig{
		Name: "#SPEED", Standard: true, Default: 1, Min: 0.5, Max: 2, Step: 0.05, Unit: "#PERCENTAGE_UNIT",
	}),
	NoteSpeed: sonolus.SliderOption(sonolus.SliderOptionConfig{
		Name: "#NOTE_SPEED", Scope: scope, Default: 5, Min: 1, Max: 12, Step: 0.1,
	}),
	NoteSize: sonolus.SliderOption(sonolus.SliderOptionConfig{
		Name: "#NOTE_SIZE", Scope: scope, Default: 1, Min: 0.1, Max: 2, Step: 0.1, Unit: "#PERCENTAGE_UNIT",
	}),
	SyncLine: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "#SIMLINE", Scope: scope, Default: true,
	}),
	NoteColor: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "Off-beat note colour", Title: titleNoteColor, Description: descNoteColor, Scope: scope, Default: true,
	}),
	LaneEffect: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "#LANE_EFFECT", Scope: scope, Default: true,
	}),
	FastSlow: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "FAST / SLOW", Title: titleFastSlow, Scope: scope, Default: true,
	}),
	ComboAP: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "All Perfect combo colour", Title: titleComboAP, Description: descComboAP, Scope: scope, Default: false,
	}),
	ComboPos: sonolus.SelectOption(sonolus.SelectOptionConfig{
		Name: "Combo position", Title: titleComboPos, Scope: scope, Default: 1,
		Values: []string{comboRightTop, comboRightMid, comboRightBot, comboLeftTop, comboLeftMid, comboLeftBot},
	}),
	Mirror: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "#MIRROR", Scope: scope, Default: false,
	}),
	SFX: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "#EFFECT", Scope: scope, Default: true,
	}),
	AutoSFX: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "#EFFECT_AUTO", Scope: scope, Default: false,
	}),
	NoteEffects: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "#NOTE_EFFECT", Scope: scope, Default: true,
	}),
	DirEffect: sonolus.SelectOption(sonolus.SelectOptionConfig{
		Name: "Directional flick effect", Title: titleDirEffect, Scope: scope, Default: 0,
		Values: []string{dirEffectLarge, dirEffectSmall},
	}),
	LineBright: sonolus.SliderOption(sonolus.SliderOptionConfig{
		Name: "Long note line brightness", Title: titleLineAlpha, Description: descLineAlpha, Scope: scope,
		Default: 80, Min: 10, Max: 100, Step: 10,
	}),
	SuddenRate: sonolus.SliderOption(sonolus.SliderOptionConfig{
		Name: "Sudden", Title: titleSudden, Description: descSudden, Scope: scope, Default: 0, Min: 0, Max: 100, Step: 1,
	}),
	// the original Light mode brightness (LiveCoreSettingsLightModeBrightness:
	// 30–100 in steps of 10, default 80)
	IllustBright: sonolus.SliderOption(sonolus.SliderOptionConfig{
		Name: "Background brightness", Title: titleIllustBright, Description: descIllustBright, Scope: scope,
		Default: 80, Min: 30, Max: 100, Step: 10,
	}),
	UI: sonolus.UIConfig{
		Scope:                         scope,
		PrimaryMetric:                 sonolus.UIMetricArcade,
		SecondaryMetric:               sonolus.UIMetricLife,
		MenuVisibility:                sonolus.UIVisibility{Scale: 1, Alpha: 1},
		JudgmentVisibility:            sonolus.UIVisibility{Scale: 1, Alpha: 0},
		ComboVisibility:               sonolus.UIVisibility{Scale: 1, Alpha: 0},
		PrimaryMetricVisibility:       sonolus.UIVisibility{Scale: 1, Alpha: 0},
		SecondaryMetricVisibility:     sonolus.UIVisibility{Scale: 1, Alpha: 0},
		ProgressVisibility:            sonolus.UIVisibility{Scale: 1, Alpha: 1},
		TutorialNavigationVisibility:  sonolus.UIVisibility{Scale: 1, Alpha: 1},
		TutorialInstructionVisibility: sonolus.UIVisibility{Scale: 1, Alpha: 1},
		JudgmentAnimation: sonolus.UIAnimation{
			Scale: sonolus.UITween{From: 1, To: 1, Ease: sonolus.UIEaseNone},
			Alpha: sonolus.UITween{From: 1, To: 1, Ease: sonolus.UIEaseNone},
		},
		ComboAnimation: sonolus.UIAnimation{
			Scale: sonolus.UITween{From: 1, To: 1, Ease: sonolus.UIEaseNone},
			Alpha: sonolus.UITween{From: 1, To: 1, Ease: sonolus.UIEaseNone},
		},
		JudgmentErrorStyle:     sonolus.UIJudgmentErrorNone,
		JudgmentErrorPlacement: sonolus.UIJudgmentErrorTop,
	},
	Replay: []string{"#NOTE_SPEED", "#NOTE_SIZE", "#MIRROR"},
}

//sonolus:level
//go:embed dev-level.json
var DevelopmentLevel sonolus.LevelFile

func main() {}

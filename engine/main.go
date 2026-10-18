// Package main is the NotGarupa Sonolus engine: a reproduction of BanG
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
	FullLength   bool
	SFX          bool
	AutoSFX      bool
	StageEffects bool
	NoteEffects  bool
	LineBright   float64
	SuddenRate   float64
	UI           sonolus.UIConfig
	Replay       []string
}

const scope = "NotGarupa"

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
		Name: "Off-beat note colour", Scope: scope, Default: true,
		Description: "Taps off the eighth-note grid use the 16th-note colour.",
	}),
	LaneEffect: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "#LANE_EFFECT", Scope: scope, Default: true,
	}),
	FastSlow: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "FAST / SLOW", Scope: scope, Default: true,
	}),
	ComboAP: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "All Perfect combo colour", Scope: scope, Default: false,
	}),
	ComboPos: sonolus.SelectOption(sonolus.SelectOptionConfig{
		Name: "Combo position", Scope: scope, Default: 1,
		Values: []string{"Right top", "Right centre", "Right bottom", "Left top", "Left centre", "Left bottom"},
	}),
	Mirror: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "#MIRROR", Scope: scope, Default: false,
	}),
	FullLength: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "Full-length song", Scope: scope, Default: false,
		Description: "Halves life damage, as for full-length songs in the original.",
	}),
	SFX: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "#EFFECT", Scope: scope, Default: true,
	}),
	AutoSFX: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "#EFFECT_AUTO", Scope: scope, Default: false,
	}),
	StageEffects: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "Stage lights", Scope: scope, Default: true,
	}),
	NoteEffects: sonolus.ToggleOption(sonolus.ToggleOptionConfig{
		Name: "#NOTE_EFFECT", Scope: scope, Default: true,
	}),
	LineBright: sonolus.SliderOption(sonolus.SliderOptionConfig{
		Name: "Long note line brightness", Scope: scope, Default: 80, Min: 10, Max: 100, Step: 10,
		Description: "Opacity of Long and Slide bodies (original live setting).",
	}),
	SuddenRate: sonolus.SliderOption(sonolus.SliderOptionConfig{
		Name: "Sudden", Scope: scope, Default: 0, Min: 0, Max: 100, Step: 1,
		Description: "Covers the upper part of the lanes (original live setting).",
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
	Replay: []string{"#NOTE_SPEED", "#NOTE_SIZE", "#MIRROR", "Full-length song"},
}

//sonolus:level
//go:embed dev-level.json
var DevelopmentLevel sonolus.LevelFile

func main() {}

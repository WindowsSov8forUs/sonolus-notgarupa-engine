//go:build play

package main

import (
	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus/play"
)

// PlayInitialization sets up layout, HUD model, life and UI (entity 0).
type PlayInitialization struct {
	play.Archetype      `archetype:"name=Initialization"`
	play.CallbackOrders `archetype:"preprocess=-10"`
	NoteCount           float64                         `archetype:"imported,name=noteCount"`
	Habahiro            bool                            `archetype:"imported,name=habahiro"`
	LaneChangeBeat      float64                         `archetype:"imported,name=laneChangeBeat,default=-1"`
	FullLength          bool                            `archetype:"imported,name=fullLength"`
	IllustBackground    bool                            `archetype:"imported,name=illustBackground"`
	FirstScored         sonolus.EntityRef[PlayNoteBase] `archetype:"imported,name=firstScored"`
}

func (n *PlayInitialization) Preprocess() {
	laneChange := -1.0
	if n.LaneChangeBeat >= 0 {
		laneChange = play.Time.BeatToTime(n.LaneChangeBeat)
	}
	initLayout(play.Screen.Rect(), Config.NoteSpeed, Config.NoteSize, Config.Mirror, n.Habahiro, laneChange)
	initHoldSE()
	initFx()
	initHud(n.NoteCount)
	Hud.FullLength = n.FullLength
	Hud.IllustBackground = n.IllustBackground
	initStageLights()
	initCurves()
	play.Background.Set(backgroundQuad())
	// Sonolus' arcade score (Level Score starts at 0); the HUD computes the
	// same score (arcadeWeight)
	play.Score.SetBase(play.BaseScore{Perfect: arcadePerfect, Great: arcadeGreat, Good: arcadeGood})
	play.Score.SetConsecutive(sonolus.JudgmentGreat, play.ConsecutiveScore{Multiplier: arcadeBonus, Step: arcadeStep, Cap: arcadeCap})

	play.Life.SetInitial(1000)
	play.Life.SetMax(2000)
	// Sonolus life (results / failure) can only be configured statically:
	// Miss costs the original damage; a Long head timeout counts on the tail
	// only (2 × ½); Slide among nodes are ÷5 per entity. Bad shares Sonolus
	// Good and cannot cost life there; the HUD gauge is exact.
	miss := play.LifeValues{Miss: lifeDamage(gradeMiss, jtNormal)}
	play.Life.SetArchetype(play.ArchetypeID[PlayTapNote](), miss)
	play.Life.SetArchetype(play.ArchetypeID[PlayFlickNote](), miss)
	play.Life.SetArchetype(play.ArchetypeID[PlayDirectionalNote](), miss)
	play.Life.SetArchetype(play.ArchetypeID[PlayLongHead](), play.LifeValues{})
	play.Life.SetArchetype(play.ArchetypeID[PlayLongTail](), miss)
	play.Life.SetArchetype(play.ArchetypeID[PlaySlideHead](), miss)
	play.Life.SetArchetype(play.ArchetypeID[PlaySlideNode](), miss)

	// The menu button lies, transparent, on the original pause button,
	// which the HUD draws (drawPauseButton).
	play.UI.SetMenu(sonolus.RuntimeUILayout{
		Anchor: pauseCenter(), Pivot: sonolus.Vec2{X: 0.5, Y: 0.5},
		Size:  sonolus.Vec2{X: 68 * Layout.U, Y: 70 * Layout.U},
		Alpha: 0,
	})
	hidden := sonolus.RuntimeUILayout{Anchor: sonolus.Vec2{X: 0, Y: 0}, Pivot: sonolus.Vec2{X: 0.5, Y: 0.5}}
	play.UI.SetJudgment(hidden)
	play.UI.SetComboValue(hidden)
	play.UI.SetComboText(hidden)
	play.UI.SetPrimaryMetricBar(hidden)
	play.UI.SetPrimaryMetricValue(hidden)
	play.UI.SetSecondaryMetricBar(hidden)
	play.UI.SetSecondaryMetricValue(hidden)
}
func (*PlayInitialization) SpawnOrder() float64 { return 1e8 }
func (*PlayInitialization) ShouldSpawn() bool   { return false }

// PlayStage draws the field, lane effects and the HUD every frame.
type PlayStage struct {
	play.Archetype      `archetype:"name=Stage"`
	play.CallbackOrders `archetype:"preprocess=100,updateSequential=-200"`
}

// Preprocess schedules the #EFFECT_AUTO hold loops in scoring-chain order
// under the cue limit (the notes record them in their preprocess).
func (*PlayStage) Preprocess() {
	if !autoSFX() {
		return
	}
	var loops holdLoopSchedule
	unit := sonolus.EntityRef[PlayInitialization]{Index: 0}.Get().FirstScored
	for k := 0; k < 100000 && unit.Index > 0; k++ {
		n := unit.Get()
		if n.HoldUntil > n.Time {
			loops.start(n.Time, n.HoldUntil)
		}
		unit = n.ScoreNext
	}
	loops.finish()
}

func (*PlayStage) SpawnOrder() float64 { return -1e9 }
func (*PlayStage) ShouldSpawn() bool   { return true }
func (*PlayStage) UpdateSequential() {
	now := play.Time.Now()
	updateLaneReserve(now)
	updateStagePose(now)
	play.Background.Set(backgroundQuad())
	if Hud.SSStart <= 0 && Hud.Score >= Hud.RankThreshold[4] && Hud.Score > 0 {
		Hud.SSStart = now
	}
	updateCover(now, false)
}

func (*PlayStage) UpdateParallel() {
	now := play.Time.Now()
	drawStageLights(now, 2)
	drawSpeakers(now, 3)
	drawPsyllium(now, 3.2) // StageFront, above the speakers, under BgCover
	drawCover(now, 4)
	drawOpeningCovers(now)
	drawField(now, lineAlpha(now))
	drawLaneEffects(now, zLaneEffect)
	drawHud(now)
}

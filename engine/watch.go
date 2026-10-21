//go:build watch

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus/watch"
)

// Watch mode replays the chart either as the original Auto Live (all
// Perfect, "AUTO" judge text) or from a replay's #JUDGMENT / #ACCURACY.
// The HUD is precomputed along the time-ordered scoring chain in preprocess
// (one snapshot per scoring unit) and looked up at runtime, so seeking needs
// no state rewind.

// WatchCursor is the runtime position in the scoring chain.
type WatchCursor struct {
	sonolus.LevelMemoryResource
	Last     float64 // entity index of the last applied unit (0: none)
	Next     float64 // entity index of the next unit (0: none)
	LastTime float64
}

var Cursor = WatchCursor{}

type WatchInitialization struct {
	watch.Archetype      `archetype:"name=Initialization"`
	watch.CallbackOrders `archetype:"preprocess=-10"`
	NoteCount            float64                          `archetype:"imported,name=noteCount"`
	Habahiro             bool                             `archetype:"imported,name=habahiro"`
	LaneChangeBeat       float64                          `archetype:"imported,name=laneChangeBeat,default=-1"`
	FirstScored          sonolus.EntityRef[WatchNoteBase] `archetype:"imported,name=firstScored"`
}

func (n *WatchInitialization) Preprocess() {
	laneChange := -1.0
	if n.LaneChangeBeat >= 0 {
		laneChange = watch.Time.BeatToTime(n.LaneChangeBeat)
	}
	initLayout(watch.Screen.Rect(), Config.NoteSpeed, Config.NoteSize, Config.Mirror, n.Habahiro, laneChange)
	initHoldSE()
	initFx()
	initHud(n.NoteCount)
	watch.Background.Set(backgroundQuad())
	// Sonolus' arcade score (Level Score starts at 0); the HUD computes the
	// same score (arcadeWeight)
	watch.Score.SetBase(watch.BaseScore{Perfect: arcadePerfect, Great: arcadeGreat, Good: arcadeGood})
	watch.Score.SetConsecutive(sonolus.JudgmentGreat, watch.ConsecutiveScore{Multiplier: arcadeBonus, Step: arcadeStep, Cap: arcadeCap})
	initStageLights()
	initCurves()

	watch.Life.SetInitial(1000)
	watch.Life.SetMax(2000)
	miss := watch.LifeValues{Miss: lifeDamage(gradeMiss, jtNormal, Config.FullLength)}
	watch.Life.SetArchetype(watch.ArchetypeID[WatchTapNote](), miss)
	watch.Life.SetArchetype(watch.ArchetypeID[WatchFlickNote](), miss)
	watch.Life.SetArchetype(watch.ArchetypeID[WatchDirectionalNote](), miss)
	watch.Life.SetArchetype(watch.ArchetypeID[WatchLongHead](), watch.LifeValues{})
	watch.Life.SetArchetype(watch.ArchetypeID[WatchLongTail](), miss)
	watch.Life.SetArchetype(watch.ArchetypeID[WatchSlideHead](), miss)
	watch.Life.SetArchetype(watch.ArchetypeID[WatchSlideNode](), miss)

	// The menu button lies, transparent, on the original pause button,
	// which the HUD draws (drawPauseButton).
	watch.UI.SetMenu(sonolus.RuntimeUILayout{
		Anchor: pauseCenter(), Pivot: sonolus.Vec2{X: 0.5, Y: 0.5},
		Size:  sonolus.Vec2{X: 68 * Layout.U, Y: 70 * Layout.U},
		Alpha: 0,
	})
	hidden := sonolus.RuntimeUILayout{Anchor: sonolus.Vec2{X: 0, Y: 0}, Pivot: sonolus.Vec2{X: 0.5, Y: 0.5}}
	watch.UI.SetJudgment(hidden)
	watch.UI.SetComboValue(hidden)
	watch.UI.SetComboText(hidden)
	watch.UI.SetPrimaryMetricBar(hidden)
	watch.UI.SetPrimaryMetricValue(hidden)
	watch.UI.SetSecondaryMetricBar(hidden)
	watch.UI.SetSecondaryMetricValue(hidden)
	watch.UI.SetProgress(sonolus.RuntimeUILayout{
		Anchor: sonolus.Vec2{X: 0, Y: Layout.Screen.B + 0.05}, Pivot: sonolus.Vec2{X: 0.5, Y: 0},
		Size: sonolus.Vec2{X: 2 * Layout.SafeR * 0.9, Y: 0.15}, Alpha: 1, Background: true,
	})
}
func (*WatchInitialization) SpawnTime() float64   { return -1e9 }
func (*WatchInitialization) DespawnTime() float64 { return -1e9 }

func isAuto() bool { return !watch.Replay.IsReplay() }

// WatchStage precomputes the HUD and SE in preprocess and draws everything
// the original draws outside the notes.
type WatchStage struct {
	watch.Archetype      `archetype:"name=Stage"`
	watch.CallbackOrders `archetype:"preprocess=100,updateSequential=-200"`
}

func (*WatchStage) Preprocess() {
	first := sonolus.EntityRef[WatchInitialization]{Index: 0}.Get().FirstScored
	var loops holdLoopSchedule
	for k := 0; k < 100000 && first.Index > 0; k++ {
		n := first.Get()
		if Config.SFX && n.HoldUntil > n.JTime {
			loops.start(n.JTime, n.HoldUntil)
		}
		j := n.judgement()
		hudApply(n.JTime, j)
		n.SCombo, n.SComboStart, n.SScore, n.SLife = Hud.Combo, Hud.ComboStart, Hud.Score, Hud.Life
		n.SAllPerfect, n.SJudge, n.STiming, n.SJudgeStart = Hud.AllPerfect, Hud.JudgeSprite, Hud.JudgeTiming, Hud.JudgeStart
		n.SAddValue, n.SAddStart = Hud.LastAdded, n.JTime
		n.SGameOver = Hud.GameOver
		if Hud.SSStart <= 0 && Hud.Score >= Hud.RankThreshold[4] && Hud.Score > 0 {
			Hud.SSStart = n.JTime
		}
		clip := judgementClip(n.JTime, j)
		if Config.SFX {
			if skillSE(j) {
				Effects.Skill.PlayScheduled(n.JTime, frameSeconds)
			}
			if clip != clipNone {
				clipOf(clip).PlayScheduled(n.JTime, frameSeconds)
			}
		}
		first = n.ScoreNext
	}
	loops.finish()
	resetHudState()
	Cursor.Next = sonolus.EntityRef[WatchInitialization]{Index: 0}.Get().FirstScored.Index
}

// watchAddPop replays a scoring unit's AddScore pop into the pool as the
// cursor passes it (rebuilt from the start after a seek).
func watchAddPop(n *WatchNoteBase) {
	if n.SAddValue > 0 {
		addPop(n.SAddValue, n.SAddStart)
	}
}

// WatchGlobals carries the watch global callbacks.
type WatchGlobals struct{ watch.GlobalCallbacks }

var WatchGlobal WatchGlobals

// UpdateSpawn is the Lifetime system's spawning clock: an entity is alive
// while it lies in [SpawnTime, DespawnTime). Without it the clock stays 0 and
// no note ever spawns.
func UpdateSpawn() float64 { return watch.Time.Now() }

func (*WatchStage) SpawnTime() float64   { return -1e9 }
func (*WatchStage) DespawnTime() float64 { return 1e9 }

func (*WatchStage) UpdateSequential() {
	now := watch.Time.Now()
	updateStagePose(now)
	watch.Background.Set(backgroundQuad())
	skip := watch.Time.Skip() || now < Cursor.LastTime
	if skip {
		Cursor.Last = 0
		Cursor.Next = sonolus.EntityRef[WatchInitialization]{Index: 0}.Get().FirstScored.Index
		for i := 0; i < 7; i++ {
			laneOff(float64(i))
		}
		clearAddPops()
	}
	// as Play and the original: the OffReserve countdown precedes this
	// update's lane events, so a pulse shows one full-bright update
	updateLaneReserve(now)
	for k := 0; k < 256 && Cursor.Next > 0; k++ {
		n := sonolus.EntityRef[WatchNoteBase]{Index: Cursor.Next}.Get()
		if n.JTime > now {
			break
		}
		watchAddPop(n)
		if !skip {
			j := n.judgement()
			Hud.GameOver = n.SGameOver
			if Config.NoteEffects {
				playJudgementParticles(j)
			}
			if j.Auto && n.ALaneOff >= 0 {
				laneOff(n.ALaneOff)
			}
			judgementLights(now, j)
			if !j.Auto && n.JFront && j.Grade != gradeMiss {
				// a replay has no touches: the touched lane lights at the
				// front judgement (Play lights it on Began, fades on Ended)
				laneOnReserve(j.Lane, now)
			}
			if j.Auto && n.ALanePulse >= 0 {
				laneOnReserve(n.ALanePulse, now)
			}
		}
		Cursor.Last = Cursor.Next
		Cursor.Next = n.ScoreNext.Index
	}
	if skip {
		for k := 0; k < 100000 && Cursor.Next > 0; k++ {
			n := sonolus.EntityRef[WatchNoteBase]{Index: Cursor.Next}.Get()
			if n.JTime > now {
				break
			}
			watchAddPop(n)
			Cursor.Last = Cursor.Next
			Cursor.Next = n.ScoreNext.Index
		}
	}
	Cursor.LastTime = now
	if Cursor.Last > 0 {
		n := sonolus.EntityRef[WatchNoteBase]{Index: Cursor.Last}.Get()
		Hud.Combo, Hud.ComboStart, Hud.Score, Hud.Life = n.SCombo, n.SComboStart, n.SScore, n.SLife
		Hud.AllPerfect, Hud.JudgeSprite, Hud.JudgeTiming, Hud.JudgeStart = n.SAllPerfect, n.SJudge, n.STiming, n.SJudgeStart
		Hud.GameOver = n.SGameOver
	} else {
		Hud.Combo, Hud.Score, Hud.Life, Hud.AllPerfect, Hud.JudgeSprite = 0, 0, 1000, true, -1
		Hud.GameOver = false
	}
	updateCover(now, skip)
}

func (*WatchStage) UpdateParallel() {
	now := watch.Time.Now()
	drawStageLights(now, 2)
	drawSpeakers(now, 3)
	drawCover(now, 4)
	drawOpeningCovers(now)
	drawField(now, lineAlpha(now))
	drawLaneEffects(now, zLaneEffect)
	drawHud(now)
	if isAuto() {
		drawAutoCaption(zHUD + 2)
	}
}

type WatchInputManager struct {
	watch.Archetype `archetype:"name=InputManager"`
}

func (*WatchInputManager) SpawnTime() float64   { return -1e9 }
func (*WatchInputManager) DespawnTime() float64 { return -1e9 }

// ---------------------------------------------------------------- timing ----

type WatchBPMChange struct {
	watch.Archetype `archetype:"name=#BPM_CHANGE"`
	Beat            float64 `archetype:"imported,name=#BEAT"`
	BPM             float64 `archetype:"imported,name=#BPM"`
}

func (*WatchBPMChange) SpawnTime() float64   { return -1e9 }
func (*WatchBPMChange) DespawnTime() float64 { return -1e9 }

type WatchTimingChange struct {
	watch.Archetype `archetype:"name=TimingChange"`
	Time            float64 `archetype:"imported,name=time"`
	Speed           float64 `archetype:"imported,name=speed"`
	Axis            float64 `archetype:"imported,name=axis"`
	Max             float64 `archetype:"imported,name=max"`
}

func (*WatchTimingChange) SpawnTime() float64   { return -1e9 }
func (*WatchTimingChange) DespawnTime() float64 { return -1e9 }

type WatchTimingGroup struct {
	watch.Archetype      `archetype:"name=TimingGroup"`
	watch.CallbackOrders `archetype:"updateSequential=-50"`
	First                sonolus.EntityRef[WatchTimingChange] `archetype:"imported,name=first"`
	Count                float64                              `archetype:"imported,name=count"`
	Cursor               float64                              `archetype:"shared"`
	AxisNow              float64                              `archetype:"shared"`
}

func (*WatchTimingGroup) SpawnTime() float64   { return -1e9 }
func (*WatchTimingGroup) DespawnTime() float64 { return 1e9 }

func (g *WatchTimingGroup) change(k float64) *WatchTimingChange {
	return sonolus.EntityRef[WatchTimingChange]{Index: g.First.Index + k}.Get()
}

func (g *WatchTimingGroup) UpdateSequential() {
	now := watch.Time.Now()
	c := g.Cursor
	for c > 0 && now < g.change(c).Time {
		c--
	}
	for c+1 < g.Count && g.change(c+1).Time <= now {
		c++
	}
	g.Cursor = c
	first := g.change(0)
	if now < first.Time {
		g.AxisNow = now - first.Time + first.Axis
	} else {
		ch := g.change(c)
		g.AxisNow = ch.Axis + ch.Speed*(now-ch.Time)
	}
}

func watchVisualSpawn(group sonolus.EntityRef[WatchTimingGroup], axisHit float64) float64 {
	g := group.Get()
	x := axisHit - Layout.Arrival
	first := g.change(0)
	if x <= first.Axis {
		return x - first.Axis + first.Time
	}
	lo, hi := 1.0, g.Count
	for lo < hi {
		mid := math.Floor((lo + hi) / 2)
		if g.change(mid).Max >= x {
			hi = mid
		} else {
			lo = mid + 1
		}
	}
	if lo < g.Count {
		p := g.change(lo - 1)
		return p.Time + (x-p.Axis)/p.Speed
	}
	last := g.change(g.Count - 1)
	if last.Speed > 0 {
		return last.Time + (x-last.Axis)/last.Speed
	}
	return 1e8
}

func watchSpawnTimeOf(beat float64, group sonolus.EntityRef[WatchTimingGroup], axisHit float64) float64 {
	t := watch.Time.BeatToTime(beat)
	visual := t - Layout.Arrival
	if group.Index > 0 {
		visual = watchVisualSpawn(group, axisHit)
	}
	return math.Min(visual, t-0.25)
}

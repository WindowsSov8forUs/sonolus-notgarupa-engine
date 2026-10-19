//go:build play

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus/play"
)

type candidate struct {
	Index, Start, Width, Time, Y float64
	Slide, Owned                 bool
}

type InputState struct {
	sonolus.LevelMemoryResource
	Candidates sonolus.VarArray[candidate]
	TouchLane  sonolus.ArrayMap[float64, float64]
	// Live hold sounds (main and intro handle IDs, 0 = free) and their
	// start order.
	Loops     [holdSlots]float64
	LoopIntro [holdSlots]float64
	LoopOrder [holdSlots]float64
	LoopSeq   float64
}

var Input = InputState{
	Candidates: sonolus.NewVarArray[candidate](64),
	TouchLane:  sonolus.NewArrayMap[float64, float64](16),
}

// PlayInputManager performs the original touch → button → note arbitration
// (ButtonManager.GetPlayButton + NoteManager.GetMoveEndTimeNearestZeroNote).
type PlayInputManager struct {
	play.Archetype      `archetype:"name=InputManager"`
	play.CallbackOrders `archetype:"updateSequential=-100,touch=-100"`
}

func (*PlayInputManager) SpawnOrder() float64 { return -1e8 }
func (*PlayInputManager) ShouldSpawn() bool   { return true }

func (*PlayInputManager) UpdateSequential() {
	Input.Candidates.Clear()
}

// candidateGrade grades a touch-down for a candidate (None if out of window).
func candidateGrade(c candidate, t float64) float64 {
	if c.Slide {
		return slideGrade(slideTau(c.Time, t))
	}
	return timeGrade(c.Time-t, 0)
}

// bestFor selects the candidate a touch would hit on lane l. Projected
// (non-lane) notes compete in the same pass when their span circle contains
// the touch; l = −1 is the pass for a touch outside every lane circle.
func bestFor(l float64, p sonolus.Vec2, t float64) int {
	ordinary, slide := -1, -1
	bestDist := 0.0
	for i, c := range Input.Candidates.Items() {
		projected := c.Start != math.Floor(c.Start) || c.Start < 0 || c.Start+c.Width > 7
		inside := projected && insideSpan(p, c.Start, c.Width, false)
		if l >= 0 && !projected {
			inside = l >= c.Start && l <= c.Start+c.Width-1
		}
		if !inside {
			continue
		}
		if !c.Slide {
			d := math.Abs(c.Time - t)
			if ordinary < 0 || d < bestDist {
				ordinary, bestDist = i, d
			}
			continue
		}
		if slide < 0 {
			if !c.Owned {
				slide = i
			}
			continue
		}
		if math.Abs(c.Y-Layout.G) < math.Abs(Input.Candidates.Get(slide).Y-Layout.G) {
			slide = i
		}
	}
	pick := ordinary
	if slide >= 0 && (ordinary < 0 || math.Abs(Input.Candidates.Get(slide).Y-Layout.G) < math.Abs(laneYForCandidate(ordinary)-Layout.G)) {
		pick = slide
	}
	if pick < 0 {
		return -1
	}
	if candidateGrade(Input.Candidates.Get(pick), t) == gradeNone {
		return -1
	}
	return pick
}

func laneYForCandidate(i int) float64 { return Input.Candidates.Get(i).Y }

func (*PlayInputManager) Touch() {
	now := play.Time.Now()
	order := 0
	for touch := range play.Touches.Values() {
		order++
		if touch.Ended {
			lane, ok := Input.TouchLane.GetOK(touch.ID)
			if ok {
				laneAnimatedOff(lane, now)
				Input.TouchLane.Delete(touch.ID)
			}
		}
		if !touch.Started {
			continue
		}
		p := touch.Position
		// nearest, second and third nearest lane inside the collision circle
		var near [3]float64
		count := 0
		for k := 0; k < 3; k++ {
			best, bestD := -1.0, Layout.Collision2
			for l := 0.0; l < 7; l++ {
				if (count > 0 && near[0] == l) || (count > 1 && near[1] == l) {
					continue
				}
				dx, dy := p.X-laneX(l), p.Y-Layout.G
				if d := dx*dx + dy*dy; d < bestD {
					best, bestD = l, d
				}
			}
			if best < 0 {
				break
			}
			near[k] = best
			count++
		}
		projected := bestFor(-1, p, touch.StartTime)
		chosen, lane := -1, -1.0
		for k := 0; k < count; k++ {
			if k == 2 && (near[0] == 0 || near[0] == 6) {
				break
			}
			c := bestFor(near[k], p, touch.StartTime)
			if c >= 0 {
				chosen, lane = c, near[k]
				break
			}
		}
		if chosen < 0 && projected >= 0 {
			chosen = projected
			cd := Input.Candidates.Get(chosen)
			lane = math.Round(receiverLane(cd.Start, cd.Width))
		}
		if count == 0 && chosen < 0 {
			continue // outside every lane circle: ignored entirely
		}
		if lane < 0 {
			lane = near[0]
		}
		laneOn(lane)
		if touch.Ended {
			laneAnimatedOff(lane, now) // began and ended in one frame: Began, then Ended
		} else if !Input.TouchLane.IsFull() {
			Input.TouchLane.Set(touch.ID, lane)
		}
		if chosen < 0 {
			// empty tap
			if Config.SFX && order <= 4 {
				Effects.EmptyTap.Play(0)
			}
			if Config.NoteEffects {
				spawnFx(fxTap, laneX(lane), durTap, durTap, false)
			}
			continue
		}
		c := Input.Candidates.Get(chosen)
		note := sonolus.EntityRef[PlayNoteBase]{Index: c.Index}.Get()
		if note.Finger >= 0 || note.Assigned >= 0 {
			continue // already owned: the touch is consumed
		}
		grade := candidateGrade(c, touch.StartTime)
		note.Assigned = touch.ID
		note.AGrade = grade
		if c.Slide {
			note.ATiming = timingNone
			if grade != gradePerfect && slideTau(c.Time, touch.StartTime) > 0 {
				note.ATiming = timingFast
			}
		} else {
			note.ATiming = timeTiming(grade, c.Time-touch.StartTime)
		}
		note.ATime = touch.StartTime
		note.AX, note.AY = p.X, p.Y
	}
}

//go:build play || watch || tutorial

package main

import (
	"math"

	"github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"
)

// LaneFxState is the tap lane beams (GamePlayButton lane effect): per lane
// 1 lit, 2 fading, 0 off, the fade start and the OffReserve counter.
type LaneFxState struct {
	sonolus.LevelMemoryResource
	On      [7]float64
	Start   [7]float64
	Reserve [7]float64
}

var LaneFx = LaneFxState{}

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
	LaneFx.On[i] = 1
}

// laneOnReserve lights a lane for one update, then fades (On + OffReserve);
// events apply before the countdown of the same update (tapLaneEffectOwner).
func laneOnReserve(lane, now float64) {
	i := laneIndex(lane)
	if i < 0 {
		return
	}
	LaneFx.On[i] = 1
	LaneFx.Reserve[i] = 1
}

// laneAnimatedOff starts the 1/6 s fade (restarts it while fading); it does
// nothing on a switched-off lane.
func laneAnimatedOff(lane, now float64) {
	i := laneIndex(lane)
	if i < 0 || LaneFx.On[i] == 0 {
		return
	}
	LaneFx.On[i] = 2
	// Animator.Play(…, 0) and the animator advances the same update
	LaneFx.Start[i] = now - frameSeconds
}

// laneOff switches a lane off immediately.
func laneOff(lane float64) {
	i := laneIndex(lane)
	if i < 0 {
		return
	}
	LaneFx.On[i] = 0
}

// updateLaneReserve counts the OffReserve counters down once per update.
func updateLaneReserve(now float64) {
	for i := 0; i < 7; i++ {
		if LaneFx.Reserve[i] > 0 {
			LaneFx.Reserve[i]--
			if LaneFx.Reserve[i] == 0 && LaneFx.On[i] != 0 {
				// the fade started with the previous update's countdown
				LaneFx.On[i] = 2
				LaneFx.Start[i] = now - frameSeconds
			}
		}
	}
}

// Lane tap effect (NoteLaneEffect_1..4): idle, or a 1/6 s fade to scale 0.7.
var laneEffectSprite = [7]int{0, 1, 2, 3, 2, 1, 0}
var laneEffectSize = [4][3]float64{{467, 500, 0.1711825579404831}, {342, 500, 0.23051762580871582}, {218, 500, 0.3616914451122284}, {154, 500, 0.5}}

func drawLaneEffects(now, z float64) {
	if !Config.LaneEffect || Layout.Habahiro {
		return // HABAHIRO lives have no lane tap beams
	}
	for lane := 0; lane < 7; lane++ {
		state := LaneFx.On[lane]
		if state == 0 {
			continue
		}
		scale, a := 1.0, 1.0
		if state == 2 {
			t := (now - LaneFx.Start[lane]) / 0.1666666716337204
			if t >= 1 {
				continue
			}
			scale = 1 - 0.3*t
			a = 1 - t
		}
		idx := laneEffectSprite[lane]
		size := laneEffectSize[idx]
		k := Layout.Rate * Layout.M / laneEffectPPU * scale
		w, h, px := size[0]*k, size[1]*k, size[2]
		x := laneX(float64(lane))
		l := x - px*w
		r := x + (1-px)*w
		if lane >= 4 { // flipX
			l, r = x+px*w, x-(1-px)*w
		}
		q := sonolus.Quad{
			BL: sonolus.Vec2{X: l, Y: Layout.G}, TL: sonolus.Vec2{X: l, Y: Layout.G + h},
			TR: sonolus.Vec2{X: r, Y: Layout.G + h}, BR: sonolus.Vec2{X: r, Y: Layout.G},
		}
		Skin.LaneEffect[idx].DrawZ(q, sonolus.ZIndex{Z1: z, Z2: float64(lane)}, a) // equal depth: slot order
	}
}

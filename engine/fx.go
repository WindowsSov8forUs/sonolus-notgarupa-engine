package main

import "github.com/WindowsSov8forUs/sonolus-go/v2/sonolus"

// Judgement particle families. Each family is a view part (BaseOpt) and
// lane-plane parts at the tilts the tap effect skins use (PlaneOpt, squashed
// by cos(tilt) on the quad's y); a particle item has some of them, at most
// two plane parts per family. initFx records which exist.
const (
	fxTap = iota
	fxGood
	fxGreat
	fxPerfect
	fxSkillGood
	fxSkillGreat
	fxSkillPerfect
	fxSwipe
	fxKeep
	fxFingerL
	fxFingerR
	fxFamilies
)

// fxPlanes is the number of plane tilts: −65, 44, 46, 48, 50, 52, 53, 55,
// 58, 60, 62, 64 degrees (PlaneOpt order); fxPlaneCos are their cosines.
const fxPlanes = 12

var fxPlaneCos = [fxPlanes]float64{0.422618, 0.71934, 0.694658, 0.669131, 0.642788, 0.615661, 0.601815, 0.573576, 0.529919, 0.5, 0.469472, 0.438371}

type FxData struct {
	sonolus.LevelDataResource
	Base [fxFamilies]bool
	// PlaneA / PlaneB: the family's plane parts (PlaneOpt index within the
	// family), −1 when absent.
	PlaneA [fxFamilies]float64
	PlaneB [fxFamilies]float64
}

var Fx = FxData{}

func initFx() {
	for f := 0; f < fxFamilies; f++ {
		Fx.Base[f] = Particles.BaseOpt[f].Exists()
		Fx.PlaneA[f], Fx.PlaneB[f] = -1, -1
		for k := 0; k < fxPlanes; k++ {
			if !Particles.PlaneOpt[f*fxPlanes+k].Exists() {
				continue
			}
			if Fx.PlaneA[f] < 0 {
				Fx.PlaneA[f] = float64(k)
			} else if Fx.PlaneB[f] < 0 {
				Fx.PlaneB[f] = float64(k)
			}
		}
	}
}

// spawnFx spawns family f's parts centred on x (plane parts with their own
// duration, which differs for looping TapKeep); the handles of absent parts
// stay 0.
func spawnFx(f int, x, dur, planeDur float64, loop bool) (sonolus.ParticleHandle, sonolus.ParticleHandle, sonolus.ParticleHandle) {
	var b, p, q sonolus.ParticleHandle
	if Fx.Base[f] {
		b = Particles.BaseOpt[f].Spawn(particleQuad(x, 1), dur, loop)
	}
	if k := int(Fx.PlaneA[f]); k >= 0 {
		p = Particles.PlaneOpt[f*fxPlanes+k].Spawn(particleQuad(x, fxPlaneCos[k]), planeDur, loop)
	}
	if k := int(Fx.PlaneB[f]); k >= 0 {
		q = Particles.PlaneOpt[f*fxPlanes+k].Spawn(particleQuad(x, fxPlaneCos[k]), planeDur, loop)
	}
	return b, p, q
}

// moveFx moves family f's live parts to x.
func moveFx(f int, x float64, b, p, q sonolus.ParticleHandle) {
	if b.ID > 0 {
		b.Move(particleQuad(x, 1))
	}
	if k := int(Fx.PlaneA[f]); k >= 0 && p.ID > 0 {
		p.Move(particleQuad(x, fxPlaneCos[k]))
	}
	if k := int(Fx.PlaneB[f]); k >= 0 && q.ID > 0 {
		q.Move(particleQuad(x, fxPlaneCos[k]))
	}
}

// destroyFx destroys the live parts.
func destroyFx(b, p, q sonolus.ParticleHandle) {
	if b.ID > 0 {
		b.Destroy()
	}
	if p.ID > 0 {
		p.Destroy()
	}
	if q.ID > 0 {
		q.Destroy()
	}
}

// fxAlive reports whether any part is live.
func fxAlive(b, p, q sonolus.ParticleHandle) bool { return b.ID > 0 || p.ID > 0 || q.ID > 0 }

// spawnKeep spawns the TapKeep of a Long of width w: the HABAHIRO width
// variant (view part + its 48° plane) or the family. Slides always use the
// family (spawnFx / moveFx).
func spawnKeep(w int, x float64) (sonolus.ParticleHandle, sonolus.ParticleHandle, sonolus.ParticleHandle) {
	if Layout.HabaFx && w >= 2 {
		var none sonolus.ParticleHandle
		b := Particles.KeepW[w-2].Spawn(particleQuad(x, 1), durKeep, true)
		p := Particles.KeepPlaneW[w-2].Spawn(particleQuad(x, planeCos48), durKeepPlane, true)
		return b, p, none
	}
	return spawnFx(fxKeep, x, durKeep, durKeepPlane, true)
}

package material

import (
	"time"

	"dasa.cc/snd"
	"golang.org/x/mobile/exp/f32"
)

var (
	ExpSig, LinSig snd.Discrete
)

func init() {
	ExpSig = snd.ExpDecay()
	ExpSig.UnitInverse()
	LinSig = snd.LinearDecay()
	LinSig.UnitInverse()
}

type Interpolator struct {
	Sig  snd.Discrete
	Dur  time.Duration
	Loop bool
}

func Animate(mat *f32.Mat4, interp Interpolator, fn func(m *f32.Mat4, dt float32)) (quit chan struct{}) {
	m := *mat // copy; translate is always relative to resting position
	quit = make(chan struct{}, 1)
	go func() {
		ticker := time.NewTicker(16 * time.Millisecond)
		start := time.Now()
		for {
			select {
			case <-quit:
				ticker.Stop()
				return
			case now := <-ticker.C:
				since := now.Sub(start)
				t := float64(since%interp.Dur) / float64(interp.Dur)
				if !interp.Loop && since >= interp.Dur {
					quit <- struct{}{}
					t = 1
				}
				dt := float32(interp.Sig.SampleUnit(t))
				fn(&m, dt)
			}
		}
	}()
	return quit
}

func AnimateRotate(angle f32.Radian, axis f32.Vec3, mat *f32.Mat4, interp Interpolator) (quit chan struct{}) {
	return Animate(mat, interp, func(m *f32.Mat4, dt float32) {
		mat.Rotate(m, f32.Radian(dt*float32(angle)), &axis)
	})
}

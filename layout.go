package material

import (
	"dasa.cc/simplex"
	"golang.org/x/mobile/exp/f32"
)

type Box struct {
	l, r, b, t, z simplex.Var
	world         f32.Mat4
}

func NewBox(prg *simplex.Program) (a Box) {
	a.l, a.r, a.b, a.t, a.z = prg.Var(1), prg.Var(1), prg.Var(1), prg.Var(1), prg.Var(1)
	return
}

func (a Box) Width(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.r}, simplex.Coef{-1, a.l}).Equal(float64(x))
}

func (a Box) Height(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.t}, simplex.Coef{-1, a.b}).Equal(float64(x))
}

func (a Box) Start(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.l}).Equal(float64(x))
}

func (a Box) End(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.r}).Equal(float64(x))
}

func (a Box) Bottom(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.b}).Equal(float64(x))
}

func (a Box) Top(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.t}).Equal(float64(x))
}

func (a Box) Z(z float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.z}).Equal(float64(z))
}

func (a Box) StartIn(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.l}, simplex.Coef{-1, b.l}).GreaterEq(float64(by))
}

func (a Box) EndIn(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.r}, simplex.Coef{-1, a.r}).GreaterEq(float64(by))
}

func (a Box) BottomIn(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.b}, simplex.Coef{-1, b.b}).GreaterEq(float64(by))
}

func (a Box) TopIn(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.t}, simplex.Coef{-1, a.t}).GreaterEq(float64(by))
}

func (a Box) CenterVerticalIn(b Box) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.b}, simplex.Coef{1, b.t}, simplex.Coef{-1, a.b}, simplex.Coef{-1, a.t})
}

func (a Box) CenterHorizontalIn(b Box) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.l}, simplex.Coef{1, b.r}, simplex.Coef{-1, a.l}, simplex.Coef{-1, a.r})
}

func (a Box) Before(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.l}, simplex.Coef{-1, a.r}).GreaterEq(float64(by))
}

func (a Box) After(b Box, by float32) simplex.Constraint {
	// TODO this is the crux of adaptive layout model, along with a Before method.
	// Consider how box a would be after box b if room, otherwise box a is below box b.
	// Note in the latter case, box a should not be aligned after box b when below.
	return simplex.Constrain(simplex.Coef{1, a.l}, simplex.Coef{-1, b.r}).GreaterEq(float64(by))
}

func (a Box) Below(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.b}, simplex.Coef{-1, a.t}).GreaterEq(float64(by))
}

func (a Box) Above(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.b}, simplex.Coef{-1, b.t}).GreaterEq(float64(by))
}

func (a Box) AlignBottoms(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.b}, simplex.Coef{-1, a.b}).GreaterEq(float64(by))
}

func (a Box) AlignTops(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.t}, simplex.Coef{-1, a.t}).GreaterEq(float64(by))
}

func (a Box) Bounds(l, r, b, t float32) []simplex.Constraint {
	return []simplex.Constraint{
		simplex.Constrain(simplex.Coef{1, a.l}).GreaterEq(float64(l)),
		simplex.Constrain(simplex.Coef{1, a.r}).LessEq(float64(r)),
		simplex.Constrain(simplex.Coef{1, a.b}).GreaterEq(float64(b)),
		simplex.Constrain(simplex.Coef{1, a.t}).LessEq(float64(t)),
	}
}

func (a *Box) UpdateWorld(prg *simplex.Program) {
	prg.For(&a.l, &a.r, &a.b, &a.t, &a.z)
	a.world.Identity()
	a.world.Translate(&a.world, float32(a.l.Val), float32(a.b.Val), 0)
	a.world.Scale(&a.world, float32(a.r.Val-a.l.Val), float32(a.t.Val-a.b.Val), 1)
	a.world[2][3] = float32(a.z.Val)
}

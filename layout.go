package material

import (
	"dasa.cc/material/simplex"
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
	return simplex.Constrain(simplex.Coef{1, a.r}, simplex.Coef{-1, a.l}).Equal(x)
}

func (a Box) Height(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.t}, simplex.Coef{-1, a.b}).Equal(x)
}

func (a Box) Start(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.l}).Equal(x)
}

func (a Box) End(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.r}).Equal(x)
}

func (a Box) Bottom(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.b}).Equal(x)
}

func (a Box) Top(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.t}).Equal(x)
}

func (a Box) Z(z float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.z}).Equal(z)
}

func (a Box) StartIn(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.l}, simplex.Coef{-1, b.l}).GreaterEq(by)
}

func (a Box) EndIn(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.r}, simplex.Coef{-1, a.r}).GreaterEq(by)
}

func (a Box) BottomIn(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.b}, simplex.Coef{-1, b.b}).GreaterEq(by)
}

func (a Box) TopIn(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.t}, simplex.Coef{-1, a.t}).GreaterEq(by)
}

func (a Box) CenterVerticalIn(b Box) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.b}, simplex.Coef{1, b.t}, simplex.Coef{-1, a.b}, simplex.Coef{-1, a.t})
}

func (a Box) CenterHorizontalIn(b Box) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.l}, simplex.Coef{1, b.r}, simplex.Coef{-1, a.l}, simplex.Coef{-1, a.r})
}

func (a Box) Before(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.l}, simplex.Coef{-1, a.r}).GreaterEq(by)
}

func (a Box) After(b Box, by float32) simplex.Constraint {
	// TODO this is the crux of adaptive layout model, along with a Before method.
	// Consider how box a would be after box b if room, otherwise box a is below box b.
	// Note in the latter case, box a should not be aligned after box b when below.
	return simplex.Constrain(simplex.Coef{1, a.l}, simplex.Coef{-1, b.r}).GreaterEq(by)
}

func (a Box) Below(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.b}, simplex.Coef{-1, a.t}).GreaterEq(by)
}

func (a Box) Above(b Box, by float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.b}, simplex.Coef{-1, b.t}).GreaterEq(by)
}

func (a Box) Bounds(l, r, b, t float32) []simplex.Constraint {
	return []simplex.Constraint{
		simplex.Constrain(simplex.Coef{1, a.l}).GreaterEq(l),
		simplex.Constrain(simplex.Coef{1, a.r}).LessEq(r),
		simplex.Constrain(simplex.Coef{1, a.b}).GreaterEq(b),
		simplex.Constrain(simplex.Coef{1, a.t}).LessEq(t),
	}
}

func (a *Box) UpdateWorld(prg *simplex.Program) {
	prg.For(&a.l, &a.r, &a.b, &a.t, &a.z)
	a.world.Identity()
	a.world.Translate(&a.world, a.l.Val, a.b.Val, 0)
	a.world.Scale(&a.world, a.r.Val-a.l.Val, a.t.Val-a.b.Val, 1)
	a.world[2][3] = a.z.Val
}

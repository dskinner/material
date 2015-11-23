package material

import (
	"dasa.cc/material/simplex"
	"golang.org/x/mobile/exp/f32"
)

type Box struct {
	l, r, b, t simplex.Var
	world      f32.Mat4
}

func NewBox(prg *simplex.Program) (a Box) {
	a.l, a.r, a.b, a.t = prg.Var(1), prg.Var(1), prg.Var(1), prg.Var(1)
	return
}

func (a Box) Bounds(l, r, b, t float32) []simplex.Constraint {
	return []simplex.Constraint{
		simplex.Constrain(simplex.Coef{1, a.l}).GreaterEq(l),
		simplex.Constrain(simplex.Coef{1, a.r}).LessEq(r),
		simplex.Constrain(simplex.Coef{1, a.b}).GreaterEq(b),
		simplex.Constrain(simplex.Coef{1, a.t}).LessEq(t),
	}
}

func (a Box) Inside(b *Box, ml, mr, mb, mt float32) []simplex.Constraint {
	return []simplex.Constraint{
		simplex.Constrain(simplex.Coef{1, a.l}, simplex.Coef{-1, b.l}).GreaterEq(ml),
		// simplex.Constrain(simplex.Coef{1, b.r}, simplex.Coef{-1, a.r}).GreaterEq(mr),
		simplex.Constrain(simplex.Coef{1, a.r}, simplex.Coef{-1, b.r}).LessEq(-mr),
		simplex.Constrain(simplex.Coef{1, a.b}, simplex.Coef{-1, b.b}).GreaterEq(mb),
		// simplex.Constrain(simplex.Coef{1, b.t}, simplex.Coef{-1, a.t}).GreaterEq(mt),
		simplex.Constrain(simplex.Coef{1, a.t}, simplex.Coef{-1, b.t}).LessEq(-mt),
	}
}

func (a Box) Width(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.r}, simplex.Coef{-1, a.l}).Equal(x)
}

func (a Box) Height(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.t}, simplex.Coef{-1, a.b}).Equal(x)
}

func (a Box) Left(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.l}).Equal(x)
}

func (a Box) Right(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.r}).Equal(x)
}

func (a Box) Bottom(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.b}).Equal(x)
}

func (a Box) Top(x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.t}).Equal(x)
}

func (a Box) LeftOf(b Box, x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, b.l}, simplex.Coef{-1, a.r}).GreaterEq(x)
}

func (a Box) RightOf(b Box, x float32) simplex.Constraint {
	return simplex.Constrain(simplex.Coef{1, a.l}, simplex.Coef{-1, b.r}).LessEq(x)
}

func (a *Box) UpdateWorld(prg *simplex.Program) {
	prg.For(&a.l, &a.r, &a.b, &a.t)
	a.world.Identity()
	a.world.Translate(&a.world, a.l.Val, a.b.Val, 0)
	a.world.Scale(&a.world, a.r.Val-a.l.Val, a.t.Val-a.b.Val, 1)
}

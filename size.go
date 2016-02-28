package material

import (
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/exp/f32"
)

var windowSize size.Event

// Uton converts unit to norm.
func Uton(u float32) float32 { return 2*u - 1 }

// Ntou converts norm to unit.
func Ntou(n float32) float32 { return (n + 1) / 2 }

// Mtoa converts mat4 to affine.
// TODO this only exists since I'm doing 2D atm
// and mat4 doesn't have inverse method.
func Mtoa(m f32.Mat4) (a f32.Affine) {
	a[0][0] = m[0][0]
	a[0][1] = m[0][1]
	a[0][2] = m[0][3]
	a[1][0] = m[1][0]
	a[1][1] = m[1][1]
	a[1][2] = m[1][3]
	return
}

func ScreenToUnit(x, y float32) (float32, float32) {
	// TODO determine y direction, don't assume
	return x / float32(windowSize.WidthPx), 1 - (y / float32(windowSize.HeightPx))
}

func ScreenToNorm(x, y float32) (float32, float32) {
	x, y = ScreenToUnit(x, y)
	return Uton(x), Uton(y)
}

func ScreenToWorld(x, y, z float32, view, proj f32.Mat4) (float32, float32) {
	x, y = ScreenToNorm(x, y)
	return NormToWorld(x, y, z, view, proj)
}

func UnitToWorld(x, y, z float32, view, proj f32.Mat4) (float32, float32) {
	wx, wy := NormToWorld(Uton(x), Uton(y), Uton(z), view, proj)
	return 1 + wx, -wy
}

func NormToWorld(x, y, z float32, view, proj f32.Mat4) (float32, float32) {
	nv := f32.Vec3{x, y, z}

	unproj := Mtoa(proj)
	unproj.Inverse(&unproj)
	unview := Mtoa(view)
	unview.Inverse(&unview)

	nv[0], nv[1] = nv.Dot(&unproj[0]), nv.Dot(&unproj[1])
	nv[0], nv[1] = nv.Dot(&unview[0]), nv.Dot(&unview[1])

	return nv[0], nv[1]
}

func NormToView(x, y, z float32, proj f32.Mat4) (float32, float32) {
	nv := f32.Vec3{x, y, z}
	unproj := Mtoa(proj)
	unproj.Inverse(&unproj)
	nv[0], nv[1] = nv.Dot(&unproj[0]), nv.Dot(&unproj[1])
	return nv[0], nv[1]
}

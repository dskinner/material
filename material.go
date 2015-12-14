package material

// TODO fonts
// https://www.mapbox.com/blog/text-signed-distance-fields/

import (
	"dasa.cc/material/glutil"
	"dasa.cc/material/icon"
	"dasa.cc/material/simplex"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

var (
	DefaultFilter = glutil.TextureFilter(gl.LINEAR, gl.LINEAR)
	DefaultWrap   = glutil.TextureWrap(gl.REPEAT, gl.REPEAT)
)

type Material struct {
	Box

	col4, col8, col12 int

	BehaviorFlags Behavior

	Texture glutil.Texture
	uvbuf   glutil.FloatBuffer
	uicon   gl.Uniform

	icx, icy float32

	cr, cg, cb, ca float32 // color for uniform

	vbuf glutil.FloatBuffer // vertices
	ibuf glutil.UintBuffer  // indices

	prg0, prg1    glutil.Program // material and shadow TODO globalize with batch op?
	ap0, ap1      gl.Attrib      // buffer pointer
	uc0, uc1      gl.Uniform     // color
	uw0, uv0, up0 gl.Uniform     // material projection
	uw1, uv1, up1 gl.Uniform     // shadow projection

	utex0 gl.Uniform
	atc0  gl.Attrib

	// TODO tmp impl
	IsCircle bool
	ucirc    gl.Uniform
}

func (mtrl *Material) Span(col4, col8, col12 int) {
	mtrl.col4, mtrl.col8, mtrl.col12 = col4, col8, col12
}

func New(ctx gl.Context, color Color) *Material {
	mtrl := &Material{
		BehaviorFlags: DescriptorRaised,
		icx:           -1,
		icy:           -1,
	}
	mtrl.cr, mtrl.cg, mtrl.cb, mtrl.ca = color.RGBA()

	// material has user-defined width and height, and precisely 1dp depth.
	mtrl.vbuf = glutil.NewFloatBuffer(ctx, []float32{
		0, 0, 0,
		0, 1, 0,
		1, 1, 0,
		1, 0, 0,
		0, 0, -1,
		0, 1, -1,
		1, 1, -1,
		1, 0, -1,
	}, gl.STATIC_DRAW)
	mtrl.ibuf = glutil.NewUintBuffer(ctx, []uint32{
		0, 2, 1, 0, 3, 2,
		2, 7, 6, 2, 3, 7,
		7, 3, 0, 7, 0, 4,
		4, 6, 7, 4, 5, 6,
		6, 1, 2, 6, 5, 1,
		1, 5, 4, 1, 4, 0,
	}, gl.STATIC_DRAW)

	n := float32(0.0234375)
	mtrl.uvbuf = glutil.NewFloatBuffer(ctx, []float32{
		0, n,
		0, 0,
		n, 0,
		n, n,
		0, n,
		0, 0,
		n, 0,
		n, n,
	}, gl.STATIC_DRAW)

	mtrl.prg0.CreateAndLink(ctx, glutil.ShaderAsset(gl.VERTEX_SHADER, "material-vert.glsl"), glutil.ShaderAsset(gl.FRAGMENT_SHADER, "material-frag.glsl"))
	mtrl.uw0 = mtrl.prg0.Uniform(ctx, "world")
	mtrl.uv0 = mtrl.prg0.Uniform(ctx, "view")
	mtrl.up0 = mtrl.prg0.Uniform(ctx, "proj")
	mtrl.uc0 = mtrl.prg0.Uniform(ctx, "color")
	mtrl.ap0 = mtrl.prg0.Attrib(ctx, "position")
	mtrl.utex0 = mtrl.prg0.Uniform(ctx, "tex0")
	mtrl.atc0 = mtrl.prg0.Attrib(ctx, "tc0")
	mtrl.uicon = mtrl.prg0.Uniform(ctx, "icon")
	mtrl.ucirc = mtrl.prg0.Uniform(ctx, "circle")

	mtrl.prg1.CreateAndLink(ctx, glutil.ShaderAsset(gl.VERTEX_SHADER, "material-shadow-vert.glsl"), glutil.ShaderAsset(gl.FRAGMENT_SHADER, "material-shadow-frag.glsl"))
	mtrl.uw1 = mtrl.prg1.Uniform(ctx, "world")
	mtrl.uv1 = mtrl.prg1.Uniform(ctx, "view")
	mtrl.up1 = mtrl.prg1.Uniform(ctx, "proj")
	mtrl.uc1 = mtrl.prg1.Uniform(ctx, "color")
	mtrl.ap1 = mtrl.prg1.Attrib(ctx, "position")
	return mtrl
}

func (mtrl *Material) SetColor(color Color) {
	mtrl.cr, mtrl.cg, mtrl.cb, mtrl.ca = color.RGBA()
}

func (mtrl *Material) SetIcon(ic icon.Icon) {
	mtrl.icx, mtrl.icy = ic.Texcoords()
}

func (mtrl *Material) Bind(lpro *simplex.Program) {
	mtrl.Box = NewBox(lpro)
}

func (mtrl *Material) World() *f32.Mat4 { return &mtrl.world }

// TODO seems to slow down goimport ...
var shdr, shdg, shdb, shda = BlueGrey900.RGBA()

func (mtrl *Material) Draw(ctx gl.Context, view, proj f32.Mat4) {
	if mtrl.BehaviorFlags&DescriptorRaised == DescriptorRaised {
		// provide larger world mat for shadows to draw within
		m := mtrl.world
		w, h := m[0][0], m[1][1]
		z := m[2][3]
		s := float32(1.25) + (z * 0.01414)
		// s := float32(2)
		m.Scale(&m, s, s, 1)
		m[0][3] -= (m[0][0] - w) / 2
		m[1][3] -= (m[1][1] - h) / 2
		// m[1][3] -= 2 // shadow y-offset

		// draw shadow
		mtrl.prg1.Use(ctx)
		mtrl.prg1.Mat4(ctx, mtrl.uw1, m)
		mtrl.prg1.Mat4(ctx, mtrl.uv1, view)
		mtrl.prg1.Mat4(ctx, mtrl.up1, proj)
		mtrl.prg1.U4f(ctx, mtrl.uc1, shdr, shdg, shdb, shda)
		mtrl.vbuf.Bind(ctx)
		mtrl.ibuf.Bind(ctx)
		mtrl.prg1.Pointer(ctx, mtrl.ap1, 3)
		mtrl.ibuf.Draw(ctx, mtrl.prg1, gl.TRIANGLES)
	}

	// draw material
	mtrl.prg0.Use(ctx)
	mtrl.prg0.Mat4(ctx, mtrl.uw0, mtrl.world)
	mtrl.prg0.Mat4(ctx, mtrl.uv0, view)
	mtrl.prg0.Mat4(ctx, mtrl.up0, proj)

	alpha := mtrl.ca
	if mtrl.BehaviorFlags&DescriptorFlat == DescriptorFlat {
		alpha = 0
	}
	mtrl.prg0.U4f(ctx, mtrl.uc0, mtrl.cr, mtrl.cg, mtrl.cb, alpha)
	mtrl.prg0.U2f(ctx, mtrl.uicon, mtrl.icx, mtrl.icy)
	if mtrl.IsCircle {
		mtrl.prg0.U1i(ctx, mtrl.ucirc, 1)
	}

	mtrl.vbuf.Bind(ctx)
	mtrl.ibuf.Bind(ctx)
	mtrl.prg0.Pointer(ctx, mtrl.ap0, 3)

	if mtrl.Texture.Value > 0 {
		mtrl.Texture.Bind(ctx, DefaultFilter, DefaultWrap)
		mtrl.prg0.U1i(ctx, mtrl.utex0, int(mtrl.Texture.Value-1))
		mtrl.uvbuf.Bind(ctx)
		mtrl.prg0.Pointer(ctx, mtrl.atc0, 2)
	}

	mtrl.ibuf.Draw(ctx, mtrl.prg0, gl.TRIANGLES)
}

func (mtrl *Material) M() *Material { return mtrl }

func (mtrl *Material) Contains(tx, ty float32) bool {
	x, y, w, h := mtrl.world[0][3], mtrl.world[1][3], mtrl.world[0][0], mtrl.world[1][1]
	return x <= tx && tx <= x+w && y <= ty && ty <= y+h
}

type Button struct {
	*Material
	OnPress func()
}

// TODO https://www.google.com/design/spec/layout/structure.html#structure-toolbars
type Toolbar struct {
	*Material
	Nav     *Button
	actions []*Button
}

func (bar *Toolbar) AddAction(btn *Button) {
	btn.BehaviorFlags = DescriptorFlat
	bar.actions = append(bar.actions, btn)
}

func (tb *Toolbar) Constraints(env *Environment) []simplex.Constraint {
	stp := env.Grid.StepSize()
	var (
		width, height float32
		btnsize       float32
	)

	switch env.Grid.Columns {
	case 4:
		width = float32(tb.col4) * stp
		height = Dp(56).Px()
		btnsize = Dp(24).Px()
	case 8:
		width = float32(tb.col8) * stp
		height = Dp(56).Px()
		btnsize = Dp(24).Px()
	case 12:
		width = float32(tb.col12) * stp
		height = Dp(64).Px()
		btnsize = Dp(32).Px()
	}
	nav := tb.Nav
	cns := []simplex.Constraint{
		tb.Width(width), tb.Height(height),
		tb.Start(env.Grid.Margin), tb.Top(float32(windowSize.HeightPx) - env.Grid.Margin),
		tb.Z(4),
		nav.Width(btnsize), nav.Height(btnsize), nav.Z(5),
		nav.StartIn(tb.Box, env.Grid.Gutter),
		nav.CenterVerticalIn(tb.Box),
	}

	for i, btn := range tb.actions {
		cns = append(cns, btn.Width(btnsize), btn.Height(btnsize), btn.Z(5), btn.CenterVerticalIn(tb.Box))
		if i == 0 {
			cns = append(cns, btn.EndIn(tb.Box, env.Grid.Gutter))
		} else {
			cns = append(cns, btn.Before(tb.actions[i-1].Box, env.Grid.Gutter))
		}
	}

	return cns
}

type NavDrawer struct {
	*Material
}

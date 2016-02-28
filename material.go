package material

// TODO fonts
// https://www.mapbox.com/blog/text-signed-distance-fields/
// https://github.com/libgdx/libgdx/wiki/Distance-field-fonts
// https://lambdacube3d.wordpress.com/2014/11/12/playing-around-with-font-rendering/

import (
	"time"

	"dasa.cc/material/glutil"
	"dasa.cc/material/icon"
	"dasa.cc/material/simplex"
	"dasa.cc/material/text"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

var (
	DefaultFilter = glutil.TextureFilter(gl.LINEAR, gl.LINEAR)
	DefaultWrap   = glutil.TextureWrap(gl.REPEAT, gl.REPEAT)

	nearestFilter = glutil.TextureFilter(gl.NEAREST, gl.NEAREST)
	glyphsFilter  = glutil.TextureFilter(gl.LINEAR_MIPMAP_LINEAR, gl.LINEAR)
)

type Text struct {
	Texture glutil.Texture
	uvbuf   glutil.FloatBuffer
	uicon   gl.Uniform

	vbuf glutil.FloatBuffer // vertices
	ibuf glutil.UintBuffer  // indices

	prg        glutil.Program
	ap         gl.Attrib  // buffer pointer
	uc         gl.Uniform // color
	uw, uv, up gl.Uniform // projection

	cr, cg, cb, ca float32 // color for uniform

	utex0 gl.Uniform // texture uniform
	atc0  gl.Attrib  // texcoords pointer

	worldfn func(*f32.Mat4)
}

func NewText(ctx gl.Context) *Text {
	txt := &Text{}
	txt.vbuf = glutil.NewFloatBuffer(ctx, []float32{
		0, 0, 0,
		0, 1, 0,
		1, 1, 0,
		1, 0, 0,
	}, gl.STATIC_DRAW)
	txt.ibuf = glutil.NewUintBuffer(ctx, []uint32{
		0, 2, 1,
		0, 3, 2,
	}, gl.STATIC_DRAW)

	nx := float32(36) / float32(512)
	ny := float32(72) / float32(512)
	txt.uvbuf = glutil.NewFloatBuffer(ctx, []float32{
		0, ny,
		0, 0,
		nx, 0,
		nx, ny,
	}, gl.STATIC_DRAW)
	txt.reloadProgram(ctx)
	return txt
}

func (txt *Text) reloadProgram(ctx gl.Context) {
	txt.prg.CreateAndLink(ctx,
		glutil.ShaderAsset(gl.VERTEX_SHADER, "material-glyphs-vert.glsl"),
		glutil.ShaderAsset(gl.FRAGMENT_SHADER, "material-glyphs-frag.glsl"))
	txt.uw = txt.prg.Uniform(ctx, "world")
	txt.uv = txt.prg.Uniform(ctx, "view")
	txt.up = txt.prg.Uniform(ctx, "proj")
	txt.uc = txt.prg.Uniform(ctx, "color")
	txt.ap = txt.prg.Attrib(ctx, "position")
	txt.utex0 = txt.prg.Uniform(ctx, "tex0")
	txt.atc0 = txt.prg.Attrib(ctx, "tc0")
	txt.uicon = txt.prg.Uniform(ctx, "icon")
}

func (txt *Text) Draw(ctx gl.Context, s string, world, view, proj f32.Mat4) {
	m := world
	if txt.worldfn != nil {
		txt.worldfn(&m)
	}
	scale := float32(58.0 / 50.0)
	m.Scale(&m, scale, scale, 1)
	m[0][0] = m[1][1] / 2

	txt.prg.Use(ctx)
	txt.prg.Mat4(ctx, txt.uv, view)
	txt.prg.Mat4(ctx, txt.up, proj)

	txt.prg.U4f(ctx, txt.uc, txt.cr, txt.cg, txt.cb, 0) // bind color

	txt.vbuf.Bind(ctx)
	txt.ibuf.Bind(ctx)
	txt.prg.Pointer(ctx, txt.ap, 3)

	if txt.Texture.Value > 0 {
		txt.Texture.Bind(ctx, glyphsFilter, DefaultWrap)
		txt.prg.U1i(ctx, txt.utex0, int(txt.Texture.Value-1))
		txt.uvbuf.Bind(ctx)
		txt.prg.Pointer(ctx, txt.atc0, 2)
	}

	for _, r := range s {
		txt.prg.Mat4(ctx, txt.uw, m)

		// TODO tmp workaround to messy glyph texture
		y := m[1][3]
		m.Translate(&m, 0, 0.5, 0) // glyph height takes up approximately half height of sprite in gen'd texture
		m[0][3] += m[1][3] - y
		m[1][3] = y

		xy := text.Texcoords[r]
		txt.prg.U2f(ctx, txt.uicon, xy[0], xy[1])
		txt.ibuf.Draw(ctx, txt.prg, gl.TRIANGLES)
	}
}

type Material struct {
	Box

	Drawer glutil.DrawerFunc

	col4, col8, col12 int

	hidden bool

	BehaviorFlags Behavior

	Texture      glutil.Texture
	GlyphTexture glutil.Texture
	uvbuf        glutil.FloatBuffer
	uicon        gl.Uniform

	icx, icy float32

	cr, cg, cb, ca     float32 // color for uniform
	cir, cig, cib, cia float32 // icon color for uniform

	vbuf glutil.FloatBuffer // vertices
	ibuf glutil.UintBuffer  // indices

	prg0, prg1    glutil.Program // material and shadow TODO globalize with batch op?
	ap0, ap1      gl.Attrib      // buffer pointer
	uc0, uc1      gl.Uniform     // color
	uw0, uv0, up0 gl.Uniform     // material projection
	uw1, uv1, up1 gl.Uniform     // shadow projection

	us1 gl.Uniform

	utex0 gl.Uniform
	atc0  gl.Attrib

	mtext *Text
	text  string

	// TODO tmp impl
	IsCircle bool
	ucirc0   gl.Uniform
	ucirc1   gl.Uniform
}

func (mtrl *Material) Span(col4, col8, col12 int) {
	mtrl.col4, mtrl.col8, mtrl.col12 = col4, col8, col12
}

func New(ctx gl.Context, color Color) *Material {
	mtrl := &Material{
		BehaviorFlags: DescriptorRaised,
		icx:           -1,
		icy:           -1,
		mtext:         NewText(ctx),
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

	mtrl.reloadProgram(ctx)
	return mtrl
}

func (mtrl *Material) reloadProgram(ctx gl.Context) {
	mtrl.prg0.CreateAndLink(ctx, glutil.ShaderAsset(gl.VERTEX_SHADER, "material-vert.glsl"), glutil.ShaderAsset(gl.FRAGMENT_SHADER, "material-frag.glsl"))
	mtrl.uw0 = mtrl.prg0.Uniform(ctx, "world")
	mtrl.uv0 = mtrl.prg0.Uniform(ctx, "view")
	mtrl.up0 = mtrl.prg0.Uniform(ctx, "proj")
	mtrl.uc0 = mtrl.prg0.Uniform(ctx, "color")
	mtrl.ap0 = mtrl.prg0.Attrib(ctx, "position")
	mtrl.utex0 = mtrl.prg0.Uniform(ctx, "tex0")
	mtrl.atc0 = mtrl.prg0.Attrib(ctx, "tc0")
	mtrl.uicon = mtrl.prg0.Uniform(ctx, "icon")
	mtrl.ucirc0 = mtrl.prg0.Uniform(ctx, "circle")

	mtrl.prg1.CreateAndLink(ctx, glutil.ShaderAsset(gl.VERTEX_SHADER, "material-shadow-vert.glsl"), glutil.ShaderAsset(gl.FRAGMENT_SHADER, "material-shadow-frag.glsl"))
	mtrl.uw1 = mtrl.prg1.Uniform(ctx, "world")
	mtrl.uv1 = mtrl.prg1.Uniform(ctx, "view")
	mtrl.up1 = mtrl.prg1.Uniform(ctx, "proj")
	mtrl.uc1 = mtrl.prg1.Uniform(ctx, "color")
	mtrl.us1 = mtrl.prg1.Uniform(ctx, "size")
	mtrl.ucirc1 = mtrl.prg1.Uniform(ctx, "circle")
	mtrl.ap1 = mtrl.prg1.Attrib(ctx, "position")

	mtrl.mtext.reloadProgram(ctx)
}

// SetColor sets background color of material unless material flags contains DescriptorFlat.
func (mtrl *Material) SetColor(color Color) {
	mtrl.cr, mtrl.cg, mtrl.cb, mtrl.ca = color.RGBA()
}

func (mtrl *Material) SetIcon(ic icon.Icon) {
	mtrl.icx, mtrl.icy = ic.Texcoords()
}

func (mtrl *Material) SetIconColor(color Color) {
	mtrl.cir, mtrl.cig, mtrl.cib, mtrl.cia = color.RGBA()
}

func (mtrl *Material) SetTextColor(color Color) {
	mtrl.mtext.cr, mtrl.mtext.cg, mtrl.mtext.cb, mtrl.mtext.ca = color.RGBA()
}

func (mtrl *Material) SetText(s string) {
	mtrl.text = s
}

func (mtrl *Material) Bind(lpro *simplex.Program) {
	mtrl.Box = NewBox(lpro)
}

func (mtrl *Material) World() *f32.Mat4 { return &mtrl.world }

func (mtrl *Material) Hidden() bool { return mtrl.hidden }

// TODO seems to slow down goimport ...
var shdr, shdg, shdb, shda = BlueGrey900.RGBA()

func min(a, b float32) float32 {
	if a < b {
		return a
	}
	return b
}

func max(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}

func (mtrl *Material) Draw(ctx gl.Context, view, proj f32.Mat4) {
	if mtrl.hidden {
		return
	}

	if mtrl.BehaviorFlags&DescriptorRaised == DescriptorRaised {
		// provide larger world mat for shadows to draw within
		m := mtrl.world
		w, h := m[0][0], m[1][1]
		z := m[2][3]

		s := float32(1.07) + (z * 0.01414)
		sw := m[0][0]*s - w
		sh := m[1][1]*s - h

		s = min(sw, sh)
		m[0][0] += s
		m[1][1] += s

		m[0][3] -= (m[0][0] - w) / 2
		m[1][3] -= (m[1][1] - h) / 2
		m[1][3] -= 2 // shadow y-offset

		// draw shadow
		mtrl.prg1.Use(ctx)
		mtrl.prg1.Mat4(ctx, mtrl.uw1, m)
		mtrl.prg1.Mat4(ctx, mtrl.uv1, view)
		mtrl.prg1.Mat4(ctx, mtrl.up1, proj)
		mtrl.prg1.U2f(ctx, mtrl.us1, w, h)
		mtrl.prg1.U4f(ctx, mtrl.uc1, shdr, shdg, shdb, shda)
		if mtrl.IsCircle {
			mtrl.prg1.U1i(ctx, mtrl.ucirc1, 1)
		}
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

	flat := mtrl.BehaviorFlags&DescriptorFlat == DescriptorFlat

	if flat {
		mtrl.prg0.U4f(ctx, mtrl.uc0, mtrl.cir, mtrl.cig, mtrl.cib, 0)
		mtrl.prg0.U2f(ctx, mtrl.uicon, mtrl.icx, mtrl.icy)
	} else {
		mtrl.prg0.U4f(ctx, mtrl.uc0, mtrl.cr, mtrl.cg, mtrl.cb, mtrl.ca)
		mtrl.prg0.U2f(ctx, mtrl.uicon, -1, -1)
	}

	if mtrl.IsCircle {
		mtrl.prg0.U1i(ctx, mtrl.ucirc0, 1)
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

	if !flat && mtrl.icx != -1 { // draw icon on background
		const scl = 0.42857142857
		m := mtrl.world
		w, h := m[0][0], m[1][1]
		sw := m[0][0]*scl - w
		sh := m[1][1]*scl - h
		s := min(sw, sh)
		m[0][0] += s
		m[1][1] += s
		m[0][3] -= (m[0][0] - w) / 2
		m[1][3] -= (m[1][1] - h) / 2
		mtrl.prg1.Mat4(ctx, mtrl.uw1, m)

		//
		mtrl.prg0.U4f(ctx, mtrl.uc0, mtrl.cir, mtrl.cig, mtrl.cib, 0)
		mtrl.prg0.U2f(ctx, mtrl.uicon, mtrl.icx, mtrl.icy)
		mtrl.ibuf.Draw(ctx, mtrl.prg0, gl.TRIANGLES)
	}

	// draw text
	mtrl.mtext.Texture = mtrl.GlyphTexture
	mtrl.mtext.Draw(ctx, mtrl.text, mtrl.world, view, proj)

	if mtrl.Drawer != nil {
		mtrl.Drawer(ctx, view, proj)
	}
}

func (mtrl *Material) M() *Material { return mtrl }

func (mtrl *Material) Contains(tx, ty float32) bool {
	x, y, w, h := mtrl.world[0][3], mtrl.world[1][3], mtrl.world[0][0], mtrl.world[1][1]
	return x <= tx && tx <= x+w && y <= ty && ty <= y+h
}

func (mtrl *Material) Constraints(env *Environment) []simplex.Constraint {
	return nil
}

type Button struct {
	*Material
	OnPress func()
	OnTouch func(touch.Event)
}

type FloatingActionButton struct {
	*Material
	Mini    bool
	OnPress func()
	OnTouch func(touch.Event)
}

func (fab *FloatingActionButton) Constraints(env *Environment) []simplex.Constraint {
	var size float32
	switch env.Grid.Columns {
	case 4, 8:
		if fab.Mini {
			size = Dp(40).Px()
		} else {
			size = Dp(56).Px()
		}
	case 12:
		if fab.Mini {
			size = Dp(48).Px() // TODO size unconfirmed
		} else {
			size = Dp(64).Px()
		}
	}
	return []simplex.Constraint{fab.Width(size), fab.Height(size), fab.Z(6)}
}

// TODO https://www.google.com/design/spec/layout/structure.html#structure-toolbars
type Toolbar struct {
	*Material
	Nav     *Button
	actions []*Button
}

func (bar *Toolbar) AddAction(btn *Button) {
	btn.BehaviorFlags = DescriptorFlat
	btn.SetIconColor(Black)
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
		tb.Width(width), tb.Height(height), tb.Z(4),
		tb.StartIn(env.Box, env.Grid.Margin), tb.TopIn(env.Box, env.Grid.Margin),
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

type Menu struct {
	*Material
	selected int
	actions  []*Button
}

func (mu *Menu) AddAction(btn *Button) {
	btn.BehaviorFlags = DescriptorFlat
	btn.SetIconColor(Black)
	btn.hidden = mu.hidden
	mu.actions = append(mu.actions, btn)
}

func (mu *Menu) ShowAt(m *f32.Mat4) {
	x := mu.Box.world[0][3]
	y := mu.Box.world[1][3]
	mu.Box.world[0][3] = m[0][3]
	mu.Box.world[1][3] = m[1][3] + m[1][1] - mu.Box.world[1][1]
	dx := mu.Box.world[0][3] - x
	dy := mu.Box.world[1][3] - y
	for _, btn := range mu.actions {
		btn.Box.world[0][3] += dx
		btn.Box.world[1][3] += dy
	}
	mu.Show()
}

func (mu *Menu) Show() {
	go func() {
		h := mu.Box.world[1][1]
		y := mu.Box.world[1][3]
		anim := Animation{
			Sig: ExpSig,
			Dur: 300 * time.Millisecond,
			Start: func() {
				mu.Box.world[1][1] = 0
				mu.Box.world[1][3] = y + h
				mu.hidden = false
				for _, btn := range mu.actions {
					btn.hidden = false
				}
			},
			Interp: func(dt float32) {
				mu.Box.world[1][1] = h * dt
				mu.Box.world[1][3] = (y + h) - h*dt
			},
			End: func() {
				mu.Box.world[1][1] = h
				mu.Box.world[1][3] = y
			},
		}
		anim.Do()
	}()
}

func (mu *Menu) Hide() {
	h := mu.Box.world[1][1]
	y := mu.Box.world[1][3]
	Animation{
		Sig: ExpSig,
		Dur: 200 * time.Millisecond,
		Interp: func(dt float32) {
			mu.Box.world[1][1] = h * (1 - dt)
			mu.Box.world[1][3] = (y + h) - h*(1-dt)
		},
		End: func() {
			mu.hidden = true
			for _, btn := range mu.actions {
				btn.hidden = true
			}
			mu.Box.world[1][1] = h
			mu.Box.world[1][3] = y
		},
	}.Do()
}

func (mu *Menu) Constraints(env *Environment) []simplex.Constraint {
	cns := []simplex.Constraint{
		mu.Width(Dp(100).Px()), mu.Z(8),
		mu.StartIn(env.Box, env.Grid.Margin), mu.TopIn(env.Box, env.Grid.Margin),
	}

	for i, btn := range mu.actions {
		cns = append(cns, btn.Width(Dp(100).Px()), btn.Height(Dp(16).Px()), btn.Z(9), btn.StartIn(mu.Box, Dp(16).Px()))
		if i == 0 {
			cns = append(cns, btn.TopIn(mu.Box, env.Grid.Gutter))
		} else {
			cns = append(cns, btn.Below(mu.actions[i-1].Box, Dp(20).Px()))
		}
	}
	cns = append(cns, mu.AlignBottoms(mu.actions[len(mu.actions)-1].Box, Dp(20).Px()))

	return cns
}

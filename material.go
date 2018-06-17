package material

import (
	"time"

	"dasa.cc/material/glutil"
	"dasa.cc/material/icon"
	"dasa.cc/simplex"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

var (
	DefaultFilter = glutil.TextureFilter(gl.LINEAR, gl.LINEAR)
	DefaultWrap   = glutil.TextureWrap(gl.REPEAT, gl.REPEAT)

	linearFilter  = glutil.TextureFilter(gl.LINEAR, gl.LINEAR)
	nearestFilter = glutil.TextureFilter(gl.NEAREST, gl.NEAREST)
	glyphsFilter  = glutil.TextureFilter(gl.LINEAR_MIPMAP_LINEAR, gl.LINEAR)
)

type Material struct {
	Box

	Drawer glutil.DrawerFunc

	col4, col8, col12 int

	hidden bool

	BehaviorFlags Behavior

	text struct {
		value      string
		height     float32
		r, g, b, a float32
	}

	icon struct {
		x, y       float32
		r, g, b, a float32
	}

	cr, cg, cb, ca float32 // color for uniform

	IsCircle  bool
	Roundness float32

	touch struct {
		state touch.Type
		x, y  float32
		start time.Time
	}

	ShowImage bool
	Rotate    float32 // Radian
}

func (mtrl *Material) Span(col4, col8, col12 int) {
	mtrl.col4, mtrl.col8, mtrl.col12 = col4, col8, col12
}

func New(ctx gl.Context, color Color) *Material {
	mtrl := &Material{
		BehaviorFlags: DescriptorRaised,
	}
	mtrl.icon.x, mtrl.icon.y = -1, -1
	mtrl.touch.state = touch.TypeEnd
	mtrl.cr, mtrl.cg, mtrl.cb, mtrl.ca = color.RGBA()

	return mtrl
}

// SetColor sets background color of material unless material flags contains DescriptorFlat.
func (mtrl *Material) SetColor(color Color) {
	mtrl.cr, mtrl.cg, mtrl.cb, mtrl.ca = color.RGBA()
}

func (mtrl *Material) SetIcon(ic icon.Icon) {
	mtrl.icon.x, mtrl.icon.y = ic.Texcoords()
}

func (mtrl *Material) SetIconColor(color Color) {
	mtrl.icon.r, mtrl.icon.g, mtrl.icon.b, mtrl.icon.a = color.RGBA()
}

func (mtrl *Material) SetTextColor(color Color) {
	mtrl.text.r, mtrl.text.g, mtrl.text.b, mtrl.text.a = color.RGBA()
}

func (mtrl *Material) SetTextHeight(h float32) {
	mtrl.text.height = h
}

func (mtrl *Material) SetText(s string) {
	mtrl.text.value = s
}

func (mtrl *Material) Bind(lpro *simplex.Program) {
	mtrl.Box = NewBox(lpro)
}

func (mtrl *Material) World() *f32.Mat4 { return &mtrl.world }

func (mtrl *Material) Hidden() bool { return mtrl.hidden }

func (mtrl *Material) M() *Material { return mtrl }

func (mtrl *Material) Contains(tx, ty float32) bool {
	x, y, w, h := mtrl.world[0][3], mtrl.world[1][3], mtrl.world[0][0], mtrl.world[1][1]
	return x <= tx && tx <= x+w && y <= ty && ty <= y+h
}

func (mtrl *Material) RelativeCoords(tx, ty float32) (float32, float32) {
	x, y, w, h := mtrl.world[0][3], mtrl.world[1][3], mtrl.world[0][0], mtrl.world[1][1]
	return (tx - x) / w, (ty - y) / h
}

func (mtrl *Material) Constraints(env *Environment) []simplex.Constraint {
	return nil
}

// TODO seems to slow down goimport ...
var shdr, shdg, shdb, shda = BlueGrey900.RGBA()

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
	fab.Roundness = size / 2 // TODO consider how this should work
	return []simplex.Constraint{fab.Width(size), fab.Height(size), fab.Z(6)}
}

// TODO https://www.google.com/design/spec/layout/structure.html#structure-toolbars
type Toolbar struct {
	*Material
	Nav     *Button
	Title   *Material
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
		titleStart    float32
	)

	switch env.Grid.Columns {
	case 4:
		width = float32(tb.col4) * stp
		height = Dp(56).Px()
		btnsize = Dp(24).Px()
		titleStart = Dp(48).Px()
	case 8:
		width = float32(tb.col8) * stp
		height = Dp(56).Px()
		btnsize = Dp(24).Px()
		titleStart = Dp(72).Px()
	case 12:
		width = float32(tb.col12) * stp
		height = Dp(64).Px()
		btnsize = Dp(32).Px()
		titleStart = Dp(72).Px()
	}
	nav := tb.Nav
	title := tb.Title
	cns := []simplex.Constraint{
		tb.Width(width), tb.Height(height), tb.Z(4),
		tb.StartIn(env.Box, env.Grid.Margin), tb.TopIn(env.Box, env.Grid.Margin),
		nav.Width(btnsize), nav.Height(btnsize), nav.Z(5),
		nav.StartIn(tb.Box, env.Grid.Gutter),
		nav.CenterVerticalIn(tb.Box),
		title.StartIn(tb.Box, titleStart), title.Before(tb.actions[len(tb.actions)-1].Box, 0),
		title.CenterVerticalIn(tb.Box), title.Height(btnsize), title.Z(5),
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

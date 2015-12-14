package material

import (
	"image"
	"log"
	"sort"

	"image/draw"
	_ "image/png"

	"dasa.cc/material/glutil"
	"dasa.cc/material/icon"
	"dasa.cc/material/simplex"

	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

var windowSize size.Event

type Dp float32

func (dp Dp) Px() float32 {
	// TODO
	// density := float32(windowSize.WidthPx) / (float32(windowSize.WidthPt) / 72)
	// return float32(dp) * (density / 160)
	return float32(dp)
}

type Sheet interface {
	Draw(ctx gl.Context, view, proj f32.Mat4)
	Bind(*simplex.Program)
	UpdateWorld(*simplex.Program)
	Contains(x, y float32) bool
	M() *Material
}

type byZ []Sheet

func (a byZ) Len() int           { return len(a) }
func (a byZ) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byZ) Less(i, j int) bool { return a[i].M().world[2][3] < a[j].M().world[2][3] }

type Environment struct {
	View f32.Mat4

	proj   f32.Mat4
	plt    Palette
	sheets []Sheet

	Grid *Grid

	lprg *simplex.Program

	icons glutil.Texture
}

func (env *Environment) LoadIcons(ctx gl.Context) {
	src, _, err := image.Decode(glutil.MustOpen("material-icons-black-mdpi.png"))
	if err != nil {
		log.Fatal(err)
	}

	r := image.Rect(0, 0, 2048, 2048)
	dst := image.NewNRGBA(r)
	// pt := image.Point{0, -(2048 - src.Bounds().Size().Y)}
	draw.Draw(dst, r, src, image.ZP, draw.Src)

	// f, _ := os.Create("debug-icons.png")
	// png.Encode(f, dst)

	env.icons.Create(ctx)
	env.icons.Bind(ctx, DefaultFilter, DefaultWrap)
	env.icons.Update(ctx, 2048, 2048, dst.Pix)
}

func (env *Environment) SetPerspective(sz size.Event) {
	windowSize = sz
	env.Grid = NewGrid()
	env.View.Identity() // TODO not here, only on creation
	env.proj.Identity()
	glutil.Perspective(&env.proj, 0, float32(sz.WidthPx), 0, float32(sz.HeightPx))
}

func (env *Environment) SetOrtho(sz size.Event) {
	windowSize = sz
	env.Grid = NewGrid()
	env.View.Identity() // TODO not here, only on creation
	env.proj.Identity()
	glutil.Ortho(&env.proj, 0, float32(sz.WidthPx), 0, float32(sz.HeightPx), 1, 10000)
	env.View.Translate(&env.View, 0, 0, -5000)
}

func (env *Environment) SetPalette(plt Palette) {
	env.plt = plt
	for _, sheet := range env.sheets {
		switch sheet := sheet.(type) {
		case Button:
			sheet.SetColor(env.plt.Primary)
		case Toolbar:
			sheet.SetColor(env.plt.Light)
		}
	}
}

func (env *Environment) StartLayout() {
	env.lprg = new(simplex.Program)
	for _, sheet := range env.sheets {
		sheet.Bind(env.lprg)
	}
}

func (env *Environment) AddConstraints(cns ...simplex.Constraint) {
	env.lprg.AddConstraints(cns...)
}

func (env *Environment) FinishLayout() {
	if err := env.lprg.Minimize(); err != nil {
		log.Println(err)
	}
	for _, sheet := range env.sheets {
		sheet.UpdateWorld(env.lprg)
	}
}

func (env *Environment) Draw(ctx gl.Context) {
	sort.Sort(byZ(env.sheets))
	for _, sheet := range env.sheets {
		sheet.M().Texture = env.icons
		sheet.Draw(ctx, env.View, env.proj)
	}
}

func (env *Environment) DrawGridDebug(ctx gl.Context) {
	env.Grid.draw(ctx, env.View, env.proj)
}

func (env *Environment) Touch(ev touch.Event) bool {
	ex, ey := ev.X, float32(windowSize.HeightPx)-ev.Y
	// for _, sheet := range env.sheets {
	for i := len(env.sheets) - 1; i >= 0; i-- {
		sheet := env.sheets[i]
		if sheet.Contains(ex, ey) {
			log.Println("Hit!", ev)
			switch sheet := sheet.(type) {
			case *Button:
				if ev.Type == touch.TypeBegin && sheet.OnPress != nil {
					sheet.OnPress()
				}
			default:
				log.Printf("Unhandled type %T\n", sheet)
			}
			return true
		}
	}
	return false
}

func (env *Environment) NewButton(ctx gl.Context) *Button {
	btn := &Button{Material: New(ctx, Black)} // TODO update constructor to remove color arg
	btn.SetColor(env.plt.Primary)
	env.sheets = append(env.sheets, btn)
	return btn
}

func (env *Environment) NewToolbar(ctx gl.Context) *Toolbar {
	bar := &Toolbar{
		Material: New(ctx, Black),
		Nav:      env.NewButton(ctx),
	}
	bar.SetColor(env.plt.Light) // create specific ColorFromPalette on each type to localize selection
	bar.Nav.BehaviorFlags = DescriptorFlat
	bar.Nav.SetIcon(icon.NavigationMenu)
	env.sheets = append(env.sheets, bar)
	return bar
}

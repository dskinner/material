package material

import (
	"bytes"
	"fmt"
	"image"
	"log"
	"sort"
	"time"
	"unicode"

	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"

	"dasa.cc/material/assets"
	"dasa.cc/material/glutil"
	"dasa.cc/material/icon"
	"dasa.cc/material/text"
	"dasa.cc/simplex"

	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

type Dp float32

func (dp Dp) Px() float32 {
	if windowSize.PixelsPerPt == 1 {
		return float32(dp)
	}
	density := windowSize.PixelsPerPt * 72
	return float32(dp) * density / 160
}

type Sheet interface {
	// Draw(ctx gl.Context, view, proj f32.Mat4)
	Bind(*simplex.Program)
	UpdateWorld(*simplex.Program)
	Contains(x, y float32) bool
	M() *Material
	Constraints(*Environment) []simplex.Constraint
	Hidden() bool
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

	Box  Box
	Grid *Grid

	lprg *simplex.Program

	icons  glutil.Texture
	glyphs glutil.Texture

	image glutil.Texture

	prg glutil.Program

	uniforms struct {
		view, proj, shadowColor gl.Uniform
		glyphs, icons, image    gl.Uniform
		glyphconf               gl.Uniform
	}

	attribs struct {
		vertex, color, dist, texcoord gl.Attrib
		touch                         gl.Attrib
	}

	buffers struct {
		verts, colors, dists, texcoords glutil.FloatBuffer
		indices                         glutil.UintBuffer
		touches                         glutil.FloatBuffer
	}

	verts, colors, dists, texcoords []float32
	indices                         []uint32
	touches                         []float32

	watchEvent chan string
	watchQuit  chan bool
}

func (env *Environment) Proj() f32.Mat4 { return env.proj }

func (env *Environment) Size() size.Event {
	return windowSize
}

func (env *Environment) WatchShaders() {
	env.watchEvent, env.watchQuit = watchShaders()
}

func (env *Environment) LoadIcons(ctx gl.Context) {
	src, _, err := image.Decode(glutil.MustOpen("material/material-icons-black-mdpi.png"))
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
	env.icons.Bind(ctx, nearestFilter, DefaultWrap)
	env.icons.Update(ctx, 0, 2048, 2048, dst.Pix)
}

var (
	imageSize        image.Point
	imageTextureSize = 2048
)

func (env *Environment) LoadImage(ctx gl.Context, name string) image.Point {
	src, _, err := image.Decode(glutil.MustOpen(name))
	if err != nil {
		log.Fatal(err)
	}

	imageSize = src.Bounds().Size()
	r := image.Rect(0, 0, imageTextureSize, imageTextureSize)
	dst := image.NewNRGBA(r)
	draw.Draw(dst, r, src, image.ZP, draw.Src)

	env.image.Create(ctx)
	env.image.Bind(ctx, nearestFilter, DefaultWrap)
	env.image.Update(ctx, 0, 2048, 2048, dst.Pix)

	return imageSize
}

func (env *Environment) LoadGlyphs(ctx gl.Context) {
	src, _, err := image.Decode(bytes.NewReader(text.Texture)) //image.Decode(glutil.MustOpen("material/glyphs.png"))
	if err != nil {
		log.Fatal(err)
	}
	env.glyphs.Create(ctx)
	env.glyphs.Bind(ctx, nearestFilter, DefaultWrap)
	switch src.(type) {
	case *image.RGBA:
		env.glyphs.Update(ctx, 0, text.TextureSize, text.TextureSize, src.(*image.RGBA).Pix)
	case *image.NRGBA:
		env.glyphs.Update(ctx, 0, text.TextureSize, text.TextureSize, src.(*image.NRGBA).Pix)
	default:
		panic(fmt.Errorf("Unhandled image type %T", src))
	}
}

func (env *Environment) Load(ctx gl.Context) {
	env.prg.CreateAndLink(ctx,
		glutil.ShaderCompile(gl.VERTEX_SHADER, "env-vert.glsl", assets.VertexShader),
		glutil.ShaderCompile(gl.FRAGMENT_SHADER, "env-frag.glsl", assets.FragmentShader))

	env.uniforms.view = env.prg.Uniform(ctx, "view")
	env.uniforms.proj = env.prg.Uniform(ctx, "proj")
	env.uniforms.shadowColor = env.prg.Uniform(ctx, "shadowColor")
	env.uniforms.glyphs = env.prg.Uniform(ctx, "texglyph")
	env.uniforms.icons = env.prg.Uniform(ctx, "texicon")
	env.uniforms.glyphconf = env.prg.Uniform(ctx, "glyphconf")
	env.uniforms.image = env.prg.Uniform(ctx, "image")

	env.attribs.vertex = env.prg.Attrib(ctx, "vertex")
	env.attribs.color = env.prg.Attrib(ctx, "color")
	env.attribs.dist = env.prg.Attrib(ctx, "dist")
	env.attribs.texcoord = env.prg.Attrib(ctx, "texcoord")
	env.attribs.touch = env.prg.Attrib(ctx, "touch")

	env.buffers.indices = glutil.NewUintBuffer(ctx, []uint32{}, gl.STREAM_DRAW)
	env.buffers.verts = glutil.NewFloatBuffer(ctx, []float32{}, gl.STREAM_DRAW)
	env.buffers.colors = glutil.NewFloatBuffer(ctx, []float32{}, gl.STREAM_DRAW)
	env.buffers.dists = glutil.NewFloatBuffer(ctx, []float32{}, gl.STREAM_DRAW)
	env.buffers.texcoords = glutil.NewFloatBuffer(ctx, []float32{}, gl.STREAM_DRAW)
	env.buffers.touches = glutil.NewFloatBuffer(ctx, []float32{}, gl.STREAM_DRAW)
}

func (env *Environment) Unload(ctx gl.Context) {
	env.prg.Delete(ctx)
	env.buffers.indices.Delete(ctx)
	env.buffers.verts.Delete(ctx)
	env.buffers.colors.Delete(ctx)
	env.buffers.dists.Delete(ctx)
	env.buffers.texcoords.Delete(ctx)
	env.buffers.touches.Delete(ctx)
	if env.glyphs.Value != 0 {
		env.glyphs.Delete(ctx)
	}
	if env.icons.Value != 0 {
		env.icons.Delete(ctx)
	}
	env.sheets = env.sheets[:0]
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

func (env *Environment) Palette() Palette { return env.plt }

func (env *Environment) SetPalette(plt Palette) {
	env.plt = plt
	for _, sheet := range env.sheets {
		switch sheet := sheet.(type) {
		case *Button:
			sheet.SetColor(env.plt.Primary)
		case *Toolbar:
			sheet.SetColor(env.plt.Light)
		}
	}
}

func (env *Environment) StartLayout() {
	env.lprg = new(simplex.Program)
	env.Box = NewBox(env.lprg)
	for _, sheet := range env.sheets {
		sheet.Bind(env.lprg)
	}
	env.AddConstraints(
		env.Box.Width(float32(windowSize.WidthPx)),
		env.Box.Height(float32(windowSize.HeightPx)),
		env.Box.Z(0),
		env.Box.Start(0), env.Box.Top(float32(windowSize.HeightPx)),
	)
	for _, sheet := range env.sheets {
		env.AddConstraints(sheet.Constraints(env)...)
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
	select {
	case <-env.watchEvent:
		env.Load(ctx)
	default:
	}

	sort.Sort(byZ(env.sheets))

	env.indices = env.indices[:0]
	env.verts = env.verts[:0]
	env.colors = env.colors[:0]
	env.dists = env.dists[:0]
	env.texcoords = env.texcoords[:0]
	env.touches = env.touches[:0]

	for i, sheet := range env.sheets {
		m := sheet.M()
		x, y, z := m.world[0][3], m.world[1][3], m.world[2][3]
		w, h := m.world[0][0], m.world[1][1]
		r := m.Roundness

		n := uint32(len(env.verts)) / 4

		if i != 0 { // degenerate triangles
			// TODO make sure last v2 matches based on how shadows are being added
			env.indices = append(env.indices,
				env.indices[len(env.indices)-2], env.indices[len(env.indices)-1], env.indices[len(env.indices)-1],
				env.indices[len(env.indices)-1], env.indices[len(env.indices)-1], n,
			)
		}

		// *** shadow layer
		if m.BehaviorFlags&DescriptorRaised == DescriptorRaised {
			// (r/w) should be a value in range [0.0..0.5] given that a value of 0.5
			// is an ellipse/circle. mapping this range to [1..3]*z provides a decent
			// default for resizing ellipses shadows for visibility given current
			// algorithm in shader.
			// s := (1 + 6*(r/w)) + z // TODO this needs harder limits based on size of material
			s := 4 + z

			// min := h
			// if w < h {
			// min = w
			// }
			// _ = min

			// s += min / 32

			ss := 2 * s
			// TODO how should roundness scale
			rr := r * ((w + ss) / w)
			// rr += s

			// rr += min / 8

			// clamp rr for circular shadows
			if rr > (w+ss)/2 {
				rr = (w + ss) / 2
			}
			// if rr > (h+ss)/2 {
			// rr = (h + ss) / 2
			// }

			x -= s
			w += ss

			y -= s * 1.5 // offset shadow from material
			h += ss

			env.indices = append(env.indices,
				n, n+2, n+1,
				n, n+3, n+2,
			)
			env.verts = append(env.verts,
				x, y, -z, rr, // v0
				x, y+h, -z, rr, // v1
				x+w, y+h, -z, rr, // v2
				x+w, y, -z, rr, // v3
			)
			env.colors = append(env.colors,
				m.cr, m.cg, m.cb, m.ca,
				m.cr, m.cg, m.cb, m.ca,
				m.cr, m.cg, m.cb, m.ca,
				m.cr, m.cg, m.cb, m.ca,
			)
			env.dists = append(env.dists,
				0.0, 0.0, w, h, // v0 left, bottom
				0.0, 1.0, w, h, // v1 left, top
				1.0, 1.0, w, h, // v2 right, top
				1.0, 0.0, w, h, // v3 right, bottom
			)
			env.texcoords = append(env.texcoords,
				-1, -1, -1, -1,
				-1, -1, -1, -1,
				-1, -1, -1, -1,
				-1, -1, -1, -1,
			)
			env.touches = append(env.touches,
				0, 0, 2, 0,
				0, 0, 2, 0,
				0, 0, 2, 0,
				0, 0, 2, 0,
			)
		}
		// *** end shadow layer

		x, y, z = m.world[0][3], m.world[1][3], m.world[2][3]
		w, h = m.world[0][0], m.world[1][1]
		n = uint32(len(env.verts)) / 4

		// sin, cos := f32.Sin(m.Rotate), f32.Cos(m.Rotate)
		// fx, fy := m.world[0][0], m.world[1][1]
		// w = fx*cos - fy*sin
		// h = fx*sin + fy*cos

		env.indices = append(env.indices,
			n, n+2, n+1, n, n+3, n+2,
			n+2, n+7, n+6, n+2, n+3, n+7,
			n+7, n+3, n, n+7, n, n+4,
			n+4, n+6, n+7, n+4, n+5, n+6,
			n+6, n+1, n+2, n+6, n+5, n+1,
			n+1, n+5, n+4, n+1, n+4, n,
		)
		env.verts = append(env.verts,
			x, y, z, r,
			x, y+h, z, r,
			x+w, y+h, z, r,
			x+w, y, z, r,
			x, y, z-1, r,
			x, y+h, z-1, r,
			x+w, y+h, z-1, r,
			x+w, y, z-1, r,
		)

		alpha := float32(0)
		if m.BehaviorFlags&DescriptorRaised == DescriptorRaised {
			alpha = m.ca
		}
		env.colors = append(env.colors,
			m.cr, m.cg, m.cb, alpha,
			m.cr, m.cg, m.cb, alpha,
			m.cr, m.cg, m.cb, alpha,
			m.cr, m.cg, m.cb, alpha,
			1, 1, 1, alpha,
			1, 1, 1, alpha,
			1, 1, 1, alpha,
			1, 1, 1, alpha,
		)
		env.dists = append(env.dists,
			0.0, 0.0, w, h, // v0 left, bottom
			0.0, 1.0, w, h, // v1 left, top
			1.0, 1.0, w, h, // v2 right, top
			1.0, 0.0, w, h, // v3 right, bottom
			0.0, 0.0, w, h, // v0 left, bottom
			0.0, 1.0, w, h, // v1 left, top
			1.0, 1.0, w, h, // v2 right, top
			1.0, 0.0, w, h, // v3 right, bottom
		)
		env.texcoords = append(env.texcoords,
			-1, -1, -1, -1,
			-1, -1, -1, -1,
			-1, -1, -1, -1,
			-1, -1, -1, -1,
			-1, -1, -1, -1,
			-1, -1, -1, -1,
			-1, -1, -1, -1,
			-1, -1, -1, -1,
		)

		ex, ey := m.touch.x, m.touch.y
		es := float32(m.touch.state)
		ed := float32(time.Since(m.touch.start) / time.Millisecond)
		env.touches = append(env.touches,
			ex, ey, es, ed,
			ex, ey, es, ed,
			ex, ey, es, ed,
			ex, ey, es, ed,
			ex, ey, es, ed,
			ex, ey, es, ed,
			ex, ey, es, ed,
			ex, ey, es, ed,
		)

		if m.icon.x != -1 {
			n = uint32(len(env.verts)) / 4
			env.indices = append(env.indices,
				n, n+2, n+1, n, n+3, n+2,
			)
			env.verts = append(env.verts,
				x, y, z, 0,
				x, y+h, z, 0,
				x+w, y+h, z, 0,
				x+w, y, z, 0,
			)
			env.colors = append(env.colors,
				m.icon.r, m.icon.g, m.icon.b, m.icon.a,
				m.icon.r, m.icon.g, m.icon.b, m.icon.a,
				m.icon.r, m.icon.g, m.icon.b, m.icon.a,
				m.icon.r, m.icon.g, m.icon.b, m.icon.a,
			)
			env.dists = append(env.dists,
				0.0, 0.0, w, h, // v0 left, bottom
				0.0, 1.0, w, h, // v1 left, top
				1.0, 1.0, w, h, // v2 right, top
				1.0, 0.0, w, h, // v3 right, bottom
			)
			s := float32(0.0234375)
			ix, iy := m.icon.x, m.icon.y
			env.texcoords = append(env.texcoords,
				ix, iy+s, 1, 0,
				ix, iy, 1, 0,
				ix+s, iy, 1, 0,
				ix+s, iy+s, 1, 0,
			)
			env.touches = append(env.touches,
				0, 0, 2, 0,
				0, 0, 2, 0,
				0, 0, 2, 0,
				0, 0, 2, 0,
			)
		}

		if m.ShowImage {
			n = uint32(len(env.verts)) / 4
			env.indices = append(env.indices,
				n, n+2, n+1, n, n+3, n+2,
			)
			env.verts = append(env.verts,
				x, y, z, 0,
				x, y+h, z, 0,
				x+w, y+h, z, 0,
				x+w, y, z, 0,
			)
			env.colors = append(env.colors,
				1, 1, 1, alpha,
				1, 1, 1, alpha,
				1, 1, 1, alpha,
				1, 1, 1, alpha,
			)
			env.dists = append(env.dists,
				0.0, 0.0, w, h, // v0 left, bottom
				0.0, 1.0, w, h, // v1 left, top
				1.0, 1.0, w, h, // v2 right, top
				1.0, 0.0, w, h, // v3 right, bottom
			)

			// imgX, imgY = 1, 1
			// proportion that image occupies of actual texture
			// texture has to be like 2048x2048
			// so if image is 800x600
			// then max X value would be 800/2048
			// and max Y value would be 600/2048
			mX := float32(imageSize.X) / float32(imageTextureSize)
			mY := float32(imageSize.Y) / float32(imageTextureSize)
			env.texcoords = append(env.texcoords,
				0, mY, 3, 0,
				0, 0, 3, 0,
				mX, 0, 3, 0,
				mX, mY, 3, 0,
			)
			env.touches = append(env.touches,
				0, 0, 3, 0,
				0, 0, 3, 0,
				0, 0, 3, 0,
				0, 0, 3, 0,
			)
		}

		// draw text
		tx, ty := m.world[0][3], m.world[1][3]
		th := m.text.height
		if th == 0 {
			th = m.world[1][1]
		}

		pad := float32(text.Pad) * (th / text.FontSize)
		ty = ty + m.world[1][1] - (text.AscentUnit * th)

		for _, r := range m.text.value {
			a := text.Bounds[r]
			ax, ay, aw, ah, aa := a[0], a[1], a[2], a[3], a[4]
			ax *= th
			ay *= th
			aw *= th
			ah *= th
			aa *= th

			if unicode.IsSpace(r) {
				if r == '\n' {
					tx = m.world[0][3]
					ty -= (text.AscentUnit * th)
				}
			} else {
				n = uint32(len(env.verts)) / 4
				env.indices = append(env.indices,
					n, n+2, n+1, n, n+3, n+2,
				)
				env.verts = append(env.verts,
					tx+ax-pad, ty-ay-pad, z, 0, // v0
					tx+ax-pad, ty-ay+ah+pad, z, 0, // v1
					tx+ax+aw+pad, ty-ay+ah+pad, z, 0, // v2
					tx+ax+aw+pad, ty-ay-pad, z, 0, // v3
				)
				env.colors = append(env.colors,
					m.text.r, m.text.g, m.text.b, m.text.a,
					m.text.r, m.text.g, m.text.b, m.text.a,
					m.text.r, m.text.g, m.text.b, m.text.a,
					m.text.r, m.text.g, m.text.b, m.text.a,
				)
				env.dists = append(env.dists,
					0.0, 0.0, aw, th,
					0.0, 1.0, aw, th,
					1.0, 1.0, aw, th,
					1.0, 0.0, aw, th,
				)
				env.touches = append(env.touches,
					0, 0, 2, 0,
					0, 0, 2, 0,
					0, 0, 2, 0,
					0, 0, 2, 0,
				)
				g := text.Texcoords[r]
				gx, gy, gw, gh := g[0], g[1], g[2], g[3]
				env.texcoords = append(env.texcoords,
					gx, gy+gh, 0, 0,
					gx, gy, 0, 0,
					gx+gw, gy, 0, 0,
					gx+gw, gy+gh, 0, 0,
				)
			}

			tx += aa
		}
	}

	env.prg.Use(ctx)
	env.prg.Mat4(ctx, env.uniforms.view, env.View)
	env.prg.Mat4(ctx, env.uniforms.proj, env.proj)
	env.prg.U4f(ctx, env.uniforms.shadowColor, shdr, shdg, shdb, shda)
	env.prg.U4f(ctx, env.uniforms.glyphconf, text.FontSize, text.Pad, 0.5, 1)

	env.buffers.texcoords.Bind(ctx)
	env.buffers.texcoords.Update(ctx, env.texcoords)
	env.prg.Pointer(ctx, env.attribs.texcoord, 4)

	env.buffers.touches.Bind(ctx)
	env.buffers.touches.Update(ctx, env.touches)
	env.prg.Pointer(ctx, env.attribs.touch, 4)

	env.buffers.dists.Bind(ctx)
	env.buffers.dists.Update(ctx, env.dists)
	env.prg.Pointer(ctx, env.attribs.dist, 4)

	env.buffers.colors.Bind(ctx)
	env.buffers.colors.Update(ctx, env.colors)
	env.prg.Pointer(ctx, env.attribs.color, 4)

	env.buffers.verts.Bind(ctx)
	env.buffers.verts.Update(ctx, env.verts)
	env.buffers.indices.Bind(ctx)
	env.buffers.indices.Update(ctx, env.indices)
	env.prg.Pointer(ctx, env.attribs.vertex, 4)

	if env.glyphs.Value != 0 {
		env.glyphs.Bind(ctx, linearFilter, DefaultWrap)
		env.prg.U1i(ctx, env.uniforms.glyphs, int(env.glyphs.Value-1))
	}

	if env.icons.Value != 0 {
		env.icons.Bind(ctx, DefaultFilter, DefaultWrap)
		env.prg.U1i(ctx, env.uniforms.icons, int(env.icons.Value-1))
	}

	if env.image.Value != 0 {
		env.image.Bind(ctx, DefaultFilter, DefaultWrap)
		env.prg.U1i(ctx, env.uniforms.image, int(env.image.Value-1))
	}

	env.buffers.indices.Draw(ctx, env.prg, gl.TRIANGLES)
}

func (env *Environment) DrawGridDebug(ctx gl.Context) {
	env.Grid.draw(ctx, env.View, env.proj)
}

func (env *Environment) Touch(ev touch.Event) bool {
	ex, ey := ev.X, float32(windowSize.HeightPx)-ev.Y
	for i := len(env.sheets) - 1; i >= 0; i-- {
		sheet := env.sheets[i]
		if !sheet.Hidden() && sheet.Contains(ex, ey) {
			mtrl := sheet.M()
			mtrl.touch.state = ev.Type
			mtrl.touch.x, mtrl.touch.y = mtrl.RelativeCoords(ex, ey)
			if ev.Type == touch.TypeBegin {
				mtrl.touch.start = time.Now()
			}

			switch sheet := sheet.(type) {
			case *Button:
				if ev.Type == touch.TypeBegin && sheet.OnPress != nil {
					sheet.OnPress()
				}
				if sheet.OnTouch != nil {
					sheet.OnTouch(ev)
				}
			case *FloatingActionButton:
				if ev.Type == touch.TypeBegin && sheet.OnPress != nil {
					sheet.OnPress()
				}
				if sheet.OnTouch != nil {
					sheet.OnTouch(ev)
				}
			default:
				log.Printf("Unhandled type %T\n", sheet)
				continue
			}
			return true
		}
	}
	return false
}

func (env *Environment) NewMaterial(ctx gl.Context) *Material {
	m := New(ctx, Black)
	m.SetColor(env.plt.Light)
	env.sheets = append(env.sheets, m)
	return m
}

func (env *Environment) NewButton(ctx gl.Context) *Button {
	btn := &Button{Material: New(ctx, Black)} // TODO update constructor to remove color arg
	btn.SetColor(env.plt.Primary)
	btn.SetIconColor(White)
	env.sheets = append(env.sheets, btn)
	return btn
}

func (env *Environment) NewFloatingActionButton(ctx gl.Context) *FloatingActionButton {
	fab := &FloatingActionButton{Material: New(ctx, Black)} // TODO update constructor to remove color arg
	fab.SetColor(env.plt.Primary)
	fab.SetIconColor(White)
	fab.IsCircle = true
	env.sheets = append(env.sheets, fab)
	return fab
}

func (env *Environment) NewToolbar(ctx gl.Context) *Toolbar {
	bar := &Toolbar{
		Material: New(ctx, Black),
		Nav:      env.NewButton(ctx),
		Title:    env.NewMaterial(ctx),
	}
	bar.SetColor(env.plt.Light) // create specific ColorFromPalette on each type to localize selection
	bar.Nav.BehaviorFlags = DescriptorFlat
	bar.Nav.SetIcon(icon.NavigationMenu)
	bar.Nav.SetIconColor(Black)
	bar.Title.BehaviorFlags = DescriptorFlat
	env.sheets = append(env.sheets, bar)
	return bar
}

func (env *Environment) NewMenu(ctx gl.Context) *Menu {
	mu := &Menu{Material: New(ctx, Black)}
	mu.SetColor(env.plt.Light)
	mu.BehaviorFlags |= VisibilityTemporary
	mu.hidden = true
	env.sheets = append(env.sheets, mu)
	return mu
}

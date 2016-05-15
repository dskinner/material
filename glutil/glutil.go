package glutil

import (
	"math"

	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

var ident f32.Mat4

func init() {
	ident.Identity()
}

// Ortho provides a general purpose orthographic projection.
// TODO probably just get rid of this and pass in zero'd out z to perspective
func Ortho(m *f32.Mat4, l, r float32, b, t float32, n, f float32) {
	m.Identity()
	m.Scale(m, 2/(r-l), 2/(t-b), 2/(f-n))
	m.Translate(m, -((l + r) / 2), -((t + b) / 2), (f+n)/2)
}

// Perspective sets m to a screen space perspective with origin at bottom-left.
func Perspective(m *f32.Mat4, l, r float32, b, t float32) {
	m.Identity()
	// TODO i think [2][2] is best as t-b, or, shortest path, but maybe worth picking something that's consistent
	// e.g. always 1000
	(*m)[0][0] = 2 / (r - l)
	(*m)[0][3] = -1 // offset for [0][0]
	(*m)[1][1] = 2 / (t - b)
	(*m)[1][3] = -1               // offset for [1][1]
	(*m)[2][2] = 2 / (r * 10)     //(r - l) // TODO should maybe pick consistent result, such as whichever is smaller; r-l or t-b
	(*m)[3][2] = -(*m)[2][2] * 10 // pronounced z effect with increased factor
}

type Drawer interface {
	Draw(ctx gl.Context, view, proj f32.Mat4)
}

type DrawerFunc func(ctx gl.Context, view, proj f32.Mat4)

func (fn DrawerFunc) Draw(ctx gl.Context, view, proj f32.Mat4) {
	fn(ctx, view, proj)
}

type FloatBuffer interface {
	Bind(gl.Context)
	Update(gl.Context, []float32)
	Draw(gl.Context, Program, gl.Enum)
	Delete(gl.Context)
}

type floatBuffer struct {
	gl.Buffer
	bin   []byte
	count int
	usage gl.Enum
}

func NewFloatBuffer(ctx gl.Context, data []float32, usage gl.Enum) FloatBuffer {
	buf := &floatBuffer{Buffer: ctx.CreateBuffer(), usage: usage}
	buf.Bind(ctx)
	buf.Update(ctx, data)
	return buf
}

func (buf *floatBuffer) Bind(ctx gl.Context) { ctx.BindBuffer(gl.ARRAY_BUFFER, buf.Buffer) }

func (buf *floatBuffer) Update(ctx gl.Context, data []float32) {
	buf.count = len(data)
	subok := len(buf.bin) > 0 && len(data)*4 <= len(buf.bin)
	if !subok {
		buf.bin = make([]byte, len(data)*4)
	}
	for i, x := range data {
		u := math.Float32bits(x)
		buf.bin[4*i+0] = byte(u >> 0)
		buf.bin[4*i+1] = byte(u >> 8)
		buf.bin[4*i+2] = byte(u >> 16)
		buf.bin[4*i+3] = byte(u >> 24)
	}
	if subok {
		ctx.BufferSubData(gl.ARRAY_BUFFER, 0, buf.bin)
	} else {
		ctx.BufferData(gl.ARRAY_BUFFER, buf.bin, buf.usage)
	}
}

func (buf *floatBuffer) Draw(ctx gl.Context, prg Program, mode gl.Enum) {
	ctx.DrawArrays(mode, 0, buf.count)
}

func (buf *floatBuffer) Delete(ctx gl.Context) {
	ctx.DeleteBuffer(buf.Buffer)
}

type UintBuffer interface {
	Bind(gl.Context)
	Update(gl.Context, []uint32)
	Draw(gl.Context, Program, gl.Enum)
	Delete(gl.Context)
}

type uintBuffer struct {
	gl.Buffer
	bin   []byte
	count int
	usage gl.Enum
}

func NewUintBuffer(ctx gl.Context, data []uint32, usage gl.Enum) UintBuffer {
	buf := &uintBuffer{Buffer: ctx.CreateBuffer(), usage: usage}
	buf.Bind(ctx)
	buf.Update(ctx, data)
	return buf
}

func (buf *uintBuffer) Bind(ctx gl.Context) { ctx.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, buf.Buffer) }

func (buf *uintBuffer) Update(ctx gl.Context, data []uint32) {
	buf.count = len(data)
	subok := len(buf.bin) > 0 && len(data)*4 <= len(buf.bin)
	if !subok {
		buf.bin = make([]byte, len(data)*4)
	}
	for i, u := range data {
		buf.bin[4*i+0] = byte(u >> 0)
		buf.bin[4*i+1] = byte(u >> 8)
		buf.bin[4*i+2] = byte(u >> 16)
		buf.bin[4*i+3] = byte(u >> 24)
	}
	if subok {
		ctx.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, buf.bin)
	} else {
		ctx.BufferData(gl.ELEMENT_ARRAY_BUFFER, buf.bin, buf.usage)
	}
}

func (buf *uintBuffer) Draw(ctx gl.Context, prg Program, mode gl.Enum) {
	ctx.DrawElements(mode, buf.count, gl.UNSIGNED_INT, 0)
}

func (buf *uintBuffer) Delete(ctx gl.Context) {
	ctx.DeleteBuffer(buf.Buffer)
}

var (
	int32v4 = make([]int32, 4)
	int32v2 = make([]int32, 2)
)

type TextureFramebuffer struct {
	fbo  Framebuffer
	tex  Texture
	w, h int

	def, filter, wrap func(gl.Context, Texture)
	withtex           func(gl.Context, Framebuffer)
}

func (buf *TextureFramebuffer) Tex() Texture { return buf.tex }

// TODO pass in an actual Texture ...
func NewTextureBuffer(ctx gl.Context, width, height int) *TextureFramebuffer {
	buf := &TextureFramebuffer{w: width, h: height}
	buf.fbo = Framebuffer{ctx.CreateFramebuffer()}
	buf.tex = Texture{ctx.CreateTexture()}

	buf.def = TextureDef(0, width, height, gl.RGBA, nil)
	buf.filter = TextureFilter(gl.LINEAR, gl.LINEAR)
	buf.wrap = TextureWrap(gl.REPEAT, gl.REPEAT)
	buf.withtex = FramebufferWithTex(buf.tex, 0, buf.def, buf.filter, buf.wrap)
	return buf
}

func (buf *TextureFramebuffer) Delete(ctx gl.Context) {
	ctx.DeleteFramebuffer(buf.fbo.Framebuffer)
	ctx.DeleteTexture(buf.tex.Texture)
}

func (buf *TextureFramebuffer) StartSample(ctx gl.Context) {
	buf.fbo.Bind(ctx, buf.withtex)
	ctx.GetIntegerv(int32v4, gl.VIEWPORT)
	ctx.Viewport(0, 0, buf.w, buf.h)
	ctx.ClearColor(0, 0, 0, 0)
	ctx.Clear(gl.COLOR_BUFFER_BIT)
}

func (buf *TextureFramebuffer) StopSample(ctx gl.Context) {
	ctx.Viewport(int(int32v4[0]), int(int32v4[1]), int(int32v4[2]), int(int32v4[3]))
	// TODO Unbind should maybe take options too?
	buf.fbo.Unbind(ctx)
	// sbuf.tex.Unbind(ctx)
}

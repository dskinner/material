package glutil

import (
	"io/ioutil"
	"log"
	"math"

	"golang.org/x/mobile/asset"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

func MustOpen(name string) asset.File {
	f, err := asset.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	return f
}

func MustReadAll(name string) []byte {
	f := MustOpen(name)
	b, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	return b
}

func ShaderAsset(typ gl.Enum, name string) func(gl.Context) gl.Shader {
	return ShaderCompile(typ, name, string(MustReadAll(name)))
}

func VertAsset(name string) func(gl.Context) gl.Shader {
	return ShaderAsset(gl.VERTEX_SHADER, name)
}

func FragAsset(name string) func(gl.Context) gl.Shader {
	return ShaderAsset(gl.FRAGMENT_SHADER, name)
}

func ShaderCompile(typ gl.Enum, name string, src string) func(gl.Context) gl.Shader {
	return func(ctx gl.Context) gl.Shader {
		s := ctx.CreateShader(typ)
		ctx.ShaderSource(s, src)
		ctx.CompileShader(s)
		if ctx.GetShaderi(s, gl.COMPILE_STATUS) == 0 {
			log.Fatalf("glutil %s: %s\n", name, ctx.GetShaderInfoLog(s))
		}
		return s
	}
}

// TODO what if Program could use reflection and map functions?
// really, it doesn't need to do functions at all! just iter over
// attribs and uniforms and match
// TODO use reflection to set values
// type TexProgram struct {
// Program
// setworld   func(f32.Mat4)
// setsampler func(int)
// world    gl.Uniform
// view     gl.Uniform
// matrix   f32.Mat4 `glsl:"gl.Uniform"` // not really necessary, could just fail if same name exists in vert and frag shader
// sampler  gl.Uniform
// position gl.Attrib
// texcoord gl.Attrib
// }
type Program struct{ gl.Program }

func (prg *Program) CreateAndLink(ctx gl.Context, compilers ...func(gl.Context) gl.Shader) {
	prg.Program = ctx.CreateProgram()
	for _, c := range compilers {
		s := c(ctx)
		ctx.AttachShader(prg.Program, s)
		defer ctx.DeleteShader(s)
	}
	ctx.LinkProgram(prg.Program)
	if ctx.GetProgrami(prg.Program, gl.LINK_STATUS) == 0 {
		log.Fatalf("program link: %s", ctx.GetProgramInfoLog(prg.Program))
	}
}

func (prg Program) Delete(ctx gl.Context) {
	ctx.DeleteProgram(prg.Program)
}

func (prg Program) Uniform(ctx gl.Context, name string) gl.Uniform {
	return ctx.GetUniformLocation(prg.Program, name)
}

func (prg Program) Attrib(ctx gl.Context, name string) gl.Attrib {
	return ctx.GetAttribLocation(prg.Program, name)
}

func (prg Program) Use(ctx gl.Context, options ...func(gl.Context, Program)) {
	ctx.UseProgram(prg.Program)
	for _, opt := range options {
		opt(ctx, prg)
	}
}

func (prg Program) Mat4(ctx gl.Context, dst gl.Uniform, src f32.Mat4) {
	ctx.UniformMatrix4fv(dst, []float32{
		src[0][0], src[1][0], src[2][0], src[3][0],
		src[0][1], src[1][1], src[2][1], src[3][1],
		src[0][2], src[1][2], src[2][2], src[3][2],
		src[0][3], src[1][3], src[2][3], src[3][3],
	})
}

func (prg Program) U1i(ctx gl.Context, dst gl.Uniform, v int) {
	ctx.Uniform1i(dst, v)
}

func (prg Program) U2i(ctx gl.Context, dst gl.Uniform, v0, v1 int) {
	ctx.Uniform2i(dst, v0, v1)
}

func (prg Program) U1f(ctx gl.Context, dst gl.Uniform, v float32) {
	ctx.Uniform1f(dst, v)
}

func (prg Program) U2f(ctx gl.Context, dst gl.Uniform, v0, v1 float32) {
	ctx.Uniform2f(dst, v0, v1)
}

func (prg Program) U4f(ctx gl.Context, dst gl.Uniform, v0, v1, v2, v3 float32) {
	ctx.Uniform4f(dst, v0, v1, v2, v3)
}

// TODO an Attrib type that describes it's format would be useful here
func (prg Program) Pointer(ctx gl.Context, a gl.Attrib, size int) {
	ctx.EnableVertexAttribArray(a)
	ctx.VertexAttribPointer(a, size, gl.FLOAT, false, 0, 0)
}

// func UniformMat4(dst gl.Uniform) func(gl.Context, f32.Mat4) {
// src := make([]float32, 16)
// return func(ctx gl.Context, m f32.Mat4) {
// for i, v := range m {
// for j, x := range v {
// src[i*4+j] = x
// }
// }
// ctx.UniformMatrix4fv(dst, src)
// }
// }

// TODO want this ???
type Uniform4fFunc func(gl.Context, f32.Vec4)

func UniformVec4(dst gl.Uniform) func(gl.Context, f32.Vec4) {
	return func(ctx gl.Context, v f32.Vec4) { ctx.Uniform4f(dst, v[0], v[1], v[2], v[3]) }
}

func UniformFloat(dst gl.Uniform) func(gl.Context, float32) {
	return func(ctx gl.Context, x float32) { ctx.Uniform1f(dst, x) }
}

type Buffer struct {
	gl.Buffer
	target gl.Enum
}

func (buf *Buffer) Create(ctx gl.Context, target gl.Enum) {
	buf.Buffer = ctx.CreateBuffer()
	// buf.Update = BufferFloatData(target, usage)
	buf.target = target
}

func (buf Buffer) Delete(ctx gl.Context) {
	ctx.DeleteBuffer(buf.Buffer)
	// buf.Update = nil
}

func (buf Buffer) Bind(ctx gl.Context, after ...func(gl.Context)) {
	ctx.BindBuffer(buf.target, buf.Buffer)
	for _, fn := range after {
		fn(ctx)
	}
}

func (buf Buffer) Draw(ctx gl.Context, mode gl.Enum, first int, count int, before ...func(gl.Context)) {
	buf.Bind(ctx, before...)
	ctx.DrawArrays(mode, first, count)
}

func (buf Buffer) DrawElements(ctx gl.Context, mode gl.Enum, count int, typ gl.Enum, offset int, before ...func(gl.Context)) {
	buf.Bind(ctx, before...)
	ctx.DrawElements(mode, count, typ, offset)
}

func VertexAttrib(a gl.Attrib, size int, typ gl.Enum, normalize bool, stride int, offset int) func(gl.Context) {
	return func(ctx gl.Context) {
		ctx.EnableVertexAttribArray(a)
		ctx.VertexAttribPointer(a, size, typ, normalize, stride, offset)
	}
}

func BufferFloatData(target, usage gl.Enum) func(gl.Context, []float32) {
	var bin []byte
	return func(ctx gl.Context, data []float32) {
		subok := len(bin) > 0 && len(data)*4 <= len(bin)
		if !subok {
			bin = make([]byte, len(data)*4)
		}
		for i, x := range data {
			u := math.Float32bits(x)
			bin[4*i+0] = byte(u >> 0)
			bin[4*i+1] = byte(u >> 8)
			bin[4*i+2] = byte(u >> 16)
			bin[4*i+3] = byte(u >> 24)
		}
		if subok {
			ctx.BufferSubData(target, 0, bin)
		} else {
			ctx.BufferData(target, bin, usage)
		}
	}
}

func BufferUintData(target, usage gl.Enum) func(gl.Context, []uint32) {
	var bin []byte
	return func(ctx gl.Context, data []uint32) {
		subok := len(bin) > 0 && len(data)*4 <= len(bin)
		if !subok {
			bin = make([]byte, len(data)*4)
		}
		for i, u := range data {
			bin[4*i+0] = byte(u >> 0)
			bin[4*i+1] = byte(u >> 8)
			bin[4*i+2] = byte(u >> 16)
			bin[4*i+3] = byte(u >> 24)
		}
		if subok {
			ctx.BufferSubData(target, 0, bin)
		} else {
			ctx.BufferData(target, bin, usage)
		}
	}
}

type Texture struct{ gl.Texture }

func (tex *Texture) Create(ctx gl.Context) {
	tex.Texture = ctx.CreateTexture()
}

func (tex Texture) Delete(ctx gl.Context) {
	ctx.DeleteTexture(tex.Texture)
}

func (tex Texture) Bind(ctx gl.Context, options ...func(gl.Context, Texture)) {
	ctx.ActiveTexture(gl.Enum(uint32(gl.TEXTURE0) + tex.Value - 1))
	ctx.BindTexture(gl.TEXTURE_2D, tex.Texture)
	for _, opt := range options {
		opt(ctx, tex)
	}
}

func (tex Texture) Unbind(ctx gl.Context) {
	ctx.BindTexture(gl.TEXTURE_2D, gl.Texture{0})
}

func (tex Texture) Update(ctx gl.Context, lvl int, width int, height int, data []byte) {
	ctx.TexImage2D(gl.TEXTURE_2D, lvl, width, height, gl.RGBA, gl.UNSIGNED_BYTE, data)
	if lvl > 0 {
		ctx.GenerateMipmap(gl.TEXTURE_2D)
	}
}

// TODO incorporate into Update, see FloatBuffer and UintBuffer
func (tex Texture) Sub(ctx gl.Context, lvl int, width int, height int, data []byte) {
	ctx.TexSubImage2D(gl.TEXTURE_2D, lvl, 0, 0, width, height, gl.RGBA, gl.UNSIGNED_BYTE, data)
	if lvl > 0 {
		ctx.GenerateMipmap(gl.TEXTURE_2D)
	}
}

func TextureDef(lvl int, width, height int, format gl.Enum, data []byte) func(gl.Context, Texture) {
	return func(ctx gl.Context, tex Texture) {
		ctx.TexImage2D(gl.TEXTURE_2D, lvl, width, height, format, gl.UNSIGNED_BYTE, data)
	}
}

func TextureFilter(min, mag int) func(gl.Context, Texture) {
	return func(ctx gl.Context, tex Texture) {
		ctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, min)
		ctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, mag)
	}
}

func TextureWrap(s, t int) func(gl.Context, Texture) {
	return func(ctx gl.Context, tex Texture) {
		ctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, s)
		ctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, t)
	}
}

type Framebuffer struct{ gl.Framebuffer }

func (fbo *Framebuffer) Create(ctx gl.Context) {
	fbo.Framebuffer = ctx.CreateFramebuffer()
}

func (fbo Framebuffer) Delete(ctx gl.Context) {
	ctx.DeleteFramebuffer(fbo.Framebuffer)
}

func (fbo Framebuffer) Bind(ctx gl.Context, options ...func(gl.Context, Framebuffer)) {
	ctx.BindFramebuffer(gl.FRAMEBUFFER, fbo.Framebuffer)
	for _, opt := range options {
		opt(ctx, fbo)
	}
}

func (fbo Framebuffer) Unbind(ctx gl.Context) {
	ctx.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{0})
}

func FramebufferTex(tex Texture, lvl int) func(gl.Context, Framebuffer) {
	return func(ctx gl.Context, fbo Framebuffer) {
		ctx.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, tex.Texture, lvl)
	}
}

func FramebufferWithTex(tex Texture, lvl int, options ...func(gl.Context, Texture)) func(gl.Context, Framebuffer) {
	fbotex := FramebufferTex(tex, lvl)
	return func(ctx gl.Context, fbo Framebuffer) {
		tex.Bind(ctx, options...)
		fbotex(ctx, fbo)
	}
}

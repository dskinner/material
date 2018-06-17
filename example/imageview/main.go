package main

import (
	"bufio"
	"fmt"
	"image"
	"log"
	"math"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"dasa.cc/material"
	"dasa.cc/snd"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

const (
	KEY_ESCAPE           = 9
	KEY_1                = 10
	KEY_2                = 11
	KEY_3                = 12
	KEY_4                = 13
	KEY_5                = 14
	KEY_6                = 15
	KEY_7                = 16
	KEY_8                = 17
	KEY_9                = 18
	KEY_0                = 19
	KEY_MINUS            = 20
	KEY_EQUAL            = 21
	KEY_BACKSPACE        = 22
	KEY_TAB              = 23
	KEY_Q                = 24
	KEY_W                = 25
	KEY_E                = 26
	KEY_R                = 27
	KEY_T                = 28
	KEY_Y                = 29
	KEY_U                = 30
	KEY_I                = 31
	KEY_O                = 32
	KEY_P                = 33
	KEY_BRACKETLEFT      = 34
	KEY_BRACKETRIGHT     = 35
	KEY_RETURN           = 36
	KEY_CONTROL_L        = 37
	KEY_A                = 38
	KEY_S                = 39
	KEY_D                = 40
	KEY_F                = 41
	KEY_G                = 42
	KEY_H                = 43
	KEY_J                = 44
	KEY_K                = 45
	KEY_L                = 46
	KEY_SEMICOLON        = 47
	KEY_APOSTROPHE       = 48
	KEY_GRAVE            = 49
	KEY_SHIFT_L          = 50
	KEY_BACKSLASH        = 51
	KEY_Z                = 52
	KEY_X                = 53
	KEY_C                = 54
	KEY_V                = 55
	KEY_B                = 56
	KEY_N                = 57
	KEY_M                = 58
	KEY_COMMA            = 59
	KEY_PERIOD           = 60
	KEY_SLASH            = 61
	KEY_SHIFT_R          = 62
	KEY_KP_MULTIPLY      = 63
	KEY_ALT_L            = 64
	KEY_SPACE            = 65
	KEY_CAPS_LOCK        = 66
	KEY_F1               = 67
	KEY_F2               = 68
	KEY_F3               = 69
	KEY_F4               = 70
	KEY_F5               = 71
	KEY_F6               = 72
	KEY_F7               = 73
	KEY_F8               = 74
	KEY_F9               = 75
	KEY_F10              = 76
	KEY_ISO_LEVEL3_SHIFT = 92
	KEY_LESS             = 94
	KEY_F11              = 95
	KEY_F12              = 96
	KEY_CONTROL_R        = 105
	KEY_PRINT            = 107
	KEY_ALT_R            = 108
	KEY_LINEFEED         = 109
	KEY_HOME             = 110
	KEY_UP               = 111
	KEY_PRIOR            = 112
	KEY_LEFT             = 113
	KEY_RIGHT            = 114
	KEY_END              = 115
	KEY_DOWN             = 116
	KEY_NEXT             = 117
	KEY_INSERT           = 118
	KEY_DELETE           = 119
)

var (
	env   = new(material.Environment)
	box   *material.Material
	sig   snd.Discrete
	quits []chan struct{}

	imageSize image.Point

	position image.Point
	rotation float32
)

func onStart(ctx gl.Context) {
	env.SetPalette(material.Palette{
		Primary: material.BlueGrey500,
		Dark:    material.BlueGrey700,
		Light:   material.BlueGrey100,
		Accent:  material.DeepOrangeA200,
	})

	quits = []chan struct{}{}

	sig = make(snd.Discrete, len(material.ExpSig))
	copy(sig, material.ExpSig)
	rsig := make(snd.Discrete, len(material.ExpSig))
	copy(rsig, material.ExpSig)
	rsig.UnitInverse()
	sig = append(sig, rsig...)
	sig.NormalizeRange(0, 1)

	ctx.Enable(gl.BLEND)
	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	ctx.Enable(gl.CULL_FACE)
	ctx.CullFace(gl.BACK)

	env.Load(ctx)
	env.LoadGlyphs(ctx)

	if len(os.Args) > 1 {
		imageSize = env.LoadImage(ctx, os.Args[1])
	}

	box = env.NewMaterial(ctx)
	box.SetColor(material.BlueGrey200)
	box.ShowImage = true
}

func onStop(ctx gl.Context) {
	env.Unload(ctx)
	box = nil
	for _, q := range quits {
		q <- struct{}{}
	}
}

var windowSize size.Event

func onLayout(sz size.Event) {
	windowSize = sz
	env.SetOrtho(sz)
	// env.SetPerspective(sz)
	env.StartLayout()

	env.AddConstraints(box.Width(100), box.Height(100), box.Z(14))

	// env.AddConstraints(
	// box.CenterVerticalIn(env.Box), box.CenterHorizontalIn(env.Box),
	// )

	box.SetTextColor(material.Black)
	box.SetTextHeight(material.Dp(24).Px())
	box.SetText("Hello World")

	for _, q := range quits {
		q <- struct{}{}
	}
	quits = quits[:0]

	// log.Println("starting layout")
	// t := time.Now()
	env.FinishLayout()
	// log.Printf("finished layout in %s\n", time.Now().Sub(t))

	// func() {
	// 	m := box.World()
	// 	x, y := m[0][3], m[1][3]
	// 	w, h := m[0][0], m[1][1]
	// 	quits = append(quits, material.Animation{
	// 		Sig:  sig,
	// 		Dur:  4000 * time.Millisecond,
	// 		Loop: true,
	// 		Interp: func(dt float32) {
	// 			m[0][0] = w + 200*dt
	// 			m[0][3] = x - 200*dt/2
	// 			box.SetText(fmt.Sprintf("w: %.2f\nh: %.2f", m[0][0], m[1][1]))
	// 		},
	// 	}.Do())
	// 	quits = append(quits, material.Animation{
	// 		Sig:  sig,
	// 		Dur:  2000 * time.Millisecond,
	// 		Loop: true,
	// 		Interp: func(dt float32) {
	// 			m[1][1] = h + 200*dt
	// 			m[1][3] = y - 200*dt/2
	// 		},
	// 	}.Do())
	// }()

	// m[0][0] = float32(imageSize.X)
	// m[0][3] = -float32(imageSize.X / 2)
	// m[1][1] = float32(imageSize.Y)
	// m[1][3] = -float32(imageSize.Y / 2)

	_ = f32.Vec3{}
	// env.View.Translate(&env.View, 400, 250, 400)
	// env.View.Rotate(&env.View, 0.785, &f32.Vec3{0, 0, 1})
	// env.View.Rotate(&env.View, 0.785, &f32.Vec3{0, 1, 0})
}

var lastpaint time.Time
var fps int

var windowFocused bool

var (
	transX, transY float32
	rotate         f32.Radian
	scale          float32 = 1
)

var transStep float32 = 0.01
var rotStep f32.Radian = 0.01
var scaleStep float32 = 0.01

func onPaint(ctx gl.Context) {
	ctx.ClearColor(material.BlueGrey100.RGBA())
	ctx.Clear(gl.COLOR_BUFFER_BIT)

	handleKeyboard()

	m := box.World()
	m.Identity()
	m[2][3] = 16

	// m[0][1] = 1
	// m[1][0] = 1

	m[0][0] = float32(imageSize.X) * (float32(windowSize.HeightPx) / float32(imageSize.Y))
	m[1][1] = float32(windowSize.HeightPx)

	// scale
	m[0][0] *= scale
	m[1][1] *= scale

	// translate
	m[0][3] = transX * 1000
	m[1][3] = transY * 1000

	// box.Rotate = float32(rotate)
	// m.Rotate(m, rotate, &f32.Vec3{0, 0, 1})

	env.SetOrtho(windowSize)
	// env.SetPerspective(windowSize)

	tx, ty := m[0][0]/2, m[1][1]/2
	env.View.Translate(&env.View, tx, ty, 0)
	env.View.Rotate(&env.View, rotate, &f32.Vec3{0, 0, 1})
	env.View.Translate(&env.View, -tx, -ty, 0)

	// env.View.Rotate(&env.View, -0.86, &f32.Vec3{0, 1, 0})
	// env.View.Rotate(&env.View, 0.785, &f32.Vec3{0, 1, 0})
	// env.View.Rotate(&env.View, 0.785, &f32.Vec3{0, 1, 0})

	// m.Translate(m, transX, transY, 0)

	// m.Mul(m, &f32.Mat4{
	// {-0.9, 0, 0, 0},
	// {0, -0.5, 0, 0},
	// {0, 0, 1, 0},
	// {0, 0, 0, 1},
	// })

	// r := float32(rotate) //* 1000
	// m[0][2] = r
	// m[3][1] = r
	// m[3][2] = r
	// m[3][3] = r

	/*
		Mat4[ 670.680,  0.000,  0.000,  190.000,
		      0.000,  503.010,  0.000,  160.000,
		      0.000,  0.000,  1.000,  16.000,
		      0.000,  0.000,  0.000,  1.000]
	*/
	// m[3][0], m[0][3] = m[0][3], m[3][0]
	// m[3][1], m[1][3] = m[1][3], m[3][1]
	// m[3][2], m[2][3] = m[2][3], m[3][2]
	// transpose(m)
	// fmt.Printf("%v\n", m)
	// m.Rotate(m, rotate, &f32.Vec3{0, 0, 1})
	// transpose(m)
	// fmt.Printf("%v\n", m)
	// m[3][0], m[0][3] = m[0][3], m[3][0]
	// m[3][1], m[1][3] = m[1][3], m[3][1]
	// m[3][2], m[2][3] = m[2][3], m[3][2]

	// aff := &f32.Affine{
	// {m[0][0], m[0][1], m[0][2]},
	// {m[1][0], m[1][1], m[0][2]},
	// }

	// aff.Rotate(aff, r)
	// m[0][0] = aff[0][0]
	// m[0][1] = aff[0][1]
	// m[0][2] = aff[0][2]
	// m[1][0] = aff[1][0]
	// m[1][1] = aff[1][1]
	// m[1][2] = aff[1][2]

	r := float32(rotate)
	// s, c := f32.Sin(r), f32.Cos(r)

	s, c := float32(math.Sin(float64(r))), float32(math.Cos(float64(r)))
	_, _ = s, c

	// s, c := f32.Sin(r), f32.Cos(r)
	// orig
	// m.Mul(m, &f32.Mat4{
	// {+c, +s, 0, 0},
	// {-s, +c, 0, 0},
	// {0, 0, 1, 0},
	// {0, 0, 0, 1},
	// })

	// rotate z
	// m.Mul(m, &f32.Mat4{
	// {c, -s, 0, 0},
	// {s, c, 0, 0},
	// {0, 0, 1, 0},
	// {0, 0, 0, 1},
	// })

	// rot := Rotate3D(r, f32.Vec3{1, 0, 0})
	// m.Mul(m, &rot)

	// m.Mul(m, &f32.Mat4{
	// 	{-1, 0, 0, 0},
	// 	{0, -1, 0, 0},
	// 	{0, 0, 1, 0},
	// 	{0, 0, 0, 1},
	// })

	// temp = x;
	// x = x * cos(angle) - y * sin(angle);
	// y = sin(angle) * temp + cos(angle) * y;

	// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
	// x, y, z := m[0][0], m[1][1], m[2][2]
	// r := float32(rotate)
	// s, c := f32.Sin(r), f32.Cos(r)
	// rx := x*c - y*s
	// ry := x*s + y*c
	// rz := z
	// m[0][0] = rx
	// m[1][1] = ry
	// m[2][2] = rz

	// rotate y
	// m.Mul(m, &f32.Mat4{
	// {+c, 0, +s, 0},
	// {0, 1, 0, 0},
	// {-s, 0, +c, 0},
	// {0, 0, 0, 1},
	// })

	// rotate x
	// m.Mul(m, &f32.Mat4{
	// {1, 0, 0, 0},
	// {0, c, -s, 0},
	// {0, s, c, 0},
	// {0, 0, 0, 1},
	// })

	env.Draw(ctx)
	now := time.Now()
	fps = int(time.Second / now.Sub(lastpaint))
	lastpaint = now
}

func Rotate3D(angle float32, axis f32.Vec3) f32.Mat4 {
	x, y, z := axis[0], axis[1], axis[2]
	s, c := float32(math.Sin(float64(angle))), float32(math.Cos(float64(angle)))
	k := 1 - c

	// return Mat4{x*x*k + c, x*y*k + z*s, x*z*k - y*s, 0, x*y*k - z*s, y*y*k + c, y*z*k + x*s, 0, x*z*k + y*s, y*z*k - x*s, z*z*k + c, 0, 0, 0, 0, 1}
	return f32.Mat4{
		{x*x*k + c, x*y*k - z*s, x*z*k + y*s, 0},
		{x*y*k + z*s, y*y*k + c, y*z*k - x*s, 0},
		{x*z*k - y*s, y*z*k + x*s, z*z*k + c, 0},
		{0, 0, 0, 1},
	}
}

func Rotate(m *f32.Mat4, angle f32.Radian, axis *f32.Vec3) {
	a := *axis
	a.Normalize()

	c, s := f32.Cos(float32(angle)), f32.Sin(float32(angle))
	d := 1 - c

	m.Mul(m, &f32.Mat4{{
		c + d*a[0]*a[0],
		0 + d*a[0]*a[1] + s*a[2],
		0 + d*a[0]*a[2] - s*a[1],
		0,
	}, {
		0 + d*a[1]*a[0] - s*a[2],
		c + d*a[1]*a[1],
		0 + d*a[1]*a[2] + s*a[0],
		0,
	}, {
		0 + d*a[2]*a[0] + s*a[1],
		0 + d*a[2]*a[1] - s*a[0],
		c + d*a[2]*a[2],
		0,
	}, {
		0, 0, 0, 1,
	}})
}

func transpose(a *f32.Mat4) {
	a01 := a[0][1]
	a[0][1] = a[1][0]
	a[1][0] = a01

	a02 := a[0][2]
	a[0][2] = a[2][0]
	a[2][0] = a02

	a03 := a[0][3]
	a[0][3] = a[3][0]
	a[3][0] = a03

	a12 := a[1][2]
	a[1][2] = a[2][1]
	a[2][1] = a12

	a13 := a[1][3]
	a[1][3] = a[3][1]
	a[3][1] = a13

	a23 := a[2][3]
	a[2][3] = a[3][2]
	a[3][2] = a23
}

var suspendKeyboard bool

func handleKeyboard() {
	keysMutex.Lock()
	defer keysMutex.Unlock()

	if keys[KEY_ESCAPE] {
		os.Exit(1)
	}

	if keys[KEY_GRAVE] && keys[KEY_1] {
		suspendKeyboard = true
	}
	if keys[KEY_GRAVE] && keys[KEY_2] {
		suspendKeyboard = false
	}
	if suspendKeyboard {
		return
	}

	if keys[KEY_A] {
		transX -= transStep
	}
	if keys[KEY_D] {
		transX += transStep
	}
	if keys[KEY_W] {
		transY += transStep
	}
	if keys[KEY_S] {
		transY -= transStep
	}
	if keys[KEY_E] {
		rotate += rotStep
	}
	if keys[KEY_Q] {
		rotate -= rotStep
	}
	if keys[KEY_Z] {
		scale -= scaleStep
	}
	if keys[KEY_X] {
		scale += scaleStep
	}
	if keys[KEY_F] {
		transX, transY = 0, 0
		rotate = 0
		scale = 1
	}
}

var keys = make(map[int]bool)
var keysMutex sync.Mutex

func main() {
	app.Main(func(a app.App) {

		go func() {
			cmd := exec.Command("xinput", "--test", "12")
			stdout, err := cmd.StdoutPipe()
			if err != nil {
				log.Fatal(err)
			}
			cmd.Start()
			scanner := bufio.NewScanner(stdout)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "key press") {
					line = strings.TrimPrefix(line, "key press")
					i, err := strconv.Atoi(strings.TrimSpace(line))
					if err != nil {
						log.Fatal(err)
					}
					keysMutex.Lock()
					keys[i] = true
					keysMutex.Unlock()
				} else if strings.HasPrefix(line, "key release") {
					line = strings.TrimPrefix(line, "key release")
					i, err := strconv.Atoi(strings.TrimSpace(line))
					if err != nil {
						log.Fatal(err)
					}
					keysMutex.Lock()
					keys[i] = false
					keysMutex.Unlock()
				}
			}
			if err := scanner.Err(); err != nil {
				fmt.Fprintln(os.Stderr, "reading command output:", err)
			}
		}()

		var glctx gl.Context
		var ticker *time.Ticker
		for ev := range a.Events() {
			switch ev := a.Filter(ev).(type) {
			case lifecycle.Event:
				switch ev.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					if ticker != nil {
						ticker.Stop()
					}
					ticker = time.NewTicker(time.Second)
					go func() {
						for range ticker.C {
							log.Printf("fps=%-4v\n", fps)
						}
					}()
					glctx = ev.DrawContext.(gl.Context)
					onStart(glctx)
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					if ticker != nil {
						ticker.Stop()
					}
					onStop(glctx)
					glctx = nil
				}
			case size.Event:
				if glctx == nil {
					a.Send(ev) // republish event until onStart is called
				} else {
					onLayout(ev)
				}
			case paint.Event:
				if glctx == nil || ev.External {
					continue
				}
				onPaint(glctx)
				a.Publish()
				a.Send(paint.Event{})
			case touch.Event:
				env.Touch(ev)
			}
		}
	})
}

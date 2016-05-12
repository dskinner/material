package main

import (
	"log"
	"time"

	"dasa.cc/material"
	"dasa.cc/snd"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/exp/f32"
	"golang.org/x/mobile/gl"
)

var (
	env   = new(material.Environment)
	boxes [9]*material.Material
	sig   snd.Discrete
	quits []chan struct{}
)

func init() {
	env.SetPalette(material.Palette{
		Primary: material.BlueGrey500,
		Dark:    material.BlueGrey700,
		Light:   material.BlueGrey100,
		Accent:  material.DeepOrangeA200,
	})

	sig = make(snd.Discrete, len(material.ExpSig))
	copy(sig, material.ExpSig)
	rsig := make(snd.Discrete, len(material.ExpSig))
	copy(rsig, material.ExpSig)
	rsig.UnitInverse()
	sig = append(sig, rsig...)
	sig.NormalizeRange(0, 1)
}

func onStart(ctx gl.Context) {
	ctx.Enable(gl.BLEND)
	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	ctx.Enable(gl.CULL_FACE)
	ctx.CullFace(gl.BACK)

	env.Load(ctx)

	for i := range boxes {
		boxes[i] = env.NewMaterial(ctx)
		boxes[i].SetColor(material.BlueGrey200)
	}
}

func onLayout(sz size.Event) {
	env.SetOrtho(sz)
	// env.SetPerspective(sz)
	env.StartLayout()

	for i, box := range boxes {
		env.AddConstraints(box.Width(100), box.Height(100), box.Z(float32(i+1)))
	}

	b, p := env.Box, env.Grid.Gutter
	env.AddConstraints(
		boxes[0].StartIn(b, p), boxes[0].TopIn(b, p),
		boxes[1].CenterHorizontalIn(b), boxes[1].TopIn(b, p),
		boxes[2].EndIn(b, p), boxes[2].TopIn(b, p),
		boxes[3].CenterVerticalIn(b), boxes[3].StartIn(b, p),
		boxes[4].CenterVerticalIn(b), boxes[4].CenterHorizontalIn(b),
		boxes[5].CenterVerticalIn(b), boxes[5].EndIn(b, p),
		boxes[6].StartIn(b, p), boxes[6].BottomIn(b, p),
		boxes[7].CenterHorizontalIn(b), boxes[7].BottomIn(b, p),
		boxes[8].EndIn(b, p), boxes[8].BottomIn(b, p),
	)

	for _, q := range quits {
		q <- struct{}{}
	}

	log.Println("starting layout")
	t := time.Now()
	env.FinishLayout()
	log.Printf("finished layout in %s\n", time.Now().Sub(t))

	func() {
		m := boxes[1].World()
		x, z := m[0][3], m[2][3]
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  2000 * time.Millisecond,
			Loop: true,
			Interp: func(dt float32) {
				m[0][3] = x + 200*dt
			},
		}.Do())
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  1000 * time.Millisecond,
			Loop: true,
			Interp: func(dt float32) {
				m[2][3] = z + 4*dt
			},
		}.Do())
	}()

	func() {
		m := boxes[2].World()
		z := m[2][3]
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  2000 * time.Millisecond,
			Loop: true,
			Interp: func(dt float32) {
				m[2][3] = z + 10*dt
			},
		}.Do())
	}()

	func() {
		m := boxes[4].World()
		x, y, z := m[0][3], m[1][3], m[2][3]
		w, h := m[0][0], m[1][1]
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  4000 * time.Millisecond,
			Loop: true,
			Interp: func(dt float32) {
				m[0][0] = w + 200*dt
				m[0][3] = x - 200*dt/2
			},
		}.Do())
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  2000 * time.Millisecond,
			Loop: true,
			Interp: func(dt float32) {
				m[1][1] = h + 200*dt
				m[1][3] = y - 200*dt/2
			},
		}.Do())
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  8000 * time.Millisecond,
			Loop: true,
			Interp: func(dt float32) {
				m[2][3] = z + 20*dt
			},
		}.Do())
	}()

	func() {
		m := boxes[6].World()
		w, h := m[0][0], m[1][1]
		z := m[2][3]
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  4000 * time.Millisecond,
			Loop: true,
			Interp: func(dt float32) {
				boxes[6].Roundness = 50 * (1 - dt)
				m[0][0] = w + 200*dt
				m[1][1] = h + 200*dt
			},
		}.Do())
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  8000 * time.Millisecond,
			Loop: true,
			Interp: func(dt float32) {
				m[2][3] = z + 7*dt
			},
		}.Do())
	}()

	func() {
		m := boxes[8].World()
		w := m[0][0]
		quits = append(quits, material.Animation{
			Sig:  sig,
			Dur:  2000 * time.Millisecond,
			Loop: true,
			Interp: func(dt float32) {
				boxes[8].Roundness = (w / 2) * dt
			},
		}.Do())
	}()

	_ = f32.Vec3{}
	// env.View.Translate(&env.View, 400, 250, 400)
	// env.View.Rotate(&env.View, 0.785, &f32.Vec3{0, 0, 1})
	// env.View.Rotate(&env.View, 0.785, &f32.Vec3{0, 1, 0})
}

var lastpaint time.Time
var fps int

func onPaint(ctx gl.Context) {
	ctx.ClearColor(material.BlueGrey100.RGBA())
	ctx.Clear(gl.COLOR_BUFFER_BIT)
	env.Draw(ctx)
	now := time.Now()
	fps = int(time.Second / now.Sub(lastpaint))
	lastpaint = now
}

func main() {
	app.Main(func(a app.App) {
		var glctx gl.Context
		for ev := range a.Events() {
			switch ev := a.Filter(ev).(type) {
			case lifecycle.Event:
				switch ev.Crosses(lifecycle.StageVisible) {
				case lifecycle.CrossOn:
					go func() {
						for range time.Tick(time.Second) {
							log.Printf("fps=%-4v\n", fps)
						}
					}()
					glctx = ev.DrawContext.(gl.Context)
					onStart(glctx)
				case lifecycle.CrossOff:
					glctx = nil
				}
			case size.Event:
				if glctx == nil {
					a.Send(ev) // republish event until onStart is called
				} else {
					onLayout(ev)
				}
			case paint.Event:
				if glctx != nil {
					onPaint(glctx)
					a.Publish()
					a.Send(paint.Event{})
				}
			}
		}
	})
}

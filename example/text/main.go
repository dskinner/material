package main

import (
	"log"
	"time"

	"dasa.cc/material"
	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/lifecycle"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/event/touch"
	"golang.org/x/mobile/gl"
)

var (
	env = new(material.Environment)

	t112, t56, t45, t34, t24, t20, t16, t14, t12 *material.Button
)

func init() {
	env.SetPalette(material.Palette{
		Primary: material.BlueGrey500,
		Dark:    material.BlueGrey700,
		Light:   material.BlueGrey100,
		Accent:  material.DeepOrangeA200,
	})
}

func onStart(ctx gl.Context) {
	ctx.Enable(gl.BLEND)
	ctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	ctx.Enable(gl.CULL_FACE)
	ctx.CullFace(gl.BACK)

	env.Load(ctx)
	env.LoadGlyphs(ctx)

	t112 = env.NewButton(ctx)
	t112.SetTextColor(material.White)
	t112.SetText("ABAAH*`e_llo |jJ go 112px")
	t112.BehaviorFlags = material.DescriptorFlat

	t56 = env.NewButton(ctx)
	t56.SetTextColor(material.White)
	t56.SetText("ABAAaBbCcDdEeFfGgHhIiJjKkLlMmNnOoPpQqRrSsTtUuVvWwXxYyZz")
	t56.BehaviorFlags = material.DescriptorFlat

	t45 = env.NewButton(ctx)
	t45.SetTextColor(material.White)
	t45.SetText("ABAAHello go 45px")
	t45.BehaviorFlags = material.DescriptorFlat

	t34 = env.NewButton(ctx)
	t34.SetTextColor(material.White)
	t34.SetText("ABAAAAHello go 34px")
	t34.BehaviorFlags = material.DescriptorFlat

	t24 = env.NewButton(ctx)
	t24.SetTextColor(material.White)
	t24.SetText("ABAAAAAAHello go 24px")
	t24.BehaviorFlags = material.DescriptorFlat

	t20 = env.NewButton(ctx)
	t20.SetTextColor(material.White)
	t20.SetText("ABAAAAAAAAHello go 20px")
	t20.BehaviorFlags = material.DescriptorFlat

	t16 = env.NewButton(ctx)
	t16.SetTextColor(material.White)
	t16.SetText("ABAAAAAAHello go 16px")
	t16.BehaviorFlags = material.DescriptorFlat

	t14 = env.NewButton(ctx)
	t14.SetTextColor(material.White)
	t14.SetText("ABAAAAAAHello go 14px")
	t14.BehaviorFlags = material.DescriptorFlat

	t12 = env.NewButton(ctx)
	t12.SetTextColor(material.White)
	t12.SetText("ABAAAAAAHello go 12px")
	t12.BehaviorFlags = material.DescriptorFlat
}

func onLayout(sz size.Event) {
	env.SetOrtho(sz)
	env.StartLayout()
	env.AddConstraints(
		t112.Width(1290), t112.Height(112), t112.Z(1), t112.StartIn(env.Box, 0), t112.TopIn(env.Box, env.Grid.Gutter),
		t56.Width(620), t56.Height(56), t56.Z(1), t56.StartIn(env.Box, env.Grid.Gutter), t56.Below(t112.Box, env.Grid.Gutter),
		t45.Width(500), t45.Height(45), t45.Z(1), t45.StartIn(env.Box, env.Grid.Gutter), t45.Below(t56.Box, env.Grid.Gutter),
		t34.Width(380), t34.Height(34), t34.Z(1), t34.StartIn(env.Box, env.Grid.Gutter), t34.Below(t45.Box, env.Grid.Gutter),
		t24.Width(270), t24.Height(24), t24.Z(1), t24.StartIn(env.Box, env.Grid.Gutter), t24.Below(t34.Box, env.Grid.Gutter),
		t20.Width(230), t20.Height(20), t20.Z(1), t20.StartIn(env.Box, env.Grid.Gutter), t20.Below(t24.Box, env.Grid.Gutter),
		t16.Width(180), t16.Height(16), t16.Z(1), t16.StartIn(env.Box, env.Grid.Gutter), t16.Below(t20.Box, env.Grid.Gutter),
		t14.Width(155), t14.Height(14), t14.Z(1), t14.StartIn(env.Box, 0), t14.Below(t16.Box, env.Grid.Gutter),
		t12.Width(135), t12.Height(12), t12.Z(1), t12.StartIn(env.Box, env.Grid.Gutter), t12.Below(t14.Box, env.Grid.Gutter),
	)
	log.Println("starting layout")
	t := time.Now()
	env.FinishLayout()
	log.Printf("finished layout in %s\n", time.Now().Sub(t))
}

var lastpaint time.Time
var fps int

func onPaint(ctx gl.Context) {
	ctx.ClearColor(material.BlueGrey500.RGBA())
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
					a.Send(paint.Event{})
				case lifecycle.CrossOff:
					env.Unload(glctx)
					glctx = nil
				}
			case size.Event:
				if glctx == nil {
					a.Send(ev) // republish event until onStart is called
				} else {
					onLayout(ev)
				}
			case touch.Event:
				env.Touch(ev)
			case paint.Event:
				if glctx == nil || ev.External {
					continue
				}
				onPaint(glctx)
				a.Publish()
				a.Send(paint.Event{})
			}
		}
	})
}

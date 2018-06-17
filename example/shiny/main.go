package main

import (
	"log"
	"reflect"
	"unsafe"

	"golang.org/x/exp/shiny/driver/gldriver"
	"golang.org/x/exp/shiny/screen"
	"golang.org/x/mobile/event/paint"
	"golang.org/x/mobile/event/size"
	"golang.org/x/mobile/gl"
)

func contextOf(w screen.Window) gl.Context {
	s := reflect.ValueOf(w).Elem()
	f := s.FieldByName("glctx")
	p := unsafe.Pointer(f.UnsafeAddr())
	q := (*gl.Context)(p)
	return *q
}

func main() {
	gldriver.Main(func(s screen.Screen) {
		w, err := s.NewWindow(nil)
		if err != nil {
			log.Fatal(err)
		}
		defer w.Release()

		var ctx gl.Context

		for {
			switch ev := w.NextEvent().(type) {
			case paint.Event:
				ctx = contextOf(w)
				log.Println(ev)
				ctx.ClearColor(1, 1, 1, 1)
				w.Publish()
				w.Send(paint.Event{})
			case size.Event:
				log.Println(ev)
				// ctx.Viewport(0, 0, ev.WidthPx, ev.HeightPx)
			}
		}
	})
}
